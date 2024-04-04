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

package ades

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

type rule struct {
	extractFrom func(step *JobStep) string
	suggestion  func(violation *Violation) string
	id          string
	title       string
	description string
}

type actionRule struct {
	appliesTo func(uses *StepUses) bool
	rule      rule
}

type stepRule struct {
	appliesTo func(step *JobStep) bool
	rule      rule
}

var actionRuleActionsGitHubScript = actionRule{
	appliesTo: func(_ *StepUses) bool {
		return true
	},
	rule: rule{
		id:    "ADES101",
		title: "Expression in 'actions/github-script' script",
		description: `
When an expression appears in a 'actions/github-script' script you can avoid any potential attacks
by extracting the expression into an environment variable and using the environment variable
instead.

For example, given the workflow snippet:

    - name: Example step
      uses: actions/github-script@v6
      with:
        script: console.log('Hello ${{ inputs.name }}')

it can be made safer by converting it into:

    - name: Example step
      uses: actions/github-script@v6
      env:
        NAME: ${{ inputs.name }} # <- Assign the expression to an environment variable
      with:
        script: console.log(` + "`" + `Hello ${process.env.NAME}` + "`" + `)
      #                     ^      ^^^^^^^^^^^^^^^^^^^
      #                     |      | Replace the expression with the environment variable
      #                     |
      #                     | Note: the use of backticks is required in this example (for interpolation)`,
		extractFrom: func(step *JobStep) string {
			return step.With["script"]
		},
		suggestion: suggestJavaScriptEnv,
	},
}

var actionRuleEriccornelissenGitTagAnnotationAction = actionRule{
	appliesTo: func(uses *StepUses) bool {
		return isBeforeVersion(uses, "v1.0.1")
	},
	rule: rule{
		id:    "ADES200",
		title: "Expression in 'ericcornelissen/git-tag-annotation-action' tag input",
		description: `
When an expression is used in the tag input for 'ericcornelissen/git-tag-annotation-action' in
v1.0.0 or earlier it may be used to execute arbitrary shell commands, see GHSA-hgx2-4pp9-357g. To
mitigate this, upgrade the action to a non-vulnerable version.`,
		extractFrom: func(step *JobStep) string {
			return step.With["tag"]
		},
		suggestion: func(_ *Violation) string {
			return "    1. Upgrade to a non-vulnerable version, see GHSA-hgx2-4pp9-357g"
		},
	},
}

var actionRuleKcebGitMessageAction = actionRule{
	appliesTo: func(uses *StepUses) bool {
		return isBeforeVersion(uses, "v1.2.0")
	},
	rule: rule{
		id:    "ADES201",
		title: "Expression in 'kceb/git-message-action' sha input",
		description: `
When an expression is used in the sha input for 'kceb/git-message-action' in v1.1.0 or earlier it
may be used to execute arbitrary shell commands (no vulnerability identifier available). To mitigate
this, upgrade the action to a non-vulnerable version.`,
		extractFrom: func(step *JobStep) string {
			return step.With["sha"]
		},
		suggestion: func(_ *Violation) string {
			return "    1. Upgrade to a non-vulnerable version, see v1.2.0 release notes"
		},
	},
}

var actionRuleRootsIssueCloserIssueCloseMessage = actionRule{
	appliesTo: func(_ *StepUses) bool {
		return true
	},
	rule: rule{
		id:    "ADES102",
		title: "Expression in 'roots/issue-closer' issue close message",
		description: `
When an expression appears in the issue close message of 'roots/issue-closer' it is interpreted as
an ES6-style template literal. You can avoid any potential attacks by extracting the expression into
an environment variable and using the environment variable instead.

For example, given the workflow snippet:

    - name: Example step
      uses: roots/issue-closer@v1
      with:
        issue-close-message: Closing ${{ github.event.issue.title }}

it can be made safer by converting it into:

    - name: Example step
      uses: roots/issue-closer@v1
      env:
        NAME: ${{ github.event.issue.title }} # <- Assign the expression to an environment variable
      with:
        issue-close-message: Closing ${process.env.NAME}
      #                              ^^^^^^^^^^^^^^^^^^^
      #                              | Replace the expression with the environment variable`,
		extractFrom: func(step *JobStep) string {
			return step.With["issue-close-message"]
		},
		suggestion: suggestJavaScriptLiteralEnv,
	},
}

var actionRuleRootsIssueCloserPrCloseMessage = actionRule{
	appliesTo: func(_ *StepUses) bool {
		return true
	},
	rule: rule{
		id:    "ADES103",
		title: "Expression in 'roots/issue-closer' pull request close message",
		description: `
When an expression appears in the pull request close message of 'roots/issue-closer' it is
interpreted as an ES6-style template literal. You can avoid any potential attacks by extracting the
expression into an environment variable and using the environment variable instead.

For example, given the workflow snippet:

    - name: Example step
      uses: roots/issue-closer@v1
      with:
        pr-close-message: Closing ${{ github.event.issue.title }}

it can be made safer by converting it into:

    - name: Example step
      uses: roots/issue-closer@v1
      env:
        NAME: ${{ github.event.issue.title }} # <- Assign the expression to an environment variable
      with:
        pr-close-message: Closing ${process.env.NAME}
      #                           ^^^^^^^^^^^^^^^^^^^
      #                           | Replace the expression with the environment variable`,
		extractFrom: func(step *JobStep) string {
			return step.With["pr-close-message"]
		},
		suggestion: suggestJavaScriptLiteralEnv,
	},
}

