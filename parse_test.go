// Copyright (C) 2023  Eric Cornelissen
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import "testing"

func TestParseWorkflowSuccess(t *testing.T) {
	testCases := []struct {
		name     string
		yaml     string
		expected Workflow
	}{
		{
			name: "Workflow with 'run:'",
			yaml: `
jobs:
  example:
    name: Example
    steps:
    - name: Checkout repository
      uses: actions/checkout@v3
      with:
        fetch-depth: 1
    - name: Echo value
      run: echo '${{ inputs.value }}'
`,
			expected: Workflow{
				Jobs: map[string]WorkflowJob{
					"example": {
						Name: "Example",
						Steps: []JobStep{
							{
								Name: "Checkout repository",
								Uses: "actions/checkout@v3",
							},
							{
								Name: "Echo value",
								Run:  "echo '${{ inputs.value }}'",
							},
						},
					},
				},
			},
		},
		{
			name: "Workflow with 'actions/github-script'",
			yaml: `
jobs:
  example:
    name: Example
    steps:
    - name: Checkout repository
      uses: actions/checkout@v3
      with:
        fetch-depth: 1
    - name: Echo value
      uses: actions/github-script@v6
      with:
        script: console.log('${{ inputs.value }}')
`,
			expected: Workflow{
				Jobs: map[string]WorkflowJob{
					"example": {
						Name: "Example",
						Steps: []JobStep{
							{
								Name: "Checkout repository",
								Uses: "actions/checkout@v3",
							},
							{
								Name: "Echo value",
								Uses: "actions/github-script@v6",
								With: StepWith{
									Script: "console.log('${{ inputs.value }}')",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "No names",
			yaml: `
jobs:
  example:
    steps:
    - uses: actions/setup-node@v3
      with:
        node-version: 20
    - run: echo ${{ inputs.value }}
`,
			expected: Workflow{
				Jobs: map[string]WorkflowJob{
					"example": {
						Name: "",
						Steps: []JobStep{
							{
								Uses: "actions/setup-node@v3",
							},
							{
								Run: "echo ${{ inputs.value }}",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			workflow, err := ParseWorkflow([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("Unexpected error: %#v", err)
			}

			if got, want := len(workflow.Jobs), len(tt.expected.Jobs); got != want {
				t.Fatalf("Unexpected number of jobs (got '%d', want '%d')", got, want)
			}

			for k, job := range workflow.Jobs {
				expected := tt.expected.Jobs[k]

				if got, want := job.Name, expected.Name; got != want {
					t.Errorf("Unexpected name for job '%s' (got '%s', want '%s')", k, got, want)
				}

				if got, want := len(job.Steps), len(expected.Steps); got != want {
					t.Fatalf("Unexpected number of steps for job '%s' (got '%d', want '%d')", job, got, want)
				}

				for i, step := range job.Steps {
					expected := expected.Steps[i]

					if got, want := step.Name, expected.Name; got != want {
						t.Errorf("Unexpected name for job '%s' step %d (got '%s', want '%s')", k, i, got, want)
					}

					if got, want := step.Run, expected.Run; got != want {
						t.Errorf("Unexpected run for job '%s' step %d (got '%s', want '%s')", k, i, got, want)
					}

					if got, want := step.Uses, expected.Uses; got != want {
						t.Errorf("Unexpected uses for job '%s' step %d (got '%s', want '%s')", k, i, got, want)
					}

					if got, want := step.With.Script, expected.With.Script; got != want {
						t.Errorf("Unexpected with for job '%s' step %d (got '%s', want '%s')", k, i, got, want)
					}
				}
			}
		})
	}
}

func TestParseWorkflowError(t *testing.T) {
	testCases := []struct {
		name string
		yaml string
	}{
		{
			name: "Invalid 'jobs' value",
			yaml: `
jobs: 3.14
`,
		},
		{
			name: "Invalid 'steps' value",
			yaml: `
jobs:
  example:
    steps: 42
`,
		},
		{
			name: "Invalid 'with' value",
			yaml: `
jobs:
  example:
    steps:
    - with: 1.618
`,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseWorkflow([]byte(tt.yaml))
			if err == nil {
				t.Fatal("Expected an error, got none")
			}
		})
	}
}

func TestParseManifestSuccess(t *testing.T) {
	testCases := []struct {
		name     string
		yaml     string
		expected Manifest
	}{
		{
			name: "Non-composite manifest",
			yaml: `
runs:
  using: node16
  main: index.js
`,
			expected: Manifest{
				Runs: ManifestRuns{
					Using: "node16",
				},
			},
		},
		{
			name: "Manifest with 'run:'",
			yaml: `
runs:
  using: composite
  steps:
  - name: Checkout repository
    uses: actions/checkout@v3
    with:
      fetch-depth: 1
  - name: Echo value
    run: echo '${{ inputs.value }}'
`,
			expected: Manifest{
				Runs: ManifestRuns{
					Using: "composite",
					Steps: []JobStep{
						{
							Name: "Checkout repository",
							Uses: "actions/checkout@v3",
						},
						{
							Name: "Echo value",
							Run:  "echo '${{ inputs.value }}'",
						},
					},
				},
			},
		},
		{
			name: "Manifest with 'actions/github-script'",
			yaml: `
runs:
  using: composite
  steps:
  - name: Checkout repository
    uses: actions/checkout@v3
    with:
      fetch-depth: 1
  - name: Echo value
    uses: actions/github-script@v6
    with:
      script: console.log('${{ inputs.value }}')
`,
			expected: Manifest{
				Runs: ManifestRuns{
					Using: "composite",
					Steps: []JobStep{
						{
							Name: "Checkout repository",
							Uses: "actions/checkout@v3",
						},
						{
							Name: "Echo value",
							Uses: "actions/github-script@v6",
							With: StepWith{
								Script: "console.log('${{ inputs.value }}')",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			manifest, err := ParseManifest([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("Unexpected error: %#v", err)
			}

			if got, want := len(manifest.Runs.Using), len(tt.expected.Runs.Using); got != want {
				t.Fatalf("Unexpected using value (got '%d', want '%d')", got, want)
			}

			if got, want := len(manifest.Runs.Steps), len(tt.expected.Runs.Steps); got != want {
				t.Fatalf("Unexpected number of steps (got '%d', want '%d')", got, want)
			}

			for i, step := range manifest.Runs.Steps {
				expected := tt.expected.Runs.Steps[i]

				if got, want := step.Name, expected.Name; got != want {
					t.Errorf("Unexpected name for step %d (got '%s', want '%s')", i, got, want)
				}

				if got, want := step.Run, expected.Run; got != want {
					t.Errorf("Unexpected run for step %d (got '%s', want '%s')", i, got, want)
				}

				if got, want := step.Uses, expected.Uses; got != want {
					t.Errorf("Unexpected uses for step %d (got '%s', want '%s')", i, got, want)
				}

				if got, want := step.With.Script, expected.With.Script; got != want {
					t.Errorf("Unexpected with for step %d (got '%s', want '%s')", i, got, want)
				}
			}
		})
	}
}

func TestParseManifestError(t *testing.T) {
	testCases := []struct {
		name string
		yaml string
	}{
		{
			name: "Invalid 'runs' value",
			yaml: `runs: 3.14`,
		},
		{
			name: "Invalid 'steps' value",
			yaml: `
runs:
  steps: 3.14
`,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseManifest([]byte(tt.yaml))
			if err == nil {
				t.Fatal("Expected an error, got none")
			}
		})
	}
}
