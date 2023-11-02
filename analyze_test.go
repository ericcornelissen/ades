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

import (
	"fmt"
	"strings"
	"testing"
	"testing/quick"
)

func TestAnalyzeManifest(t *testing.T) {
	type TestCase struct {
		name     string
		manifest Manifest
		expected int
	}

	testCases := []TestCase{
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
			violations := analyzeManifest(&tt.manifest)

			if got, want := len(violations), tt.expected; got != want {
				t.Fatalf("Unexpected number of violations (got '%d', want '%d')", got, want)
			}
		})
	}
}

func TestAnalyzeWorkflow(t *testing.T) {
	type TestCase struct {
		name     string
		workflow Workflow
		expected int
	}

	testCases := []TestCase{
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
			violations := analyzeWorkflow(&tt.workflow)

			if got, want := len(violations), tt.expected; got != want {
				t.Fatalf("Unexpected number of violations (got '%d', want '%d')", got, want)
			}
		})
	}
}

func TestAnalyzeJob(t *testing.T) {
	type Expectation struct {
		count int
		id    string
	}

	type TestCase struct {
		name     string
		id       string
		job      WorkflowJob
		expected Expectation
	}

	testCases := []TestCase{
		{
			name: "Safe unnamed job",
			id:   "job-id",
			job: WorkflowJob{
				Name: "",
				Steps: []JobStep{
					{
						Name: "Unnamed Example",
						Run:  "",
					},
				},
			},
			expected: Expectation{
				count: 0,
				id:    "job-id",
			},
		},
		{
			name: "Safe named job",
			id:   "job-id",
			job: WorkflowJob{
				Name: "Safe",
				Steps: []JobStep{
					{
						Name: "Named example",
						Run:  "",
					},
				},
			},
			expected: Expectation{
				count: 0,
				id:    "job-id",
			},
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
			expected: Expectation{
				count: 1,
				id:    "job-id",
			},
		},
		{
			name: "Named job with unsafe step",
			id:   "job-id",
			job: WorkflowJob{
				Name: "Unsafe",
				Steps: []JobStep{
					{
						Name: "Example",
						Run:  "echo ${{ inputs.value }}",
					},
				},
			},
			expected: Expectation{
				count: 1,
				id:    "Unsafe",
			},
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
			expected: Expectation{
				count: 1,
				id:    "job-id",
			},
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
			expected: Expectation{
				count: 1,
				id:    "Unsafe",
			},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			violations := analyzeJob(tt.id, &tt.job)

			if got, want := len(violations), tt.expected.count; got != want {
				t.Fatalf("Unexpected number of violations (got '%d', want '%d')", got, want)
			}

			for i, violation := range violations {
				if got, want := violation.jobId, tt.expected.id; got != want {
					t.Errorf("Unexpected job ID for violation %d (got '%s', want '%s')", i, got, want)
				}
			}
		})
	}
}

