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
	"flag"
	"fmt"
	"os"
	"path"
)

const (
	exitSuccess  = 0
	exitError    = 1
	exitProblems = 2
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
	flag.Usage = func() { usage() }
	flag.Parse()

	if *flagLegal {
		legal()
		os.Exit(exitSuccess)
	}

	if *flagVersion {
		version()
		os.Exit(exitSuccess)
	}

	targets, err := getTargets(flag.Args())
	if err != nil {
		fmt.Printf("Unexpected error getting working directory: %s", err)
		os.Exit(exitError)
	}

	problems, err := false, nil
	for i, target := range targets {
		if len(targets) > 1 {
			fmt.Println("Scanning", target)
		}

		targetHasProblems, targetErr := run(target)
		if targetErr != nil {
			err = targetErr
			fmt.Printf("An unexpected error occurred: %s\n", targetErr)
		}

		if targetHasProblems {
			problems = true
		}

		if len(targets) > 1 && i < len(targets)-1 {
			fmt.Println( /* empty line */ )
		}
	}

	switch {
	case err != nil:
		os.Exit(exitError)
	case problems:
		os.Exit(exitProblems)
	default:
		os.Exit(exitSuccess)
	}
}

func run(wd string) (hasProblems bool, err error) {
	if data, err := os.ReadFile(path.Join(wd, "action.yml")); err == nil {
		manifest, err := parseManifest(data)
		if err != nil {
			return hasProblems, err
		}

		problems := processManifest(&manifest)
		if cnt := len(problems); cnt > 0 {
			hasProblems = true
			fmt.Printf("Detected %d problem(s) in 'action.yml':\n", cnt)
			for _, problem := range problems {
				fmt.Println("  ", problem)
			}
		}
	}

	workflowsDir := path.Join(wd, ".github", "workflows")
	workflows, err := os.ReadDir(workflowsDir)
	if err != nil {
		return hasProblems, fmt.Errorf("could not read workflows directory: %v", err)
	}

	for _, entry := range workflows {
		if entry.Type().IsDir() {
			continue
		}

		if path.Ext(entry.Name()) != ".yml" {
			continue
		}

		workflowPath := path.Join(workflowsDir, entry.Name())
		data, err := os.ReadFile(workflowPath)
		if err != nil {
			fmt.Printf("Could not read %s: %v\n", entry.Name(), err)
			continue
		}

		workflow, err := parseWorkflow(data)
		if err != nil {
			fmt.Printf("Could not parse %s: %v\n", entry.Name(), err)
			continue
		}

		problems := processWorkflow(&workflow)
		if cnt := len(problems); cnt > 0 {
			hasProblems = true
			fmt.Printf("Detected %d problem(s) in '%s':\n", cnt, entry.Name())
			for _, problem := range problems {
				fmt.Println("  ", problem)
			}
		}
	}

	return hasProblems, nil
}

func getTargets(argv []string) ([]string, error) {
	if len(argv) > 0 {
		return argv, nil
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		return []string{wd}, err
	}
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
	fmt.Println("v23.08")
}
