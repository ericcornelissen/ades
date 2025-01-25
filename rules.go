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
	"strings"

	"golang.org/x/mod/semver"
)

type rule struct {
	extractFrom func(step *JobStep) string
	fix         func(violation *Violation) []fix
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

type fix struct {
	// New is the replacement string to fix a violation.
	New string

	// Old is a regular expression to search and replace with in order to fix a violation.
	Old regexp.Regexp
}

var actionRuleActionsGitHubScript = actionRule{
	appliesTo: func(_ *StepUses) bool {
		return true
	},
	rule: rule{
		id:    "ADES101",
		title: "Expression in 'actions/github-script' script",
		description: `
When an expression appears in a 'actions/github-script' script you can avoid potential attacks by
extracting the expression into an environment variable and using the environment variable instead.

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
	},
}

var actionRuleAtlassianGajiraCreate = actionRule{
	appliesTo: func(uses *StepUses) bool {
		return isBeforeVersion(uses, "v2.0.1")
	},
	rule: rule{
		id:    "ADES202",
		title: "Expression in 'atlassian/gajira-create' summary input",
		description: `
When an expression is used in the summary input for 'atlassian/gajira-create' in v2.0.0 or earlier
it may be used to execute arbitrary JavaScript code, see GHSA-4xqx-pqpj-9fqw. To mitigate this,
upgrade the action to a non-vulnerable version.`,
		extractFrom: func(step *JobStep) string {
			return step.With["summary"]
		},
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
an ES6-style template literal. You can avoid potential attacks by extracting the expression into an
environment variable and using the environment variable instead.

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
		fix: func(violation *Violation) []fix {
			var step JobStep
			switch source := (violation.source).(type) {
			case *Manifest:
				step = source.Runs.Steps[violation.stepIndex]
			case *Workflow:
				step = source.Jobs[violation.jobKey].Steps[violation.stepIndex]
			}

			name := getVariableNameForExpression(violation.Problem)
			if _, ok := step.Env[name]; ok {
				return nil
			}

			fixes := fixAddEnvVar(step, name, violation.Problem)
			fixes = append(fixes, fixReplaceIn(
				step.With["issue-close-message"],
				violation.Problem,
				fmt.Sprintf("${process.env.%s}", name),
			))

			return fixes
		},
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
interpreted as an ES6-style template literal. You can avoid potential attacks by extracting the
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
	},
}

var actionRules = map[string][]actionRule{
	"actions/github-script": {
		actionRuleActionsGitHubScript,
	},
	"atlassian/gajira-create": {
		actionRuleAtlassianGajiraCreate,
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

var stepRuleRunDefault = stepRule{
	appliesTo: func(step *JobStep) bool {
		return len(step.Run) != 0 && !isShellPython(step)
	},
	rule: rule{
		id:    "ADES100",
		title: "Expression in 'run:' directive",
		description: `
When an expression appears in a 'run:' directive you can avoid potential attacks by extracting the
expression into an environment variable and using the environment variable instead.

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
	},
}

var stepRuleRunPython = stepRule{
	appliesTo: func(step *JobStep) bool {
		return len(step.Run) != 0 && isShellPython(step)
	},
	rule: rule{
		id:    "ADES105",
		title: "Expression in 'run:' directive with Python",
		description: `
When an expression appears in a 'run:' directive you can avoid potential attacks by extracting the
expression into an environment variable and using the environment variable instead.

For example, given the workflow snippet:

    - name: Example step
      shell: python
      run: |
        print('Hello ${{ inputs.name }}')

it can be made safer by converting it into:

    - name: Example step
      shell: python
      env:
        NAME: ${{ inputs.name }} # <- Assign the expression to an environment variable
      run: |
        print(f'Hello {os.getenv("HOME")}')
      #       ^        ^^^^^^^^^^^^^^^^^
      #       |        | Replace the expression with the environment variable
      #       |
      #       | Note: the use of the literal prefix 'f' is required for interpolation`,
		extractFrom: func(step *JobStep) string {
			return step.Run
		},
	},
}

var stepRules = []stepRule{
	stepRuleRunDefault,
	stepRuleRunPython,
}

func isBeforeVersion(uses *StepUses, version string) bool {
	ref := uses.Ref
	if !semver.IsValid(ref) {
		ref = uses.Annotation
		if !semver.IsValid(ref) {
			return false
		}
	}

	switch {
	case semver.Canonical(ref) == ref:
		return semver.Compare(ref, version) < 0
	case semver.MajorMinor(ref) == ref:
		return semver.Compare(ref, semver.MajorMinor(version)) < 0
	default:
		return semver.Compare(ref, semver.Major(version)) < 0
	}
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

// Fix produces a set of fixes to address the violation if possible. If the return value is nil the
// violation cannot be fixed automatically.
func Fix(violation *Violation) ([]fix, error) {
	ruleId := violation.RuleId
	r, err := findRule(ruleId)
	if err != nil {
		return nil, err
	}

	if r.fix == nil {
		return nil, nil
	}

	return r.fix(violation), nil
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

func fixAddEnvVar(step JobStep, name, value string) []fix {
	if step.Env == nil {
		return []fix{
			{
				Old: *regexp.MustCompile(fmt.Sprintf(`\n(\s+)uses:\s*%s.*?\n`, step.Uses)),
				New: fmt.Sprintf("${0}${1}env:\n${1}  %s: %s\n", name, value),
			},
			{
				Old: *regexp.MustCompile(fmt.Sprintf(`\n(\s+)-(\s+)uses:\s*%s.*?\n`, step.Uses)),
				New: fmt.Sprintf("${0}${1} ${2}env:\n${1} ${2}  %s: %s\n", name, value),
			},
		}
	} else {
		var sb strings.Builder
		sb.WriteString(`env:\s*\n(?:`)
		for k, v := range step.Env {
			sb.WriteString(fmt.Sprintf(`(\s*)%s\s*:\s*%s\s*\n|`, k, v))
		}
		sb.WriteString(`)+`)

		return []fix{
			{
				Old: *regexp.MustCompile(sb.String()),
				New: fmt.Sprintf("${0}${1}%s: %s\n", name, value),
			},
		}
	}
}

func fixReplaceIn(s, old, new string) fix {
	return fix{
		Old: *regexp.MustCompile(regexp.QuoteMeta(s)),
		New: strings.ReplaceAll(s, old, new),
	}
}

func getVariableNameForExpression(expression string) string {
	name := expression[strings.LastIndex(expression, ".")+1:]
	name = strings.TrimRight(name, "}")
	name = strings.TrimSpace(name)
	return strings.ToUpper(name)
}

func isShellPython(step *JobStep) bool {
	return step.Shell == "python"
}
