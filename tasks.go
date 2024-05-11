// MIT No Attribution
//
// Copyright (c) 2024 Eric Cornelissen
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

//go:build tasks

package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	DEFAULT_CONTAINER_ENGINE = "docker"
	ENV_CONTAINER_ENGINE     = "CONTAINER_ENGINE"
	ENV_CONTAINER_TAG        = "CONTAINER_TAG"
)

var (
	buildAllDir = "_compiled"
	webDir      = "web"
)

var (
	permFile fs.FileMode = 0o664
	permDir  fs.FileMode = 0o755
)

// Audit the codebase.
func TaskAudit(t *T) error {
	return t.Run(
		TaskAuditCapabilities,
		TaskAuditVulnerabilities,
	)
}

// Audit for capabilities.
func TaskAuditCapabilities(t *T) error {
	t.Log("Checking capabilities...")
	return t.Exec(`
		go run github.com/google/capslock/cmd/capslock
			-packages ./...
			-noisy
			-output=compare capabilities.json
	`)
}

// Audit for known vulnerabilities.
func TaskAuditVulnerabilities(t *T) error {
	t.Log("Checking vulnerabilities...")
	return t.Exec(`go run golang.org/x/vuln/cmd/govulncheck .`)
}

// Build the ades binary for the current platform.
func TaskBuild(t *T) error {
	t.Log("Building...")
	return t.Exec(`go build ./cmd/ades`)
}

// Build the ades binary for all supported platforms.
func TaskBuildAll(t *T) error {
	type Target struct {
		GOOS   string
		GOARCH string
	}

	var (
		osMac     = "darwin"
		osLinux   = "linux"
		osWindows = "windows"
		arch386   = "386"
		archAmd64 = "amd64"
		archArm   = "arm"
		archArm64 = "arm64"
	)

	var targets = []Target{
		{GOOS: osMac, GOARCH: archAmd64},
		{GOOS: osMac, GOARCH: archArm64},
		{GOOS: osLinux, GOARCH: arch386},
		{GOOS: osLinux, GOARCH: archAmd64},
		{GOOS: osLinux, GOARCH: archArm},
		{GOOS: osLinux, GOARCH: archArm64},
		{GOOS: osWindows, GOARCH: arch386},
		{GOOS: osWindows, GOARCH: archAmd64},
		{GOOS: osWindows, GOARCH: archArm},
		{GOOS: osWindows, GOARCH: archArm64},
	}

	t.Log("Building (all platforms)...")
	if err := os.RemoveAll(buildAllDir); err != nil {
		return err
	}
	if err := os.Mkdir(buildAllDir, permDir); err != nil {
		return err
	}

	archives := make([]string, len(targets))
	for i, target := range targets {
		fmt.Printf("Compiling for %s/%s...\n", target.GOOS, target.GOARCH)

		executable := "ades"
		if target.GOOS == osWindows {
			executable = "ades.exe"
		}

		archiveCmd := "tar -czf"
		if target.GOOS == osWindows {
			archiveCmd = "zip -9q"
		}

		archiveExt := "tar.gz"
		if target.GOOS == osWindows {
			archiveExt = "zip"
		}

		archiveFile := fmt.Sprintf("ades_%s_%s.%s", target.GOOS, target.GOARCH, archiveExt)
		archives[i] = archiveFile

		var (
			compile = fmt.Sprintf(
				"env GOOS=%s GOARCH=%s go build -o %s ./cmd/ades",
				target.GOOS,
				target.GOARCH,
				executable,
			)
			archive = fmt.Sprintf(
				"%s '_compiled/%s' %s",
				archiveCmd,
				archiveFile,
				executable,
			)
		)

		if err := t.Exec(compile, archive); err != nil {
			return err
		}
	}

	t.Log("Computing checksums...")
	t.Cd("_compiled")
	out, err := t.ExecS(`shasum --algorithm 512 ` + strings.Join(archives, " "))
	if err != nil {
		return err
	}

	return os.WriteFile("./_compiled/checksums-shas512.txt", []byte(out), permFile)
}

// Reset the project to a clean state.
func TaskClean(t *T) error {
	var (
		items = []string{
			"_compiled/",
			"web/node_modules/",
			"web/ades.wasm",
			"web/COPYING.txt",
			"web/wasm_exec.js",
			"ades",
			"ades.exe",
			"cover.html",
			"cover.out",
		}
		clean = "git clean -fx " + strings.Join(items, " ")
	)

	t.Log("Cleaning...")
	return t.Exec(clean)
}

