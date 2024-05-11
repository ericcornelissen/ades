<!-- SPDX-License-Identifier: CC0-1.0 -->

# Contributing Guidelines

The maintainers of `ades` welcome contributions and corrections. This includes improvements to the
documentation or code base, tests, bug fixes, and implementations of new features. We recommend you
open an issue before making any substantial changes so you can be sure your work won't be rejected.
But for small changes, such as fixing a typo, you can open a Pull Request directly.

If you decide to make a contribution, please read the [DCO] and use the following workflow:

- Fork the repository.
- Create a new branch from the latest `main`.
- Make your changes on the new branch.
- Commit, with [signoff], to the new branch and push the commit(s).
- Open a Pull Request against `main`.

[dco]: ./DCO.txt
[signoff]: https://git-scm.com/docs/git-commit#Documentation/git-commit.txt---signoff

---

## Tasks

This project uses a custom Go-based task runner to run common tasks. To get started run:

```shell
go run tasks.go
```

For example, you can run one task as:

```shell
go run tasks.go verify
```

We recommend configuring the following command alias:

```shell
alias gask='go run tasks.go'
```

Which would allow you to run:

```shell
gask verify
```

---

## Adding a Rule

To add a rule you need to add some information and logic to the `rules.go` file, and corresponding
tests to the `rules_test.go` file. The details depend on whether it's a rule for actions (i.e. steps
with `uses:`) or for other steps, but defining the rule is the same regardless.

To define a rule you need to create an instance of the `rule` type. This involves giving the rule an
id, title, and description as well as a function to extract what needs to be analyzed and a function
that builds a suggestion for fixing a violation. The id, title, and description are simple text
values. The extraction function needs to return a string to be analyzed for a given `JobStep`. The
suggestion functions needs to return a suggestion string for a given `Violation`. Lastly, you need
to add the rule to the `RULES.md` documentation file in the same format as the `--explain` output.

Note that if multiple things could be checked for one action or step construct, they should be
defined as separate rules.

### Action Rules

If the rule is for an action you need to specify when the rule applies, usually in terms of the ref
that is used. Defining this is part of the `actionRule` type. You also need to add the action (if it
isn't already) and rule to the `actionRules` map.

### Step Rules

If the rule is for other steps you need to specify when the rule applies. Defining this is part of
the `stepRule` type. You also need to add the rule to the `stepRules` slice.

### Testing

Every new rule needs to be tested. The rule id, title, and description are tested automatically. The
`appliesTo`, `extractFrom`, and `suggestion` functions require dedicated unit tests. For this, it is
recommended to follow the lead of the tests for existing rules. Additionally, a test case should be
added to the `test/rules.txtar` test file.
