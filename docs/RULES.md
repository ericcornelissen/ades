<!-- SPDX-License-Identifier: GFDL-1.3-or-later -->

# Rules

All rules supported by `ades` are listed and explained in this document, including an example of how
to address it.

## <a id="ADES100"></a> ADES100 - Expression in `run:` directive

When an expression appears in a `run:` directive you can avoid potential attacks by extracting the
expression into an environment variable and using the environment variable instead.

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

Note that the changes depend on the runner and shell being used. For example, on Windows (or when
using `shell: powershell`) the environment variable must be accessed as `$Env:NAME`.

## <a id="ADES101"></a> ADES101 - Expression in `actions/github-script` script

When an expression appears in a `actions/github-script` script you can avoid potential attacks by
extracting the expression into an environment variable and using the environment variable instead.

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

## <a id="ADES102"></a> ADES102 - Expression in `issue-close-message` input of `roots/issue-closer-action`

When an expression appears in the `issue-close-message` input of `roots/issue-closer-action` it is
interpreted as an ES6-style template literal. You can avoid potential attacks by extracting the
expression into an environment variable and using the environment variable instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: roots/issue-closer-action@v1
  with:
    issue-close-message: Closing ${{ github.event.issue.title }}
```

it can be made safer by converting it into:

```yaml
- name: Example step
  uses: roots/issue-closer-action@v1
  env:
    NAME: ${{ github.event.issue.title }} # <- Assign the expression to an environment variable
  with:
    issue-close-message: Closing ${process.env.NAME}
  #                              ^^^^^^^^^^^^^^^^^^^
  #                              | Replace the expression with the environment variable
```

## <a id="ADES103"></a> ADES103 - Expression in `pr-close-message` input of `roots/issue-closer-action`

When an expression appears in the `pr-close-message` input of `roots/issue-closer-action` it is
interpreted as an ES6-style template literal. You can avoid potential attacks by extracting the
expression into an environment variable and using the environment variable instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: roots/issue-closer-action@v1
  with:
    pr-close-message: Closing ${{ github.event.issue.title }}
```

it can be made safer by converting it into:

```yaml
- name: Example step
  uses: roots/issue-closer-action@v1
  env:
    NAME: ${{ github.event.issue.title }} # <- Assign the expression to an environment variable
  with:
    pr-close-message: Closing ${process.env.NAME}
  #                           ^^^^^^^^^^^^^^^^^^^
  #                           | Replace the expression with the environment variable
```

## <a id="ADES104"></a> ADES104 - Expression in `cmd` input of `sergeysova/jq-action`

When an expression appears in the `cmd` input of `sergeysova/jq-action` you can avoid any potential
attack by extracting the expression into an environment variable and using the environment variable
instead.

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
    cmd: jq .version "$FILE" -r
  #                 / ^^^^^
  #                 | | Replace the expression with the environment variable
  #                 |
  #                 | Note: use double quotes to avoid argument splitting
```

## <a id="ADES105"></a> ADES105 - Expression in `run` input of `addnab/docker-run-action`

When an expression appears in the `run` input of `addnab/docker-run-action` you can avoid any
potential attack by removing the expression. There is no safe way to use untrusted inputs here
without risking injection.

Do NOT pass environment variables into the container through the action's options input. This opens
up alternative attack vectors because the options are not validated.

## <a id="ADES106"></a> ADES106 - Expression in `expression` input of `cardinalby/js-eval-action`

When an expression appears in the `expression` input of `cardinalby/js-eval-action` you can avoid
any potential attack by extracting the expression into an environment variable and using the
environment variable instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: cardinalby/js-eval-action@v1
  with:
    expression: 1 + parseInt(${{ inputs.value }})
```

it can be made safer by converting it into:

```yaml
- name: Example step
  uses: cardinalby/js-eval-action@v1
  env:
    VALUE: ${{ inputs.value }} # <- Assign the expression to an environment variable
  with:
    expression: 1 + parseInt(env.VALUE)
  #                          ^^^^^^^^^
  #                          | Replace the expression with the environment variable
```