// Check license compliance.
func TaskCompliance(t *T) error {
	var (
		licenses = []string{
			"BSD-3-Clause",
			"GPL-3.0",
			"MIT",
		}
		licenseCheck = fmt.Sprintf(
			"go run github.com/google/go-licenses check --allowed_licenses %s ./...",
			strings.Join(licenses, ","),
		)
	)

	t.Log("Checking license compliance...")
	return t.Exec(licenseCheck)
}

// Build the ades container for the current platform.
func TaskContainer(t *T) error {
	var (
		engine = t.Env(ENV_CONTAINER_ENGINE, DEFAULT_CONTAINER_ENGINE)
		tag    = t.Env(ENV_CONTAINER_TAG, "latest")
		build  = fmt.Sprintf(
			"%s build --file Containerfile --tag ericornelissen/ades:%s .",
			engine,
			tag,
		)
	)

	return t.Exec(build)
}

// Run all tests and generate a coverage report.
func TaskCoverage(t *T) error {
	t.Log("Generating coverage report...")
	return t.Exec(
		"go test -coverprofile cover.out ./...",
		"go tool cover -html cover.out -o cover.html",
	)
}

// Run an ephemeral development environment container.
func TaskDevEnv(t *T) error {
	wd, _ := os.Getwd()

	if err := TaskDevImg(t); err != nil {
		return err
	}

	var (
		engine = t.Env(ENV_CONTAINER_ENGINE, DEFAULT_CONTAINER_ENGINE)
		build  = fmt.Sprintf(
			"%s run -it --rm --workdir '/ades' --mount 'type=bind,source=%s,target=/ades' --name ades-dev-env ades-dev-img",
			engine,
			wd,
		)
	)

	return t.Exec(build)
}

// Build a development environment container image.
func TaskDevImg(t *T) error {
	var (
		engine = t.Env(ENV_CONTAINER_ENGINE, DEFAULT_CONTAINER_ENGINE)
		build  = fmt.Sprintf(
			"%s build --file 'Containerfile.dev' --tag ades-dev-img .",
			engine,
		)
	)

	return t.Exec(build)
}

// Run the project on itself.
func TaskDogfeed(t *T) error {
	return t.Exec(`go run ./cmd/ades`)
}

// Format the source code.
func TaskFormat(t *T) error {
	t.Log("Formatting...")
	return t.Exec(
		"gofmt -w .",
		"gofmt -w -r 'interface{} -> any' .",
		"go mod tidy",
		"go run golang.org/x/tools/cmd/goimports -w .",
	)
}

// Check the source code formatting.
func TaskFormatCheck(t *T) error {
	t.Log("Checking formatting...")

	out, err := t.ExecS(
		"gofmt -l .",
		"gofmt -l -r 'interface{} -> any' .",
		"go run golang.org/x/tools/cmd/goimports -l .",
	)
	if err != nil {
		return err
	} else if out != "" {
		return errors.New("not formatted")
	}

	return nil
}

// Initiate a new release.
func TaskRelease(t *T) error {
	var (
		baseBranch = "main"
		bumpBranch = "version-bump"
	)

	t.Log("Checking repository state...")
	if out, err := t.ExecS(`git branch --show-current`); err != nil {
		return err
	} else if out != baseBranch {
		return errors.New("not on " + baseBranch)
	}

	if out, err := t.ExecS(`git status --porcelain`); err != nil {
		return err
	} else if out != "" {
		return errors.New("workspace is dirty")
	}

	if _, err := t.ExecS(`git fetch`); err != nil {
		return err
	}

	head, err := t.ExecS(`git rev-parse HEAD`)
	if err != nil {
		return err
	}

	fetchHead, err := t.ExecS(`git rev-parse FETCH_HEAD`)
	if err != nil {
		return err
	} else if head != fetchHead {
		return errors.New("branch is not up-to-date")
	}

	t.Log("Preparing for version bump...")
	date, err := t.ExecS(`date '+%y.%m'`)
	if err != nil {
		return err
	}

	err = t.Exec(
		`sed -i cmd/ades/main.go -e 's/versionString := "v[0-9][0-9][.][0-9][0-9]"/versionString := "v`+date+`"/'`,
		`sed -i test/flags-info.txtar -e "s/stdout 'v[0-9][0-9][.][0-9][0-9]'/stdout 'v`+date+`'/"`,
	)
	if err != nil {
		return err
	}

	t.Log("Committing and pushing version bump...")
	err = t.Exec(
		`git checkout -b `+bumpBranch,
		`git add 'cmd/ades/main.go' 'test/flags-info.txtar'`,
		`git commit --signoff --message 'version bump'`,
		`git push origin `+bumpBranch,
	)
	if err != nil {
		return err
	}

	t.Log("Next steps...")
	fmt.Println("Next, you should open a Pull Request to merge the branch " + bumpBranch + " into")
	fmt.Println(baseBranch + " and merge it if all checks succeeds. After merging run:")
	fmt.Println()
	fmt.Println("    git checkout " + baseBranch)
	fmt.Println("    git pull origin " + baseBranch)
	fmt.Println("    git tag v" + date)
	fmt.Println("    git push origin v" + date)
	fmt.Println()
	fmt.Println("After that a release should be created automatically. If not, follow the release")
	fmt.Println("guidelines in RELEASE.md.")

	return nil
}

