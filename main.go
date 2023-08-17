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
	exitSuccess = 0
	exitError   = 1
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

	wd, err := getwd(flag.Args())
	if err != nil {
		fmt.Printf("Unexpected error getting working directory: %s", err)
		os.Exit(exitError)
	}

	if err := run(wd); err != nil {
		fmt.Printf("An unexpected error occurred: %s\n", err)
		os.Exit(exitError)
	}

	os.Exit(exitSuccess)
}

func run(wd string) error {
	workflowsDir := path.Join(wd, ".github", "workflows")
	workflows, err := os.ReadDir(workflowsDir)
	if err != nil {
		return fmt.Errorf("could not read workflows directory: %v", err)
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

		workflow, err := parse(data)
		if err != nil {
			fmt.Printf("Could not parse %s: %v\n", entry.Name(), err)
			continue
		}

		problems := processWorkflow(&workflow)
		if cnt := len(problems); cnt > 0 {
			fmt.Printf("Detected %d problem(s) in '%s':\n", cnt, entry.Name())
			for _, problem := range problems {
				fmt.Println("  ", problem)
			}
		}
	}

	return nil
}

func getwd(argv []string) (string, error) {
	if len(argv) > 0 {
		return argv[0], nil
	} else {
		return os.Getwd()
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

  ades [path]

Flags:

  --help      Show this help message and exit.
  --legal     Show legal information and exit.
  --version   Show the program version and exit.`)
}

func version() {
	fmt.Println("v23.08")
}
