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
	"os"
	"path"
	"regexp"
)

const (
	exitSuccess    = 0
	exitError      = 1
	exitViolations = 2
)

var (
	flagLegal = flag.Bool(
		"legal",
		false,
		"Show legal information and exit",
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

	targets, err := getTargets(flag.Args())
	if err != nil {
		fmt.Printf("Unexpected error getting working directory: %s", err)
		return exitError
	}

	hasViolations, hasError := false, false
	for i, target := range targets {
		if len(targets) > 1 {
			fmt.Println("Scanning", target)
		}

		violations, err := analyzeTarget(target)
		if err == nil {
			printViolations(violations)
		} else {
			fmt.Printf("An unexpected error occurred: %s\n", err)
			hasError = true
		}

		for _, fileVioviolations := range violations {
			if len(fileVioviolations) > 0 {
				hasViolations = true
			}
		}

		if len(targets) > 1 && i < len(targets)-1 {
			fmt.Println( /* empty line */ )
		}
	}

	switch {
	case hasError:
		return exitError
	case hasViolations:
		return exitViolations
	default:
		return exitSuccess
	}
}

func analyzeTarget(target string) (map[string][]Violation, error) {
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

		violations := make(map[string][]Violation)
		violations[target] = fileViolations
		return violations, nil
	}
}

func analyzeRepository(target string) (map[string][]Violation, error) {
	violations := make(map[string][]Violation)

	if fileViolations, err := tryManifest(path.Join(target, "action.yml")); err == nil {
		violations["action.yml"] = fileViolations
	} else if !errors.Is(err, ErrNotRead) {
		fmt.Printf("Could not process manifest 'action.yml': %v\n", err)
	}

	if fileViolations, err := tryManifest(path.Join(target, "action.yaml")); err == nil {
		violations["action.yaml"] = fileViolations
	} else if !errors.Is(err, ErrNotRead) {
		fmt.Printf("Could not process manifest 'action.yaml': %v\n", err)
	}

	workflowsDir := path.Join(target, ".github", "workflows")
	workflows, err := os.ReadDir(workflowsDir)
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

		workflowPath := path.Join(workflowsDir, entry.Name())
		if workflowViolations, err := tryWorkflow(workflowPath); err == nil {
			targetRelativePath := path.Join(".github", "workflows", entry.Name())
			violations[targetRelativePath] = workflowViolations
		} else {
			fmt.Printf("Could not process workflow %s: %v\n", entry.Name(), err)
		}
	}

	return violations, nil
}

func analyzeFile(target string) ([]Violation, error) {
	if matched, _ := regexp.MatchString("action.ya?ml", target); matched {
		return tryManifest(target)
	} else {
		return tryWorkflow(target)
	}
}

var (
	ErrNotRead   = errors.New("not found")
	ErrNotParsed = errors.New("not parsed")
)

func tryManifest(manifestPath string) ([]Violation, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrNotRead, err)
	}

	manifest, err := ParseManifest(data)
	if err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrNotParsed, err)
	}

	return analyzeManifest(&manifest), nil
}

func tryWorkflow(workflowPath string) ([]Violation, error) {
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrNotRead, err)
	}

	workflow, err := ParseWorkflow(data)
	if err != nil {
		return nil, fmt.Errorf("%w (%v)", ErrNotParsed, err)
	}

	return analyzeWorkflow(&workflow), nil
}

func printViolations(violations map[string][]Violation) {
	for file, fileViolations := range violations {
		if cnt := len(fileViolations); cnt > 0 {
			fmt.Printf("Detected %d violation(s) in '%s':\n", cnt, file)
			for _, violation := range fileViolations {
				if violation.jobId == "" {
					fmt.Printf("   step %s has '%s'\n", violation.stepId, violation.problem)
				} else {
					fmt.Printf("   job %s, step %s has '%s'\n", violation.jobId, violation.stepId, violation.problem)
				}
			}
		}
	}
}

func getTargets(argv []string) ([]string, error) {
	if len(argv) == 0 {
		wd, err := os.Getwd()
		return []string{wd}, err
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
  --legal     Show legal information and exit.
  --version   Show the program version and exit.

Exit Codes:

  0   Success
  1   Unexpected error
  2   Problems detected`)
}

func version() {
	fmt.Println("v23.10")
}
