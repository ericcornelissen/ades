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

	"github.com/ericcornelissen/go-gha-models"
	"golang.org/x/mod/semver"
)

type rule struct {
	extractFrom func(step *gha.Step) string
	fix         func(violation *Violation) []fix
	id          string
	title       string
	description string
}

type actionRule struct {
	appliesTo func(uses *gha.Uses) bool
	rule      rule
}

type stepRule struct {
	appliesTo func(step *gha.Step) bool
	rule      rule
}

type fix struct {
	// New is the replacement string to fix a violation.
	New string

	// Old is a regular expression to search and replace with in order to fix a violation.
	Old regexp.Regexp
}

var actionRule8398a7ActionSlack = actionRule{
	appliesTo: func(_ *gha.Uses) bool {
		return true
	},
	rule: rule{
		id:    "ADES107",
		title: "Expression in 'custom_payload' input of '8398a7/action-slack'",
		description: `
When an expression appears in the 'custom_payload' input of '8398a7/action-slack' you can avoid any
potential attack by extracting the expression into an environment variable and using the environment
variable instead.

For example, given the workflow snippet:

    - name: Example step
      uses: 8398a7/action-slack@v3
      with:
        custom_payload: |
          { attachments: [{ color: '${{ inputs.color }}' }] }

it can be made safer by converting it into:

    - name: Example step
      uses: 8398a7/action-slack@v3
      env:
        COLOR: ${{ inputs.color }} # <- Assign the expression to an environment variable
      with:
        custom_payload: |
          { attachments: [{ color: process.env.COLOR }] }
      #                            ^^^^^^^^^^^^^^^^^
      #                            | Replace the expression with the environment variable
`,
		extractFrom: func(step *gha.Step) string {
			return step.With["custom_payload"]
		},
	},
}

