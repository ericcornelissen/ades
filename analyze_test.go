// Copyright (C) 2023-2026  Eric Cornelissen
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

package ades

import (
	"io"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/ericcornelissen/go-gha-models"
)

func TestAnalyzeRepo(t *testing.T) {
	type TestCase struct {
		fsys    fs.FS
		matcher ExprMatcher
		want    int
	}

	testCases := map[string]TestCase{
		"Repository with one workflow": {
			fsys: fstest.MapFS{
				".github/workflows/example.yml": &fstest.MapFile{
					Data: []byte(`name: Example workflow with a ADES100 violation
on: [push]
jobs:
  example:
    runs-on: ubuntu-latest
    steps:
    - run: echo 'Hello ${{ inputs.name }}'
`),
				},
			},
			matcher: AllMatcher,
			want:    1,
		},
		"Repository with two workflows": {
			fsys: fstest.MapFS{
				".github/workflows/ades100.yml": &fstest.MapFile{
					Data: []byte(`name: Example workflow with a ADES100 violation
on: [push]
jobs:
  example:
    runs-on: ubuntu-latest
    steps:
    - run: echo 'Hello ${{ inputs.name }}'
`),
				},
				".github/workflows/ades101.yml": &fstest.MapFile{
					Data: []byte(`name: Example workflow with a ADES101 violation
on: [push]
jobs:
  example:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/github-script@v6
      with:
        script: console.log('Hello ${{ inputs.name }}')
`),
				},
			},
			matcher: AllMatcher,
			want:    2,
		},
		"Repository with manifest in root": {
			fsys: fstest.MapFS{
				"action.yml": &fstest.MapFile{
					Data: []byte(`name: Example action
description: An example action.

runs:
  using: composite
  steps:
  - run: echo 'Hello, ${{ inputs.name }}!'
`),
				},
			},
			matcher: AllMatcher,
			want:    1,
		},
		"Repository with manifest in directory": {
			fsys: fstest.MapFS{
				"example/action.yml": &fstest.MapFile{
					Data: []byte(`name: Example action
description: An example action.

runs:
  using: composite
  steps:
  - run: echo 'Hello, ${{ inputs.name }}!'
`),
				},
			},
			matcher: AllMatcher,
			want:    1,
		},
		"Skip .git directory": {
			fsys: fstest.MapFS{
				".git/action.yml": &fstest.MapFile{
					Data: []byte(`name: Example action
description: An example action.

runs:
  using: composite
  steps:
  - run: echo 'Hello, ${{ inputs.name }}!'
`),
				},
			},
			matcher: AllMatcher,
			want:    0,
		},
		"Skip irrelevant file that cannot be opened": {
			fsys: FailOpenFS{
				FS: fstest.MapFS{
					".github/workflows/example.yml": &fstest.MapFile{
						Data: []byte(`name: Purely illustrative`),
					},
					"wham.mp3": &fstest.MapFile{},
				},
				Fail: map[string]error{
					"wham.mp3": fs.ErrNotExist,
				},
			},
			matcher: AllMatcher,
			want:    0,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			report, err := AnalyzeRepo(tt.fsys, AllMatcher)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			count := 0
			for _, v := range report {
				count += len(v)
			}

			if got, want := count, tt.want; got != want {
				t.Fatalf("unexpected report size (got %d, want %d; %v)", got, want, report)
			}
		})
	}

	errCases := map[string]TestCase{
		"Cannot open dir": {
			fsys: FailOpenFS{
				FS: fstest.MapFS{
					".github/": &fstest.MapFile{},
				},
				Fail: map[string]error{
					".github": fs.ErrPermission,
				},
			},
			matcher: AllMatcher,
		},
		"Cannot open workflow": {
			fsys: FailOpenFS{
				FS: fstest.MapFS{
					".github/workflows/example.yml": &fstest.MapFile{},
				},
				Fail: map[string]error{
					".github/workflows/example.yml": fs.ErrNotExist,
				},
			},
			matcher: AllMatcher,
		},
		"Cannot open manifest": {
			fsys: FailOpenFS{
				FS: fstest.MapFS{
					"action.yml": &fstest.MapFile{},
				},
				Fail: map[string]error{
					"action.yml": fs.ErrNotExist,
				},
			},
			matcher: AllMatcher,
		},
		"Cannot read workflow": {
			fsys: FailReadFS{
				FS: fstest.MapFS{
					".github/workflows/example.yml": &fstest.MapFile{},
				},
				Fail: map[string]error{
					".github/workflows/example.yml": io.ErrNoProgress,
				},
			},
			matcher: AllMatcher,
		},
		"Cannot read manifest": {
			fsys: FailReadFS{
				FS: fstest.MapFS{
					"action.yml": &fstest.MapFile{},
				},
				Fail: map[string]error{
					"action.yml": io.ErrNoProgress,
				},
			},
			matcher: AllMatcher,
		},
		"Corrupt workflow": {
			fsys: fstest.MapFS{
				".github/workflows/example.yml": &fstest.MapFile{
					Data: []byte(`* this is not YAML`),
				},
			},
			matcher: AllMatcher,
		},
		"Corrupt manifest": {
			fsys: fstest.MapFS{
				"action.yml": &fstest.MapFile{
					Data: []byte(`* this is not YAML`),
				},
			},
			matcher: AllMatcher,
		},
	}

	for name, tt := range errCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, err := AnalyzeRepo(tt.fsys, AllMatcher); err == nil {
				t.Fatal("Unexpected success")
			}
		})
	}
}

