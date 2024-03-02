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
	"regexp"
	"testing"
	"testing/quick"
)

func TestActionRuleActionsGithubScript(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		f := func(uses StepUses, ref string) bool {
			uses.Name = "actions/github-script"
			uses.Ref = ref
			return actionRuleActionsGitHubScript.appliesTo(&uses)
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		withScript := func(step JobStep, script string) bool {
			step.With["script"] = script
			return actionRuleActionsGitHubScript.rule.extractFrom(&step) == script
		}
		if err := quick.Check(withScript, nil); err != nil {
			t.Error(err)
		}

		withoutScript := func(step JobStep) bool {
			delete(step.With, "script")
			return actionRuleActionsGitHubScript.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(withoutScript, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Suggestion", func(t *testing.T) {
		violation := Violation{
			JobId:   "4",
			StepId:  "2",
			Problem: "${{ foo.bar }}",
			RuleId:  "ADES101",
		}

		got := actionRuleActionsGitHubScript.rule.suggestion(&violation)
		want := `    1. Set ` + "`" + `BAR: ${{ foo.bar }}` + "`" + ` in the step's ` + "`" + `env` + "`" + ` map
    2. Replace all occurrences of ` + "`" + `${{ foo.bar }}` + "`" + ` by ` + "`" + `process.env.BAR` + "`" + `
       (make sure to keep the behavior of the script the same)`

		if got != want {
			t.Errorf("Unexpected suggestion (got %q, want %q)", got, want)
		}
	})
}

func TestActionRuleEriccornelissenGitTagAnnotationAction(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		type TestCase struct {
			name string
			uses StepUses
			want bool
		}

		testCases := []TestCase{
			{
				name: "Last vulnerable version",
				uses: StepUses{
					Ref: "v1.0.0",
				},
				want: true,
			},
			{
				name: "Old version",
				uses: StepUses{
					Ref: "v0.0.9",
				},
				want: true,
			},
			{
				name: "New version",
				uses: StepUses{
					Ref: "v1.0.1",
				},
				want: false,
			},
		}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				if got, want := actionRuleEriccornelissenGitTagAnnotationAction.appliesTo(&tt.uses), tt.want; got != want {
					t.Fatalf("Unexpected result for %s, got %t", tt.uses.Ref, want)
				}
			})
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		withTag := func(step JobStep, tag string) bool {
			step.With["tag"] = tag
			return actionRuleEriccornelissenGitTagAnnotationAction.rule.extractFrom(&step) == tag
		}
		if err := quick.Check(withTag, nil); err != nil {
			t.Error(err)
		}

		withoutTag := func(step JobStep) bool {
			delete(step.With, "tag")
			return actionRuleEriccornelissenGitTagAnnotationAction.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(withoutTag, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Suggestion", func(t *testing.T) {
		violation := Violation{
			JobId:   "4",
			StepId:  "2",
			Problem: "${{ foo.bar }}",
			RuleId:  "ADES101",
		}

		got := actionRuleEriccornelissenGitTagAnnotationAction.rule.suggestion(&violation)
		want := "    1. Upgrade to a non-vulnerable version, see GHSA-hgx2-4pp9-357g"

		if got != want {
			t.Errorf("Unexpected suggestion (got %q, want %q)", got, want)
		}
	})
}

func TestActionRuleKcebGitMessageAction(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		type TestCase struct {
			name string
			uses StepUses
			want bool
		}

		testCases := []TestCase{
			{
				name: "Last vulnerable version",
				uses: StepUses{
					Ref: "v1.1.0",
				},
				want: true,
			},
			{
				name: "Old version",
				uses: StepUses{
					Ref: "v1.0.0",
				},
				want: true,
			},
			{
				name: "New version",
				uses: StepUses{
					Ref: "v1.2.0",
				},
				want: false,
			},
		}

		for _, tt := range testCases {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				if got, want := actionRuleKcebGitMessageAction.appliesTo(&tt.uses), tt.want; got != want {
					t.Fatalf("Unexpected result for %s, got %t", tt.uses.Ref, want)
				}
			})
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		withSha := func(step JobStep, tag string) bool {
			step.With["sha"] = tag
			return actionRuleKcebGitMessageAction.rule.extractFrom(&step) == tag
		}
		if err := quick.Check(withSha, nil); err != nil {
			t.Error(err)
		}

		withoutSha := func(step JobStep) bool {
			delete(step.With, "sha")
			return actionRuleKcebGitMessageAction.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(withoutSha, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Suggestion", func(t *testing.T) {
		violation := Violation{
			JobId:   "4",
			StepId:  "2",
			Problem: "${{ foo.bar }}",
			RuleId:  "ADES101",
		}

		got := actionRuleKcebGitMessageAction.rule.suggestion(&violation)
		want := "    1. Upgrade to a non-vulnerable version, see v1.2.0 release notes"

		if got != want {
			t.Errorf("Unexpected suggestion (got %q, want %q)", got, want)
		}
	})
}

func TestActionRulesRootsIssueCloser(t *testing.T) {
	t.Run("issue-close-message", func(t *testing.T) {
		t.Run("Applies to", func(t *testing.T) {
			f := func(uses StepUses, ref string) bool {
				uses.Name = "roots/issue-closer"
				uses.Ref = ref
				return actionRuleRootsIssueCloserIssueCloseMessage.appliesTo(&uses)
			}

			if err := quick.Check(f, nil); err != nil {
				t.Error(err)
			}
		})

		t.Run("Extract from", func(t *testing.T) {
			withIssueCloseMessage := func(step JobStep, message string) bool {
				step.With["issue-close-message"] = message
				return actionRuleRootsIssueCloserIssueCloseMessage.rule.extractFrom(&step) == message
			}
			if err := quick.Check(withIssueCloseMessage, nil); err != nil {
				t.Error(err)
			}

			withoutIssueCloseMessage := func(step JobStep) bool {
				delete(step.With, "issue-close-message")
				return actionRuleRootsIssueCloserIssueCloseMessage.rule.extractFrom(&step) == ""
			}
			if err := quick.Check(withoutIssueCloseMessage, nil); err != nil {
				t.Error(err)
			}
		})

		t.Run("Suggestion", func(t *testing.T) {
			violation := Violation{
				JobId:   "3",
				StepId:  "14",
				Problem: "${{ hello.world }}",
				RuleId:  "ADES101",
			}

			got := actionRuleRootsIssueCloserIssueCloseMessage.rule.suggestion(&violation)
			want := `    1. Set ` + "`" + `WORLD: ${{ hello.world }}` + "`" + ` in the step's ` + "`" + `env` + "`" + ` map
    2. Replace all occurrences of ` + "`" + `${{ hello.world }}` + "`" + ` by ` + "`" + `${process.env.WORLD}` + "`" + `
       (make sure to keep the behavior of the script the same)`

			if got != want {
				t.Errorf("Unexpected suggestion (got %q, want %q)", got, want)
			}
		})
	})

	t.Run("pr-close-message", func(t *testing.T) {
		t.Run("Applies to", func(t *testing.T) {
			f := func(uses StepUses, ref string) bool {
				uses.Name = "roots/issue-closer"
				uses.Ref = ref
				return actionRuleRootsIssueCloserPrCloseMessage.appliesTo(&uses)
			}

			if err := quick.Check(f, nil); err != nil {
				t.Error(err)
			}
		})

		t.Run("Extract from", func(t *testing.T) {
			withIssueCloseMessage := func(step JobStep, message string) bool {
				step.With["pr-close-message"] = message
				return actionRuleRootsIssueCloserPrCloseMessage.rule.extractFrom(&step) == message
			}
			if err := quick.Check(withIssueCloseMessage, nil); err != nil {
				t.Error(err)
			}

			withoutIssueCloseMessage := func(step JobStep) bool {
				delete(step.With, "pr-close-message")
				return actionRuleRootsIssueCloserPrCloseMessage.rule.extractFrom(&step) == ""
			}
			if err := quick.Check(withoutIssueCloseMessage, nil); err != nil {
				t.Error(err)
			}
		})

		t.Run("Suggestion", func(t *testing.T) {
			violation := Violation{
				JobId:   "3",
				StepId:  "14",
				Problem: "${{ hello.world }}",
				RuleId:  "ADES101",
			}

			got := actionRuleRootsIssueCloserPrCloseMessage.rule.suggestion(&violation)
			want := `    1. Set ` + "`" + `WORLD: ${{ hello.world }}` + "`" + ` in the step's ` + "`" + `env` + "`" + ` map
    2. Replace all occurrences of ` + "`" + `${{ hello.world }}` + "`" + ` by ` + "`" + `${process.env.WORLD}` + "`" + `
       (make sure to keep the behavior of the script the same)`

			if got != want {
				t.Errorf("Unexpected suggestion (got %q, want %q)", got, want)
			}
		})
	})
}