var actionRuleSergeysovaJqAction = actionRule{
	appliesTo: func(_ *StepUses) bool {
		return true
	},
	rule: rule{
		id:    "ADES104",
		title: "Expression in 'sergeysova/jq-action' command",
		description: `
When an expression appears in the command input of 'sergeysova/jq-action' you can avoid any
potential attack by extracting the expression into an environment variable and using the environment
variable instead.

For example, given the workflow snippet:

    - name: Example step
      uses: sergeysova/jq-action@v2
      with:
        cmd: jq .version ${{ github.event.inputs.file }} -r

it can be made safer by converting it into:

    - name: Example step
      uses: sergeysova/jq-action@v2
      env:
        FILE: ${{ github.event.inputs.file }} # <- Assign the expression to an environment variable
      with:
      #                  | Note: use double quotes to avoid argument splitting
      #                  v
        cmd: jq .version "$FILE" -r
      #                   ^^^^^
      #                   | Replace the expression with the environment variable`,
		extractFrom: func(step *JobStep) string {
			return step.With["cmd"]
		},
		suggestion: suggestShellEnv,
	},
}

var actionRules = map[string][]actionRule{
	"actions/github-script": {
		actionRuleActionsGitHubScript,
	},
	"ericcornelissen/git-tag-annotation-action": {
		actionRuleEriccornelissenGitTagAnnotationAction,
	},
	"kceb/git-message-action": {
		actionRuleKcebGitMessageAction,
	},
	"roots/issue-closer": {
		actionRuleRootsIssueCloserIssueCloseMessage,
		actionRuleRootsIssueCloserPrCloseMessage,
	},
	"sergeysova/jq-action": {
		actionRuleSergeysovaJqAction,
	},
}

var stepRuleRun = stepRule{
	appliesTo: func(step *JobStep) bool {
		return len(step.Run) > 0
	},
	rule: rule{
		id:    "ADES100",
		title: "Expression in 'run:' directive",
		description: `
When an expression appears in a 'run:' directive you can avoid any potential attacks by extracting
the expression into an environment variable and using the environment variable instead.

For example, given the workflow snippet:

    - name: Example step
      run: |
        echo 'Hello ${{ inputs.name }}'

it can be made safer by converting it into:

    - name: Example step
      env:
        NAME: ${{ inputs.name }} # <- Assign the expression to an environment variable
      run: |
        echo "Hello $NAME"
      #      ^      ^^^^^
      #      |      | Replace the expression with the environment variable
      #      |
      #      | Note: the use of double quotes is required in this example (for interpolation)`,
		extractFrom: func(step *JobStep) string {
			return step.Run
		},
		suggestion: suggestShellEnv,
	},
}

var stepRules = []stepRule{
	stepRuleRun,
}

func isBeforeVersion(uses *StepUses, version string) bool {
	// This comparison checks that the `Ref` is a semantic version string, which is currently the
	// only supported type of `Ref`.
	if semver.Canonical(uses.Ref) != uses.Ref {
		return false
	}

	return semver.Compare(version, uses.Ref) > 0
}

// Explain returns an explanation for a rule.
func Explain(ruleId string) (string, error) {
	r, err := findRule(ruleId)
	if err != nil {
		return "", err
	}

	explanation := fmt.Sprintf("%s - %s\n%s", r.id, r.title, r.description)
	return explanation, nil
}

// Suggestion returns a suggestion for the violation.
func Suggestion(violation *Violation) (string, error) {
	ruleId := violation.RuleId
	r, err := findRule(ruleId)
	if err != nil {
		return "", err
	}

	return r.suggestion(violation), nil
}

func findRule(ruleId string) (rule, error) {
	for _, rs := range actionRules {
		for _, r := range rs {
			if r.rule.id == ruleId {
				return r.rule, nil
			}
		}
	}

	for _, r := range stepRules {
		if r.rule.id == ruleId {
			return r.rule, nil
		}
	}

	return rule{}, fmt.Errorf("unknown rule %q", ruleId)
}

func suggestJavaScriptEnv(violation *Violation) string {
	return suggestUseEnv(violation.Problem, "process.env.", "")
}

func suggestJavaScriptLiteralEnv(violation *Violation) string {
	return suggestUseEnv(violation.Problem, "${process.env.", "}")
}

func suggestShellEnv(violation *Violation) string {
	return suggestUseEnv(violation.Problem, "$", "")
}

func suggestUseEnv(problem, pre, post string) string {
	var sb strings.Builder

	name := getVariableNameForExpression(problem)
	replacement := pre + name + post

	sb.WriteString(fmt.Sprintf("    1. Set `%s: %s` in the step's `env` map", name, problem))
	sb.WriteRune('\n')
	sb.WriteString(fmt.Sprintf("    2. Replace all occurrences of `%s` by `%s`", problem, replacement))
	sb.WriteRune('\n')
	sb.WriteString("       (make sure to keep the behavior of the script the same)")

	return sb.String()
}

func getVariableNameForExpression(expression string) string {
	name := expression[strings.LastIndex(expression, ".")+1:]
	name = strings.TrimRight(name, "}")
	name = strings.TrimSpace(name)
	return strings.ToUpper(name)
}