var actionRuleActionsGitHubScript = actionRule{
	appliesTo: func(_ *gha.Uses) bool {
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
      #                     | Note: the use of backticks is required in this example (for interpolation)
`,
		extractFrom: func(step *gha.Step) string {
			return step.With["script"]
		},
	},
}

var actionRuleAddnabDockerRunAction = actionRule{
	appliesTo: func(_ *gha.Uses) bool {
		return true
	},
	rule: rule{
		id:    "ADES105",
		title: "Expression in 'run' input of 'addnab/docker-run-action'",
		description: `
When an expression appears in the 'run' input of 'addnab/docker-run-action' you can avoid any
potential attack by removing the expression. There is no safe way to use untrusted inputs here
without risking injection.

Do NOT pass environment variables into the container through the action's options input. This opens
up alternative attack vectors because the options are not validated.
`,
		extractFrom: func(step *gha.Step) string {
			return step.With["run"]
		},
	},
}

var actionRuleAppleboySshAction = actionRule{
	appliesTo: func(_ *gha.Uses) bool {
		return true
	},
	rule: rule{
		id:    "ADES108",
		title: "Expression in 'script' input of 'appleboy/ssh-action'",
		description: `
When an expression appears in the 'script' input of 'appleboy/ssh-action' you can avoid any
potential attack by extracting the expression into an environment variable and using the environment
variable instead.

For example, given the workflow snippet:

    - name: Example step
      uses: appleboy/ssh-action@v1
      with:
        script: echo 'Hello ${{ inputs.name }}'

it can be made safer by converting it into:

    - name: Example step
      uses: appleboy/ssh-action@v1
      env:
        NAME: ${{ inputs.name }} # <- Assign the expression to an environment variable
      with:
        envs: NAME # <- Pass the environment variable through SSH
        script: echo "Hello $NAME"
      #              ^      ^^^^^
      #              |      | Replace the expression with the environment variable
      #              |
      #              | Note: the use of double quotes is required in this example (for interpolation)
`,
		extractFrom: func(step *gha.Step) string {
			return step.With["script"]
		},
	},
}

var actionRuleAtlassianGajiraCreate = actionRule{
	appliesTo: func(uses *gha.Uses) bool {
		return isBeforeVersion(uses, "v2.0.1")
	},
	rule: rule{
		id:    "ADES202",
		title: "Expression in 'summary' input of 'atlassian/gajira-create'",
		description: `
When an expression is used in the 'summary' input of 'atlassian/gajira-create' in v2.0.0 or earlier
it may be used to execute arbitrary JavaScript code, see GHSA-4xqx-pqpj-9fqw. To mitigate this,
upgrade the action to a non-vulnerable version.
`,
		extractFrom: func(step *gha.Step) string {
			return step.With["summary"]
		},
	},
}

var actionRuleCardinalbyJsEvalAction = actionRule{
	appliesTo: func(_ *gha.Uses) bool {
		return true
	},
	rule: rule{
		id:    "ADES106",
		title: "Expression in 'expression' input of 'cardinalby/js-eval-action'",
		description: `
When an expression appears in the 'expression' input of 'cardinalby/js-eval-action' you can avoid
any potential attack by extracting the expression into an environment variable and using the
environment variable instead.

For example, given the workflow snippet:

    - name: Example step
      uses: cardinalby/js-eval-action@v1
      with:
        expression: 1 + parseInt(${{ inputs.value }})

it can be made safer by converting it into:

    - name: Example step
      uses: cardinalby/js-eval-action@v1
      env:
        VALUE: ${{ inputs.value }} # <- Assign the expression to an environment variable
      with:
        expression: 1 + parseInt(env.VALUE)
      #                          ^^^^^^^^^
      #                          | Replace the expression with the environment variable
`,
		extractFrom: func(step *gha.Step) string {
			return step.With["expression"]
		},
		fix: func(violation *Violation) []fix {
			var step gha.Step
			switch source := (violation.source).(type) {
			case *gha.Manifest:
				step = source.Runs.Steps[violation.stepIndex]
			case *gha.Workflow:
				step = source.Jobs[violation.jobKey].Steps[violation.stepIndex]
			}

			name := getVariableNameForExpression(violation.Problem)
			if _, ok := step.Env[name]; ok {
				return nil
			}

			fixes := fixAddEnvVar(step, name, violation.Problem)
			fixes = append(fixes, fixReplaceIn(
				step.With["expression"],
				violation.Problem,
				"env."+name,
			))

			return fixes
		},
	},
}

var actionRuleEriccornelissenGitTagAnnotationAction = actionRule{
	appliesTo: func(uses *gha.Uses) bool {
		return isBeforeVersion(uses, "v1.0.1")
	},
	rule: rule{
		id:    "ADES200",
		title: "Expression in 'tag' input of 'ericcornelissen/git-tag-annotation-action'",
		description: `
When an expression is used in the 'tag' input of 'ericcornelissen/git-tag-annotation-action' in
v1.0.0 or earlier it may be used to execute arbitrary shell commands, see GHSA-hgx2-4pp9-357g. To
mitigate this, upgrade the action to a non-vulnerable version.
`,
		extractFrom: func(step *gha.Step) string {
			return step.With["tag"]
		},
	},
}

var actionRuleFishShopSyntaxCheck = actionRule{
	appliesTo: func(uses *gha.Uses) bool {
		return isBeforeVersion(uses, "v1.6.12")
	},
	rule: rule{
		id:    "ADES206",
		title: "Expression in 'pattern' input of 'fish-shop/syntax-check'",
		description: `
When an expression is used in the 'pattern' input of 'fish-shop/syntax-check' in v1.6.11 or earlier
it may be used to execute arbitrary shell commands, see GHSA-xj87-mqvh-88w2. To mitigate this,
upgrade the action to a non-vulnerable version.
`,
		extractFrom: func(step *gha.Step) string {
			return step.With["pattern"]
		},
	},
}

var actionRuleKcebGitMessageAction = actionRule{
	appliesTo: func(uses *gha.Uses) bool {
		return isBeforeVersion(uses, "v1.2.0")
	},
	rule: rule{
		id:    "ADES201",
		title: "Expression in 'sha' input of 'kceb/git-message-action'",
		description: `
When an expression is used in the 'sha' input of 'kceb/git-message-action' in v1.1.0 or earlier it
may be used to execute arbitrary shell commands (no vulnerability identifier available). To mitigate
this, upgrade the action to a non-vulnerable version.
`,
		extractFrom: func(step *gha.Step) string {
			return step.With["sha"]
		},
	},
}

var actionRuleLycheeverseLycheeAction = actionRule{
	appliesTo: func(uses *gha.Uses) bool {
		return isBeforeVersion(uses, "v2.0.2")
	},
	rule: rule{
		id:    "ADES204",
		title: "Expression in 'lycheeVersion' input of 'lycheeverse/lychee'",
		description: `
When an expression is used in the 'lycheeVersion' input of 'lycheeverse/lychee' in v2.0.1 or earlier
it may be used to execute arbitrary shell commands, see GHSA-65rg-554r-9j5x. To mitigate this,
upgrade the action to a non-vulnerable version.
`,
		extractFrom: func(step *gha.Step) string {
			return step.With["lycheeVersion"]
		},
	},
}

var actionRuleOziProjectPublish = actionRule{
	appliesTo: func(uses *gha.Uses) bool {
		return isAtOrAfterVersion(uses, "v1.13.2") && isBeforeVersion(uses, "v1.13.6")
	},
	rule: rule{
		id:    "ADES205",
		title: "Expression in 'pull-request-body' input of 'OZI-Project/publish'",
		description: `
When an expression is used in the 'pull-request-body' input of 'OZI-Project/publish' between v1.13.2
and v1.13.5 it may be used to execute arbitrary shell commands, see GHSA-2487-9f55-2vg9. To mitigate
this, upgrade the action to a non-vulnerable version.
`,
		extractFrom: func(step *gha.Step) string {
			return step.With["pull-request-body"]
		},
	},
}

var actionRuleRootsIssueCloserActionIssueCloseMessage = actionRule{
	appliesTo: func(_ *gha.Uses) bool {
		return true
	},
	rule: rule{
		id:    "ADES102",
		title: "Expression in 'issue-close-message' input of 'roots/issue-closer-action'",
		description: `
When an expression appears in the 'issue-close-message' input of 'roots/issue-closer-action' it is
interpreted as an ES6-style template literal. You can avoid potential attacks by extracting the
expression into an environment variable and using the environment variable instead.

For example, given the workflow snippet:

    - name: Example step
      uses: roots/issue-closer-action@v1
      with:
        issue-close-message: Closing ${{ github.event.issue.title }}

it can be made safer by converting it into:

    - name: Example step
      uses: roots/issue-closer-action@v1
      env:
        NAME: ${{ github.event.issue.title }} # <- Assign the expression to an environment variable
      with:
        issue-close-message: Closing ${process.env.NAME}
      #                              ^^^^^^^^^^^^^^^^^^^
      #                              | Replace the expression with the environment variable
`,
		extractFrom: func(step *gha.Step) string {
			return step.With["issue-close-message"]
		},
		fix: func(violation *Violation) []fix {
			var step gha.Step
			switch source := (violation.source).(type) {
			case *gha.Manifest:
				step = source.Runs.Steps[violation.stepIndex]
			case *gha.Workflow:
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

var actionRuleRootsIssueCloserActionPrCloseMessage = actionRule{
	appliesTo: func(_ *gha.Uses) bool {
		return true
	},
	rule: rule{
		id:    "ADES103",
		title: "Expression in 'pr-close-message' input of 'roots/issue-closer-action'",
		description: `
When an expression appears in the 'pr-close-message' input of 'roots/issue-closer-action' it is
interpreted as an ES6-style template literal. You can avoid potential attacks by extracting the
expression into an environment variable and using the environment variable instead.

For example, given the workflow snippet:

    - name: Example step
      uses: roots/issue-closer-action@v1
      with:
        pr-close-message: Closing ${{ github.event.issue.title }}

it can be made safer by converting it into:

    - name: Example step
      uses: roots/issue-closer-action@v1
      env:
        NAME: ${{ github.event.issue.title }} # <- Assign the expression to an environment variable
      with:
        pr-close-message: Closing ${process.env.NAME}
      #                           ^^^^^^^^^^^^^^^^^^^
      #                           | Replace the expression with the environment variable
`,
		extractFrom: func(step *gha.Step) string {
			return step.With["pr-close-message"]
		},
	},
}

var actionRuleSergeysovaJqAction = actionRule{
	appliesTo: func(_ *gha.Uses) bool {
		return true
	},
	rule: rule{
		id:    "ADES104",
		title: "Expression in 'cmd' input of 'sergeysova/jq-action'",
		description: `
When an expression appears in the 'cmd' input of 'sergeysova/jq-action' you can avoid any potential
attack by extracting the expression into an environment variable and using the environment variable
instead.

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
      #                   | Replace the expression with the environment variable
`,
		extractFrom: func(step *gha.Step) string {
			return step.With["cmd"]
		},
	},
}

var actionRuleSonarSourceSonarqubeScanAction = actionRule{
	appliesTo: func(uses *gha.Uses) bool {
		return isAtOrAfterVersion(uses, "v4.0.0") && isBeforeVersion(uses, "v5.3.1")
	},
	rule: rule{
		id:    "ADES203",
		title: "Expression in 'args' input of 'SonarSource/sonarqube-scan-action'",
		description: `
When an expression is used in the 'args' input of 'SonarSource/sonarqube-scan-action' between v4.0.0
and v5.3.0 it may be used to execute arbitrary shell commands, see GHSA-f79p-9c5r-xg88. To mitigate
this, upgrade the action to a non-vulnerable version.
`,
		extractFrom: func(step *gha.Step) string {
			return step.With["args"]
		},
	},
}

var actionRules = map[string][]actionRule{
	"8398a7/action-slack": {
		actionRule8398a7ActionSlack,
	},
	"actions/github-script": {
		actionRuleActionsGitHubScript,
	},
	"addnab/docker-run-action": {
		actionRuleAddnabDockerRunAction,
	},
	"appleboy/ssh-action": {
		actionRuleAppleboySshAction,
	},
	"atlassian/gajira-create": {
		actionRuleAtlassianGajiraCreate,
	},
	"cardinalby/js-eval-action": {
		actionRuleCardinalbyJsEvalAction,
	},
	"ericcornelissen/git-tag-annotation-action": {
		actionRuleEriccornelissenGitTagAnnotationAction,
	},
	"fish-shop/syntax-check": {
		actionRuleFishShopSyntaxCheck,
	},
	"kceb/git-message-action": {
		actionRuleKcebGitMessageAction,
	},
	"lycheeverse/lychee": {
		actionRuleLycheeverseLycheeAction,
	},
	"ozi-project/publish": {
		actionRuleOziProjectPublish,
	},
	"roots/issue-closer-action": {
		actionRuleRootsIssueCloserActionIssueCloseMessage,
		actionRuleRootsIssueCloserActionPrCloseMessage,
	},
	"roots/issue-closer": {
		actionRuleRootsIssueCloserActionIssueCloseMessage,
		actionRuleRootsIssueCloserActionPrCloseMessage,
	},
	"sergeysova/jq-action": {
		actionRuleSergeysovaJqAction,
	},
	"sonarsource/sonarqube-scan-action": {
		actionRuleSonarSourceSonarqubeScanAction,
	},
}

var stepRuleRun = stepRule{
	appliesTo: func(step *gha.Step) bool {
		return len(step.Run) > 0
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
      #      | Note: the use of double quotes is required in this example (for interpolation)

Note that the changes depend on the runner and shell being used. For example, on Windows (or when
using 'shell: powershell') the environment variable must be accessed as '$Env:NAME'.
`,
		extractFrom: func(step *gha.Step) string {
			return step.Run
		},
	},
}