// Check if the build is reproducible.
func TaskReproducible(t *T) error {
	var (
		build    = "go build ./cmd/ades"
		checksum = "shasum --algorithm 512 ades"
	)

	t.Log("Initial build...")
	checksum1, err := t.ExecS(build, checksum)
	if err != nil {
		return err
	}

	t.Log("Reproducing build...")
	checksum2, err := t.ExecS(build, checksum)
	if err != nil {
		return err
	}

	if checksum1 != checksum2 {
		return errors.New("Build did not reproduce")
	}

	return nil
}

// Run all tests.
func TaskTest(t *T) error {
	t.Log("Testing...")
	return t.Exec(`go test ./...`)
}

// Run mutation tests.
func TaskTestMutation(t *T) error {
	t.Log("Mutation testing...")
	return t.Exec(`go test -tags=mutation`)
}

// Run tests in a random order.
func TaskTestRandomized(t *T) error {
	t.Log("Testing (random order)...")
	return t.Exec(`go test -shuffle=on ./...`)
}

// Update the capability snapshot to the project's current capabilities.
func TaskUpdateCapabilities(t *T) error {
	t.Log("Updating capabilities...")
	stdout, err := t.ExecS(`
		go run github.com/google/capslock/cmd/capslock
			-packages ./...
			-noisy
			-output json
	`)
	if err != nil {
		return err
	}

	return os.WriteFile("./capabilities.json", []byte(stdout), permFile)
}

// Verify project is in a good state.
func TaskVerify(t *T) error {
	return t.Run(
		TaskBuild,
		TaskCompliance,
		TaskFormatCheck,
		TaskTest,
		TaskDogfeed,
		TaskVet,
	)
}

// Vet the source code.
func TaskVet(t *T) error {
	t.Log("Vetting...")
	return t.Exec(
		"go vet ./...",
		"go run 4d63.com/gochecknoinits ./...",
		"go run github.com/alexkohler/dogsled/cmd/dogsled -set_exit_status ./...",
		"go run github.com/alexkohler/nakedret/v2/cmd/nakedret -l 0 ./...",
		"go run github.com/alexkohler/prealloc -set_exit_status ./...",
		"go run github.com/alexkohler/unimport ./...",
		"go run github.com/butuzov/ireturn/cmd/ireturn ./...",
		"go run github.com/catenacyber/perfsprint ./...",
		"go run github.com/dkorunic/betteralign/cmd/betteralign ./...",
		"go run github.com/go-critic/go-critic/cmd/gocritic check ./...",
		"go run github.com/gordonklaus/ineffassign ./...",
		"go run github.com/jgautheron/goconst/cmd/goconst -numbers -set-exit-status -ignore 'web' ./...",
		"go run github.com/jgautheron/goconst/cmd/goconst -numbers -set-exit-status ./web/...",
		"go run github.com/kisielk/errcheck ./...",
		"go run github.com/kunwardeep/paralleltest -i -ignoreloopVar ./...",
		"go run github.com/mdempsky/unconvert ./...",
		"go run github.com/nishanths/exhaustive/cmd/exhaustive ./...",
		"go run github.com/polyfloyd/go-errorlint ./...",
		"go run github.com/remyoudompheng/go-misc/deadcode .",
		"go run github.com/remyoudompheng/go-misc/deadcode ./cmd/ades",
		"go run github.com/remyoudompheng/go-misc/deadcode ./web",
		"go run github.com/rhysd/actionlint/cmd/actionlint",
		"go run github.com/tetafro/godot/cmd/godot .",
		"go run github.com/tomarrell/wrapcheck/v2/cmd/wrapcheck ./...",
		"go run github.com/ultraware/whitespace/cmd/whitespace ./...",
		"go run go.uber.org/nilaway/cmd/nilaway ./...",
		"go run golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow ./...",
		"go run honnef.co/go/tools/cmd/staticcheck ./...",
		"go run mvdan.cc/unparam ./...",
	)
}