## <a id="ADES107"></a> ADES107 - Expression in `custom_payload` input of `8398a7/action-slack`

When an expression appears in the `custom_payload` input of `8398a7/action-slack` you can avoid any
potential attack by extracting the expression into an environment variable and using the environment
variable instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: 8398a7/action-slack@v3
  with:
    custom_payload: |
      { attachments: [{ color: '${{ inputs.color }}' }] }
```

it can be made safer by converting it into:

```yaml
- name: Example step
  uses: 8398a7/action-slack@v3
  env:
    COLOR: ${{ inputs.color }} # <- Assign the expression to an environment variable
  with:
    custom_payload: |
      { attachments: [{ color: process.env.COLOR }] }
  #                            ^^^^^^^^^^^^^^^^^
  #                            | Replace the expression with the environment variable
```

## <a id="ADES108"></a> ADES108 - Expression in `script` input of `appleboy/ssh-action`

When an expression appears in the `script` input of `appleboy/ssh-action` you can avoid any
potential attack by extracting the expression into an environment variable and using the environment
variable instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: appleboy/ssh-action@v1
  with:
    script: echo 'Hello ${{ inputs.name }}'
```

it can be made safer by converting it into:

```yaml
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
```

## <a id="ADES109"></a> ADES109 - Expression in `script` input of `jannekem/run-python-script-action`

When an expression appears in the `script` input of `jannekem/run-python-script-action` you can
avoid any potential attack by extracting the expression into an environment variable and using the
environment variable instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: jannekem/run-python-script-action@v1
  with:
    script: print("Hello ${{ inputs.name }}")
```

it can be made safer by converting it into:

```yaml
- name: Example step
  uses: jannekem/run-python-script-action@v1
  env:
    NAME: ${{ inputs.name }} # <- Assign the expression to an environment variable
  with:
    script: print(f"Hello {os.getenv('NAME')}")
  #               ^       ^^^^^^^^^^^^^^^^^^^
  #               |       | Replace the expression with the environment variable
  #               |
  #               | Note: the use of string interpolation is required in this example
```

## <a id="ADES110"></a> ADES110 - Expression in `script` input of `Amadevus/pwsh-script`

When an expression appears in the `script` input of `Amadevus/pwsh-script` you can avoid any
potential attack by extracting the expression into an environment variable and using the environment
variable instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: Amadevus/pwsh-script@v2
  with:
    script: |
      Write-Output 'Hello ${{ inputs.name }}'
```

it can be made safer by converting it into:

```yaml
- name: Example step
  uses: Amadevus/pwsh-script@v2
  env:
    NAME: ${{ inputs.name }} # <- Assign the expression to an environment variable
  with:
    script: |
      Write-Output "Hello $env:NAME"
  #                ^      ^^^^^^^^^
  #                |      | Replace the expression with the environment variable
  #                |
  #                | Note: the use of double quotes is required in this example (for interpolation)
```

## <a id="ADES111"></a> ADES111 - Expression in `cmd` input of `mikefarah/yq`

When an expression appears in the `cmd` input of `mikefarah/yq` you can avoid any potential attack
by extracting the expression into an environment variable and using the environment variable
instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: mikefarah/yq@master
  with:
    cmd: yq '${{ inputs.query }}' 'config.yml'
```

it can be made safer by converting it into:

```yaml
- name: Example step
  uses: mikefarah/yq@master
  env:
    QUERY: ${{ inputs.query }} # <- Assign the expression to an environment variable
  with:
    cmd: yq "$QUERY" 'config.yml'
  #        / ^^^^^^
  #        | | Replace the expression with the environment variable
  #        |
  #        | Note: the use of double quotes is required in this example (for interpolation)
```

## <a id="ADES112"></a> ADES112 - Expression in `cmd` input of `devorbitus/yq-action-output`

When an expression appears in the `cmd` input of `devorbitus/yq-action-output` you can avoid any
potential attack by extracting the expression into an environment variable and using the environment
variable instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: devorbitus/yq-action-output@v1.1
  with:
    cmd: yq eval '${{ inputs.query }}' 'config.yml'
```