func TestAnalyzeManifest(t *testing.T) {
	type TestCase struct {
		manifest gha.Manifest
		matcher  ExprMatcher
		want     int
	}

	testCases := map[string]TestCase{
		"Non-composite manifest": {
			manifest: gha.Manifest{
				Runs: gha.Runs{
					Using: "node16",
					Steps: []gha.Step{
						{
							Name: "Example unsafe",
							Run:  "echo ${{ inputs.value }}",
						},
					},
				},
			},
			matcher: AllMatcher,
			want:    0,
		},
		"Safe manifest": {
			manifest: gha.Manifest{
				Runs: gha.Runs{
					Using: "composite",
					Steps: []gha.Step{
						{
							Name: "Example",
							Run:  "",
						},
					},
				},
			},
			matcher: AllMatcher,
			want:    0,
		},
		"Problem in first of two steps in manifest": {
			manifest: gha.Manifest{
				Runs: gha.Runs{
					Using: "composite",
					Steps: []gha.Step{
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
			matcher: AllMatcher,
			want:    1,
		},
		"Problem in second of two steps in manifest": {
			manifest: gha.Manifest{
				Runs: gha.Runs{
					Using: "composite",
					Steps: []gha.Step{
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
			matcher: AllMatcher,
			want:    1,
		},
		"Problem in all steps in manifest": {
			manifest: gha.Manifest{
				Runs: gha.Runs{
					Using: "composite",
					Steps: []gha.Step{
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
			matcher: AllMatcher,
			want:    2,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			violations := AnalyzeManifest(&tt.manifest, tt.matcher)
			if got, want := len(violations), tt.want; got != want {
				t.Fatalf("Unexpected number of violations (got %d, want %d)", got, want)
			}

			for _, violation := range violations {
				if got, want := violation.source, &tt.manifest; got != want {
					t.Errorf("Violation source is not the manifest (got %v, want %v)", got, want)
				}
			}
		})
	}

	t.Run("nil pointer", func(t *testing.T) {
		violations := AnalyzeManifest(nil, AllMatcher)
		if got, want := len(violations), 0; got != want {
			t.Fatalf("Unexpected number of violations (got %d, want %d)", got, want)
		}
	})
}

func TestAnalyzeWorkflow(t *testing.T) {
	type TestCase struct {
		workflow gha.Workflow
		matcher  ExprMatcher
		want     int
	}

	testCases := map[string]TestCase{
		"Safe workflow": {
			workflow: gha.Workflow{
				Jobs: map[string]gha.Job{
					"safe": {
						Name: "Safe",
						Steps: []gha.Step{
							{
								Name: "Example",
								Run:  "",
							},
						},
					},
				},
			},
			matcher: AllMatcher,
			want:    0,
		},
		"Problem in first of two jobs in workflow": {
			workflow: gha.Workflow{
				Jobs: map[string]gha.Job{
					"unsafe": {
						Name: "Unsafe",
						Steps: []gha.Step{
							{
								Name: "Example",
								Run:  "echo ${{ inputs.value }}",
							},
						},
					},
					"safe": {
						Name: "Safe",
						Steps: []gha.Step{
							{
								Name: "Example",
								Run:  "echo 'Hello world!'",
							},
						},
					},
				},
			},
			matcher: AllMatcher,
			want:    1,
		},
		"Problem in second of two jobs in workflow": {
			workflow: gha.Workflow{
				Jobs: map[string]gha.Job{
					"safe": {
						Name: "Safe",
						Steps: []gha.Step{
							{
								Name: "Example",
								Run:  "echo 'Hello world!'",
							},
						},
					},
					"unsafe": {
						Name: "Unsafe",
						Steps: []gha.Step{
							{
								Name: "Example",
								Run:  "echo ${{ inputs.value }}",
							},
						},
					},
				},
			},
			matcher: AllMatcher,
			want:    1,
		},
		"Problem in all jobs in workflow": {
			workflow: gha.Workflow{
				Jobs: map[string]gha.Job{
					"unsafe": {
						Name: "Unsafe",
						Steps: []gha.Step{
							{
								Name: "Greeting",
								Run:  "echo 'Hello ${{ inputs.name }}!'",
							},
						},
					},
					"more-unsafe": {
						Name: "More Unsafe",
						Steps: []gha.Step{
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
			matcher: AllMatcher,
			want:    3,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			violations := AnalyzeWorkflow(&tt.workflow, tt.matcher)
			if got, want := len(violations), tt.want; got != want {
				t.Fatalf("Unexpected number of violations (got %d, want %d)", got, tt.want)
			}

			for _, violation := range violations {
				if got, want := violation.source, &tt.workflow; got != want {
					t.Errorf("Violation source is not the workflow (got %v, want %v)", got, want)
				}
			}
		})
	}

	t.Run("nil pointer", func(t *testing.T) {
		violations := AnalyzeWorkflow(nil, AllMatcher)
		if got, want := len(violations), 0; got != want {
			t.Fatalf("Unexpected number of violations (got %d, want %d)", got, want)
		}
	})
}

func TestAnalyzeJob(t *testing.T) {
	type TestCase struct {
		id        string
		job       gha.Job
		matcher   ExprMatcher
		wantCount int
		wantId    string
	}

	testCases := map[string]TestCase{
		"Safe unnamed job": {
			id: "job-id",
			job: gha.Job{
				Name: "",
				Steps: []gha.Step{
					{
						Name: "Unnamed Example",
						Run:  "",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 0,
		},
		"Safe named job": {
			id: "job-id",
			job: gha.Job{
				Name: "Safe",
				Steps: []gha.Step{
					{
						Name: "Named example",
						Run:  "",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 0,
		},
		"Unnamed job with unsafe step": {
			id: "job-id",
			job: gha.Job{
				Name: "",
				Steps: []gha.Step{
					{
						Name: "Example",
						Run:  "echo ${{ inputs.value }}",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 1,
			wantId:    "job-id",
		},
		"Named job with unsafe step": {
			id: "job-id",
			job: gha.Job{
				Name: "Unsafe",
				Steps: []gha.Step{
					{
						Name: "Example",
						Run:  "echo ${{ inputs.value }}",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 1,
			wantId:    "Unsafe",
		},
		"Unnamed job with unsafe and safe steps": {
			id: "job-id",
			job: gha.Job{
				Name: "",
				Steps: []gha.Step{
					{
						Name: "Checkout repository",
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
			matcher:   AllMatcher,
			wantCount: 1,
			wantId:    "job-id",
		},
		"Named job with unsafe and safe steps": {
			job: gha.Job{
				Name: "Unsafe",
				Steps: []gha.Step{
					{
						Name: "Checkout repository",
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
			matcher:   AllMatcher,
			wantCount: 1,
			wantId:    "Unsafe",
		},
		"Job with two unsafe expressions in one step": {
			job: gha.Job{
				Name: "Unsafe",
				Steps: []gha.Step{
					{
						Name: "Checkout repository",
					},
					{
						Name: "Example",
						Run:  "echo ${{ inputs.foo }}\necho ${{ inputs.bar }}",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 2,
			wantId:    "Unsafe",
		},
		"Job with two unsafe expressions in separate steps": {
			job: gha.Job{
				Name: "Unsafe",
				Steps: []gha.Step{
					{
						Name: "Checkout repository",
					},
					{
						Name: "Example",
						Run:  "echo ${{ inputs.foo }}",
					},
					{
						Name: "Example",
						Run:  "echo ${{ inputs.bar }}",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 2,
			wantId:    "Unsafe",
		},
		"matrix safe": {
			job: gha.Job{
				Name: "Safe matrix",
				Strategy: gha.Strategy{
					Matrix: []map[string]any{
						{
							"field": []any{
								"safe",
								"also safe",
							},
						},
					},
				},
				Steps: []gha.Step{
					{
						Run: "echo ${{ matrix.field }}",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 0,
		},
		"matrix unsafe": {
			job: gha.Job{
				Name: "Unsafe matrix",
				Strategy: gha.Strategy{
					Matrix: []map[string]any{
						{
							"field": []any{
								"${{ inputs.unsafe }}",
								"${{ inputs.also-unsafe }}",
							},
						},
					},
				},
				Steps: []gha.Step{
					{
						Run: "echo ${{ matrix.field }}",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 1,
			wantId:    "Unsafe matrix",
		},
		"matrix nested safe": {
			job: gha.Job{
				Name: "Safe nested matrix",
				Strategy: gha.Strategy{
					Matrix: []map[string]any{
						{
							"foo": map[string]any{
								"bar": "safe",
							},
						},
					},
				},
				Steps: []gha.Step{
					{
						Run: "echo ${{ matrix.foo.bar }}",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 0,
		},
		"matrix nested unsafe": {
			job: gha.Job{
				Name: "Unsafe nested matrix",
				Strategy: gha.Strategy{
					Matrix: []map[string]any{
						{
							"foo": map[string]any{
								"bar": "${{ inputs.unsafe }}",
							},
						},
					},
				},
				Steps: []gha.Step{
					{
						Run: "echo ${{ matrix.foo.bar }}",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 1,
			wantId:    "Unsafe nested matrix",
		},
		"matrix safe combined with something unsafe": {
			job: gha.Job{
				Name: "Safe matrix",
				Strategy: gha.Strategy{
					Matrix: []map[string]any{
						{
							"field": "safe",
						},
					},
				},
				Steps: []gha.Step{
					{
						Run: "echo ${{ matrix.field || inputs.unsafe }}",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 1,
			wantId:    "Safe matrix",
		},
		"matrix safe and something unsafe": {
			job: gha.Job{
				Name: "Safe matrix",
				Strategy: gha.Strategy{
					Matrix: []map[string]any{
						{
							"field": "safe",
						},
					},
				},
				Steps: []gha.Step{
					{
						Run: "echo ${{ matrix.field }} ${{ inputs.unsafe }}",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 1,
			wantId:    "Safe matrix",
		},
		"matrix missing": {
			job: gha.Job{
				Name: "Incomplete matrix",
				Steps: []gha.Step{
					{
						Run: "echo ${{ matrix.field }}",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 1,
			wantId:    "Incomplete matrix",
		},
		"matrix incomplete access": {
			job: gha.Job{
				Name: "Incomplete access",
				Strategy: gha.Strategy{
					Matrix: []map[string]any{
						{
							"foo": map[string]any{
								"bar": "${{ inputs.unsafe }}",
							},
						},
					},
				},
				Steps: []gha.Step{
					{
						Run: "echo ${{ matrix.foo }}",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 1,
			wantId:    "Incomplete access",
		},
		"matrix multiple in one expression": {
			job: gha.Job{
				Name: "Partially unsafe matrix",
				Strategy: gha.Strategy{
					Matrix: []map[string]any{
						{
							"foo": map[string]any{
								"foo": "safe",
								"bar": "${{ inputs.unsafe }}",
							},
						},
					},
				},
				Steps: []gha.Step{
					{
						Run: "echo ${{ matrix.foo || matrix.bar }}",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 1,
			wantId:    "Partially unsafe matrix",
		},
		"matrix value safe expression": {
			job: gha.Job{
				Name: "Safe matrix",
				Strategy: gha.Strategy{
					Matrix: []map[string]any{
						{
							"field": []any{
								"safe",
								"${{ 3.14 }}",
							},
						},
					},
				},
				Steps: []gha.Step{
					{
						Run: "echo ${{ matrix.field }}",
					},
				},
			},
			matcher:   AllMatcher,
			wantCount: 0,
		},
		"matrix conservatively safe": {
			job: gha.Job{
				Name: "Conservatively safe matrix",
				Strategy: gha.Strategy{
					Matrix: []map[string]any{
						{
							"field": []any{
								"${{ foo.bar }}",
							},
						},
					},
				},
				Steps: []gha.Step{
					{
						Run: "echo ${{ matrix.field }}",
					},
				},
			},
			matcher:   ConservativeMatcher,
			wantCount: 0,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			violations := analyzeJob(tt.id, &tt.job, tt.matcher)
			if got, want := len(violations), tt.wantCount; got != want {
				t.Fatalf("Unexpected number of violations (got %d, want %d) %v", got, want, violations)
			}

			for i, violation := range violations {
				if got, want := violation.jobKey, tt.id; got != want {
					t.Errorf("Unexpected job key for violation %d (got %q, want %q)", i, got, want)
				}

				if got, want := violation.JobId, tt.wantId; got != want {
					t.Errorf("Unexpected job ID for violation %d (got %q, want %q)", i, got, want)
				}
			}
		})
	}
}

func TestAnalyzeStep(t *testing.T) {
	type TestCase struct {
		id         int
		step       gha.Step
		matcher    ExprMatcher
		wantCount  int
		wantStepId string
	}

	testCases := map[string]TestCase{
		"Unnamed step that does nothing": {
			step: gha.Step{
				Name: "",
			},
			matcher:   AllMatcher,
			wantCount: 0,
		},
		"Named step that does nothing": {
			step: gha.Step{
				Name: "Doesn't run",
			},
			matcher:   AllMatcher,
			wantCount: 0,
		},
		"Unnamed step without violation": {
			step: gha.Step{
				Name: "",
				Run:  "echo 'Hello world!'",
			},
			matcher:   AllMatcher,
			wantCount: 0,
		},
		"Named step without violation": {
			step: gha.Step{
				Name: "Run something",
				Run:  "echo 'Hello world!'",
			},
			matcher:   AllMatcher,
			wantCount: 0,
		},
		"Unnamed step with one violation": {
			id: 42,
			step: gha.Step{
				Name: "",
				Run:  "echo 'Hello ${{ inputs.name }}!'",
			},
			matcher:    AllMatcher,
			wantCount:  1,
			wantStepId: "#42",
		},
		"Named step with one violation": {
			step: gha.Step{
				Name: "Greet person",
				Run:  "echo 'Hello ${{ inputs.name }}!'",
			},
			matcher:    AllMatcher,
			wantCount:  1,
			wantStepId: "Greet person",
		},
		"Unnamed step with two violation": {
			id: 3,
			step: gha.Step{
				Name: "",
				Run:  "echo 'Hello ${{ inputs.name }}! How is your ${{ steps.id.outputs.day }}'",
			},
			matcher:    AllMatcher,
			wantCount:  2,
			wantStepId: "#3",
		},
		"Named step with two violation": {
			id: 1,
			step: gha.Step{
				Name: "Greet person today",
				Run:  "echo 'Hello ${{ inputs.name }}! How is your ${{ steps.id.outputs.day }}'",
			},
			matcher:    AllMatcher,
			wantCount:  2,
			wantStepId: "Greet person today",
		},
		"Uses step with unknown action": {
			id: 1,
			step: gha.Step{
				Uses: gha.Uses{
					Name: "this/is-not",
					Ref:  "a-real-action",
				},
			},
			matcher:   AllMatcher,
			wantCount: 0,
		},
		"Uses step with known action, no violations": {
			id: 1,
			step: gha.Step{
				Uses: gha.Uses{
					Name: "actions/github-script",
					Ref:  "v6",
				},
			},
			matcher:   AllMatcher,
			wantCount: 0,
		},
		"Uses step with known action, one violation": {
			id: 1,
			step: gha.Step{
				Uses: gha.Uses{
					Name: "actions/github-script",
					Ref:  "v6",
				},
				With: map[string]string{
					"script": "console.log('Hello ${{ inputs.name }}')",
				},
			},
			matcher:    AllMatcher,
			wantCount:  1,
			wantStepId: "#1",
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			violations := analyzeStep(tt.id, &tt.step, tt.matcher)
			if got, want := len(violations), tt.wantCount; got != want {
				t.Fatalf("Unexpected number of violations (got %d, want %d)", got, want)
			}

			for i, violation := range violations {
				if got, want := violation.stepIndex, tt.id; got != want {
					t.Errorf("Unexpected step index for violation #%d (got %q, want %q)", i, got, want)
				}

				if got, want := violation.StepId, tt.wantStepId; got != want {
					t.Errorf("Unexpected step id for violation #%d (got %q, want %q)", i, got, want)
				}
			}
		})
	}
}

func TestAnalyzeString(t *testing.T) {
	type TestCase struct {
		value   string
		matcher ExprMatcher
		want    []Violation
	}

	testCases := map[string]TestCase{
		"Simple and safe": {
			value:   "echo 'Hello world!'",
			matcher: AllMatcher,
			want:    []Violation{},
		},
		"Multiline and safe": {
			value:   "echo 'Hello'\necho 'world!'",
			matcher: AllMatcher,
			want:    []Violation{},
		},
		"One violations": {
			value:   "echo 'Hello ${{ inputs.name }}!'",
			matcher: AllMatcher,
			want: []Violation{
				{Problem: "${{ inputs.name }}"},
			},
		},
		"Two violations": {
			value:   "echo '${{ inputs.greeting }} ${{ inputs.name }}!'",
			matcher: AllMatcher,
			want: []Violation{
				{Problem: "${{ inputs.greeting }}"},
				{Problem: "${{ inputs.name }}"},
			},
		},
		"Safe expressions": {
			value:   "echo '${{ true }} or ${{ false }}!'",
			matcher: AllMatcher,
			want:    nil,
		},
		"Partially safe expressions": {
			value:   "echo '${{ true || inputs.sha }}'",
			matcher: AllMatcher,
			want: []Violation{
				{Problem: "${{ true || inputs.sha }}"},
			},
		},
		"Safe and unsafe expressions": {
			value:   "echo '${{ true }} ${{ inputs.name }}'",
			matcher: AllMatcher,
			want: []Violation{
				{Problem: "${{ inputs.name }}"},
			},
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			violations := analyzeString(tt.value, tt.matcher)
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

type FailOpenFS struct {
	fs.FS
	Fail map[string]error
}

func (f FailOpenFS) Open(name string) (fs.File, error) {
	if err, ok := f.Fail[name]; ok {
		return nil, err
	}

	return f.FS.Open(name)
}

type FailReadFS struct {
	fs.FS
	Fail map[string]error
}

func (f FailReadFS) Open(name string) (fs.File, error) {
	file, err := f.FS.Open(name)
	if err != nil {
		return nil, err
	}

	if err, ok := f.Fail[name]; ok {
		return FailReadFile{file, err}, nil
	}

	return file, nil
}

type FailReadFile struct {
	fs.File
	err error
}

func (f FailReadFile) Read(p []byte) (int, error) {
	return 0, f.err
}
