// Copyright (C) 2024  Eric Cornelissen
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
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ericcornelissen/ades"
)

const (
	exitSuccess = iota
	exitError
	exitViolations
)

var (
	flagConservative = flag.Bool(
		"conservative",
		false,
		"Only report expressions known to be controllable by attackers",
	)
	flagExplain = flag.String(
		"explain",
		"",
		"Explain the given violation",
	)
	flagFix = flag.Bool(
		"fix-experimental",
		false,
		"Automatically fix violations if possible",
	)
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
	flagSuggestions = flag.Bool(
		"suggestions",
		false,
		"Show suggested fixes inline",
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

	if *flagExplain != "" {
		explanation, err := ades.Explain(*flagExplain)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unknown rule %q\n", *flagExplain)
			return exitError
		} else {
			fmt.Println(explanation)
			return exitSuccess
		}
	}

	targets := flag.Args()
	if len(targets) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get working directory, use explicit target instead (error: %s)\n", err)
			return exitError
		}

		targets = []string{wd}
	}

	var (
		ok     bool
		report map[string]map[string][]ades.Violation
	)

	if targets[0] == "-" {
		targets = []string{"stdin"}
		report, ok = runOnStdin()
	} else {
		report, ok = runOnTargets(targets)
	}

	if !ok {
		return exitError
	}

	if *flagFix {
		if err := fix(report); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitError
		}
	}

	if *flagJson {
		fmt.Println(printJson(report))
	} else {
		for i, target := range targets {
			if i > 0 {
				fmt.Println( /* empty line between targets */ )
			}
			if len(targets) > 1 {
				fmt.Printf("[%s]\n", target)
			}

			violations := report[target]
			fmt.Print(printViolations(violations, *flagSuggestions))
		}
	}

	for _, targetViolations := range report {
		for _, fileViolations := range targetViolations {
			if len(fileViolations) > 0 {
				return exitViolations
			}
		}
	}

	return exitSuccess
}

func runOnStdin() (map[string]map[string][]ades.Violation, bool) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Printf("Could not read from stdin (error: %s)\n", err)
		return nil, false
	}

	violations := make(map[string][]ades.Violation)
	if workflowViolations, err := tryWorkflow(data); err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse input (error: %s)\n", err)
		return nil, false
	} else if len(workflowViolations) != 0 {
		violations["stdin"] = workflowViolations
	} else {
		manifestViolations, _ := tryManifest(data)
		violations["stdin"] = manifestViolations
	}

	report := make(map[string]map[string][]ades.Violation)
	report["stdin"] = violations

	return report, true
}

func runOnTargets(targets []string) (map[string]map[string][]ades.Violation, bool) {
	report, hasError := make(map[string]map[string][]ades.Violation), false
	for _, target := range targets {
		violations, err := runOnTarget(target)
		if err == nil {
			for file, fileViolations := range violations {
				targetViolations, ok := report[target]
				if !ok {
					targetViolations = make(map[string][]ades.Violation)
					report[target] = targetViolations
				}

				targetViolations[file] = fileViolations
			}
		} else {
			fmt.Fprintf(os.Stderr, "An unexpected error occurred: %s\n", err)
			hasError = true
		}
	}

	return report, !hasError
}

func runOnTarget(target string) (map[string][]ades.Violation, error) {
	stat, err := os.Stat(target)
	if err != nil {
		return nil, fmt.Errorf("could not process %s: %v", target, err)
	}

	if stat.IsDir() {
		return runOnRepository(target)
	} else {
		fileViolations, err := runOnFile(target)
		if err != nil {
			return nil, err
		}

		violations := make(map[string][]ades.Violation)
		violations[target] = fileViolations
		return violations, nil
	}
}

const (
	githubDir    = ".github"
	workflowsDir = "workflows"
)

