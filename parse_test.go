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
	type TestCase struct {
		name string
		yaml string
		want Workflow
	}

	testCases := []TestCase{
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
			want: Workflow{
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
			want: Workflow{
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
			want: Workflow{
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

			if got, want := len(workflow.Jobs), len(tt.want.Jobs); got != want {
				t.Fatalf("Unexpected number of jobs (got %d, want %d)", got, want)
			}

			for k, job := range workflow.Jobs {
				want := tt.want.Jobs[k]

				if got, want := job.Name, want.Name; got != want {
					t.Errorf("Unexpected name for job %q (got %q, want %q)", k, got, want)
				}

				if got, want := len(job.Steps), len(want.Steps); got != want {
					t.Fatalf("Unexpected number of steps for job %q (got %d, want %d)", job, got, want)
				}

				for i, step := range job.Steps {
					want := want.Steps[i]

					if got, want := step.Name, want.Name; got != want {
						t.Errorf("Unexpected name for job %q step %d (got %q, want %q)", k, i, got, want)
					}

					if got, want := step.Run, want.Run; got != want {
						t.Errorf("Unexpected run for job %q step %d (got %q, want %q)", k, i, got, want)
					}

					if got, want := step.Uses, want.Uses; got != want {
						t.Errorf("Unexpected uses for job %q step %d (got %q, want %q)", k, i, got, want)
					}

					if got, want := step.With.Script, want.With.Script; got != want {
						t.Errorf("Unexpected with for job %q step %d (got %q, want %q)", k, i, got, want)
					}
				}
			}
		})
	}
}

func TestParseWorkflowError(t *testing.T) {
	type TestCase struct {
		name string
		yaml string
	}

	testCases := []TestCase{
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
	type TestCase struct {
		name string
		yaml string
		want Manifest
	}

	testCases := []TestCase{
		{
			name: "Non-composite manifest",
			yaml: `
runs:
  using: node16
  main: index.js
`,
			want: Manifest{
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
			want: Manifest{
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
			want: Manifest{
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

			if got, want := len(manifest.Runs.Using), len(tt.want.Runs.Using); got != want {
				t.Fatalf("Unexpected using value (got %d, want %d)", got, want)
			}

			if got, want := len(manifest.Runs.Steps), len(tt.want.Runs.Steps); got != want {
				t.Fatalf("Unexpected number of steps (got %d, want %d)", got, want)
			}

			for i, step := range manifest.Runs.Steps {
				want := tt.want.Runs.Steps[i]

				if got, want := step.Name, want.Name; got != want {
					t.Errorf("Unexpected name for step %d (got %q, want %q)", i, got, want)
				}

				if got, want := step.Run, want.Run; got != want {
					t.Errorf("Unexpected run for step %d (got %q, want %q)", i, got, want)
				}

				if got, want := step.Uses, want.Uses; got != want {
					t.Errorf("Unexpected uses for step %d (got %q, want %q)", i, got, want)
				}

				if got, want := step.With.Script, want.With.Script; got != want {
					t.Errorf("Unexpected with for step %d (got %q, want %q)", i, got, want)
				}
			}
		})
	}
}

func TestParseManifestError(t *testing.T) {
	type TestCase struct {
		name string
		yaml string
	}

	testCases := []TestCase{
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
