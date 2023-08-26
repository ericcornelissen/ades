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

func TestProcessManifest(t *testing.T) {
	testCases := []struct {
		name     string
		manifest Manifest
		expected int
	}{
		{
			name: "Non-composite manifest",
			manifest: Manifest{
				Runs: ManifestRuns{
					Using: "node16",
				},
			},
			expected: 0,
		},
		{
			name: "Safe manifest",
			manifest: Manifest{
				Runs: ManifestRuns{
					Using: "composite",
					Steps: []JobStep{
						{
							Name: "Example",
							Run:  "",
						},
					},
				},
			},
			expected: 0,
		},
		{
			name: "Problem in first of two steps in manifest",
			manifest: Manifest{
				Runs: ManifestRuns{
					Using: "composite",
					Steps: []JobStep{
						{
							Name: "Example unsafe",
							Run:  "echo ${{ inputs.value }}",
						},
						{
							Name: "Example safe",
							Run:  "echo 'Hello world!'",
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "Problem in second of two steps in manifest",
			manifest: Manifest{
				Runs: ManifestRuns{
					Using: "composite",
					Steps: []JobStep{
						{
							Name: "Example safe",
							Run:  "echo 'Hello world!'",
						},
						{
							Name: "Example unsafe",
							Run:  "echo ${{ inputs.value }}",
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "Problem in all steps in manifest",
			manifest: Manifest{
				Runs: ManifestRuns{
					Using: "composite",
					Steps: []JobStep{
						{
							Name: "Greeting",
							Run:  "echo 'Hello ${{ inputs.name }}!'",
						},
						{
							Name: "Today is",
							Run:  "echo ${{ steps.id.outputs.day }}",
						},
					},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			problems := processManifest(&tt.manifest)

			if got, want := len(problems), tt.expected; got != want {
				t.Fatalf("Unexpected number of problems (got '%d', want '%d')", got, want)
			}
		})
	}
}

func TestProcessWorkflow(t *testing.T) {
	testCases := []struct {
		name     string
		workflow Workflow
		expected int
	}{
		{
			name: "Safe workflow",
			workflow: Workflow{
				Jobs: map[string]WorkflowJob{
					"safe": {
						Name: "Safe",
						Steps: []JobStep{
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
				Jobs: map[string]WorkflowJob{
					"unsafe": {
						Name: "Unsafe",
						Steps: []JobStep{
							{
								Name: "Example",
								Run:  "echo ${{ inputs.value }}",
							},
						},
					},
					"safe": {
						Name: "Safe",
						Steps: []JobStep{
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
				Jobs: map[string]WorkflowJob{
					"safe": {
						Name: "Safe",
						Steps: []JobStep{
							{
								Name: "Example",
								Run:  "echo 'Hello world!'",
							},
						},
					},
					"unsafe": {
						Name: "Unsafe",
						Steps: []JobStep{
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
			name: "Problem in all jobs in workflow",
			workflow: Workflow{
				Jobs: map[string]WorkflowJob{
					"unsafe": {
						Name: "Unsafe",
						Steps: []JobStep{
							{
								Name: "Greeting",
								Run:  "echo 'Hello ${{ inputs.name }}!'",
							},
						},
					},
					"more-unsafe": {
						Name: "More Unsafe",
						Steps: []JobStep{
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
		job      WorkflowJob
		expected int
	}{
		{
			name: "Safe unnamed job",
			job: WorkflowJob{
				Name: "",
				Steps: []JobStep{
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
			job: WorkflowJob{
				Name: "Safe",
				Steps: []JobStep{
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
			job: WorkflowJob{
				Name: "",
				Steps: []JobStep{
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
			job: WorkflowJob{
				Name: "Unsafe",
				Steps: []JobStep{
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
			job: WorkflowJob{
				Name: "",
				Steps: []JobStep{
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
			job: WorkflowJob{
				Name: "Unsafe",
				Steps: []JobStep{
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
	type TestCase struct {
		name     string
		id       int
		step     JobStep
		expected []Problem
	}

	runTestCases := []TestCase{
		{
			name: "Unnamed step with no run value",
			step: JobStep{
				Name: "",
				Run:  "",
			},
			expected: []Problem{},
		},
		{
			name: "Named step with no run value",
			step: JobStep{
				Name: "Doesn't run",
				Run:  "",
			},
			expected: []Problem{},
		},
		{
			name: "Unnamed step with safe run value",
			step: JobStep{
				Name: "",
				Run:  "echo 'Hello world!'",
			},
			expected: []Problem{},
		},
		{
			name: "Named step with safe run value",
			step: JobStep{
				Name: "Run something",
				Run:  "echo 'Hello world!'",
			},
			expected: []Problem{},
		},
		{
			name: "Unnamed run with one expression",
			id:   42,
			step: JobStep{
				Name: "",
				Run:  "echo 'Hello ${{ inputs.name }}!'",
			},
			expected: []Problem{
				{
					stepId:  "#42",
					problem: "${{ inputs.name }}",
				},
			},
		},
		{
			name: "Named run with one expression",
			step: JobStep{
				Name: "Greet person",
				Run:  "echo 'Hello ${{ inputs.name }}!'",
			},
			expected: []Problem{
				{
					stepId:  "'Greet person'",
					problem: "${{ inputs.name }}",
				},
			},
		},
		{
			name: "Unnamed run with two expressions",
			id:   3,
			step: JobStep{
				Name: "",
				Run:  "echo 'Hello ${{ inputs.name }}! How is your ${{ steps.id.outputs.day }}'",
			},
			expected: []Problem{
				{
					stepId:  "#3",
					problem: "${{ inputs.name }}",
				},
				{
					stepId:  "#3",
					problem: "${{ steps.id.outputs.day }}",
				},
			},
		},
		{
			name: "Named run with two expressions",
			id:   1,
			step: JobStep{
				Name: "Greet person today",
				Run:  "echo 'Hello ${{ inputs.name }}! How is your ${{ steps.id.outputs.day }}'",
			},
			expected: []Problem{
				{
					stepId:  "'Greet person today'",
					problem: "${{ inputs.name }}",
				},
				{
					stepId:  "'Greet person today'",
					problem: "${{ steps.id.outputs.day }}",
				},
			},
		},
	}

	actionsGitHubScriptCases := []TestCase{
		{
			name: "Unnamed step using another action",
			step: JobStep{
				Name: "",
				Uses: "ericcornelissen/non-existent-action",
			},
			expected: []Problem{},
		},
		{
			name: "Named step using another action",
			step: JobStep{
				Name: "Doesn't run",
				Uses: "ericcornelissen/non-existent-action",
			},
			expected: []Problem{},
		},
		{
			name: "Unnamed step with safe script",
			step: JobStep{
				Name: "",
				Uses: "actions/github-script@v6",
				With: StepWith{
					Script: "console.log('Hello world!')",
				},
			},
			expected: []Problem{},
		},
		{
			name: "Named step with safe script",
			step: JobStep{
				Name: "Run something",
				Uses: "actions/github-script@v6",
				With: StepWith{
					Script: "console.log('Hello world!')",
				},
			},
			expected: []Problem{},
		},
		{
			name: "Unnamed step with unsafe script, one expression",
			id:   42,
			step: JobStep{
				Name: "",
				Uses: "actions/github-script@v6",
				With: StepWith{
					Script: "console.log('Hello ${{ inputs.name }}!')",
				},
			},
			expected: []Problem{
				{
					stepId:  "#42",
					problem: "${{ inputs.name }}",
				},
			},
		},
		{
			name: "Named step with unsafe script, one expression",
			step: JobStep{
				Name: "Greet person",
				Uses: "actions/github-script@v6",
				With: StepWith{
					Script: "console.log('Hello ${{ inputs.name }}!')",
				},
			},
			expected: []Problem{
				{
					stepId:  "'Greet person'",
					problem: "${{ inputs.name }}",
				},
			},
		},
		{
			name: "Unnamed step with unsafe script, two expression",
			id:   3,
			step: JobStep{
				Name: "",
				Uses: "actions/github-script@v6",
				With: StepWith{
					Script: "console.log('Hello ${{ inputs.name }}! How is your ${{ steps.id.outputs.day }}')",
				},
			},
			expected: []Problem{
				{
					stepId:  "#3",
					problem: "${{ inputs.name }}",
				},
				{
					stepId:  "#3",
					problem: "${{ steps.id.outputs.day }}",
				},
			},
		},
		{
			name: "Named run with two expressions",
			id:   1,
			step: JobStep{
				Name: "Greet person today",
				Uses: "actions/github-script@v6",
				With: StepWith{
					Script: "console.log('Hello ${{ inputs.name }}! How is your ${{ steps.id.outputs.day }}')",
				},
			},
			expected: []Problem{
				{
					stepId:  "'Greet person today'",
					problem: "${{ inputs.name }}",
				},
				{
					stepId:  "'Greet person today'",
					problem: "${{ steps.id.outputs.day }}",
				},
			},
		},
	}

	var allTestCases []TestCase
	allTestCases = append(allTestCases, runTestCases...)
	allTestCases = append(allTestCases, actionsGitHubScriptCases...)

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
