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

import "fmt"

func explain(violationId string) (explanation string, err error) {
	switch violationId {
	case expressionInRunScriptId:
		explanation = explainAdes100()
	case expressionInActionsGithubScriptId:
		explanation = explainAdes101()
	case expressionInGitTagAnnotationActionTagInputId:
		explanation = explainAdes200()
	case expressionInGitMessageActionShaInputId:
		explanation = explainAdes201()
	default:
		err = fmt.Errorf("unknown rule %q", violationId)
	}

	return explanation, err
}

func explainAdes100() string {
	return fmt.Sprintln(`ADES100 - Expression in 'run:' directive

When a workflow expression appears in a 'run:' directive you can avoid any potential attacks by
extracting the expression into an environment variable and using the environment variable instead.

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
      #      | Note: the use of double quotes is required in this example (for interpolation)`)
}

func explainAdes101() string {
	return fmt.Sprintln(`ADES101 - Expression in 'actions/github-script' script

When a workflow expression appears in a 'actions/github-script' script you can avoid any potential
attacks by extracting the expression into an environment variable and using the environment variable
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
      #                     | Note: the use of backticks is required in this example (for interpolation)`)
}

func explainAdes200() string {
	return fmt.Sprintln(`ADES200 - Expression in 'ericcornelissen/git-tag-annotation-action' tag input

When a workflow expression is used in the tag input for 'ericcornelissen/git-tag-annotation-action'
in v1.0.0 or earlier it may be used to execute arbitrary shell commands, see GHSA-hgx2-4pp9-357g. To
mitigate this, upgrade the action to a non-vulnerable version.`)
}

func explainAdes201() string {
	return fmt.Sprintln(`ADES201 - Expression in 'kceb/git-message-action' sha input

When a workflow expression is used in the sha input for 'kceb/git-message-action' in v1.1.0 or
earlier it may be used to execute arbitrary shell commands (no vulnerability identifier available).
To mitigate this, upgrade the action to a non-vulnerable version.`)
}
