<!-- SPDX-License-Identifier: GFDL-1.3-or-later -->

# Rules

All rules supported by `ades` are listed and explained in this document, including an example of how
to address it.

## <a id="ADES100"></a> ADES100 - Expression in `run:` directive

When an expression appears in a `run:` directive you can avoid any potential attacks by extracting
the expression into an environment variable and using the environment variable instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  run: |
    echo 'Hello ${{ inputs.name }}'
```

it can be made safer by converting it into:

```yaml
- name: Example step
  env:
    NAME: ${{ inputs.name }} # <- Assign the expression to an environment variable
  run: |
    echo "Hello $NAME"
#        ^      ^^^^^
#        |      | Replace the expression with the environment variable
#        |
#        | Note: the use of double quotes is required in this example (for interpolation)
```

## <a id="ADES101"></a> ADES101 - Expression in `actions/github-script` script

When an expression appears in a `actions/github-script` script you can avoid any potential attacks
by extracting the expression into an environment variable and using the environment variable
instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: actions/github-script@v6
  with:
    script: console.log('Hello ${{ inputs.name }}')
```

it can be made safer by converting it into:

```yaml
- name: Example step
  uses: actions/github-script@v6
  env:
    NAME: ${{ inputs.name }} # <- Assign the expression to an environment variable
  with:
    script: console.log(`Hello ${process.env.NAME}`)
#                       ^      ^^^^^^^^^^^^^^^^^^^
#                       |      | Replace the expression with the environment variable
#                       |
#                       | Note: the use of backticks is required in this example (for interpolation)
```

## <a id="ADES102"></a> ADES102 - Expression in `roots/issue-closer` issue close message

When an expression appears in the issue close message of `roots/issue-closer` it is interpreted as
an ES6-style template literal. You can avoid any potential attacks by extracting the expression into
an environment variable and using the environment variable instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: roots/issue-closer@v1
  with:
    issue-close-message: Closing ${{ github.event.issue.title }}
```

it can be made safer by converting it into:

```yaml
- name: Example step
  uses: roots/issue-closer@v1
  env:
    NAME: ${{ github.event.issue.title }} # <- Assign the expression to an environment variable
  with:
    issue-close-message: Closing ${process.env.NAME}
  #                              ^^^^^^^^^^^^^^^^^^^
  #                              | Replace the expression with the environment variable
```

## <a id="ADES103"></a> ADES103 - Expression in `roots/issue-closer` pull request close message

When an expression appears in the pull request close message of `roots/issue-closer` it is
interpreted as an ES6-style template literal. You can avoid any potential attacks by extracting the
expression into an environment variable and using the environment variable instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: roots/issue-closer@v1
  with:
    pr-close-message: Closing ${{ github.event.issue.title }}
```

it can be made safer by converting it into:

```yaml
- name: Example step
  uses: roots/issue-closer@v1
  env:
    NAME: ${{ github.event.issue.title }} # <- Assign the expression to an environment variable
  with:
    pr-close-message: Closing ${process.env.NAME}
  #                           ^^^^^^^^^^^^^^^^^^^
  #                           | Replace the expression with the environment variable
```

## <a id="ADES104"></a> ADES104 - Expression in `sergeysova/jq-action` command

When an expression appears in the command input of `sergeysova/jq-action` you can avoid any
potential attack by extracting the expression into an environment variable and using the environment
variable instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: sergeysova/jq-action@v2
  with:
    cmd: jq .version ${{ github.event.inputs.file }} -r
```

it can be made safer by converting it into:

```yaml
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
```

## <a id="ADES200"></a> ADES200 - Expression in `ericcornelissen/git-tag-annotation-action` tag input

When an expression is used in the tag input for `ericcornelissen/git-tag-annotation-action` in
v1.0.0 or earlier it may be used to execute arbitrary shell commands, see [GHSA-hgx2-4pp9-357g]. To
mitigate this, upgrade the action to a non-vulnerable version.

[GHSA-hgx2-4pp9-357g]: https://github.com/ericcornelissen/git-tag-annotation-action/security/advisories/GHSA-hgx2-4pp9-357g

## <a id="ADES201"></a> ADES201 - Expression in `kceb/git-message-action` sha input

When an expression is used in the sha input for `kceb/git-message-action` in v1.1.0 or earlier it
may be used to execute arbitrary shell commands (no vulnerability identifier available). To mitigate
this, upgrade the action to a non-vulnerable version.
