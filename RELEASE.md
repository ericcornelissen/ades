<!-- SPDX-License-Identifier: CC0-1.0 -->

# Release Guidelines

To release a new version of the `ades` project follow the description in this file.

## Preferred

Run `make release` and follow the instructions it gives.

## Fallback

The following is a full step-by-step walkthrough of how to create a release for `ades` (using v23.12
as an example):

1. Make sure that your local copy of the repository is up-to-date, sync:

   ```shell
   git checkout main
   git pull origin main
   ```

   Or clone:

   ```shell
   git clone git@github.com:ericcornelissen/ades.git
   ```

1. Update the version number following to the current year-month pair in the `version` function in
   `main.go` (and update the tests correspondingly):

   ```diff
     func version() {
   -   versionString := "v23.11"
   +   versionString := "v23.12"
       fmt.Println(versionString)
     }
   ```

   Single-digit months should be prefixed with a `0` (for example for January `24.01`).

1. Commit the changes to a new branch and push using:

   ```shell
   git checkout -b version-bump
   git add 'main.go' 'test/flags-info.txtar'
   git commit --signoff --message 'version bump'
   git push origin version-bump
   ```

1. Create a Pull Request to merge the new branch into `main`.

1. Merge the Pull Request if the changes look OK and all continuous integration checks are passing.

1. Immediately after the Pull Request is merged, sync the `main` branch:

   ```shell
   git checkout main
   git pull origin main
   ```

1. Create a [git tag] for the new version and push it:

   ```shell
   git tag v23.12
   git push origin v23.12
   ```

   > **Note** At this point, the continuous delivery automation may pick up and complete the release
   > process. If not, or only partially, continue following the remaining steps.

1. Create a [GitHub Release] for the [git tag] of the new release. The release title should be
   "Release {_version_}" (e.g. "Release v23.12"). The release text should be "{_version_}" (e.g.
   "v23.12"). The release artifact should follow the previous release as closely as possible.

1. Publish to [Docker Hub], first with a version tag:

   ```shell
   make container CONTAINER_TAG=v23.12
   docker push ericornelissen/ades:v23.12
   ```

   then the `latest` tag:

   ```shell
   make container CONTAINER_TAG=latest
   docker push ericornelissen/ades:latest
   ```

[docker hub]: https://hub.docker.com/
[git tag]: https://git-scm.com/book/en/v2/Git-Basics-Tagging
[github release]: https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository
