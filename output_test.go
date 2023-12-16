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
	"testing"
)

func TestPrintJson(t *testing.T) {
	type TestCase struct {
		name       string
		violations func() map[string]map[string][]violation
		want       string
	}

	testCases := []TestCase{
		{
			name: "No targets",
			violations: func() map[string]map[string][]violation {
				return make(map[string]map[string][]violation)
			},
			want: `{"problems":[]}`,
		},
		{
			name: "target without files",
			violations: func() map[string]map[string][]violation {
				m := make(map[string]map[string][]violation)
				m["foobar"] = make(map[string][]violation)
				return m
			},
			want: `{"problems":[]}`,
		},
		{
			name: "target with files without violations",
			violations: func() map[string]map[string][]violation {
				m := make(map[string]map[string][]violation)
				m["foo"] = make(map[string][]violation)
				m["foo"]["bar"] = make([]violation, 0)
				return m
			},
			want: `{"problems":[]}`,
		},
		{
			name: "target with files with violations",
			violations: func() map[string]map[string][]violation {
				m := make(map[string]map[string][]violation)
				m["foo"] = make(map[string][]violation)
				m["foo"]["bar"] = make([]violation, 1)
				m["foo"]["bar"][0] = violation{
					jobId:   "4",
					stepId:  "2",
					problem: "${{ foo.bar }}",
				}
				return m
			},
			want: `{"problems":[{"target":"foo","file":"bar","job":"4","step":"2","problem":"${{ foo.bar }}"}]}`,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got, want := printJson(tt.violations()), tt.want; got != want {
				t.Fatalf("Unexpected JSON output (got %q, want %q)", got, want)
			}
		})
	}
}

func TestPrintViolations(t *testing.T) {
	type TestCase struct {
		name            string
		violations      func() map[string][]violation
		want            string
		wantSuggestions string
	}

	testCases := []TestCase{
		{
			name: "No files",
			violations: func() map[string][]violation {
				return make(map[string][]violation)
			},
			want:            ``,
			wantSuggestions: ``,
		},
		{
			name: "File without violations",
			violations: func() map[string][]violation {
				m := make(map[string][]violation)
				m["workflow.yml"] = make([]violation, 0)
				return m
			},
			want:            ``,
			wantSuggestions: ``,
		},
		{
			name: "Workflow with violation in run script",
			violations: func() map[string][]violation {
				m := make(map[string][]violation)
				m["workflow.yml"] = make([]violation, 1)
				m["workflow.yml"][0] = violation{
					jobId:   "4",
					stepId:  "2",
					problem: "${{ foo.bar }}",
					kind:    expressionInRunScript,
				}
				return m
			},
			want: `Detected 1 violation(s) in "workflow.yml":
  job "4", step "2" has "${{ foo.bar }}" (ADES100)
`,
			wantSuggestions: `Detected 1 violation(s) in "workflow.yml":
  job "4", step "2" has "${{ foo.bar }}", suggestion:
    1. Set ` + "`" + `BAR: ${{ foo.bar }}` + "`" + ` in the step's ` + "`" + `env` + "`" + ` map
    2. Replace all occurrences of ` + "`" + `${{ foo.bar }}` + "`" + ` by ` + "`" + `$BAR` + "`" + `
       (make sure to keep the behavior of the script the same)
`,
		},
		{
			name: "Workflow with violation in actions/github-script",
			violations: func() map[string][]violation {
				m := make(map[string][]violation)
				m["workflow.yml"] = make([]violation, 1)
				m["workflow.yml"][0] = violation{
					jobId:   "4",
					stepId:  "2",
					problem: "${{ foo.bar }}",
					kind:    expressionInActionsGithubScript,
				}
				return m
			},
			want: `Detected 1 violation(s) in "workflow.yml":
  job "4", step "2" has "${{ foo.bar }}" (ADES101)
`,
			wantSuggestions: `Detected 1 violation(s) in "workflow.yml":
  job "4", step "2" has "${{ foo.bar }}", suggestion:
    1. Set ` + "`" + `BAR: ${{ foo.bar }}` + "`" + ` in the step's ` + "`" + `env` + "`" + ` map
    2. Replace all occurrences of ` + "`" + `${{ foo.bar }}` + "`" + ` by ` + "`" + `process.env.BAR` + "`" + `
       (make sure to keep the behavior of the script the same)
`,
		},

		{
			name: "Manifest with violation in run script",
			violations: func() map[string][]violation {
				m := make(map[string][]violation)
				m["action.yml"] = make([]violation, 1)
				m["action.yml"][0] = violation{
					stepId:  "2",
					problem: "${{ foo.bar }}",
					kind:    expressionInRunScript,
				}
				return m
			},
			want: `Detected 1 violation(s) in "action.yml":
  step "2" has "${{ foo.bar }}" (ADES100)
`,
			wantSuggestions: `Detected 1 violation(s) in "action.yml":
  step "2" has "${{ foo.bar }}", suggestion:
    1. Set ` + "`" + `BAR: ${{ foo.bar }}` + "`" + ` in the step's ` + "`" + `env` + "`" + ` map
    2. Replace all occurrences of ` + "`" + `${{ foo.bar }}` + "`" + ` by ` + "`" + `$BAR` + "`" + `
       (make sure to keep the behavior of the script the same)
`,
		},
		{
			name: "Manifest with violation in actions/github-script",
			violations: func() map[string][]violation {
				m := make(map[string][]violation)
				m["action.yml"] = make([]violation, 1)
				m["action.yml"][0] = violation{
					stepId:  "2",
					problem: "${{ foo.bar }}",
					kind:    expressionInActionsGithubScript,
				}
				return m
			},
			want: `Detected 1 violation(s) in "action.yml":
  step "2" has "${{ foo.bar }}" (ADES101)
`,
			wantSuggestions: `Detected 1 violation(s) in "action.yml":
  step "2" has "${{ foo.bar }}", suggestion:
    1. Set ` + "`" + `BAR: ${{ foo.bar }}` + "`" + ` in the step's ` + "`" + `env` + "`" + ` map
    2. Replace all occurrences of ` + "`" + `${{ foo.bar }}` + "`" + ` by ` + "`" + `process.env.BAR` + "`" + `
       (make sure to keep the behavior of the script the same)
`,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got, want := printViolations(tt.violations(), false), tt.want; got != want {
				t.Errorf("Unexpected output (got %q, want %q)", got, want)
			}

			if got, want := printViolations(tt.violations(), true), tt.wantSuggestions; got != want {
				t.Errorf("Unexpected output with suggestions (got %q, want %q)", got, want)
			}
		})
	}
}