// Build the ades web application.
func TaskWebBuild(t *T) error {
	goroot, err := t.ExecS("go env GOROOT")
	if err != nil {
		return err
	}

	var (
		buildWasm    = "env GOOS=js GOARCH=wasm go build -o ades.wasm"
		copyLicense  = "cp ../COPYING.txt ./COPYING.txt"
		copyWasmExec = fmt.Sprintf("cp %s/misc/wasm/wasm_exec.js ./wasm_exec.js", goroot)
	)

	t.Log("Building webapp...")
	t.Cd(webDir)
	return t.Exec(buildWasm, copyLicense, copyWasmExec)
}

// Serve the ades web application.
func TaskWebServe(t *T) error {
	if err := t.Run(TaskWebBuild); err != nil {
		return err
	}

	t.Log("Serving locally...")
	t.Cd(webDir)
	if err := t.Exec("npm install"); err != nil {
		return err
	}

	return t.Exec("npx http-server . --port 8080")
}

// -------------------------------------------------------------------------------------------------

// T is a type passed to Task functions to perform common tasks.
type T struct {
	dir string
}

// Task is a function that performs a task.
type Task func(t *T) error

// Cd changes the directory in which the task operates.
func (t *T) Cd(dir string) {
	t.dir = dir
}

// Env returns the value of the environment variable identified by key, or the fallback value.
func (t *T) Env(key, fallback string) string {
	if value, present := os.LookupEnv(key); present {
		return value
	} else {
		return fallback
	}
}

// Env returns the value of the environment variable identified by key, or the fallback value.
func (t *T) Run(tasks ...Task) error {
	for _, task := range tasks {
		var tt T
		if err := task(&tt); err != nil {
			return err
		}
	}

	return nil
}

// Exec executes the commands printing to stdout.
func (t *T) Exec(commands ...string) error {
	return t.ExecF(os.Stdout, commands...)
}