var stepRules = []stepRule{
	stepRuleRun,
}

func getRef(uses *gha.Uses) (string, bool) {
	ref := uses.Ref
	if !semver.IsValid(ref) {
		ref = uses.Annotation
		if !semver.IsValid(ref) {
			return "", false
		}
	}

	return ref, true
}

func isAtOrAfterVersion(uses *gha.Uses, version string) bool {
	ref, ok := getRef(uses)
	if !ok {
		return false
	}

	switch {
	case semver.Canonical(ref) == ref:
		return semver.Compare(ref, version) >= 0
	case semver.MajorMinor(ref) == ref:
		return semver.Compare(ref, semver.MajorMinor(version)) >= 0
	default:
		return semver.Compare(ref, semver.Major(version)) >= 0
	}
}

func isBeforeVersion(uses *gha.Uses, version string) bool {
	ref, ok := getRef(uses)
	if !ok {
		return false
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
	ruleId = strings.ToUpper(ruleId)
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

func fixAddEnvVar(step gha.Step, name, value string) []fix {
	if step.Env == nil {
		return []fix{
			{
				Old: *regexp.MustCompile(fmt.Sprintf(`\n(\s+)uses:\s*%s@%s.*?\n`, step.Uses.Name, step.Uses.Ref)),
				New: fmt.Sprintf("${0}${1}env:\n${1}  %s: %s\n", name, value),
			},
			{
				Old: *regexp.MustCompile(fmt.Sprintf(`\n(\s+)-(\s+)uses:\s*%s@%s.*?\n`, step.Uses.Name, step.Uses.Ref)),
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