it can be made safer by converting it into:

```yaml
- name: Example step
  uses: devorbitus/yq-action-output@v1.1
  env:
    QUERY: ${{ inputs.query }} # <- Assign the expression to an environment variable
  with:
    cmd: yq eval "$QUERY" 'config.yml'
  #             / ^^^^^^
  #             | | Replace the expression with the environment variable
  #             |
  #             | Note: the use of double quotes is required in this example (for interpolation)
```

## <a id="ADES113"></a> ADES113 - Expression in `inlineScript` input of `azure/powershell`

When an expression appears in the `inlineScript` input of `azure/powershell` you can avoid any
potential attack by extracting the expression into an environment variable and using the environment
variable instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: azure/powershell@v2.0.0
  with:
    azPSVersion: latest
    inlineScript: |
      Write-Output 'Hello ${{ inputs.name }}'
```

it can be made safer by converting it into:

```yaml
- name: Example step
  uses: azure/powershell@v2.0.0
  env:
    NAME: ${{ inputs.name }} # <- Assign the expression to an environment variable
  with:
    azPSVersion: latest
    inlineScript: |
      Write-Output "Hello $env:NAME"
  #                ^      ^^^^^^^^^
  #                |      | Replace the expression with the environment variable
  #                |
  #                | Note: the use of double quotes is required in this example (for interpolation)
```

## <a id="ADES114"></a> ADES114 - Expression in `item_exec` input of `gautamkrishnar/blog-post-workflow`

When an expression appears in the `item_exec` input of `gautamkrishnar/blog-post-workflow` you can
avoid any potential attack by extracting the expression into an environment variable and using the
environment variable instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: gautamkrishnar/blog-post-workflow@1.9.4
  with:
    item_exec: |
      post.includes('${{ inputs.substr }}')
```

it can be made safer by converting it into:

```yaml
- name: Example step
  uses: gautamkrishnar/blog-post-workflow@1.9.4
  env:
    SUBSTR: ${{ inputs.substr }} # <- Assign the expression to an environment variable
  with:
    item_exec: |
      post.includes(`${process.env.SUBSTR}`)
  #                / ^ ^^^^^^^^^^^^^^^^^^
  #                | | | Replace the expression with the environment variable
  #                | |
  #                | | Note: the use of ${...} is required in this example (for interpolating)
  #                |
  #                | Note: the use of backticks is required in this example (for interpolation)
```

## <a id="ADES115"></a> ADES115 - Expression in `inlineScript` input of `azure/cli`

When an expression appears in the `inlineScript` input of `azure/cli` you can avoid any potential
attack by extracting the expression into an environment variable and using the environment variable
instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: azure/cli@v2.2.0
  with:
    inlineScript: |
      az vm create --name '${{ inputs.name }}'
```

it can be made safer by converting it into:

```yaml
- name: Example step
  uses: azure/cli@v2.2.0
  env:
    NAME: ${{ inputs.name }} # <- Assign the expression to an environment variable
  with:
    item_exec: |
      az vm create --name "$NAME"
  #                      / ^^^^^
  #                      | | Replace the expression with the environment variable
  #                      |
  #                      | Note: the use of double quotes is required in this example (for interpolation)
```

## <a id="ADES116"></a> ADES116 - Expression in `overwrite` input of `Skitionek/notify-microsoft-teams`

When an expression appears in the `overwrite` input of `Skitionek/notify-microsoft-teams` you can
avoid any potential attack by extracting the expression into an environment variable and using the
environment variable instead.

For example, given the workflow snippet:

```yaml
- name: Example step
  uses: Skitionek/notify-microsoft-teams@v1.0.8
  with:
    overwrite: |
      {title: "${{ inputs.title }}"}
```

it can be made safer by converting it into:

```yaml
- name: Example step
  uses: Skitionek/notify-microsoft-teams@v1.0.8
  env:
    TITLE: ${{ inputs.title }} # <- Assign the expression to an environment variable
  with:
    overwrite: |
      {title: `${process.env.TITLE}`}
  #           ^  ^^^^^^^^^^^^^^^^^
  #           |  | Replace the expression with the environment variable
  #           |
  #           | Note: the use of backticks is required in this example (for interpolation)
