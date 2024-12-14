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

	"github.com/ericcornelissen/ades"
)

func TestPrintJson(t *testing.T) {
	type TestCase struct {
		name       string
		violations func() map[string]map[string][]ades.Violation
		want       string
	}

	testCases := []TestCase{
		{
			name: "No targets",
			violations: func() map[string]map[string][]ades.Violation {
				return make(map[string]map[string][]ades.Violation)
			},
			want: `{"problems":[]}`,
		},
		{
			name: "target without files",
			violations: func() map[string]map[string][]ades.Violation {
				m := make(map[string]map[string][]ades.Violation)
				m["foobar"] = make(map[string][]ades.Violation)
				return m
			},
			want: `{"problems":[]}`,
		},
		{
			name: "target with files without violations",
			violations: func() map[string]map[string][]ades.Violation {
				m := make(map[string]map[string][]ades.Violation)
				m["foo"] = make(map[string][]ades.Violation)
				m["foo"]["bar"] = make([]ades.Violation, 0)
				return m
			},
			want: `{"problems":[]}`,
		},
		{
			name: "target with files with violations",
			violations: func() map[string]map[string][]ades.Violation {
				m := make(map[string]map[string][]ades.Violation)
				m["foo"] = make(map[string][]ades.Violation)
				m["foo"]["bar"] = make([]ades.Violation, 1)
				m["foo"]["bar"][0] = ades.Violation{
					JobId:   "4",
					StepId:  "2",
					Problem: "${{ foo.bar }}",
				}
				return m
			},
			want: `{"problems":[{"target":"foo","file":"bar","job":"4","step":"2","problem":"${{ foo.bar }}"}]}`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got, want := printJson(tt.violations()), tt.want; got != want {
				t.Fatalf("Unexpected JSON output (got %q, want %q)", got, want)
			}
		})
	}
}

func TestPrintProjectViolations(t *testing.T) {
	type TestCase struct {
		violations func() map[string][]ades.Violation
		want       string
	}

	testCases := map[string]TestCase{
		"No files": {
			violations: func() map[string][]ades.Violation {
				return make(map[string][]ades.Violation)
			},
			want: `Ok
`,
		},
		"File without violations": {
			violations: func() map[string][]ades.Violation {
				m := make(map[string][]ades.Violation, 1)
				m["workflow.yml"] = make([]ades.Violation, 0)
				return m
			},
			want: `Ok
`,
		},
		"Workflow with a violation": {
			violations: func() map[string][]ades.Violation {
				m := make(map[string][]ades.Violation, 1)
				m["workflow.yml"] = []ades.Violation{
					{
						JobId:   "4",
						StepId:  "2",
						Problem: "${{ foo.bar }}",
						RuleId:  "ADES100",
					},
				}

				return m
			},
			want: `Detected 1 violation(s) in "workflow.yml":
  1 in job "4":
    step "2" contains "${{ foo.bar }}" (ADES100)
`,
		},
		"Workflow with multiple violations in the same job": {
			violations: func() map[string][]ades.Violation {
				m := make(map[string][]ades.Violation, 1)
				m["workflow.yml"] = []ades.Violation{
					{
						JobId:   "3",
						StepId:  "6",
						Problem: "${{ foo.bar }}",
						RuleId:  "ADES100",
					},
					{
						JobId:   "3",
						StepId:  "14",
						Problem: "${{ foo.baz }}",
						RuleId:  "ADES101",
					},
				}

				return m
			},
			want: `Detected 2 violation(s) in "workflow.yml":
  2 in job "3":
    step "6" contains "${{ foo.bar }}" (ADES100)
    step "14" contains "${{ foo.baz }}" (ADES101)
`,
		},
		"Workflow with multiple violations in different jobs": {
			violations: func() map[string][]ades.Violation {
				m := make(map[string][]ades.Violation, 1)
				m["workflow.yml"] = []ades.Violation{
					{
						JobId:   "4",
						StepId:  "2",
						Problem: "${{ foo.bar }}",
						RuleId:  "ADES100",
					},
					{
						JobId:   "3",
						StepId:  "14",
						Problem: "${{ foo.baz }}",
						RuleId:  "ADES101",
					},
				}

				return m
			},
			want: `Detected 2 violation(s) in "workflow.yml":
  1 in job "3":
    step "14" contains "${{ foo.baz }}" (ADES101)
  1 in job "4":
    step "2" contains "${{ foo.bar }}" (ADES100)
`,
		},
		"Manifest with a violation": {
			violations: func() map[string][]ades.Violation {
				m := make(map[string][]ades.Violation, 1)
				m["action.yml"] = []ades.Violation{
					{
						StepId:  "7",
						Problem: "${{ foo.bar }}",
						RuleId:  "ADES100",
					},
				}

				return m
			},
			want: `Detected 1 violation(s) in "action.yml":
    step "7" contains "${{ foo.bar }}" (ADES100)
`,
		},
		"Manifest with multiple violations": {
			violations: func() map[string][]ades.Violation {
				m := make(map[string][]ades.Violation, 1)
				m["action.yml"] = []ades.Violation{
					{
						StepId:  "4",
						Problem: "${{ foo.bar }}",
						RuleId:  "ADES100",
					},
					{
						StepId:  "2",
						Problem: "${{ foo.baz }}",
						RuleId:  "ADES101",
					},
				}

				return m
			},
			want: `Detected 2 violation(s) in "action.yml":
    step "4" contains "${{ foo.bar }}" (ADES100)
    step "2" contains "${{ foo.baz }}" (ADES101)
`,
		},
		"Project with multiple workflows": {
			violations: func() map[string][]ades.Violation {
				m := make(map[string][]ades.Violation, 2)
				m["workflow-a.yml"] = []ades.Violation{
					{
						JobId:   "4",
						StepId:  "2",
						Problem: "${{ foo.bar }}",
						RuleId:  "ADES100",
					},
				}
				m["workflow-b.yml"] = []ades.Violation{
					{
						JobId:   "3",
						StepId:  "14",
						Problem: "${{ foo.baz }}",
						RuleId:  "ADES101",
					},
				}

				return m
			},
			want: `Detected 1 violation(s) in "workflow-a.yml":
  1 in job "4":
    step "2" contains "${{ foo.bar }}" (ADES100)
Detected 1 violation(s) in "workflow-b.yml":
  1 in job "3":
    step "14" contains "${{ foo.baz }}" (ADES101)
`,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got, want := printProjectViolations(tt.violations()), tt.want; got != want {
				t.Errorf("Unexpected output\n=== GOT ===\n%s\n=== WANT ===\n%s", got, want)
			}
		})
	}
}
