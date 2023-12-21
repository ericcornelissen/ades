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
	"testing"
	"testing/quick"
)

func TestViolationKindString(t *testing.T) {
	type TestCase struct {
		kind violationKind
		want string
	}

	testCases := []TestCase{
		{
			kind: expressionInRunScript,
			want: expressionInRunScriptId,
		},
		{
			kind: expressionInActionsGithubScript,
			want: expressionInActionsGithubScriptId,
		},
		{
			kind: expressionInGitTagAnnotationActionTagInput,
			want: expressionInGitTagAnnotationActionTagInputId,
		},
		{
			kind: expressionInGitMessageActionShaInput,
			want: expressionInGitMessageActionShaInputId,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(fmt.Sprint(tt.kind), func(t *testing.T) {
			t.Parallel()

			if got, want := tt.kind.String(), tt.want; got != want {
				t.Errorf("Unexpected result (got %q, want %q)", got, want)
			}
		})
	}
}

func TestAnalyzeManifest(t *testing.T) {
	type TestCase struct {
		name     string
		manifest Manifest
		want     int
	}

	testCases := []TestCase{
		{
			name: "Non-composite manifest",
			manifest: Manifest{
				Runs: ManifestRuns{
					Using: "node16",
					Steps: []JobStep{
						{
							Name: "Example unsafe",
							Run:  "echo ${{ inputs.value }}",
						},
					},
				},
			},
			want: 0,
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
			want: 0,
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
			want: 1,
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
			want: 1,
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
			want: 2,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			violations := analyzeManifest(&tt.manifest)
			if got, want := len(violations), tt.want; got != want {
				t.Fatalf("Unexpected number of violations (got %d, want %d)", got, want)
			}
		})
	}

	t.Run("nil pointer", func(t *testing.T) {
		violations := analyzeManifest(nil)
		if got, want := len(violations), 0; got != want {
			t.Fatalf("Unexpected number of violations (got %d, want %d)", got, want)
		}
	})
}

func TestAnalyzeWorkflow(t *testing.T) {
	type TestCase struct {
		name     string
		workflow Workflow
		want     int
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
			want: 0,
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
			want: 1,
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
			want: 1,
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
			want: 3,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			violations := analyzeWorkflow(&tt.workflow)
			if got, want := len(violations), tt.want; got != want {
				t.Fatalf("Unexpected number of violations (got %d, want %d)", got, tt.want)
			}
		})
	}

	t.Run("nil pointer", func(t *testing.T) {
		violations := analyzeWorkflow(nil)
		if got, want := len(violations), 0; got != want {
			t.Fatalf("Unexpected number of violations (got %d, want %d)", got, want)
		}
	})
}

func TestAnalyzeJob(t *testing.T) {
	type TestCase struct {
		name      string
		id        string
		job       WorkflowJob
		wantCount int
		wantId    string
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
			wantCount: 0,
			wantId:    "job-id",
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
			wantCount: 0,
			wantId:    "job-id",
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
			wantCount: 1,
			wantId:    "job-id",
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
			wantCount: 1,
			wantId:    "Unsafe",
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
			wantCount: 1,
			wantId:    "job-id",
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
			wantCount: 1,
			wantId:    "Unsafe",
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			violations := analyzeJob(tt.id, &tt.job)
			if got, want := len(violations), tt.wantCount; got != want {
				t.Fatalf("Unexpected number of violations (got %d, want %d)", got, want)
			}

			for i, v := range violations {
				if got, want := v.jobId, tt.wantId; got != want {
					t.Errorf("Unexpected job ID for violation %d (got %q, want %q)", i, got, want)
				}
			}
		})
	}
}