func runOnRepository(target string) (map[string][]ades.Violation, error) {
	violations := make(map[string][]ades.Violation)

	fsys := os.DirFS(target)
	_ = fs.WalkDir(fsys, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if entry.IsDir() {
			if path == ".git" {
				return fs.SkipDir
			}

			return nil
		}

		if name := filepath.Base(path); !ghaManifestFileRegExp.MatchString(name) {
			return nil
		}

		fullPath := filepath.Join(target, path)
		if fileViolations, err := runOnFile(fullPath); err == nil {
			violations[path] = fileViolations
		} else {
			fmt.Fprintf(os.Stderr, "Could not process manifest %q: %v\n", path, err)
		}

		return nil
	})

	workflowsPath := filepath.Join(githubDir, workflowsDir)
	_ = fs.WalkDir(fsys, workflowsPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if entry.IsDir() {
			if path == workflowsPath {
				return nil
			}

			return fs.SkipDir
		}

		if ext := filepath.Ext(entry.Name()); ext != ".yml" && ext != ".yaml" {
			return nil
		}

		fullPath := filepath.Join(target, path)
		if workflowViolations, err := runOnFile(fullPath); err == nil {
			violations[path] = workflowViolations
		} else {
			fmt.Fprintf(os.Stderr, "Could not process workflow %s: %v\n", entry.Name(), err)
		}

		return nil
	})

	return violations, nil
}

var (
	errNotFound  = errors.New("not found")
	errNotParsed = errors.New("not parsed")

	ghaManifestFileRegExp = regexp.MustCompile("action.ya?ml")
)

func runOnFile(target string) ([]ades.Violation, error) {
	absolutePath, err := filepath.Abs(target)
	if err != nil {
		return nil, errors.Join(errNotFound, err)
	}

	data, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, errors.Join(errNotFound, err)
	}

	switch {
	case strings.HasSuffix(absolutePath, filepath.Join(githubDir, workflowsDir, filepath.Base(target))):
		return tryWorkflow(data)
	case ghaManifestFileRegExp.MatchString(target):
		return tryManifest(data)
	default:
		return tryWorkflow(data)
	}
}

func tryManifest(data []byte) ([]ades.Violation, error) {
	manifest, err := ades.ParseManifest(data)
	if err != nil {
		return nil, errors.Join(errNotParsed, err)
	}

	var matcher ades.ExprMatcher
	if *flagConservative {
		matcher = ades.ConservativeMatcher
	} else {
		matcher = ades.AllMatcher
	}

	return ades.AnalyzeManifest(&manifest, matcher), nil
}

func tryWorkflow(data []byte) ([]ades.Violation, error) {
	workflow, err := ades.ParseWorkflow(data)
	if err != nil {
		return nil, errors.Join(errNotParsed, err)
	}

	var matcher ades.ExprMatcher
	if *flagConservative {
		matcher = ades.ConservativeMatcher
	} else {
		matcher = ades.AllMatcher
	}

	return ades.AnalyzeWorkflow(&workflow, matcher), nil
}

func fix(report map[string]map[string][]ades.Violation) error {
	for _, report := range report {
		for file, violations := range report {
			fh, err := os.OpenFile(file, os.O_RDWR, 0o644)
			if err != nil {
				return fmt.Errorf("cannot open %q: %v", file, err)
			} else {
				defer func() { _ = fh.Close() }()
			}

			raw, err := io.ReadAll(fh)
			if err != nil {
				return fmt.Errorf("cannot read %q: %v", file, err)
			}

			var unfixed []ades.Violation
			for _, violation := range violations {
				fixes, _ := ades.Fix(&violation)
				if fixes != nil {
					for _, fix := range fixes {
						raw = fix.Old.ReplaceAll(raw, []byte(fix.New))
					}
				} else {
					unfixed = append(unfixed, violation)
				}
			}

			if _, err := fh.WriteAt(raw, 0); err != nil {
				return fmt.Errorf("cannot write to %q: %v", file, err)
			} else {
				_ = fh.Close()
			}

			report[file] = unfixed
		}
	}

	return nil
}

func legal() {
	fmt.Println(`ades  Copyright (C) 2024  Eric Cornelissen
This program comes with ABSOLUTELY NO WARRANTY; see the GPL v3.0 for details.
This is free software, and you are welcome to redistribute it under certain
conditions; see the GPL v3.0 for details.`)
}

func usage() {
	fmt.Println(`find dangerous uses of expressions in GitHub Action workflows

Usage:

  ades [path]...

Flags:

  -conservative       Only report expressions known to be controllable by attackers.
  -explain ADESxxx    Explain the given violation.
  -fix-experimental   Automatically fix violations if possible.
  -help               Show this help message and exit.
  -json               Output results in JSON format.
  -legal              Show legal information and exit.
  -suggestions        Show suggested fixes inline.
  -version            Show the program version and exit.
  -                   Read workflow or manifest from stdin.

Exit Codes:

  0   Success
  1   Unexpected error
  2   Problems detected`)
}

func version() {
	versionString := "v24.03"
	fmt.Println(versionString)
}
