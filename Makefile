# MIT No Attribution
#
# Copyright (c) 2024 Eric Cornelissen
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

CONTAINER_ENGINE?=docker
CONTAINER_TAG?=latest

.PHONY: default
default:
	@printf "Usage: make <command>\n\n"
	@printf "Commands:\n"
	@awk -F ':(.*)## ' '/^[a-zA-Z0-9%\\\/_.-]+:(.*)##/ { \
		printf "  \033[36m%-30s\033[0m %s\n", $$1, $$NF \
	}' $(MAKEFILE_LIST)

.PHONY: audit
audit: ## Audit for vulnerabilities
	@echo 'Checking vulnerabilities...'
	@go run golang.org/x/vuln/cmd/govulncheck .

.PHONY: build
build: ## Build the ades binary for the current platform
	@echo 'Building...'
	@go build .

.PHONY: clean
clean: ## Reset the project to a clean state
	@echo 'Cleaning...'
	@git clean -fx \
		_compiled/ \
		ades \
		cover.*

.PHONY: compliance
compliance: ## Check license compliance
	@echo 'Checking license compliance...'
	@go run github.com/google/go-licenses check \
		--allowed_licenses BSD-3-Clause,GPL-3.0,MIT \
		.

.PHONY: container
container: ## Build the ades container for the current platform
	@$(CONTAINER_ENGINE) build \
		--file Containerfile \
		--tag ericornelissen/ades:$(CONTAINER_TAG) \
		.

.PHONY: coverage
coverage: ## Run all tests and generate a coverage report
	@echo 'Testing...'
	@go test -coverprofile cover.out .
	@echo 'Generating coverage report...'
	@go tool cover -html cover.out -o cover.html

.PHONY: dev-env dev-img
dev-env: dev-img ## Run an ephemeral development environment container
	@$(CONTAINER_ENGINE) run -it \
		--rm \
		--workdir '/ades' \
		--mount "type=bind,source=$(shell pwd),target=/ades" \
		--name 'ades-dev-env' \
		'ades-dev-img'

dev-img: ## Build a development environment container image
	@$(CONTAINER_ENGINE) build \
		--file 'Containerfile.dev' \
		--tag 'ades-dev-img' \
		.

.PHONY: fmt fmt-check
fmt: ## Format the source code
	@echo 'Formatting...'
	@gofmt -w .
	@gofmt -w -r 'interface{} -> any' .
	@go mod tidy
	@go run golang.org/x/tools/cmd/goimports -w .

fmt-check: ## Check the source code formatting
	@echo 'Checking formatting...'
	@test -z "$$(gofmt -l .)"
	@test -z "$$(go run golang.org/x/tools/cmd/goimports -l .)"

.PHONY: release release-compile
release:
	@echo 'On main and not dirty?'
	@test "$$(git branch --show-current)" = 'main'
	@test "$$(git status --porcelain)" = ''

	@echo 'Is main up-to-date?'
	@git fetch
	@test "$$(git rev-parse HEAD)" = "$$(git rev-parse FETCH_HEAD)"

	@echo 'Preparing for version bump...'
	@sed -i main.go -e "s/versionString := \"v[0-9][0-9][.][0-9][0-9]\"/versionString := \"v$$(date '+%y.%m')\"/"
	@sed -i test/flags-info.txtar -e "s/stdout 'v[0-9][0-9][.][0-9][0-9]'/stdout 'v$$(date '+%y.%m')'/"

	@echo 'Committing version bump...'
	@git checkout -b version-bump
	@git add 'main.go' 'test/flags-info.txtar'
	@git commit --signoff --message 'version bump'

	@echo 'Pushing version-bump branch...'
	@git push origin version-bump

	@echo ''
	@echo 'Next, you should open a Pull Request to merge the branch version-bump into main and'
	@echo 'merge it if all checks succeeds. After merging run:'
	@echo ''
	@echo '    git checkout main'
	@echo '    git pull origin main'
	@echo "    git tag v$$(date '+%y.%m')"
	@echo "    git push origin v$$(date '+%y.%m')"
	@echo ''
	@echo 'After that a release should be created automatically. If not, follow the instructions in'
	@echo 'RELEASE.md.'