```

## <a id="ADES200"></a> ADES200 - Expression in `tag` input of `ericcornelissen/git-tag-annotation-action`

When an expression is used in the `tag` input of `ericcornelissen/git-tag-annotation-action` in
v1.0.0 or earlier it may be used to execute arbitrary shell commands, see [GHSA-hgx2-4pp9-357g]. To
mitigate this, upgrade the action to a non-vulnerable version.

[GHSA-hgx2-4pp9-357g]: https://github.com/ericcornelissen/git-tag-annotation-action/security/advisories/GHSA-hgx2-4pp9-357g

## <a id="ADES201"></a> ADES201 - Expression in `sha` input of `kceb/git-message-action`

When an expression is used in the `sha` input of `kceb/git-message-action` in v1.1.0 or earlier it
may be used to execute arbitrary shell commands (no vulnerability identifier available). To mitigate
this, upgrade the action to a non-vulnerable version.

## <a id="ADES202"></a> ADES202 - Expression in `summary` input of `atlassian/gajira-create`

When an expression is used in the `summary` input of `atlassian/gajira-create` in v2.0.0 or earlier
it may be used to execute arbitrary JavaScript code, see [GHSA-4xqx-pqpj-9fqw]. To mitigate this,
upgrade the action to a non-vulnerable version.

[GHSA-4xqx-pqpj-9fqw]: https://github.com/advisories/GHSA-4xqx-pqpj-9fqw

## <a id="ADES203"></a> ADES203 - Expression in `args` input of `SonarSource/sonarqube-scan-action`

When an expression is used in the `args` input of `SonarSource/sonarqube-scan-action` between v4.0.0
and v5.3.0 it may be used to execute arbitrary shell commands, see [GHSA-f79p-9c5r-xg88]. To
mitigate this, upgrade the action to a non-vulnerable version.

[GHSA-f79p-9c5r-xg88]: https://github.com/advisories/GHSA-f79p-9c5r-xg88

## <a id="ADES204"></a> ADES204 - Expression in `lycheeVersion` input of `lycheeverse/lychee`

When an expression is used in the `lycheeVersion` input of `lycheeverse/lychee` in v2.0.1 or earlier
it may be used to execute arbitrary shell commands, see [GHSA-65rg-554r-9j5x]. To mitigate this,
upgrade the action to a non-vulnerable version.

[GHSA-65rg-554r-9j5x]: https://github.com/advisories/GHSA-65rg-554r-9j5x

## <a id="ADES205"></a> ADES205 - Expression in `pull-request-body` input of `OZI-Project/publish`

When an expression is used in the `pull-request-body` input of `OZI-Project/publish` between v1.13.2
and v1.13.5 it may be used to execute arbitrary shell commands, see [GHSA-2487-9f55-2vg9]. To
mitigate this, upgrade the action to a non-vulnerable version.

[GHSA-2487-9f55-2vg9]: https://github.com/advisories/GHSA-2487-9f55-2vg9

## <a id="ADES206"></a> ADES206 - Expression in `pattern` input of `fish-shop/syntax-check`

When an expression is used in the `pattern` input of `fish-shop/syntax-check` in v1.6.11 or earlier
it may be used to execute arbitrary shell commands, see [GHSA-xj87-mqvh-88w2]. To mitigate this,
upgrade the action to a non-vulnerable version.

[GHSA-xj87-mqvh-88w2]: https://github.com/advisories/GHSA-xj87-mqvh-88w2

## <a id="ADES207"></a> ADES207 - Expression in _any_ input of `aquasecurity/trivy-action`

When an expression is used in _any_ input of `aquasecurity/trivy-action` starting from v0.31.0 up
to v0.33.1 it may be used to execute arbitrary shell commands, see [GHSA-9p44-j4g5-cfx5]. To
mitigate this, upgrade the action to a non-vulnerable version.

[GHSA-9p44-j4g5-cfx5]: https://github.com/advisories/GHSA-9p44-j4g5-cfx5
