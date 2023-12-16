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
build: ## Build a binary for the current platform
	@echo 'Building...'
	@go build .

.PHONY: clean
clean: ## Reset the project to a clean state
	@echo 'Cleaning...'
	@git clean -fx \
		ades \
		cover.*

.PHONY: container
container:
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
	@go test -v -tags=mutation

.PHONY: vet
vet: ## Vet the source code
	@echo 'Vetting...'
	@go vet .
	@go run 4d63.com/gochecknoinits .
	@go run github.com/alexkohler/dogsled/cmd/dogsled .
	@go run github.com/alexkohler/nakedret/v2/cmd/nakedret .
	@go run github.com/alexkohler/prealloc .
	@go run github.com/alexkohler/unimport .
	@go run github.com/go-critic/go-critic/cmd/gocritic check .
	@go run github.com/gordonklaus/ineffassign .
	@go run github.com/jgautheron/goconst/cmd/goconst .
	@go run github.com/kisielk/errcheck .
	@go run github.com/kyoh86/looppointer/cmd/looppointer .
	@go run github.com/mdempsky/unconvert .
	@go run github.com/nishanths/exhaustive/cmd/exhaustive .
	@go run github.com/polyfloyd/go-errorlint .
	@go run github.com/remyoudompheng/go-misc/deadcode .
	@go run github.com/tetafro/godot/cmd/godot .
	@go run github.com/tomarrell/wrapcheck/v2/cmd/wrapcheck .
	@go run golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow .
	@go run go.uber.org/nilaway/cmd/nilaway .
	@go run honnef.co/go/tools/cmd/staticcheck .
	@go run mvdan.cc/unparam .

.PHONY: verify
verify: build fmt-check test run vet ## Verify project is in a good state