// ExecF executes the commands writing stdout to buf.
func (t *T) ExecF(buf io.Writer, commands ...string) error {
	for _, commandStr := range commands {
		commandName, args := t.parseCommand(commandStr)

		cmd := exec.Command(commandName, args...)
		cmd.Dir = t.dir
		cmd.Stdin = os.Stdin
		cmd.Stdout = buf
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

// ExecS executes the commands returning stdout as a string.
func (t *T) ExecS(commands ...string) (string, error) {
	buf := new(bytes.Buffer)
	err := t.ExecF(buf, commands...)
	return strings.TrimSpace(buf.String()), err
}

// Log prints the messages as a line in bold. Useful to delineate steps in a task.
func (t *T) Log(msgs ...string) {
	fmt.Print("\033[1m")
	for _, msg := range msgs {
		fmt.Print(msg)
	}
	fmt.Println("\033[0m")
}

func (t *T) parseCommand(command string) (string, []string) {
	commandExp := regexp.MustCompile(`'((?:\'|[^'])+?)'|"((?:\"|[^"])+?)"|(\S+)`)
	matches := commandExp.FindAllStringSubmatch(command, -1)
	parsed := make([]string, len(matches))
	for i, match := range matches {
		if match[1] != "" {
			parsed[i] = match[1]
		} else if match[2] != "" {
			parsed[i] = match[2]
		} else {
			parsed[i] = match[3]
		}
	}

	return parsed[0], parsed[1:]
}

func main() {
	type internalTask struct {
		desc string
		name string
	}

	var (
		taskFnPrefix   = "Task"
		exprCapital    = regexp.MustCompile(`(.)([A-Z])`)
		exprHyphenated = regexp.MustCompile(`(^|-)[a-z]`)
	)

	var (
		typeCheckTaskParams = func(params []*ast.Field) bool {
			if len(params) != 1 {
				return false
			}

			paramType, ok := params[0].Type.(*ast.StarExpr)
			if !ok {
				return false
			}

			paramTypeIdent, ok := paramType.X.(*ast.Ident)
			if !ok || paramTypeIdent.Name != "T" {
				return false
			}

			return true
		}
		typeCheckTaskResults = func(results []*ast.Field) bool {
			if len(results) != 1 {
				return false
			}

			_, ok := results[0].Type.(ast.Expr)
			return ok
		}
	)

	var (
		parse = func() ([]internalTask, error) {
			file, err := parser.ParseFile(token.NewFileSet(), "tasks.go", nil, parser.ParseComments)
			if err != nil {
				return nil, fmt.Errorf("could not parse file: %s", err)
			}

			tasks := make([]internalTask, 0)
			for _, decl := range file.Decls {
				// Check the declaration type, only functions can be tasks
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}

				// Check for the task prefix, which marks a runnable task
				fnName := fn.Name.Name
				if !strings.HasPrefix(fnName, taskFnPrefix) {
					continue
				}

				// Check that the function signature is correct
				if ok := typeCheckTaskParams(fn.Type.Params.List); !ok {
					return nil, fmt.Errorf("wrong signature for %q, should accept '*T'", fnName)
				}
				if ok := typeCheckTaskResults(fn.Type.Results.List); !ok {
					return nil, fmt.Errorf("wrong signature for %q, should return 'error'", fnName)
				}

				// Convert the function name to a task name
				name := strings.TrimPrefix(fnName, taskFnPrefix)
				name = exprCapital.ReplaceAllString(name, "${1}-${2}")
				name = strings.ToLower(name)

				// Extract task description as the first line of the doc comment
				desc := fn.Doc.Text()
				if eol := strings.IndexRune(desc, '\n'); eol != -1 {
					desc = desc[0:eol]
				}

				tasks = append(tasks, internalTask{desc, name})
			}

			return tasks, nil
		}
		build = func(tasks []string) ([]byte, error) {
			wd, err := os.Getwd()
			if err != nil {
				return nil, errors.New("could not get the current working directory")
			}

			original, err := os.ReadFile("./tasks.go")
			if err != nil {
				return nil, errors.New("could not read the task file")
			}

			var sb strings.Builder
			sb.WriteString(`func main() {var t T;`)
			for _, taskName := range tasks {
				name := exprHyphenated.ReplaceAllStringFunc(taskName, strings.ToUpper)
				name = strings.ReplaceAll(name, "-", "")

				sb.WriteString(fmt.Sprintf(`t.Cd("%s");`, wd))
				sb.WriteString(fmt.Sprintf(`if err := Task%s(&t); err != nil {`, name))
				sb.WriteString(`fmt.Fprintln(os.Stderr);`)
				sb.WriteString(`exitCode := 1;`)
				sb.WriteString(`if exitErr, ok := err.(*exec.ExitError); ok {`)
				sb.WriteString(`exitCode = exitErr.ExitCode()`)
				sb.WriteString(`} else {`)
				sb.WriteString(`fmt.Fprintf(os.Stderr, "Error: %v\n", err)`)
				sb.WriteString(`};`)
				sb.WriteString(fmt.Sprintf(`fmt.Fprintln(os.Stderr, "Task '%s' failed");`, taskName))
				sb.WriteString(`os.Exit(exitCode)`)
				sb.WriteString(`};`)
			}
			sb.WriteRune('}')

			var (
				exprMain         = regexp.MustCompile(`func main\(\) \{\n([^\n]*\n)+\}`)
				exprUnusedImport = regexp.MustCompile(`	"go/[a-z]*"\n`)
			)

			runner := exprMain.ReplaceAll(original, []byte(sb.String()))
			runner = exprUnusedImport.ReplaceAll(runner, []byte{})
			return runner, nil
		}
		run = func(tasks []string) (int, error) {
			runner, err := build(tasks)
			if err != nil {
				return 2, err
			}

			wd, err := os.MkdirTemp(os.TempDir(), "go-task-*")
			if err != nil {
				return 2, errors.New("could not create a temporary working directory")
			}
			defer os.RemoveAll(wd)

			workerBin := fmt.Sprintf("%s%ctask-runner", wd, os.PathSeparator)
			workerSrc := workerBin + ".go"
			os.WriteFile(workerSrc, runner, 0o666)

			cmd := exec.Command("go", "build", "-o", workerBin, workerSrc)
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil {
				return 2, fmt.Errorf("could not build the task runner: %v", err)
			}

			cmd = exec.Command(workerBin)
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					return exitErr.ExitCode(), nil
				} else {
					return 2, fmt.Errorf("unexpected execution error: %v", err)
				}
			}

			return 0, nil
		}
	)

	tasks, err := parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Syntax error: %s\n", err)
		os.Exit(2)
	}

	if len(os.Args) < 2 {
		fmt.Println("usage:\n  go run tasks.go <task1> [task2...]")
		fmt.Println()
		fmt.Println("tasks:")
		for _, task := range tasks {
			fmt.Printf("  %s\n    %s\n", task.name, task.desc)
		}

		os.Exit(0)
	}

	for _, taskName := range os.Args[1:] {
		found := false
		for _, task := range tasks {
			found = (taskName == task.name) || found
		}

		if !found {
			fmt.Fprintf(os.Stderr, "Task not found: %q\n", taskName)
			os.Exit(2)
		}
	}

	exitCode, err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	os.Exit(exitCode)
}
