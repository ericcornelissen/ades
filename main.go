// Copyright (C) 2023  Eric Cornelissen
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	exitSuccess = iota
	exitError
	exitViolations
)

var (
	flagJson = flag.Bool(
		"json",
		false,
		"Output results as JSON",
	)
	flagLegal = flag.Bool(
		"legal",
		false,
		"Show legal information and exit",
	)
	flagStdin = flag.Bool(
		"stdin",
		false,
		"Read workflow or manifest from stdin",
	)
	flagVersion = flag.Bool(
		"version",
		false,
		"Show the program version and exit",
	)
)

func main() {
	os.Exit(run())
}

func run() int {
	flag.Usage = func() { usage() }
	flag.Parse()

	if *flagLegal {
		legal()
		return exitSuccess
	}

	if *flagVersion {
		version()
		return exitSuccess
	}

	if *flagStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return exitError
		}

		violations := make(map[string][]violation)
		if workflowViolations, err := tryWorkflow(data); err != nil {
			return exitError
		} else if len(workflowViolations) > 0 {
			violations["stdin"] = workflowViolations
		} else if manifestViolations, err := tryManifest(data); err != nil {
			return exitError
		} else if len(manifestViolations) > 0 {
			violations["stdin"] = manifestViolations
		}

		if len(violations) > 0 {
			if *flagJson {
				report := make(map[string]map[string][]violation)
				report["stdin"] = violations
				fmt.Println(printJson(report))
			} else {
				fmt.Print(printViolations(violations))
			}

			return exitViolations
		}

		return exitSuccess
	}

	targets, err := getTargets(flag.Args())
	if err != nil {
		fmt.Printf("Unexpected error getting working directory: %s", err)
		return exitError
	}

	violations, hasError := make(map[string]map[string][]violation), false
	for i, target := range targets {
		if len(targets) > 1 && !(*flagJson) {
			fmt.Println("Scanning", target)
		}

		targetViolations, err := analyzeTarget(target)
		if err == nil {
			if !(*flagJson) {
				fmt.Print(printViolations(targetViolations))

				if i < len(targets)-1 {
					fmt.Println( /* empty line between targets */ )
				}
			}

			for file, fileViolations := range targetViolations {
				if len(fileViolations) > 0 {
					targetViolations, ok := violations[target]
					if !ok {
						targetViolations = make(map[string][]violation)
						violations[target] = targetViolations
					}

					targetViolations[file] = fileViolations
				}
			}
		} else {
			fmt.Printf("An unexpected error occurred: %s\n", err)
			hasError = true
		}
	}

	if *flagJson {
		fmt.Println(printJson(violations))
	}

	switch {
	case hasError:
		return exitError
	case len(violations) > 0:
		return exitViolations
	default:
		return exitSuccess
	}
}

func analyzeTarget(target string) (map[string][]violation, error) {
	stat, err := os.Stat(target)
	if err != nil {
		return nil, fmt.Errorf("could not process %s: %v", target, err)
	}

	if stat.IsDir() {
		return analyzeRepository(target)
	} else {
		fileViolations, err := analyzeFile(target)
		if err != nil {
			return nil, err
		}

		violations := make(map[string][]violation)
		violations[target] = fileViolations
		return violations, nil
	}
}

const (
	githubDir    = ".github"
	workflowsDir = "workflows"
)

var (
	manifestExpr = regexp.MustCompile("action.ya?ml")
)

func analyzeRepository(target string) (map[string][]violation, error) {
	violations := make(map[string][]violation)

	if fileViolations, err := analyzeFile(path.Join(target, "action.yml")); err == nil {
		violations["action.yml"] = fileViolations
	} else if !errors.Is(err, errNotFound) {
		fmt.Printf("Could not process manifest 'action.yml': %v\n", err)
	}

	if fileViolations, err := analyzeFile(path.Join(target, "action.yaml")); err == nil {
		violations["action.yaml"] = fileViolations
	} else if !errors.Is(err, errNotFound) {
		fmt.Printf("Could not process manifest 'action.yaml': %v\n", err)
	}

	workflowsPath := path.Join(target, githubDir, workflowsDir)
	workflows, err := os.ReadDir(workflowsPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return violations, fmt.Errorf("could not read workflows directory: %v", err)
	}

	for _, entry := range workflows {
		if entry.IsDir() {
			continue
		}

		if ext := path.Ext(entry.Name()); ext != ".yml" && ext != ".yaml" {
			continue
		}

		workflowPath := path.Join(workflowsPath, entry.Name())
		if workflowViolations, err := analyzeFile(workflowPath); err == nil {
			targetRelativePath := path.Join(githubDir, workflowsDir, entry.Name())
			violations[targetRelativePath] = workflowViolations
		} else {
			fmt.Printf("Could not process workflow %s: %v\n", entry.Name(), err)
		}
	}

	return violations, nil
}

var (
	errNotFound  = errors.New("not found")
	errNotParsed = errors.New("not parsed")
)

func analyzeFile(target string) ([]violation, error) {
	absolutePath, err := filepath.Abs(target)
	if err != nil {
		return nil, errors.Join(errNotFound, err)
	}

	data, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, errors.Join(errNotFound, err)
	}

	switch {
	case strings.HasSuffix(absolutePath, path.Join(githubDir, workflowsDir, path.Base(target))):
		return tryWorkflow(data)
	case manifestExpr.MatchString(target):
		return tryManifest(data)
	default:
		return tryWorkflow(data)
	}
}

func tryManifest(data []byte) ([]violation, error) {
	manifest, err := ParseManifest(data)
	if err != nil {
		return nil, errors.Join(errNotParsed, err)
	}

	return analyzeManifest(&manifest), nil
}

func tryWorkflow(data []byte) ([]violation, error) {
	workflow, err := ParseWorkflow(data)
	if err != nil {
		return nil, errors.Join(errNotParsed, err)
	}

	return analyzeWorkflow(&workflow), nil
}

func getTargets(argv []string) ([]string, error) {
	if len(argv) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("could not get cwd: %v", err)
		}

		return []string{wd}, nil
	}

	return argv, nil
}

func legal() {
	fmt.Println(`ades  Copyright (C) 2023  Eric Cornelissen
This program comes with ABSOLUTELY NO WARRANTY; see the GPL v3.0 for details.
This is free software, and you are welcome to redistribute it under certain
conditions; see the GPL v3.0 for details.`)
}

func usage() {
	fmt.Println(`find problematic use of template variables in GitHub Action workflows

Usage:

  ades [path]...

Flags:

  --help      Show this help message and exit.
  --json      Output results in JSON format.
  --legal     Show legal information and exit.
  --stdin     Read workflow or manifest from stdin.
  --version   Show the program version and exit.

Exit Codes:

  0   Success
  1   Unexpected error
  2   Problems detected`)
}

func version() {
	fmt.Println("v23.11")
}
