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
	"strings"
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

func TestPrintViolations(t *testing.T) {
	type TestCase struct {
		name            string
		violations      func() map[string][]ades.Violation
		want            string
		wantSuggestions string
	}

	testCases := []TestCase{
		{
			name: "No files",
			violations: func() map[string][]ades.Violation {
				return make(map[string][]ades.Violation)
			},
			want: `Ok
`,
			wantSuggestions: `Ok
`,
		},
		{
			name: "File without violations",
			violations: func() map[string][]ades.Violation {
				m := make(map[string][]ades.Violation)
				m["workflow.yml"] = make([]ades.Violation, 0)
				return m
			},
			want: `Ok
`,
			wantSuggestions: `Ok
`,
		},
		{
			name: "Workflow with a violation",
			violations: func() map[string][]ades.Violation {
				m := make(map[string][]ades.Violation)
				m["workflow.yml"] = make([]ades.Violation, 1)
				m["workflow.yml"][0] = ades.Violation{
					JobId:   "4",
					StepId:  "2",
					Problem: "${{ foo.bar }}",
					RuleId:  "ADES100",
				}
				return m
			},
			want: `Detected 1 violation(s) in "workflow.yml":
  job "4", step "2" has "${{ foo.bar }}" (ADES100)
`,
			wantSuggestions: `Detected 1 violation(s) in "workflow.yml":
  job "4", step "2" has "${{ foo.bar }}", suggestion:`,
		},
		{
			name: "Manifest with a violation",
			violations: func() map[string][]ades.Violation {
				m := make(map[string][]ades.Violation)
				m["action.yml"] = make([]ades.Violation, 1)
				m["action.yml"][0] = ades.Violation{
					StepId:  "2",
					Problem: "${{ foo.bar }}",
					RuleId:  "ADES100",
				}
				return m
			},
			want: `Detected 1 violation(s) in "action.yml":
  step "2" has "${{ foo.bar }}" (ADES100)
`,
			wantSuggestions: `Detected 1 violation(s) in "action.yml":
  step "2" has "${{ foo.bar }}", suggestion:`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got, want := printViolations(tt.violations(), false), tt.want; got != want {
				t.Errorf("Unexpected output (got %q, want %q)", got, want)
			}

			if got, prefix := printViolations(tt.violations(), true), tt.wantSuggestions; !strings.HasPrefix(got, prefix) {
				t.Errorf("Unexpected prefix for output with suggestions (got %q, want %q)", got, prefix)
			}
		})
	}
}