func TestAnalyzeStep(t *testing.T) {
	type TestCase struct {
		name string
		id   int
		step JobStep
		want []violation
	}

	runTestCases := []TestCase{
		{
			name: "Unnamed step with no run value",
			step: JobStep{
				Name: "",
				Run:  "",
			},
			want: []violation{},
		},
		{
			name: "Named step with no run value",
			step: JobStep{
				Name: "Doesn't run",
				Run:  "",
			},
			want: []violation{},
		},
		{
			name: "Unnamed step with safe run value",
			step: JobStep{
				Name: "",
				Run:  "echo 'Hello world!'",
			},
			want: []violation{},
		},
		{
			name: "Named step with safe run value",
			step: JobStep{
				Name: "Run something",
				Run:  "echo 'Hello world!'",
			},
			want: []violation{},
		},
		{
			name: "Unnamed run with one expression",
			id:   42,
			step: JobStep{
				Name: "",
				Run:  "echo 'Hello ${{ inputs.name }}!'",
			},
			want: []violation{
				{
					stepId:  "#42",
					problem: "${{ inputs.name }}",
					kind:    expressionInRunScript,
				},
			},
		},
		{
			name: "Named run with one expression",
			step: JobStep{
				Name: "Greet person",
				Run:  "echo 'Hello ${{ inputs.name }}!'",
			},
			want: []violation{
				{
					stepId:  "Greet person",
					problem: "${{ inputs.name }}",
					kind:    expressionInRunScript,
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
			want: []violation{
				{
					stepId:  "#3",
					problem: "${{ inputs.name }}",
					kind:    expressionInRunScript,
				},
				{
					stepId:  "#3",
					problem: "${{ steps.id.outputs.day }}",
					kind:    expressionInRunScript,
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
			want: []violation{
				{
					stepId:  "Greet person today",
					problem: "${{ inputs.name }}",
					kind:    expressionInRunScript,
				},
				{
					stepId:  "Greet person today",
					problem: "${{ steps.id.outputs.day }}",
					kind:    expressionInRunScript,
				},
			},
		},
	}

	actionsGitHubScriptCases := []TestCase{
		{
			name: "Unnamed step using another action",
			step: JobStep{
				Name: "",
				Uses: "ericcornelissen/non-existent-action@1.0.0",
			},
			want: []violation{},
		},
		{
			name: "Named step using another action",
			step: JobStep{
				Name: "Doesn't run",
				Uses: "ericcornelissen/non-existent-action@1.0.0",
			},
			want: []violation{},
		},
		{
			name: "Unnamed step with safe script",
			step: JobStep{
				Name: "",
				Uses: "actions/github-script@v6",
				With: map[string]string{
					"script": "console.log('Hello world!')",
				},
			},
			want: []violation{},
		},
		{
			name: "Named step with safe script",
			step: JobStep{
				Name: "Run something",
				Uses: "actions/github-script@v6",
				With: map[string]string{
					"script": "console.log('Hello world!')",
				},
			},
			want: []violation{},
		},
		{
			name: "Unnamed step with unsafe script, one expression",
			id:   42,
			step: JobStep{
				Name: "",
				Uses: "actions/github-script@v6",
				With: map[string]string{
					"script": "console.log('Hello ${{ inputs.name }}!')",
				},
			},
			want: []violation{
				{
					stepId:  "#42",
					problem: "${{ inputs.name }}",
					kind:    expressionInActionsGithubScript,
				},
			},
		},
		{
			name: "Named step with unsafe script, one expression",
			step: JobStep{
				Name: "Greet person",
				Uses: "actions/github-script@v6",
				With: map[string]string{
					"script": "console.log('Hello ${{ inputs.name }}!')",
				},
			},
			want: []violation{
				{
					stepId:  "Greet person",
					problem: "${{ inputs.name }}",
					kind:    expressionInActionsGithubScript,
				},
			},
		},
		{
			name: "Unnamed step with unsafe script, two expression",
			id:   3,
			step: JobStep{
				Name: "",
				Uses: "actions/github-script@v6",
				With: map[string]string{
					"script": "console.log('Hello ${{ inputs.name }}! How is your ${{ steps.id.outputs.day }}')",
				},
			},
			want: []violation{
				{
					stepId:  "#3",
					problem: "${{ inputs.name }}",
					kind:    expressionInActionsGithubScript,
				},
				{
					stepId:  "#3",
					problem: "${{ steps.id.outputs.day }}",
					kind:    expressionInActionsGithubScript,
				},
			},
		},
		{
			name: "Named run with two expressions",
			id:   1,
			step: JobStep{
				Name: "Greet person today",
				Uses: "actions/github-script@v6",
				With: map[string]string{
					"script": "console.log('Hello ${{ inputs.name }}! How is your ${{ steps.id.outputs.day }}')",
				},
			},
			want: []violation{
				{
					stepId:  "Greet person today",
					problem: "${{ inputs.name }}",
					kind:    expressionInActionsGithubScript,
				},
				{
					stepId:  "Greet person today",
					problem: "${{ steps.id.outputs.day }}",
					kind:    expressionInActionsGithubScript,
				},
			},
		},
	}

	actionTestCases := []TestCase{
		{
			name: "git-tag-annotation-action, vulnerable version",
			step: JobStep{
				Name: "Vulnerable",
				Uses: "ericcornelissen/git-tag-annotation-action@v1.0.0",
				With: map[string]string{
					"tag": "${{ inputs.tag }}",
				},
			},
			want: []violation{
				{
					stepId:  "Vulnerable",
					problem: "${{ inputs.tag }}",
					kind:    expressionInGitTagAnnotationActionTagInput,
				},
			},
		},
		{
			name: "git-tag-annotation-action, old version",
			step: JobStep{
				Name: "Old",
				Uses: "ericcornelissen/git-tag-annotation-action@v0.0.9",
				With: map[string]string{
					"tag": "${{ inputs.tag }}",
				},
			},
			want: []violation{
				{
					stepId:  "Old",
					problem: "${{ inputs.tag }}",
					kind:    expressionInGitTagAnnotationActionTagInput,
				},
			},
		},
		{
			name: "git-tag-annotation-action, fixed version",
			step: JobStep{
				Name: "Fixed",
				Uses: "ericcornelissen/git-tag-annotation-action@v1.0.1",
				With: map[string]string{
					"tag": "${{ inputs.tag }}",
				},
			},
			want: []violation{},
		},
	}

	gitMessageActionTestCases := []TestCase{
		{
			name: "git-message-action, vulnerable version without vulnerable input",
			step: JobStep{
				Name: "Vulnerable",
				Uses: "kceb/git-message-action@v1.1.0",
				With: map[string]string{},
			},
			want: []violation{},
		},
		{
			name: "git-message-action, vulnerable version with vulnerable input",
			step: JobStep{
				Name: "Vulnerable",
				Uses: "kceb/git-message-action@v1.1.0",
				With: map[string]string{
					"sha": "${{ inputs.sha }}",
				},
			},
			want: []violation{
				{
					stepId:  "Vulnerable",
					problem: "${{ inputs.sha }}",
					kind:    expressionInGitMessageActionShaInput,
				},
			},
		},
		{
			name: "git-tag-annotation-action, old version without vulnerable input",
			step: JobStep{
				Name: "Old",
				Uses: "kceb/git-message-action@v1.0.0",
				With: map[string]string{},
			},
			want: []violation{},
		},
		{
			name: "git-tag-annotation-action, old version with vulnerable input",
			step: JobStep{
				Name: "Old",
				Uses: "kceb/git-message-action@v1.0.0",
				With: map[string]string{
					"sha": "${{ inputs.sha }}",
				},
			},
			want: []violation{
				{
					stepId:  "Old",
					problem: "${{ inputs.sha }}",
					kind:    expressionInGitMessageActionShaInput,
				},
			},
		},
		{
			name: "git-message-action, fixed version without vulnerable input",
			step: JobStep{
				Name: "Fixed",
				Uses: "kceb/git-message-action@v1.2.0",
				With: map[string]string{},
			},
			want: []violation{},
		},
		{
			name: "git-message-action, fixed version with vulnerable input",
			step: JobStep{
				Name: "Fixed",
				Uses: "kceb/git-message-action@v1.2.0",
				With: map[string]string{
					"sha": "${{ inputs.sha }}",
				},
			},
			want: []violation{},
		},
	}

	var allTestCases []TestCase
	allTestCases = append(allTestCases, runTestCases...)
	allTestCases = append(allTestCases, actionsGitHubScriptCases...)
	allTestCases = append(allTestCases, actionTestCases...)
	allTestCases = append(allTestCases, gitMessageActionTestCases...)

	for _, tt := range allTestCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			violations := analyzeStep(tt.id, &tt.step)
			if got, want := len(violations), len(tt.want); got != want {
				t.Fatalf("Unexpected number of violations (got %d, want %d)", got, want)
			}

			for i, v := range violations {
				if got, want := v, tt.want[i]; got != want {
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
				t.Errorf("Run step not identified: %q", tt)
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

func TestIsBeforeOrAtVersion(t *testing.T) {
	type TestCase struct {
		name    string
		uses    StepUses
		version string
		want    bool
	}

	testCases := []TestCase{
		{
			name: "At, full semantic version",
			uses: StepUses{
				Ref: "v1.0.0",
			},
			version: "v1.0.0",
			want:    true,
		},
		{
			name: "Before, full semantic version",
			uses: StepUses{
				Ref: "v0.1.0",
			},
			version: "v1.0.0",
			want:    true,
		},
		{
			name: "After, full semantic version",
			uses: StepUses{
				Ref: "v1.0.1",
			},
			version: "v1.0.0",
			want:    false,
		},
		{
			name: "SHA",
			uses: StepUses{
				Ref: "21fa0360d55070a1d6b999d027db44cc21a7b48d",
			},
			version: "v1.0.0",
			want:    false,
		},

		{
			name: "Major version only",
			uses: StepUses{
				Ref: "v1",
			},
			version: "v1.0.0",
			want:    false,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got, want := isBeforeOrAtVersion(tt.uses, tt.version), tt.want; got != want {
				t.Errorf("Wrong answer for given %s compared to %s (got %t, want %t)", tt.uses.Ref, tt.version, got, want)
			}
		})
	}
}