func TestAnalyzeStep(t *testing.T) {
	type TestCase struct {
		name     string
		id       int
		step     JobStep
		expected []Violation
	}

	runTestCases := []TestCase{
		{
			name: "Unnamed step with no run value",
			step: JobStep{
				Name: "",
				Run:  "",
			},
			expected: []Violation{},
		},
		{
			name: "Named step with no run value",
			step: JobStep{
				Name: "Doesn't run",
				Run:  "",
			},
			expected: []Violation{},
		},
		{
			name: "Unnamed step with safe run value",
			step: JobStep{
				Name: "",
				Run:  "echo 'Hello world!'",
			},
			expected: []Violation{},
		},
		{
			name: "Named step with safe run value",
			step: JobStep{
				Name: "Run something",
				Run:  "echo 'Hello world!'",
			},
			expected: []Violation{},
		},
		{
			name: "Unnamed run with one expression",
			id:   42,
			step: JobStep{
				Name: "",
				Run:  "echo 'Hello ${{ inputs.name }}!'",
			},
			expected: []Violation{
				{
					stepId:  "#42",
					problem: "${{ inputs.name }}",
					kind:    ExpressionInRunScript,
				},
			},
		},
		{
			name: "Named run with one expression",
			step: JobStep{
				Name: "Greet person",
				Run:  "echo 'Hello ${{ inputs.name }}!'",
			},
			expected: []Violation{
				{
					stepId:  "Greet person",
					problem: "${{ inputs.name }}",
					kind:    ExpressionInRunScript,
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
			expected: []Violation{
				{
					stepId:  "#3",
					problem: "${{ inputs.name }}",
					kind:    ExpressionInRunScript,
				},
				{
					stepId:  "#3",
					problem: "${{ steps.id.outputs.day }}",
					kind:    ExpressionInRunScript,
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
			expected: []Violation{
				{
					stepId:  "Greet person today",
					problem: "${{ inputs.name }}",
					kind:    ExpressionInRunScript,
				},
				{
					stepId:  "Greet person today",
					problem: "${{ steps.id.outputs.day }}",
					kind:    ExpressionInRunScript,
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
			expected: []Violation{},
		},
		{
			name: "Named step using another action",
			step: JobStep{
				Name: "Doesn't run",
				Uses: "ericcornelissen/non-existent-action",
			},
			expected: []Violation{},
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
			expected: []Violation{},
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
			expected: []Violation{},
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
			expected: []Violation{
				{
					stepId:  "#42",
					problem: "${{ inputs.name }}",
					kind:    ExpressionInActionsGithubScript,
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
			expected: []Violation{
				{
					stepId:  "Greet person",
					problem: "${{ inputs.name }}",
					kind:    ExpressionInActionsGithubScript,
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
			expected: []Violation{
				{
					stepId:  "#3",
					problem: "${{ inputs.name }}",
					kind:    ExpressionInActionsGithubScript,
				},
				{
					stepId:  "#3",
					problem: "${{ steps.id.outputs.day }}",
					kind:    ExpressionInActionsGithubScript,
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
			expected: []Violation{
				{
					stepId:  "Greet person today",
					problem: "${{ inputs.name }}",
					kind:    ExpressionInActionsGithubScript,
				},
				{
					stepId:  "Greet person today",
					problem: "${{ steps.id.outputs.day }}",
					kind:    ExpressionInActionsGithubScript,
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
			violations := analyzeStep(tt.id, &tt.step)
			if got, want := len(violations), len(tt.expected); got != want {
				t.Fatalf("Unexpected number of violations (got '%d', want '%d')", got, want)
			}

			for i, violation := range violations {
				if got, want := violation, tt.expected[i]; got != want {
					t.Errorf("Unexpected #%d violation (got '%v', want '%v')", i, got, want)
				}
			}
		})
	}
}

func TestIsRunStep(t *testing.T) {
	t.Run("Run step", func(t *testing.T) {
		testCases := []JobStep{
			{
				Run: "echo 'Hello world!'",
			},
			{
				Run: "echo 'Hello'\necho 'world!'",
			},
			{
				Run: "a",
			},
		}

		for _, tt := range testCases {
			tt := tt
			if !isRunStep(&tt) {
				t.Errorf("Run step not identified: '%s'", tt)
			}
		}

		f := func(step JobStep, run string) bool {
			if len(run) == 0 {
				return true
			}

			step.Run = run
			return isRunStep(&step)
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Non-run step", func(t *testing.T) {
		f := func(step JobStep) bool {
			step.Run = ""
			return !isRunStep(&step)
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestIsActionsGitHubScriptStep(t *testing.T) {
	t.Run("Script step", func(t *testing.T) {
		f := func(step JobStep, ref string) bool {
			step.Uses = fmt.Sprintf("actions/github-script@%s", ref)
			return isActionsGitHubScriptStep(&step)
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Non-script step", func(t *testing.T) {
		f := func(step JobStep) bool {
			if strings.HasPrefix(step.Uses, "actions/github-script@") {
				return true
			}

			return !isActionsGitHubScriptStep(&step)
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})
}
