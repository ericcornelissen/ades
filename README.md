# Actions Dangerous Expressions Scanner

A simple tool to find dangerous uses of GitHub Actions [Workflow expression]s.

## Usage

Run the tool from the root of a GitHub repository:

```shell
ades .
```

and it will report all detected dangerous uses of workflow expressions.

### Features

- Scan workflow files and action manifests.
- Report dangerous uses of workflow expressions in [`run:`] directives.
- Report dangerous uses of workflow expressions in [`actions/github-script`] scripts.

## Background

A [workflow expression] is a string like:

```text
${{ <expression> }}
```

that may appear in a GitHub Actions workflow and is filled in at runtime. If the value is controlled
by a malicious actor it could be used to hijack the continuous integration pipeline of a repository.
GitHub [blogged about this problem] in August of 2023, and the 2023 publication [ARGUS: A Framework
for Staged Static Taint Analysis of GitHub Workflows and Actions] analyzes the problem in depth
using advanced methods.

This project aims to provide a far simpler tool aimed at helping developers avoid the problem
altogether. Instead of reporting on problematic uses of workflow expressions, `ades` reports on all
potentially dangerous uses of workflow expressions, nudging developers to use safe alternatives
from the get-go.

The motivation behind this is twofold. First, it makes the tool much simpler and faster. Second, it
acknowledges that software development is dynamic and making changes some time after introduction
can be difficult - (guaranteed) reporting the violations when the code is being written simplifies
the mitigation process.

## License

The software is available under the `GPL-3.0-or-later` license, see [COPYING.txt] for the full
license text. The documentation is available under the `GFDL-1.3-or-later` license, see [GNU Free
Documentation License v1.3] for the full license text.

[`actions/github-script`]: https://github.com/actions/github-script
[`run:`]: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idstepsrun
[argus: a framework for staged static taint analysis of github workflows and actions]:https://www.usenix.org/conference/usenixsecurity23/presentation/muralee
[blogged about this problem]: https://github.blog/2023-08-09-four-tips-to-keep-your-github-actions-workflows-secure/#1-dont-use-syntax-in-the-run-section-to-avoid-unexpected-substitution-behavior
[copying.txt]: ./COPYING.txt
[gnu free documentation license v1.3]: https://www.gnu.org/licenses/fdl-1.3.en.html
[workflow expression]: https://docs.github.com/en/actions/learn-github-actions/expressions
