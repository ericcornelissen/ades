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

func TestProcessWorkflow(t *testing.T) {
	testCases := []struct {
		name     string
		workflow Workflow
		expected int
	}{
		{
			name: "Safe workflow",
			workflow: Workflow{
				Jobs: map[string]Job{
					"safe": {
						Name: "Safe",
						Steps: []Step{
							{
								Name: "Example",
								Run:  "",
							},
						},
					},
				},
			},
			expected: 0,
		},
		{
			name: "Problem in first of two jobs in workflow",
			workflow: Workflow{
				Jobs: map[string]Job{
					"unsafe": {
						Name: "Unsafe",
						Steps: []Step{
							{
								Name: "Example",
								Run:  "echo ${{ inputs.value }}",
							},
						},
					},
					"safe": {
						Name: "Safe",
						Steps: []Step{
							{
								Name: "Example",
								Run:  "echo 'Hello world!'",
							},
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "Problem in second of two jobs in workflow",
			workflow: Workflow{
				Jobs: map[string]Job{
					"safe": {
						Name: "Safe",
						Steps: []Step{
							{
								Name: "Example",
								Run:  "echo 'Hello world!'",
							},
						},
					},
					"unsafe": {
						Name: "Unsafe",
						Steps: []Step{
							{
								Name: "Example",
								Run:  "echo ${{ inputs.value }}",
							},
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "Problem in ll jobs in workflow",
			workflow: Workflow{
				Jobs: map[string]Job{
					"unsafe": {
						Name: "Unsafe",
						Steps: []Step{
							{
								Name: "Greeting",
								Run:  "echo 'Hello ${{ inputs.name }}!'",
							},
						},
					},
					"more-unsafe": {
						Name: "More Unsafe",
						Steps: []Step{
							{
								Name: "Example",
								Run:  "echo ${{ inputs.value }}",
							},
							{
								Name: "Today is",
								Run:  "echo ${{ steps.id.outputs.day }}",
							},
						},
					},
				},
			},
			expected: 3,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			problems := processWorkflow(&tt.workflow)

			if got, want := len(problems), tt.expected; got != want {
				t.Fatalf("Unexpected number of problems (got '%d', want '%d')", got, want)
			}
		})
	}
}

func TestProcessJob(t *testing.T) {
	testCases := []struct {
		name     string
		id       string
		job      Job
		expected int
	}{
		{
			name: "Safe unnamed job",
			job: Job{
				Name: "",
				Steps: []Step{
					{
						Name: "Unnamed Example",
						Run:  "",
					},
				},
			},
			expected: 0,
		},
		{
			name: "Safe named job",
			job: Job{
				Name: "Safe",
				Steps: []Step{
					{
						Name: "Named example",
						Run:  "",
					},
				},
			},
			expected: 0,
		},
		{
			name: "Unnamed job with unsafe step",
			id:   "job-id",
			job: Job{
				Name: "",
				Steps: []Step{
					{
						Name: "Example",
						Run:  "echo ${{ inputs.value }}",
					},
				},
			},
			expected: 1,
		},
		{
			name: "Named job with unsafe step",
			job: Job{
				Name: "Unsafe",
				Steps: []Step{
					{
						Name: "Example",
						Run:  "echo ${{ inputs.value }}",
					},
				},
			},
			expected: 1,
		},
		{
			name: "Unnamed job with unsafe and safe steps",
			id:   "job-id",
			job: Job{
				Name: "",
				Steps: []Step{
					{
						Name: "Checkout repository",
						Run:  "",
					},
					{
						Name: "Example",
						Run:  "echo ${{ inputs.value }}",
					},
					{
						Name: "Run tests",
						Run:  "make test suite=$SUITE",
					},
				},
			},
			expected: 1,
		},
		{
			name: "Named job with unsafe and safe steps",
			job: Job{
				Name: "Unsafe",
				Steps: []Step{
					{
						Name: "Checkout repository",
						Run:  "",
					},
					{
						Name: "Example",
						Run:  "echo ${{ inputs.value }}",
					},
					{
						Name: "Run tests",
						Run:  "make test suite=$SUITE",
					},
				},
			},
			expected: 1,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			problems := processJob(tt.id, &tt.job)

			if got, want := len(problems), tt.expected; got != want {
				t.Fatalf("Unexpected number of problems (got '%d', want '%d')", got, want)
			}
		})
	}
}

func TestProcessStep(t *testing.T) {
	runTestCases := []struct {
		name     string
		id       int
		step     Step
		expected []string
	}{
		{
			name: "Unnamed step with no run value",
			step: Step{
				Name: "",
				Run:  "",
			},
			expected: []string{},
		},
		{
			name: "Named step with no run value",
			step: Step{
				Name: "Doesn't run",
				Run:  "",
			},
			expected: []string{},
		},
		{
			name: "Unnamed step with safe run value",
			step: Step{
				Name: "",
				Run:  "echo 'Hello world!'",
			},
			expected: []string{},
		},
		{
			name: "Named step with safe run value",
			step: Step{
				Name: "Run something",
				Run:  "echo 'Hello world!'",
			},
			expected: []string{},
		},
		{
			name: "Unnamed run with one expression",
			id:   42,
			step: Step{
				Name: "",
				Run:  "echo 'Hello ${{ inputs.name }}!'",
			},
			expected: []string{
				"step #42 has '${{ inputs.name }}' in run",
			},
		},
		{
			name: "Named run with one expression",
			step: Step{
				Name: "Greet person",
				Run:  "echo 'Hello ${{ inputs.name }}!'",
			},
			expected: []string{
				"step 'Greet person' has '${{ inputs.name }}' in run",
			},
		},
		{
			name: "Unnamed run with two expressions",
			id:   3,
			step: Step{
				Name: "",
				Run:  "echo 'Hello ${{ inputs.name }}! How is your ${{ steps.id.outputs.day }}'",
			},
			expected: []string{
				"step #3 has '${{ inputs.name }}' in run",
				"step #3 has '${{ steps.id.outputs.day }}' in run",
			},
		},
		{
			name: "Named run with two expressions",
			id:   1,
			step: Step{
				Name: "Greet person today",
				Run:  "echo 'Hello ${{ inputs.name }}! How is your ${{ steps.id.outputs.day }}'",
			},
			expected: []string{
				"step 'Greet person today' has '${{ inputs.name }}' in run",
				"step 'Greet person today' has '${{ steps.id.outputs.day }}' in run",
			},
		},
	}

	actionsGitHubScriptCases := []struct {
		name     string
		id       int
		step     Step
		expected []string
	}{
		{
			name: "Unnamed step using another action",
			step: Step{
				Name: "",
				Uses: "ericcornelissen/non-existent-action",
			},
			expected: []string{},
		},
		{
			name: "Named step using another action",
			step: Step{
				Name: "Doesn't run",
				Uses: "ericcornelissen/non-existent-action",
			},
			expected: []string{},
		},
		{
			name: "Unnamed step with safe script",
			step: Step{
				Name: "",
				Uses: "actions/github-script@v6",
				With: StepWith{
					Script: "console.log('Hello world!')",
				},
			},
			expected: []string{},
		},
		{
			name: "Named step with safe script",
			step: Step{
				Name: "Run something",
				Uses: "actions/github-script@v6",
				With: StepWith{
					Script: "console.log('Hello world!')",
				},
			},
			expected: []string{},
		},
		{
			name: "Unnamed step with unsafe script, one expression",
			id:   42,
			step: Step{
				Name: "",
				Uses: "actions/github-script@v6",
				With: StepWith{
					Script: "console.log('Hello ${{ inputs.name }}!')",
				},
			},
			expected: []string{
				"step #42 has '${{ inputs.name }}' in script",
			},
		},
		{
			name: "Named step with unsafe script, one expression",
			step: Step{
				Name: "Greet person",
				Uses: "actions/github-script@v6",
				With: StepWith{
					Script: "console.log('Hello ${{ inputs.name }}!')",
				},
			},
			expected: []string{
				"step 'Greet person' has '${{ inputs.name }}' in script",
			},
		},
		{
			name: "Unnamed step with unsafe script, two expression",
			id:   3,
			step: Step{
				Name: "",
				Uses: "actions/github-script@v6",
				With: StepWith{
					Script: "console.log('Hello ${{ inputs.name }}! How is your ${{ steps.id.outputs.day }}')",
				},
			},
			expected: []string{
				"step #3 has '${{ inputs.name }}' in script",
				"step #3 has '${{ steps.id.outputs.day }}' in script",
			},
		},
		{
			name: "Named run with two expressions",
			id:   1,
			step: Step{
				Name: "Greet person today",
				Uses: "actions/github-script@v6",
				With: StepWith{
					Script: "console.log('Hello ${{ inputs.name }}! How is your ${{ steps.id.outputs.day }}')",
				},
			},
			expected: []string{
				"step 'Greet person today' has '${{ inputs.name }}' in script",
				"step 'Greet person today' has '${{ steps.id.outputs.day }}' in script",
			},
		},
	}

	allTestCases := append(runTestCases, actionsGitHubScriptCases...)
	for _, tt := range allTestCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			problems := processStep(tt.id, &tt.step)
			if got, want := len(problems), len(tt.expected); got != want {
				t.Fatalf("Unexpected number of problems (got '%d', want '%d')", got, want)
			}

			for i, problem := range problems {
				if got, want := problem, tt.expected[i]; got != want {
					t.Errorf("Unexpected #%d problem (got '%s', want '%s')", i, got, want)
				}
			}
		})
	}
}
