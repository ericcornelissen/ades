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

.PHONY: default
default:
	@printf "Usage: make <command>\n\n"
	@printf "Commands:\n"
	@awk -F ':(.*)## ' '/^[a-zA-Z0-9%\\\/_.-]+:(.*)##/ { \
		printf "  \033[36m%-30s\033[0m %s\n", $$1, $$NF \
	}' $(MAKEFILE_LIST)

.PHONY: build
build: wasm_exec.js ## Build the webapp
	@echo 'Building...'
	@GOOS=js GOARCH=wasm go build \
		-o ades.wasm
	@cp ../COPYING.txt ./COPYING.txt

.PHONY: clean
clean: ## Clean the webapp directory
	@echo 'Cleaning...'
	@git clean -fx \
		node_modules/ \
		*.wasm \
		COPYING.txt \
		wasm_exec.js

.PHONY: serve
serve: build node_modules ## Serve the webapp locally
	@echo 'Serving locally...'
	@npx http-server . \
		--port 8080

node_modules: .npmrc package*.json
	@npm clean-install

wasm_exec.js:
	@cp "$(shell go env GOROOT)/misc/wasm/wasm_exec.js" ./wasm_exec.js
