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

fmt-check: ## Check the source code formatting
	@echo 'Checking formatting...'
	@gofmt -l .

.PHONY: run
run: ## Run the project on itself
	@go run .

.PHONY: test
test: ## Run all tests
	@echo 'Testing...'
	@go test .

.PHONY: vet
vet: ## Vet the source code
	@echo 'Vetting...'
	@go vet .

.PHONY: verify
verify: build fmt-check test run vet ## Verify project is in a good state
