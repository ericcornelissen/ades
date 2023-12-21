<!-- SPDX-License-Identifier: GFDL-1.3-or-later -->

# Rules

All rules supported by `ades` are listed and explained in this document, including an example of how
to address it.

## ADES100 - Expression in `run:` directive

When a workflow expression appears in a `run:` directive you can avoid any potential attacks by
extracting the expression into an environment variable and using the environment variable instead.

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

## ADES101 - Expression in `actions/github-script` script

When a workflow expression appears in a `actions/github-script` script you can avoid any potential
attacks by extracting the expression into an environment variable and using the environment variable
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

## ADES200 - Expression in `ericcornelissen/git-tag-annotation-action` tag input

When a workflow expression is used in the tag input for `ericcornelissen/git-tag-annotation-action`
in v1.0.0 or earlier it may be used to execute arbitrary shell commands, see [GHSA-hgx2-4pp9-357g].
To avoid this, upgrade the action to a non-vulnerable version.

[GHSA-hgx2-4pp9-357g]: https://github.com/ericcornelissen/git-tag-annotation-action/security/advisories/GHSA-hgx2-4pp9-357g

## ADES201 - Expression in `kceb/git-message-action` sha input

When a workflow expression is used in the sha input for `kceb/git-message-action` in v1.1.0 or
earlier it may be used to execute arbitrary shell commands (no vulnerability identifier available).
To mitigate this, upgrade the action to a non-vulnerable version.
