<!-- SPDX-License-Identifier: GFDL-1.3-or-later -->

# Actions Dangerous Expressions Scanner

A simple tool to find dangerous uses of [GitHub Actions Expression]s.

Expressions in GitHub Actions, e.g. `${{ <expression> }}`, may appear in a GitHub Actions workflow
or manifest and are filled in at runtime. If the value is controlled by an attacker it could be used
to hijack the continuous integration pipeline of a repository. A more detailed description of the
problem is given by GitHub in "[Understanding the risk of script injections]".

`ades` helps you **find** and **resolve** dangerous uses of GitHub Actions Expressions in workflows
and manifests.

[github actions expression]: https://docs.github.com/en/actions/learn-github-actions/expressions
[understanding the risk of script injections]: https://docs.github.com/en/actions/security-for-github-actions/security-guides/security-hardening-for-github-actions#understanding-the-risk-of-script-injections

## Overview

- [Getting Started](#getting-started)
  - [Installation](#installation)
    - [Binary](#binary)
    - [Docker / Podman](#docker--podman)
    - [Go](#go)
  - [Usage](#usage)
- [Features](#features)
  - [Rules](#rules)
  - [JSON Output](#json-output)
- [Philosophy](#philosophy)
- [Related Work](#related-work)
- [License](#license)

## Getting Started

### Installation

#### Binary

Download the binary for your platform manually from the [latest release] or using the CLI, for
example using the [`gh` CLI]:

```shell
gh release download --repo ericcornelissen/ades --pattern ades_linux_amd64.tar.gz
```

Validate the provenance of the release you downloaded:

```shell
gh attestation verify --owner ericcornelissen ades_linux_amd64.tar.gz
```

Unpack the archive to get the binary out:

```shell
tar -xf ades_linux_amd64.tar.gz
```

Then add it to your `PATH` and run it:

```shell
ades -version
```

Or, without adding it to your `PATH`:

```shell
./ades -version
```

[`gh` cli]: https://cli.github.com/
[latest release]: https://github.com/ericcornelissen/ades/releases

#### Docker / Podman

Install the `ades` container by pulling it:

```shell
docker pull docker.io/ericornelissen/ades:latest
```

Validate the provenance of the container using [cosign]:

```shell
cosign verify \
  --certificate-identity-regexp 'https://github.com/ericcornelissen/ades/.+' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  docker.io/ericornelissen/ades:latest
```

Then run it using:

```shell
docker run --rm --volume $PWD:/src docker.io/ericornelissen/ades -version
```

You can set up an alias for convenience:

```shell
alias ades='docker run --rm --volume $PWD:/src docker.io/ericornelissen/ades'
```

> **NOTE:** To use [Podman] instead of [Docker] you can replace `docker` by `podman`.

[cosign]: https://github.com/sigstore/cosign
[docker]: https://www.docker.com/
[podman]: https://podman.io/

#### Go

Fetch and run `ades` from source using the [Go] CLI:

```shell
go run github.com/ericcornelissen/ades/cmd/ades@latest -version
```

Or integrate it into a Go project as a tool (after which it can be run without `@latest`):

```shell
go get -tool github.com/ericcornelissen/ades
```

[go]: https://go.dev/

### Usage

Run `ades` from the root of a GitHub repository and it will report all dangerous uses of GitHub
Actions Expressions for the project:

```shell
ades
```

Alternatively, specify any number of projects to scan, and it well report for each:

```shell
ades project-a project-b
```

If you need more information, ask for help:

```shell
ades -help
```

## Features

- Scans workflow files and action manifests.
- Reports dangerous uses of expressions in [`run:`] directives, [`actions/github-script`] scripts,
  and known problematic action inputs.
- Report dangerous uses of expressions in known vulnerable actions.
- Provides suggested fixes and _(experimental)_ fully automated fixes.
- Configurable sensitivity.
- Machine & human readable output formats.

[`actions/github-script`]: https://github.com/actions/github-script
[`run:`]: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsrun

### Rules

See [RULES.md].

[rules.md]: ./RULES.md

### JSON output

The `-json` flag can be used to get the scan results in JSON format. This can be used by machines to
parse the results to process them for other purposes. The schema is defined in [`schema.json`] and
it is intended to be stable from one version to the next for longer periods of time.

[`schema.json`]: ./schema.json

## Philosophy

This project aims to provide a tool aimed at helping developers avoid the problem of injection
through expressions altogether. Instead of reporting on known problematic uses of expressions,
`ades` reports on all potentially dangerous uses of expressions, nudging developers to use safe
alternatives from the get-go.

The motivation behind this is twofold. First, it makes the tool much simpler and faster. Second, it
acknowledges that software development is a dynamic process and that future changes can make an
expression that is safe today unsafe. Moreover, fixing a workflow while creating it is easier now
than it is later.

## Related Work

### [ARGUS: A Framework for Staged Static Taint Analysis of GitHub Workflows and Actions]

A research tool aimed at finding problematic expression in GitHub Action Workflows and Actions. It
performs taint analysis to track known problematic expressions across workflows, steps, and jobs and
into and out of JavaScript Actions. Because of the taint analysis it will report fewer expressions
than `ades` (fewer _false positives_), but it might also miss some problematic expressions (more
_false negatives_).

### [Automatic Security Assessment of GitHub Actions Workflows]

A research tool aimed at finding misconfigurations in GitHub Action Workflows (not Actions). It
includes looking for problematic expression in `run:` scripts. It only reports on the use of known
problematic expression in `run:` scripts. Because it considers fewer expressions problematic it will
report fewer expressions overall (fewer _false positives_), but it might also miss other problematic
expressions in `run:` scripts and will completely miss others, for example expressions in
`actions/github-script` scripts.

### [Ghast]

A tool to find misconfigurations in GitHub Actions Workflows (not Actions). Among other checks it
looks for a couple known problematic uses of expressions involving the `github` context. It also
steers users away from using inline scripts, recommending local Actions instead. As a result it will
report fewer expressions overall (fewer _false positives_) but miss some (more _false negatives_).

### [`poutine`]

A tool that aims to find misconfigurations in CI/CD pipeline configurations including GitHub Actions
Workflows. Among other checks it looks for a couple known problematic uses of expressions involving
the `github` context. As a result it will report fewer expressions overall (fewer _false positives_)
but also miss some (more _false negatives_).

### [Raven]

A tool aimed at finding misconfigurations in GitHub Actions Workflows (not Actions). Among other
checks it looks for a couple known problematic uses of expressions involving the `github` context.
As a result it will report fewer expressions overall (fewer _false positives_) but miss some (more
_false negatives_).

### [`zizmor`]

A tool that aims to find security issues in GitHub Actions CI/CD setups. It reports various kinds of
potential security problems including dangerous uses of expressions ("template injection"). Similar
to `ades`, it will report on most uses of expressions but only in `run:` and `actions/github-script`
scripts except for a small allowlist of known safe expressions. It does distinguish between
expressions known to be attacker controlled and only potentially attacker controlled with different
severities.

[argus: a framework for staged static taint analysis of github workflows and actions]: https://www.usenix.org/conference/usenixsecurity23/presentation/muralee
[automatic security assessment of github actions workflows]: https://dl.acm.org/doi/abs/10.1145/3560835.3564554
[ghast]: https://github.com/bin3xish477/ghast
[`poutine`]: https://github.com/boostsecurityio/poutine
[raven]: https://github.com/CycodeLabs/raven
[`zizmor`]: https://github.com/woodruffw/zizmor

### Others

There is other work being done in the scope of GitHub Actions security that does not focus on GitHub
Actions Expression but is still worth mentioning:

#### Tooling

- [`actionlint`]: General purpose linter for GitHub Actions users.
- [`aeisenberg/codeql-actions-queries`]: A CodeQl query pack for writing reusable GitHub Actions.
- [CodeQL support for GitHub Actions]: CodeQL queries for GitHub Actions workflows.
- [StepSecurity]: Runtime protection for GitHub Action users.

[`actionlint`]: https://github.com/rhysd/actionlint
[`aeisenberg/codeql-actions-queries`]: https://github.com/aeisenberg/codeql-actions-queries
[codeql support for github actions]: https://docs.github.com/en/code-security/code-scanning/managing-your-code-scanning-configuration/actions-built-in-queries
[stepsecurity]: https://www.stepsecurity.io/

#### Research

- [Ambush From All Sides: Understanding Security Threats in Open-Source Software CI/CD Pipelines]
- [A Preliminary Study of GitHub Actions Dependencies]
- [Catching Smells in the Act: A GitHub Actions Workflow Investigation]
- [Characterizing the Security of Github CI Workflows]
- [Continuous Intrusion: Characterizing the Security of Continuous Integration Services]
- [GitHub Actions Attack Diagram]
- [Living Off the Pipeline]
- [Mitigating Security Issues in GitHub Actions]
- [On the outdatedness of workflows in the GitHub Actions ecosystem]
- [Quantifying Security Issues in Reusable JavaScript Actions in GitHub Workflows]

[ambush from all sides: understanding security threats in open-source software ci/cd pipelines]: https://ieeexplore.ieee.org/abstract/document/10061526
[a preliminary study of github actions dependencies]: https://ceur-ws.org/Vol-3483/paper7.pdf
[catching smells in the act: a github actions workflow investigation]: https://ieeexplore.ieee.org/abstract/document/10795325
[characterizing the security of github ci workflows]: https://www.usenix.org/conference/usenixsecurity22/presentation/koishybayev
[continuous intrusion: characterizing the security of continuous integration services]: https://ieeexplore.ieee.org/abstract/document/10179471
[github actions attack diagram]: https://github.com/jstawinski/GitHub-Actions-Attack-Diagram
[living off the pipeline]: https://boostsecurityio.github.io/lotp/#github-actions
[mitigating security issues in gitHub actions]: https://dl.acm.org/doi/abs/10.1145/3643662.3643961
[on the outdatedness of workflows in the github actions ecosystem]: https://www.sciencedirect.com/science/article/pii/S0164121223002224
[quantifying security issues in reusable javascript actions in github workflows]: https://dl.acm.org/doi/abs/10.1145/3643991.3644899

## License

The software is available under the `GPL-3.0-or-later` license, see [COPYING.txt] for the full
license text. The documentation is available under the `GFDL-1.3-or-later` license, see [GNU Free
Documentation License v1.3] for the full license text.

[copying.txt]: ./COPYING.txt
[gnu free documentation license v1.3]: https://www.gnu.org/licenses/fdl-1.3.en.html
