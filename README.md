<!-- SPDX-License-Identifier: GFDL-1.3-or-later -->

# Actions Dangerous Expressions Scanner

A simple tool to find dangerous uses of GitHub Actions [Workflow expression]s.

## Usage

Run the tool from the root of a GitHub repository:

```shell
ades .
```

and it will report all detected dangerous uses of workflow expressions.

You can also use the containerized version of the CLI, for example using Docker:

```shell
docker run --rm --volume $PWD:/src docker.io/ericornelissen/ades .
```

Or you can use Go to build from source and run the CLI directly, for example using `go run`:

```shell
go run github.com/ericcornelissen/ades/cmd/ades@latest .
```

### Features

- Scan workflow files and action manifests.
- Report dangerous uses of workflow expressions in [`run:`] directives.
- Report dangerous uses of workflow expressions in [`actions/github-script`] scripts.
- _(Experimental)_ Report dangerous uses of workflow expressions in known vulnerable actions.
- Configurable sensitivity.
- Provides suggested fixes.
- Machine & human readable output formats.

### Rules

See [RULES.md].

### JSON output

The `-json` flag can be used to get the scan results in JSON format. This can be used by machines to
parse the results to process them for other purposes. The schema is defined in [`schema.json`]. The
schema is intended to be stable from one version to the next for longer periods of time.

## Background

A [workflow expression] is a string like:

```text
${{ <expression> }}
```

that may appear in a GitHub Actions workflow and is filled in at runtime. If the value is controlled
by a malicious actor it could be used to hijack the continuous integration pipeline of a repository.
GitHub [blogged about this problem] in August of 2023.

## Philosophy

This project aims to provide a tool aimed at helping developers avoid the problem of injection
through expressions altogether. Instead of reporting on problematic uses of workflow expressions,
`ades` reports on all potentially dangerous uses of workflow expressions, nudging developers to use
safe alternatives from the get-go.

The motivation behind this is twofold. First, it makes the tool much simpler and faster. Second, it
acknowledges that software development is dynamic and making changes some time after a piece of code
was introduced can be harder when compared to when the code is being written - thus reporting when a
dangerous expression is introduces simplifies the mitigation process.

## Related Work

### [ARGUS: A Framework for Staged Static Taint Analysis of GitHub Workflows and Actions]

A research tool aimed at finding problematic expression in GitHub Action Workflows and Actions,
similar to `ades`.

Performs taint analysis tracking known problematic expressions across workflows, steps, and jobs and
into and out of JavaScript Actions. It only takes into account known problematic expressions that
use `github` context values. Because of the taint analysis it will report fewer expressions than
`ades` (fewer _false positives_), but it might also miss some problematic expressions (more _false
negatives_).

It may find problematic expressions as a result of a (unknown) vulnerability in an Action, which
`ades` won't report because it is considered out of scope (and arguably better left for dedicated
tooling).

Lastly, a seemingly unrelated change in a workflow might change the result of the taint analysis and
result in a new warning, thus requiring constant usage of ARGUS, which is relatively expensive.

### [Automatic Security Assessment of GitHub Actions Workflows]

A research tool aimed at finding misconfigurations in GitHub Action Workflows (not Actions). It
includes looking for problematic expression in `run:` scripts, which is also covered by `ades`.

When it reports on problematic expression in `run:` scripts it only considers known problematic
expression that use `github` context values. Because it considers fewer expressions problematic it
will report fewer expressions overall (fewer _false positives_), but it might also miss other
problematic expressions in `run:` scripts and will completely miss others, for example in
`actions/github-script` scripts, when compared to `ades` (more _false positives_).

### [CycodeLabs/raven]

An open source tool developed by a commercial company. It aims to find misconfigurations in GitHub
Actions Workflows (not Actions). Among other checks it looks for a couple known problematic uses of
expressions involving the `github` context. As a result it will report fewer expressions overall
(fewer _false positives_) but miss many more compared to `ades` (more _false positives_).

### Other

There's other work being done in the scope of securing GitHub Actions Workflows and Actions that do
not focus on expression but are still worth mentioning:

- [`aeisenberg/codeql-actions-queries` (CodeQL queries for GitHub Actions)]
- [A Preliminary Study of GitHub Actions Dependencies]
- [Characterizing the Security of Github CI Workflows]
- [On the outdatedness of workflows in the GitHub Actions ecosystem]
- [StepSecurity]

## License

The software is available under the `GPL-3.0-or-later` license, see [COPYING.txt] for the full
license text. The documentation is available under the `GFDL-1.3-or-later` license, see [GNU Free
Documentation License v1.3] for the full license text.

[`actions/github-script`]: https://github.com/actions/github-script
[`aeisenberg/codeql-actions-queries` (codeql queries for github actions)]: https://github.com/aeisenberg/codeql-actions-queries
[`run:`]: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsrun
[`schema.json`]: ./schema.json
[a preliminary study of github actions dependencies]: https://ceur-ws.org/Vol-3483/paper7.pdf
[argus: a framework for staged static taint analysis of github workflows and actions]: https://www.usenix.org/conference/usenixsecurity23/presentation/muralee
[automatic security assessment of github actions workflows]: https://dl.acm.org/doi/abs/10.1145/3560835.3564554
[blogged about this problem]: https://github.blog/2023-08-09-four-tips-to-keep-your-github-actions-workflows-secure/#1-dont-use-syntax-in-the-run-section-to-avoid-unexpected-substitution-behavior
[characterizing the security of github ci workflows]: https://www.usenix.org/conference/usenixsecurity22/presentation/koishybayev
[copying.txt]: ./COPYING.txt
[cycodelabs/raven]: https://github.com/CycodeLabs/raven
[gnu free documentation license v1.3]: https://www.gnu.org/licenses/fdl-1.3.en.html
[on the outdatedness of workflows in the github actions ecosystem]: https://www.sciencedirect.com/science/article/pii/S0164121223002224
[rules.md]: ./RULES.md
[stepsecurity]: https://www.stepsecurity.io/
[workflow expression]: https://docs.github.com/en/actions/learn-github-actions/expressions