release-compile:
	@mkdir _compiled/

	@echo 'Compiling for darwin/amd64...'
	@env GOOS=darwin GOARCH=amd64 go build -o 'ades'
	@tar -czf 'ades_darwin_amd64.tar.gz' 'ades'
	@mv 'ades_darwin_amd64.tar.gz' '_compiled/'

	@echo 'Compiling for darwin/arm64...'
	@env GOOS=darwin GOARCH=arm64 go build -o 'ades'
	@tar -czf 'ades_darwin_arm64.tar.gz' 'ades'
	@mv 'ades_darwin_arm64.tar.gz' '_compiled/'

	@echo 'Compiling for linux/386...'
	@env GOOS=linux GOARCH=386 go build -o 'ades'
	@tar -czf 'ades_linux_386.tar.gz' 'ades'
	@mv 'ades_linux_386.tar.gz' '_compiled/'

	@echo 'Compiling for linux/amd64...'
	@env GOOS=linux GOARCH=amd64 go build -o 'ades'
	@tar -czf 'ades_linux_amd64.tar.gz' 'ades'
	@mv 'ades_linux_amd64.tar.gz' '_compiled/'

	@echo 'Compiling for linux/arm...'
	@env GOOS=linux GOARCH=arm go build -o 'ades'
	@tar -czf 'ades_linux_arm.tar.gz' 'ades'
	@mv 'ades_linux_arm.tar.gz' '_compiled/'

	@echo 'Compiling for linux/arm64...'
	@env GOOS=linux GOARCH=arm64 go build -o 'ades'
	@tar -czf 'ades_linux_arm64.tar.gz' 'ades'
	@mv 'ades_linux_arm64.tar.gz' '_compiled/'

	@echo 'Compiling for windows/386...'
	@env GOOS=windows GOARCH=386 go build -o 'ades'
	@mv 'ades' 'ades.exe'
	@zip -9q 'ades_windows_386.zip' 'ades.exe'
	@mv 'ades_windows_386.zip' '_compiled/'

	@echo 'Compiling for windows/amd64...'
	@env GOOS=windows GOARCH=amd64 go build -o 'ades'
	@mv 'ades' 'ades.exe'
	@zip -9q 'ades_windows_amd64.zip' 'ades.exe'
	@mv 'ades_windows_amd64.zip' '_compiled/'

	@echo 'Compiling for windows/arm...'
	@env GOOS=windows GOARCH=arm go build -o 'ades'
	@mv 'ades' 'ades.exe'
	@zip -9q 'ades_windows_arm.zip' 'ades.exe'
	@mv 'ades_windows_arm.zip' '_compiled/'

	@echo 'Compiling for windows/arm64...'
	@env GOOS=windows GOARCH=arm64 go build -o 'ades'
	@mv 'ades' 'ades.exe'
	@zip -9q 'ades_windows_arm64.zip' 'ades.exe'
	@mv 'ades_windows_arm64.zip' '_compiled/'

	@echo 'Computing checksums...'
	@cd _compiled/ && shasum --algorithm 512 ./* >'checksums-sha512.txt'

.PHONY: run
run: ## Run the project on itself
	@go run .

.PHONY: test
test: ## Run all tests
	@echo 'Testing...'
	@go test .

.PHONY: test-mutation
test-mutation: ## Run mutation tests
	@echo 'Mutation testing...'
	@go test -tags=mutation

.PHONY: test-randomized
test-randomized: ## Run tests in a random order
	@echo 'Testing (random order)...'
	@go test -shuffle=on .

.PHONY: vet
vet: ## Vet the source code
	@echo 'Vetting...'
	@go vet .
	@go run 4d63.com/gochecknoinits .
	@go run github.com/alexkohler/dogsled/cmd/dogsled -set_exit_status .
	@go run github.com/alexkohler/nakedret/v2/cmd/nakedret -l 0 .
	@go run github.com/alexkohler/prealloc -set_exit_status .
	@go run github.com/alexkohler/unimport .
	@go run github.com/butuzov/ireturn/cmd/ireturn .
	@go run github.com/catenacyber/perfsprint .
	@go run github.com/dkorunic/betteralign/cmd/betteralign .
	@go run github.com/go-critic/go-critic/cmd/gocritic check .
	@go run github.com/gordonklaus/ineffassign .
	@go run github.com/jgautheron/goconst/cmd/goconst -set-exit-status .
	@go run github.com/kisielk/errcheck .
	@go run github.com/kunwardeep/paralleltest -i .
	@go run github.com/kyoh86/looppointer/cmd/looppointer .
	@go run github.com/mdempsky/unconvert .
	@go run github.com/nishanths/exhaustive/cmd/exhaustive .
	@go run github.com/polyfloyd/go-errorlint .
	@go run github.com/remyoudompheng/go-misc/deadcode .
	@go run github.com/tetafro/godot/cmd/godot .
	@go run github.com/tomarrell/wrapcheck/v2/cmd/wrapcheck .
	@go run github.com/ultraware/whitespace/cmd/whitespace ./...
	@go run go.uber.org/nilaway/cmd/nilaway .
	@go run golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow .
	@go run honnef.co/go/tools/cmd/staticcheck .
	@go run mvdan.cc/unparam .

.PHONY: verify
verify: build compliance fmt-check test run vet ## Verify project is in a good state