func TestActionRuleSergeysovaJqAction(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		f := func(uses StepUses) bool {
			uses.Name = "sergeysova/jq-action"
			return actionRuleSergeysovaJqAction.appliesTo(&uses)
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		withCmd := func(step JobStep, cmd string) bool {
			step.With["cmd"] = cmd
			return actionRuleSergeysovaJqAction.rule.extractFrom(&step) == cmd
		}
		if err := quick.Check(withCmd, nil); err != nil {
			t.Error(err)
		}

		withoutCmd := func(step JobStep) bool {
			delete(step.With, "cmd")
			return actionRuleSergeysovaJqAction.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(withoutCmd, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Suggestion", func(t *testing.T) {
		violation := Violation{
			JobId:   "3",
			StepId:  "14",
			Problem: "${{ github.event.inputs.file }}",
			RuleId:  "ADES101",
		}

		got := actionRuleSergeysovaJqAction.rule.suggestion(&violation)
		want := `    1. Set ` + "`" + `FILE: ${{ github.event.inputs.file }}` + "`" + ` in the step's ` + "`" + `env` + "`" + ` map
    2. Replace all occurrences of ` + "`" + `${{ github.event.inputs.file }}` + "`" + ` by ` + "`" + `$FILE` + "`" + `
       (make sure to keep the behavior of the script the same)`

		if got != want {
			t.Errorf("Unexpected suggestion (got %q, want %q)", got, want)
		}
	})
}

func TestStepRuleRun(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		runSteps := func(step JobStep, run string) bool {
			if len(run) == 0 {
				return true
			}

			step.Run = run
			return stepRuleRun.appliesTo(&step)
		}
		if err := quick.Check(runSteps, nil); err != nil {
			t.Error(err)
		}

		nonRunStep := func(step JobStep) bool {
			step.Run = ""
			return !stepRuleRun.appliesTo(&step)
		}
		if err := quick.Check(nonRunStep, nil); err != nil {
			t.Error(err)
		}

		if !stepRuleRun.appliesTo(&JobStep{Run: "a"}) {
			t.Error("Should apply to extremely short scripts, but didn't")
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		f := func(step JobStep, run string) bool {
			step.Run = run
			return stepRuleRun.rule.extractFrom(&step) == run
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Suggestion", func(t *testing.T) {
		violation := Violation{
			JobId:   "4",
			StepId:  "2",
			Problem: "${{ foo.bar }}",
			RuleId:  "ADES101",
		}

		got := stepRuleRun.rule.suggestion(&violation)
		want := `    1. Set ` + "`" + `BAR: ${{ foo.bar }}` + "`" + ` in the step's ` + "`" + `env` + "`" + ` map
    2. Replace all occurrences of ` + "`" + `${{ foo.bar }}` + "`" + ` by ` + "`" + `$BAR` + "`" + `
       (make sure to keep the behavior of the script the same)`

		if got != want {
			t.Errorf("Unexpected suggestion (got %q, want %q)", got, want)
		}
	})
}

func TestAllRules(t *testing.T) {
	testCases := make([]rule, 0)
	for _, rules := range actionRules {
		for _, rule := range rules {
			testCases = append(testCases, rule.rule)
		}
	}
	for _, rule := range stepRules {
		testCases = append(testCases, rule.rule)
	}

	t.Run("id", func(t *testing.T) {
		idExpr := regexp.MustCompile(`ADES\d{3}`)

		ids := make([]string, 0)
		for _, tt := range testCases {
			id := tt.id
			ids = append(ids, id)

			t.Run(tt.title, func(t *testing.T) {
				t.Parallel()

				if !idExpr.MatchString(id) {
					t.Errorf("The ID did not match the expected format (got %q)", id)
				}
			})
		}

		t.Run("unique", func(t *testing.T) {
			uniqueIds := make(map[string]any, len(ids))
			for _, id := range ids {
				if _, ok := uniqueIds[id]; ok {
					t.Errorf("Found repeated ID %q", id)
				}

				uniqueIds[id] = true
			}
		})
	})

	t.Run("description", func(t *testing.T) {
		for _, tt := range testCases {
			t.Run(tt.title, func(t *testing.T) {
				t.Parallel()

				if len(tt.description) == 0 {
					t.Error("The description must not be empty")
				}
			})
		}
	})

	t.Run("title", func(t *testing.T) {
		for _, tt := range testCases {
			t.Run(tt.id, func(t *testing.T) {
				t.Parallel()

				if len(tt.title) == 0 {
					t.Error("The title must not be empty")
				}
			})
		}
	})
}

func TestExplainRule(t *testing.T) {
	t.Run("Existing rules", func(t *testing.T) {
		for _, rs := range actionRules {
			for _, r := range rs {
				tt := r.rule.id
				t.Run(tt, func(t *testing.T) {
					t.Parallel()

					explanation, err := explainRule(tt)
					if err != nil {
						t.Fatalf("Unexpected error: %#v", err)
					}

					if explanation == "" {
						t.Error("Unexpected empty explanation")
					}
				})
			}
		}

		for _, r := range stepRules {
			tt := r.rule.id
			t.Run(tt, func(t *testing.T) {
				t.Parallel()

				explanation, err := explainRule(tt)
				if err != nil {
					t.Fatalf("Unexpected error: %#v", err)
				}

				if explanation == "" {
					t.Error("Unexpected empty explanation")
				}
			})
		}
	})

	t.Run("Missing rules", func(t *testing.T) {
		testCases := []string{
			"ADES000",
			"foobar",
		}

		for _, tt := range testCases {
			t.Run(tt, func(t *testing.T) {
				t.Parallel()

				_, err := explainRule(tt)
				if err == nil {
					t.Fatal("Expected an error, got none")
				}
			})
		}
	})
}

func TestFindRule(t *testing.T) {
	t.Run("Existing rules", func(t *testing.T) {
		for _, rs := range actionRules {
			for _, r := range rs {
				tt := r.rule.id
				t.Run(tt, func(t *testing.T) {
					t.Parallel()

					r, ok := findRule(tt)
					if !ok {
						t.Fatalf("Couldn't find rule %q", tt)
					}

					if r.id != tt {
						t.Errorf("Unexpected rule found: %#v", r)
					}
				})
			}
		}

		for _, r := range stepRules {
			tt := r.rule.id
			t.Run(tt, func(t *testing.T) {
				t.Parallel()

				r, ok := findRule(tt)
				if !ok {
					t.Fatalf("Couldn't find rule %q", tt)
				}

				if r.id != tt {
					t.Errorf("Unexpected rule found: %#v", r)
				}
			})
		}
	})

	t.Run("Missing rules", func(t *testing.T) {
		testCases := []string{
			"ADES000",
			"foobar",
		}

		for _, tt := range testCases {
			t.Run(tt, func(t *testing.T) {
				t.Parallel()

				_, ok := findRule(tt)
				if ok {
					t.Fatalf("Expectedly found a rule for %q", tt)
				}
			})
		}
	})
}

func TestIsBeforeVersion(t *testing.T) {
	type TestCase struct {
		name    string
		uses    StepUses
		version string
		want    bool
	}

	testCases := []TestCase{
		{
			name: "Same version, full semantic version",
			uses: StepUses{
				Ref: "v1.0.0",
			},
			version: "v1.0.0",
			want:    false,
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got, want := isBeforeVersion(&tt.uses, tt.version), tt.want; got != want {
				t.Errorf("Wrong answer for given %s compared to %s (got %t, want %t)", tt.uses.Ref, tt.version, got, want)
			}
		})
	}
}
