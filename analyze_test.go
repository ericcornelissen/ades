// Copyright (C) 2024  Eric Cornelissen
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
	"testing"
)

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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			violations := AnalyzeManifest(&tt.manifest)
			if got, want := len(violations), tt.want; got != want {
				t.Fatalf("Unexpected number of violations (got %d, want %d)", got, want)
			}
		})
	}

	t.Run("nil pointer", func(t *testing.T) {
		violations := AnalyzeManifest(nil)
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			violations := AnalyzeWorkflow(&tt.workflow)
			if got, want := len(violations), tt.want; got != want {
				t.Fatalf("Unexpected number of violations (got %d, want %d)", got, tt.want)
			}
		})
	}

	t.Run("nil pointer", func(t *testing.T) {
		violations := AnalyzeWorkflow(nil)
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			violations := analyzeJob(tt.id, &tt.job)
			if got, want := len(violations), tt.wantCount; got != want {
				t.Fatalf("Unexpected number of violations (got %d, want %d)", got, want)
			}

			for i, violation := range violations {
				if got, want := violation.JobId, tt.wantId; got != want {
					t.Errorf("Unexpected job ID for violation %d (got %q, want %q)", i, got, want)
				}
			}
		})
	}
}

func TestAnalyzeStep(t *testing.T) {
	type TestCase struct {
		name       string
		id         int
		step       JobStep
		wantCount  int
		wantStepId string
	}

	testCases := []TestCase{
		{
			name: "Unnamed step that does nothing",
			step: JobStep{
				Name: "",
			},
			wantCount: 0,
		},
		{
			name: "Named step that does nothing",
			step: JobStep{
				Name: "Doesn't run",
			},
			wantCount: 0,
		},
		{
			name: "Unnamed step without violation",
			step: JobStep{
				Name: "",
				Run:  "echo 'Hello world!'",
			},
			wantCount: 0,
		},
		{
			name: "Named step without violation",
			step: JobStep{
				Name: "Run something",
				Run:  "echo 'Hello world!'",
			},
			wantCount: 0,
		},
		{
			name: "Unnamed step with one violation",
			id:   42,
			step: JobStep{
				Name: "",
				Run:  "echo 'Hello ${{ inputs.name }}!'",
			},
			wantCount:  1,
			wantStepId: "#42",
		},
		{
			name: "Named step with one violation",
			step: JobStep{
				Name: "Greet person",
				Run:  "echo 'Hello ${{ inputs.name }}!'",
			},
			wantCount:  1,
			wantStepId: "Greet person",
		},
		{
			name: "Unnamed step with two violation",
			id:   3,
			step: JobStep{
				Name: "",
				Run:  "echo 'Hello ${{ inputs.name }}! How is your ${{ steps.id.outputs.day }}'",
			},
			wantCount:  2,
			wantStepId: "#3",
		},
		{
			name: "Named step with two violation",
			id:   1,
			step: JobStep{
				Name: "Greet person today",
				Run:  "echo 'Hello ${{ inputs.name }}! How is your ${{ steps.id.outputs.day }}'",
			},
			wantCount:  2,
			wantStepId: "Greet person today",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			violations := analyzeStep(tt.id, &tt.step)
			if got, want := len(violations), tt.wantCount; got != want {
				t.Fatalf("Unexpected number of violations (got %d, want %d)", got, want)
			}

			for i, violation := range violations {
				if got, want := violation.StepId, tt.wantStepId; got != want {
					t.Errorf("Unexpected step id for violation #%d (got %q, want %q)", i, got, want)
				}
			}
		})
	}
}

func TestAnalyzeString(t *testing.T) {
	type TestCase struct {
		name  string
		value string
		want  []Violation
	}

	testCases := []TestCase{
		{
			name:  "Simple and safe",
			value: "echo 'Hello world!'",
			want:  []Violation{},
		},
		{
			name:  "Multiline and safe",
			value: "echo 'Hello'\necho 'world!'",
			want:  []Violation{},
		},
		{
			name:  "One violations",
			value: "echo 'Hello ${{ input.name }}!'",
			want: []Violation{
				{Problem: "${{ input.name }}"},
			},
		},
		{
			name:  "Two violations",
			value: "echo '${{ input.greeting }} ${{ input.name }}!'",
			want: []Violation{
				{Problem: "${{ input.greeting }}"},
				{Problem: "${{ input.name }}"},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			violations := analyzeString(tt.value)
			if got, want := len(violations), len(tt.want); got != want {
				t.Fatalf("Unexpected number of violations (got %d, want %d)", got, want)
			}

			for i, violation := range violations {
				if got, want := violation, tt.want[i]; got != want {
					t.Errorf("Unexpected #%d violation (got '%v', want '%v')", i, got, want)
				}
			}
		})
	}
}
