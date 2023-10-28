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
	@govulncheck .

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

.PHONY: coverage
coverage: ## Run all tests and generate a coverage report
	@echo 'Testing...'
	@go test -coverprofile cover.out .
	@echo 'Generating coverage report...'
	@go tool cover -html cover.out -o cover.html

.PHONY: fmt fmt-check
fmt: ## Format the source code
	@echo 'Formatting...'
	@go fmt .
	@go mod tidy
	@go run golang.org/x/tools/cmd/goimports@v0.14.0 -w .

fmt-check: ## Check the source code formatting
	@echo 'Checking formatting...'
	@test -z "$$(gofmt -l .)"
	@test -z "$$(go run golang.org/x/tools/cmd/goimports@v0.14.0 -l .)"

.PHONY: run
run: ## Run the project on itself
	@go run .

.PHONY: test
test: ## Run all tests
	@echo 'Testing...'
	@go test .
	@echo 'Validating JSON schema...'
	@go run github.com/santhosh-tekuri/jsonschema/cmd/jv@f2cc8ae -assertformat schema.json

.PHONY: test-mutation
test-mutation: ## Run mutation tests
	@echo 'Mutation testing...'
	@go test -v -tags=mutation

.PHONY: vet
vet: ## Vet the source code
	@echo 'Vetting...'
	@go vet .
	@go run 4d63.com/gochecknoinits@25bb07f .
	@go run github.com/alexkohler/dogsled/cmd/dogsled@34d2ab9 .
	@go run github.com/alexkohler/nakedret/v2/cmd/nakedret@v2.0.1 .
	@go run github.com/alexkohler/prealloc@v1.0.0 .
	@go run github.com/alexkohler/unimport@e6f2b2e .
	@go run github.com/go-critic/go-critic/cmd/gocritic@v0.9.0 check .
	@go run github.com/gordonklaus/ineffassign@0e73809 .
	@go run github.com/jgautheron/goconst/cmd/goconst@v1.6.0 .
	@go run github.com/kisielk/errcheck@v1.6.3 .
	@go run github.com/kyoh86/looppointer/cmd/looppointer@v0.2.1 .
	@go run github.com/mdempsky/unconvert@4157069 .
	@go run github.com/nishanths/exhaustive/cmd/exhaustive@v0.11.0 .
	@go run github.com/polyfloyd/go-errorlint@v1.4.5 .
	@go run github.com/remyoudompheng/go-misc/deadcode@2d6ac65 .
	@go run github.com/tetafro/godot/cmd/godot@v1.4.15 .
	@go run github.com/tomarrell/wrapcheck/v2/cmd/wrapcheck@v2.8.1 .
	@go run golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@2f9d82f .
	@go run honnef.co/go/tools/cmd/staticcheck@v0.4.6 .
	@go run mvdan.cc/unparam@3ee2d22 .

.PHONY: verify
verify: build fmt-check test run vet ## Verify project is in a good state
