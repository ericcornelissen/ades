// Copyright (C) 2023-2025  Eric Cornelissen
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
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/quick"

	"github.com/ericcornelissen/go-gha-models"
)

func TestActionRule8398a7ActionSlack(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		f := func(uses gha.Uses) bool {
			uses.Name = "8398a7/action-slack"
			return actionRule8398a7ActionSlack.appliesTo(&uses)
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["custom_payload"] = value
			return actionRule8398a7ActionSlack.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "custom_payload")
			return actionRule8398a7ActionSlack.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRuleActionsGithubScript(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		f := func(uses gha.Uses) bool {
			uses.Name = "actions/github-script"
			return actionRuleActionsGitHubScript.appliesTo(&uses)
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["script"] = value
			return actionRuleActionsGitHubScript.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "script")
			return actionRuleActionsGitHubScript.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRuleAddnabDockerRunAction(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		f := func(uses gha.Uses) bool {
			uses.Name = "addnab/docker-run-action"
			return actionRuleAddnabDockerRunAction.appliesTo(&uses)
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["run"] = value
			return actionRuleAddnabDockerRunAction.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "run")
			return actionRuleAddnabDockerRunAction.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRuleAmadevusPwshScript(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		f := func(uses gha.Uses) bool {
			uses.Name = "Amadevus/pwsh-script"
			return actionRuleAmadevusPwshScript.appliesTo(&uses)
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["script"] = value
			return actionRuleAmadevusPwshScript.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "script")
			return actionRuleAmadevusPwshScript.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRuleAppleboySshAction(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		f := func(uses gha.Uses) bool {
			uses.Name = "appleboy/ssh-action"
			return actionRuleAppleboySshAction.appliesTo(&uses)
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["script"] = value
			return actionRuleAppleboySshAction.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "script")
			return actionRuleAppleboySshAction.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRuleAtlassianGajiraCreate(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		type TestCase struct {
			uses gha.Uses
			want bool
		}

		testCases := map[string]TestCase{
			"Last vulnerable version": {
				uses: gha.Uses{
					Ref: "v2.0.0",
				},
				want: true,
			},
			"Old version": {
				uses: gha.Uses{
					Ref: "v1.0.1",
				},
				want: true,
			},
			"First fixed version": {
				uses: gha.Uses{
					Ref: "v2.0.1",
				},
				want: false,
			},
			"New version": {
				uses: gha.Uses{
					Ref: "v3.0.0",
				},
				want: false,
			},
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				if got, want := actionRuleAtlassianGajiraCreate.appliesTo(&tt.uses), tt.want; got != want {
					t.Fatalf("Unexpected result for %s (got %t, want %t)", tt.uses.Ref, got, want)
				}
			})
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["summary"] = value
			return actionRuleAtlassianGajiraCreate.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "summary")
			return actionRuleAtlassianGajiraCreate.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRuleCardinalbyJsEvalAction(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		f := func(uses gha.Uses) bool {
			uses.Name = "cardinalby/js-eval-action"
			return actionRuleCardinalbyJsEvalAction.appliesTo(&uses)
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["expression"] = value
			return actionRuleCardinalbyJsEvalAction.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "expression")
			return actionRuleCardinalbyJsEvalAction.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRuleEriccornelissenGitTagAnnotationAction(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		type TestCase struct {
			uses gha.Uses
			want bool
		}

		testCases := map[string]TestCase{
			"Last vulnerable version": {
				uses: gha.Uses{
					Ref: "v1.0.0",
				},
				want: true,
			},
			"Old version": {
				uses: gha.Uses{
					Ref: "v0.0.9",
				},
				want: true,
			},
			"First fixed version": {
				uses: gha.Uses{
					Ref: "v1.0.1",
				},
				want: false,
			},
			"New version": {
				uses: gha.Uses{
					Ref: "v1.1.0",
				},
				want: false,
			},
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				if got, want := actionRuleEriccornelissenGitTagAnnotationAction.appliesTo(&tt.uses), tt.want; got != want {
					t.Fatalf("Unexpected result for %s (got %t, want %t)", tt.uses.Ref, got, want)
				}
			})
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["tag"] = value
			return actionRuleEriccornelissenGitTagAnnotationAction.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "tag")
			return actionRuleEriccornelissenGitTagAnnotationAction.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRuleFishShopSyntaxCheck(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		type TestCase struct {
			uses gha.Uses
			want bool
		}

		testCases := map[string]TestCase{
			"Last vulnerable version": {
				uses: gha.Uses{
					Ref: "v1.6.11",
				},
				want: true,
			},
			"First fixed version": {
				uses: gha.Uses{
					Ref: "v1.6.12",
				},
				want: false,
			},
			"New version": {
				uses: gha.Uses{
					Ref: "v2.0.0",
				},
				want: false,
			},
			"Old version": {
				uses: gha.Uses{
					Ref: "v1.0.0",
				},
				want: true,
			},
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				if got, want := actionRuleFishShopSyntaxCheck.appliesTo(&tt.uses), tt.want; got != want {
					t.Fatalf("Unexpected result for %s (got %t, want %t)", tt.uses.Ref, got, want)
				}
			})
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["pattern"] = value
			return actionRuleFishShopSyntaxCheck.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "pattern")
			return actionRuleFishShopSyntaxCheck.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRuleJannekemRunPythonScriptAction(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		f := func(uses gha.Uses) bool {
			uses.Name = "jannekem/run-python-script-action"
			return actionRuleJannekemRunPythonScriptAction.appliesTo(&uses)
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["script"] = value
			return actionRuleJannekemRunPythonScriptAction.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "script")
			return actionRuleJannekemRunPythonScriptAction.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRuleKcebGitMessageAction(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		type TestCase struct {
			uses gha.Uses
			want bool
		}

		testCases := map[string]TestCase{
			"Last vulnerable version": {
				uses: gha.Uses{
					Ref: "v1.1.0",
				},
				want: true,
			},
			"Old version": {
				uses: gha.Uses{
					Ref: "v1.0.0",
				},
				want: true,
			},
			"First fixed version": {
				uses: gha.Uses{
					Ref: "v1.2.0",
				},
				want: false,
			},
			"New version": {
				uses: gha.Uses{
					Ref: "v1.3.0",
				},
				want: false,
			},
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				if got, want := actionRuleKcebGitMessageAction.appliesTo(&tt.uses), tt.want; got != want {
					t.Fatalf("Unexpected result for %s (got %t, want %t)", tt.uses.Ref, got, want)
				}
			})
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["sha"] = value
			return actionRuleKcebGitMessageAction.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "sha")
			return actionRuleKcebGitMessageAction.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRuleLycheeverseLycheeAction(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		type TestCase struct {
			uses gha.Uses
			want bool
		}

		testCases := map[string]TestCase{
			"Last vulnerable version": {
				uses: gha.Uses{
					Ref: "v2.0.1",
				},
				want: true,
			},
			"Old version": {
				uses: gha.Uses{
					Ref: "v1.6.1",
				},
				want: true,
			},
			"First fixed version": {
				uses: gha.Uses{
					Ref: "v2.0.2",
				},
				want: false,
			},
			"New version": {
				uses: gha.Uses{
					Ref: "v3.0.0",
				},
				want: false,
			},
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				if got, want := actionRuleLycheeverseLycheeAction.appliesTo(&tt.uses), tt.want; got != want {
					t.Fatalf("Unexpected result for %s (got %t, want %t)", tt.uses.Ref, got, want)
				}
			})
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["lycheeVersion"] = value
			return actionRuleLycheeverseLycheeAction.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "lycheeVersion")
			return actionRuleLycheeverseLycheeAction.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRuleMikefarahYq(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		f := func(uses gha.Uses) bool {
			uses.Name = "mikefarah/yq"
			return actionRuleMikefarahYq.appliesTo(&uses)
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["cmd"] = value
			return actionRuleMikefarahYq.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "cmd")
			return actionRuleMikefarahYq.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRuleOziProjectPublish(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		type TestCase struct {
			uses gha.Uses
			want bool
		}

		testCases := map[string]TestCase{
			"Last vulnerable version": {
				uses: gha.Uses{
					Ref: "v1.13.5",
				},
				want: true,
			},
			"Earliest vulnerable version version": {
				uses: gha.Uses{
					Ref: "v1.13.2",
				},
				want: true,
			},
			"First fixed version": {
				uses: gha.Uses{
					Ref: "v1.13.6",
				},
				want: false,
			},
			"New version": {
				uses: gha.Uses{
					Ref: "v2.0.0",
				},
				want: false,
			},
			"Old version": {
				uses: gha.Uses{
					Ref: "v1.0.0",
				},
				want: false,
			},
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				if got, want := actionRuleOziProjectPublish.appliesTo(&tt.uses), tt.want; got != want {
					t.Fatalf("Unexpected result for %s (got %t, want %t)", tt.uses.Ref, got, want)
				}
			})
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["pull-request-body"] = value
			return actionRuleOziProjectPublish.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "pull-request-body")
			return actionRuleOziProjectPublish.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRuleRootsIssueCloserAction(t *testing.T) {
	t.Run("issue-close-message", func(t *testing.T) {
		t.Run("Applies to", func(t *testing.T) {
			f := func(uses gha.Uses) bool {
				uses.Name = "roots/issue-closer"
				return actionRuleRootsIssueCloserActionIssueCloseMessage.appliesTo(&uses)
			}

			if err := quick.Check(f, nil); err != nil {
				t.Error(err)
			}
		})

		t.Run("Extract from", func(t *testing.T) {
			with := func(step gha.Step, value string) bool {
				step.With["issue-close-message"] = value
				return actionRuleRootsIssueCloserActionIssueCloseMessage.rule.extractFrom(&step) == value
			}
			if err := quick.Check(with, nil); err != nil {
				t.Error(err)
			}

			without := func(step gha.Step) bool {
				delete(step.With, "issue-close-message")
				return actionRuleRootsIssueCloserActionIssueCloseMessage.rule.extractFrom(&step) == ""
			}
			if err := quick.Check(without, nil); err != nil {
				t.Error(err)
			}
		})
	})

	t.Run("pr-close-message", func(t *testing.T) {
		t.Run("Applies to", func(t *testing.T) {
			f := func(uses gha.Uses) bool {
				uses.Name = "roots/issue-closer"
				return actionRuleRootsIssueCloserActionPrCloseMessage.appliesTo(&uses)
			}

			if err := quick.Check(f, nil); err != nil {
				t.Error(err)
			}
		})

		t.Run("Extract from", func(t *testing.T) {
			with := func(step gha.Step, value string) bool {
				step.With["pr-close-message"] = value
				return actionRuleRootsIssueCloserActionPrCloseMessage.rule.extractFrom(&step) == value
			}
			if err := quick.Check(with, nil); err != nil {
				t.Error(err)
			}

			without := func(step gha.Step) bool {
				delete(step.With, "pr-close-message")
				return actionRuleRootsIssueCloserActionPrCloseMessage.rule.extractFrom(&step) == ""
			}
			if err := quick.Check(without, nil); err != nil {
				t.Error(err)
			}
		})
	})
}

func TestActionRuleSergeysovaJqAction(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		f := func(uses gha.Uses) bool {
			uses.Name = "sergeysova/jq-action"
			return actionRuleSergeysovaJqAction.appliesTo(&uses)
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["cmd"] = value
			return actionRuleSergeysovaJqAction.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "cmd")
			return actionRuleSergeysovaJqAction.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRuleSonarSourceSonarqubeScanAction(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		type TestCase struct {
			uses gha.Uses
			want bool
		}

		testCases := map[string]TestCase{
			"Old unaffected version": {
				uses: gha.Uses{
					Ref: "v3.1.0",
				},
				want: false,
			},
			"First vulnerable version": {
				uses: gha.Uses{
					Ref: "v4.0.0",
				},
				want: true,
			},
			"Middle vulnerable version": {
				uses: gha.Uses{
					Ref: "v4.2.0",
				},
				want: true,
			},
			"Last vulnerable version": {
				uses: gha.Uses{
					Ref: "v5.3.0",
				},
				want: true,
			},
			"First fixed version": {
				uses: gha.Uses{
					Ref: "v5.3.1",
				},
				want: false,
			},
			"New version": {
				uses: gha.Uses{
					Ref: "v6.0.0",
				},
				want: false,
			},
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				if got, want := actionRuleSonarSourceSonarqubeScanAction.appliesTo(&tt.uses), tt.want; got != want {
					t.Fatalf("Unexpected result for %s (got %t, want %t)", tt.uses.Ref, got, want)
				}
			})
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		with := func(step gha.Step, value string) bool {
			step.With["args"] = value
			return actionRuleSonarSourceSonarqubeScanAction.rule.extractFrom(&step) == value
		}
		if err := quick.Check(with, nil); err != nil {
			t.Error(err)
		}

		without := func(step gha.Step) bool {
			delete(step.With, "args")
			return actionRuleSonarSourceSonarqubeScanAction.rule.extractFrom(&step) == ""
		}
		if err := quick.Check(without, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestStepRuleRun(t *testing.T) {
	t.Run("Applies to", func(t *testing.T) {
		runSteps := func(step gha.Step, run string) bool {
			if len(run) == 0 {
				return true
			}

			step.Run = run
			return stepRuleRun.appliesTo(&step)
		}
		if err := quick.Check(runSteps, nil); err != nil {
			t.Error(err)
		}

		nonRunStep := func(step gha.Step) bool {
			step.Run = ""
			return !stepRuleRun.appliesTo(&step)
		}
		if err := quick.Check(nonRunStep, nil); err != nil {
			t.Error(err)
		}

		if !stepRuleRun.appliesTo(&gha.Step{Run: "a"}) {
			t.Error("Should apply to extremely short scripts, but didn't")
		}
	})

	t.Run("Extract from", func(t *testing.T) {
		f := func(step gha.Step, run string) bool {
			step.Run = run
			return stepRuleRun.rule.extractFrom(&step) == run
		}

		if err := quick.Check(f, nil); err != nil {
			t.Error(err)
		}
	})
}

func TestActionRules(t *testing.T) {
	t.Run("key", func(t *testing.T) {
		for tt := range actionRules {
			t.Run(tt, func(t *testing.T) {
				t.Parallel()

				if got, want := tt, strings.ToLower(tt); got != want {
					t.Errorf("action name is not lowercase (got %q)", got)
				}
			})
		}
	})
}

func TestAllRules(t *testing.T) {
	testCases := allRules()

	t.Run("id", func(t *testing.T) {
		idExpr := regexp.MustCompile(`ADES\d{3}`)

		for _, tt := range testCases {
			t.Run(tt.title, func(t *testing.T) {
				t.Parallel()

				if !idExpr.MatchString(tt.id) {
					t.Errorf("The ID did not match the expected format (got %q)", tt.id)
				}
			})
		}

		t.Run("unique", func(t *testing.T) {
			ids := make(map[string]rule, len(testCases))
			for _, tt := range testCases {
				if got, ok := ids[tt.id]; ok && tt.title != got.title {
					t.Errorf("Found repeated ID %q", tt.id)
				}

				ids[tt.id] = tt
			}
		})
	})

	t.Run("description", func(t *testing.T) {
		for _, tt := range testCases {
			t.Run(tt.id, func(t *testing.T) {
				t.Parallel()

				if len(tt.description) == 0 {
					t.Error("The description must not be empty")
				}

				if !strings.HasPrefix(tt.description, "\n") {
					t.Error("The description must start with a newline")
				}

				if !strings.HasSuffix(tt.description, "\n") {
					t.Error("The description must end with a newline")
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

func TestExplain(t *testing.T) {
	t.Run("Existing rules", func(t *testing.T) {
		testCases := allRules()

		for _, tt := range testCases {
			t.Run(tt.id, func(t *testing.T) {
				t.Parallel()

				explanation, err := Explain(tt.id)
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
		type TestCase = string

		testCases := map[string]TestCase{
			"valid but unknown id": "ADES000",
			"invalid id":           "foobar",
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := Explain(tt)
				if err == nil {
					t.Fatal("Expected an error, got none")
				}
			})
		}
	})
}

func TestFix(t *testing.T) {
	t.Run("Existing rules", func(t *testing.T) {
		testCases := allRules()

		for _, tt := range testCases {
			t.Run(tt.id, func(t *testing.T) {
				t.Parallel()

				violation := Violation{
					RuleId: tt.id,
				}

				_, err := Fix(&violation)
				if err != nil {
					t.Fatalf("Unexpected error: %#v", err)
				}
			})
		}
	})

	t.Run("Missing rules", func(t *testing.T) {
		type TestCase = string

		testCases := map[string]TestCase{
			"valid but unknown id": "ADES000",
			"invalid id":           "foobar",
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				violation := Violation{
					RuleId: tt,
				}

				_, err := Fix(&violation)
				if err == nil {
					t.Fatal("Expected an error, got none")
				}
			})
		}
	})
}

func TestFindRule(t *testing.T) {
	t.Run("Action rules", func(t *testing.T) {
		for _, testCases := range actionRules {
			for _, tt := range testCases {
				t.Run(tt.rule.id, func(t *testing.T) {
					t.Parallel()

					r, err := findRule(tt.rule.id)
					if err != nil {
						t.Fatalf("Couldn't find rule %q", tt.rule.id)
					}

					if r.id != tt.rule.id {
						t.Errorf("Unexpected rule found: %#v", r)
					}
				})
			}
		}
	})

	t.Run("Step rules", func(t *testing.T) {
		testCases := stepRules

		for _, tt := range testCases {
			t.Run(tt.rule.id, func(t *testing.T) {
				t.Parallel()

				r, err := findRule(tt.rule.id)
				if err != nil {
					t.Fatalf("Couldn't find rule %q", tt.rule.id)
				}

				if r.id != tt.rule.id {
					t.Errorf("Unexpected rule found: %#v", r)
				}
			})
		}
	})

	t.Run("Missing rules", func(t *testing.T) {
		type TestCase = string

		testCases := map[string]TestCase{
			"valid but unknown id": "ADES000",
			"invalid id":           "foobar",
		}

		for name, tt := range testCases {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := findRule(tt)
				if err == nil {
					t.Fatalf("Expectedly found a rule for %q", tt)
				}
			})
		}
	})
}

func TestIsAtOrAfterVersion(t *testing.T) {
	type TestCase struct {
		uses    gha.Uses
		version string
		want    bool
	}

	testCases := map[string]TestCase{
		"Full version, exact same version": {
			uses: gha.Uses{
				Ref: "v1.2.3",
			},
			version: "v1.2.3",
			want:    true,
		},
		"Full version, earlier major version": {
			uses: gha.Uses{
				Ref: "v0.1.0",
			},
			version: "v1.2.3",
			want:    false,
		},
		"Full version, earlier minor version": {
			uses: gha.Uses{
				Ref: "v1.1.0",
			},
			version: "v1.2.3",
			want:    false,
		},
		"Full version, earlier patch version": {
			uses: gha.Uses{
				Ref: "v1.2.1",
			},
			version: "v1.2.3",
			want:    false,
		},
		"Full version, later major version": {
			uses: gha.Uses{
				Ref: "v2.0.0",
			},
			version: "v1.2.3",
			want:    true,
		},
		"Full version, later minor version": {
			uses: gha.Uses{
				Ref: "v1.3.0",
			},
			version: "v1.2.3",
			want:    true,
		},
		"Full version, later patch version": {
			uses: gha.Uses{
				Ref: "v1.2.4",
			},
			version: "v1.2.3",
			want:    true,
		},
		"Major version only, earlier major version": {
			uses: gha.Uses{
				Ref: "v1",
			},
			version: "v2.1.0",
			want:    false,
		},
		"Major version only, same major version": {
			uses: gha.Uses{
				Ref: "v2",
			},
			version: "v2.1.0",
			want:    true,
		},
		"Major version only, later major version": {
			uses: gha.Uses{
				Ref: "v3",
			},
			version: "v2.1.0",
			want:    true,
		},
		"Major+minor version, earlier major version and earlier minor version": {
			uses: gha.Uses{
				Ref: "v1.1",
			},
			version: "v2.2.1",
			want:    false,
		},
		"Major+minor version, earlier major version and same minor version": {
			uses: gha.Uses{
				Ref: "v1.2",
			},
			version: "v2.2.1",
			want:    false,
		},
		"Major+minor version, earlier major version and later minor version": {
			uses: gha.Uses{
				Ref: "v1.3",
			},
			version: "v2.2.1",
			want:    false,
		},
		"Major+minor version, same major version and earlier minor version": {
			uses: gha.Uses{
				Ref: "v2.1",
			},
			version: "v2.2.1",
			want:    false,
		},
		"Major+minor version, same major version and same minor version": {
			uses: gha.Uses{
				Ref: "v2.2",
			},
			version: "v2.2.1",
			want:    true,
		},
		"Major+minor version, same major version and later minor version": {
			uses: gha.Uses{
				Ref: "v2.3",
			},
			version: "v2.2.1",
			want:    true,
		},
		"Major+minor version, later major version and earlier minor version": {
			uses: gha.Uses{
				Ref: "v3.1",
			},
			version: "v2.2.1",
			want:    true,
		},
		"Major+minor version, later major version and same minor version": {
			uses: gha.Uses{
				Ref: "v3.2",
			},
			version: "v2.2.1",
			want:    true,
		},
		"Major+minor version, later major version and later minor version": {
			uses: gha.Uses{
				Ref: "v3.3",
			},
			version: "v2.2.1",
			want:    true,
		},
		"SHA without annotation": {
			uses: gha.Uses{
				Ref: "21fa0360d55070a1d6b999d027db44cc21a7b48d",
			},
			version: "v1.0.0",
			want:    false,
		},
		"SHA with annotation that is not a version": {
			uses: gha.Uses{
				Ref:        "21fa0360d55070a1d6b999d027db44cc21a7b48d",
				Annotation: "I'm just a comment",
			},
			version: "v1.0.0",
			want:    false,
		},
		"SHA with annotation, later version": {
			uses: gha.Uses{
				Ref:        "21fa0360d55070a1d6b999d027db44cc21a7b48d",
				Annotation: "v1.1.0",
			},
			version: "v1.0.0",
			want:    true,
		},
		"SHA with annotation, same version": {
			uses: gha.Uses{
				Ref:        "21fa0360d55070a1d6b999d027db44cc21a7b48d",
				Annotation: "v1.0.0",
			},
			version: "v1.0.0",
			want:    true,
		},
		"SHA with annotation, earlier version": {
			uses: gha.Uses{
				Ref:        "21fa0360d55070a1d6b999d027db44cc21a7b48d",
				Annotation: "v0.1.0",
			},
			version: "v1.0.0",
			want:    false,
		},
		"semver ref and annotation, ref later": {
			uses: gha.Uses{
				Ref:        "v1.1.0",
				Annotation: "v0.1.0",
			},
			version: "v1.0.0",
			want:    true,
		},
		"semver ref and annotation, ref earlier": {
			uses: gha.Uses{
				Ref:        "v0.1.0",
				Annotation: "v1.1.0",
			},
			version: "v1.0.0",
			want:    false,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got, want := isAtOrAfterVersion(&tt.uses, tt.version), tt.want; got != want {
				ref := tt.uses.Ref
				if tt.uses.Annotation != "" {
					ref = fmt.Sprintf("%s (%s)", tt.uses.Ref, tt.uses.Annotation)
				}

				t.Errorf("Wrong answer for given %s compared to %s (got %t, want %t)", ref, tt.version, got, want)
			}
		})
	}
}

func TestIsBeforeVersion(t *testing.T) {
	type TestCase struct {
		uses    gha.Uses
		version string
		want    bool
	}

	testCases := map[string]TestCase{
		"Full version, exact same version": {
			uses: gha.Uses{
				Ref: "v1.2.3",
			},
			version: "v1.2.3",
			want:    false,
		},
		"Full version, earlier major version": {
			uses: gha.Uses{
				Ref: "v0.1.0",
			},
			version: "v1.2.3",
			want:    true,
		},
		"Full version, earlier minor version": {
			uses: gha.Uses{
				Ref: "v1.1.0",
			},
			version: "v1.2.3",
			want:    true,
		},
		"Full version, earlier patch version": {
			uses: gha.Uses{
				Ref: "v1.2.1",
			},
			version: "v1.2.3",
			want:    true,
		},
		"Full version, later major version": {
			uses: gha.Uses{
				Ref: "v2.0.0",
			},
			version: "v1.2.3",
			want:    false,
		},
		"Full version, later minor version": {
			uses: gha.Uses{
				Ref: "v1.3.0",
			},
			version: "v1.2.3",
			want:    false,
		},
		"Full version, later patch version": {
			uses: gha.Uses{
				Ref: "v1.2.4",
			},
			version: "v1.2.3",
			want:    false,
		},
		"Major version only, earlier major version": {
			uses: gha.Uses{
				Ref: "v1",
			},
			version: "v2.1.0",
			want:    true,
		},
		"Major version only, same major version": {
			uses: gha.Uses{
				Ref: "v2",
			},
			version: "v2.1.0",
			want:    false,
		},
		"Major version only, later major version": {
			uses: gha.Uses{
				Ref: "v3",
			},
			version: "v2.1.0",
			want:    false,
		},
		"Major+minor version, earlier major version and earlier minor version": {
			uses: gha.Uses{
				Ref: "v1.1",
			},
			version: "v2.2.1",
			want:    true,
		},
		"Major+minor version, earlier major version and same minor version": {
			uses: gha.Uses{
				Ref: "v1.2",
			},
			version: "v2.2.1",
			want:    true,
		},
		"Major+minor version, earlier major version and later minor version": {
			uses: gha.Uses{
				Ref: "v1.3",
			},
			version: "v2.2.1",
			want:    true,
		},
		"Major+minor version, same major version and earlier minor version": {
			uses: gha.Uses{
				Ref: "v2.1",
			},
			version: "v2.2.1",
			want:    true,
		},
		"Major+minor version, same major version and same minor version": {
			uses: gha.Uses{
				Ref: "v2.2",
			},
			version: "v2.2.1",
			want:    false,
		},
		"Major+minor version, same major version and later minor version": {
			uses: gha.Uses{
				Ref: "v2.3",
			},
			version: "v2.2.1",
			want:    false,
		},
		"Major+minor version, later major version and earlier minor version": {
			uses: gha.Uses{
				Ref: "v3.1",
			},
			version: "v2.2.1",
			want:    false,
		},
		"Major+minor version, later major version and same minor version": {
			uses: gha.Uses{
				Ref: "v3.2",
			},
			version: "v2.2.1",
			want:    false,
		},
		"Major+minor version, later major version and later minor version": {
			uses: gha.Uses{
				Ref: "v3.3",
			},
			version: "v2.2.1",
			want:    false,
		},
		"SHA without annotation": {
			uses: gha.Uses{
				Ref: "21fa0360d55070a1d6b999d027db44cc21a7b48d",
			},
			version: "v1.0.0",
			want:    false,
		},
		"SHA with annotation that is not a version": {
			uses: gha.Uses{
				Ref:        "21fa0360d55070a1d6b999d027db44cc21a7b48d",
				Annotation: "I'm just a comment",
			},
			version: "v1.0.0",
			want:    false,
		},
		"SHA with annotation, later version": {
			uses: gha.Uses{
				Ref:        "21fa0360d55070a1d6b999d027db44cc21a7b48d",
				Annotation: "v1.1.0",
			},
			version: "v1.0.0",
			want:    false,
		},
		"SHA with annotation, same version": {
			uses: gha.Uses{
				Ref:        "21fa0360d55070a1d6b999d027db44cc21a7b48d",
				Annotation: "v1.0.0",
			},
			version: "v1.0.0",
			want:    false,
		},
		"SHA with annotation, earlier version": {
			uses: gha.Uses{
				Ref:        "21fa0360d55070a1d6b999d027db44cc21a7b48d",
				Annotation: "v0.1.0",
			},
			version: "v1.0.0",
			want:    true,
		},
		"semver ref and annotation, ref later": {
			uses: gha.Uses{
				Ref:        "v1.1.0",
				Annotation: "v0.1.0",
			},
			version: "v1.0.0",
			want:    false,
		},
		"semver ref and annotation, ref earlier": {
			uses: gha.Uses{
				Ref:        "v0.1.0",
				Annotation: "v1.1.0",
			},
			version: "v1.0.0",
			want:    true,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got, want := isBeforeVersion(&tt.uses, tt.version), tt.want; got != want {
				ref := tt.uses.Ref
				if tt.uses.Annotation != "" {
					ref = fmt.Sprintf("%s (%s)", tt.uses.Ref, tt.uses.Annotation)
				}

				t.Errorf("Wrong answer for given %s compared to %s (got %t, want %t)", ref, tt.version, got, want)
			}
		})
	}
}

func TestFixAddEnvVar(t *testing.T) {
	type (
		TestWant struct {
			old []string
			new string
		}
		TestCase struct {
			step  gha.Step
			name  string
			value string
			want  []TestWant
		}
	)

	testCases := map[string]TestCase{
		"no environment variables yet": {
			step: gha.Step{
				Uses: gha.Uses{
					Name: "foo/bar",
					Ref:  "v1",
				},
				Env: nil,
			},
			name:  "no",
			value: "env yet",
			want: []TestWant{
				{
					old: []string{`\n(\s+)uses:\s*foo/bar@v1.*?\n`},
					new: "${0}${1}env:\n${1}  no: env yet\n",
				},
				{
					old: []string{`\n(\s+)-(\s+)uses:\s*foo/bar@v1.*?\n`},
					new: "${0}${1} ${2}env:\n${1} ${2}  no: env yet\n",
				},
			},
		},
		"one environment variable already": {
			step: gha.Step{
				Env: map[string]string{
					"foo": "bar",
				},
			},
			name:  "one",
			value: "already",
			want: []TestWant{
				{
					old: []string{`env:\s*\n(?:(\s*)foo\s*:\s*bar\s*\n|)+`},
					new: "${0}${1}one: already\n",
				},
			},
		},
		"two environment variables already": {
			step: gha.Step{
				Env: map[string]string{
					"foo":   "bar",
					"hello": "world!",
				},
			},
			name:  "two",
			value: "already",
			want: []TestWant{
				{
					old: []string{
						`env:\s*\n(?:(\s*)foo\s*:\s*bar\s*\n|(\s*)hello\s*:\s*world!\s*\n|)+`,
						`env:\s*\n(?:(\s*)hello\s*:\s*world!\s*\n|(\s*)foo\s*:\s*bar\s*\n|)+`,
					},
					new: "${0}${1}two: already\n",
				},
			},
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fs := fixAddEnvVar(tt.step, tt.name, tt.value)
			for i, f := range fs {
				want := tt.want[i]

				if got, want := f.Old.String(), want.old; !slices.Contains(want, got) {
					t.Errorf("Incorrect %dth old string (got %q, want one of %v)", i, got, want)
				}

				if got, want := f.New, want.new; got != want {
					t.Errorf("Incorrect %dth new string (got %q, want %q)", i, got, want)
				}
			}
		})
	}
}

func TestFixReplaceIn(t *testing.T) {
	type (
		TestWant struct {
			old string
			new string
		}
		TestCase struct {
			s    string
			old  string
			new  string
			want TestWant
		}
	)

	testCases := map[string]TestCase{
		"basic example": {
			s:   "hello foobar world!",
			old: "foobar",
			new: "foobaz",
			want: TestWant{
				old: "hello foobar world!",
				new: "hello foobaz world!",
			},
		},
		"example with regular expression meta characters": {
			s:   "Hello world! (Hola mundo!)",
			old: "!",
			new: "",
			want: TestWant{
				old: `Hello world! \(Hola mundo!\)`,
				new: "Hello world (Hola mundo)",
			},
		},
		"example where replacing is a no-op": {
			s:   "This does not contain the string to replace",
			old: "foobar",
			new: "foobaz",
			want: TestWant{
				old: "This does not contain the string to replace",
				new: "This does not contain the string to replace",
			},
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			f := fixReplaceIn(tt.s, tt.old, tt.new)
			if got, want := f.Old.String(), tt.want.old; got != want {
				t.Errorf("Incorrect old string (got %q, want %q)", got, want)
			}

			if got, want := f.New, tt.want.new; got != want {
				t.Errorf("Incorrect new string (got %q, want %q)", got, want)
			}
		})
	}
}

func allRules() []rule {
	rules := make([]rule, 0)

	for _, rs := range actionRules {
		for _, r := range rs {
			rules = append(rules, r.rule)
		}
	}

	for _, r := range stepRules {
		rules = append(rules, r.rule)
	}

	return rules
}
