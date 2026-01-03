// Copyright (C) 2023-2026  Eric Cornelissen
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

package ades

import (
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/ericcornelissen/go-gha-models"
)

// Violation contain information on problematic GitHub Actions Expressions found in a workflow or
// manifest.
type Violation struct {
	// Source is a reference to the Workflow or Manifest struct in which the violation occurs.
	source any

	// JobId is an identifier of a job in a GitHub Actions workflow, either the name or key.
	//
	// This will be the zero value if the violation is for a GitHub Actions manifest.
	JobId string

	// StepId is the identifier of a step in a GitHub Actions workflow or manifest, either the name
	// or index.
	StepId string

	// Problem is the problematic GitHub Actions Expression as observed in the workflow or manifest.
	Problem string

	// RuleId is the identifier of the ades rule that produced the violation.
	RuleId string

	// JobKey is the key of the job in which the violation occurs. Different from JobId in that it
	// always uniquely identifies the job.
	//
	// This will be the zero value if the violation is for a GitHub Actions manifest.
	jobKey string

	// StepIndex is the index of the step in which the violation occurs. Different from StepId in
	// that it always uniquely identifies the step.
	stepIndex int
}

// AnalyzeRepo analyzes a GitHub repository for problematic GitHub Actions Expressions in manifests
// and workflows.
func AnalyzeRepo(fsys fs.FS, matcher ExprMatcher) (map[string][]Violation, error) {
	report := make(map[string][]Violation)
	return report, fs.WalkDir(fsys, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if path == ".git" {
				return fs.SkipDir
			}

			return nil
		}

		var (
			dir  = filepath.Dir(path)
			base = filepath.Base(path)
			ext  = filepath.Ext(path)

			isManifest = base == "action.yml" || base == "action.yaml"
			isWorkflow = dir == ".github/workflows" && (ext == ".yml" || ext == ".yaml")
		)

		if !isManifest && !isWorkflow {
			return nil
		}

		file, err := fsys.Open(path)
		if err != nil {
			return fmt.Errorf("could not open %q: %v", path, err)
		}
		defer func() { _ = file.Close() }()

		content, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("could not read %q: %v", path, err)
		}

		if isWorkflow {
			workflow, err := gha.ParseWorkflow(content)
			if err != nil {
				return fmt.Errorf("could not process workflow %q: %v", path, err)
			}

			report[path] = AnalyzeWorkflow(&workflow, matcher)
		} else {
			manifest, err := gha.ParseManifest(content)
			if err != nil {
				return fmt.Errorf("could not process manifest %q: %v", path, err)
			}

			report[path] = AnalyzeManifest(&manifest, matcher)
		}

		return nil
	})
}

// AnalyzeManifest analyzes a GitHub Actions manifest for problematic GitHub Actions Expressions.
func AnalyzeManifest(manifest *gha.Manifest, matcher ExprMatcher) []Violation {
	violations := make([]Violation, 0)
	if manifest == nil {
		return violations
	}

	if manifest.Runs.Using != "composite" {
		return violations
	}

	for _, violation := range analyzeSteps(manifest.Runs.Steps, matcher) {
		violation.source = manifest
		violations = append(violations, violation)
	}

	return violations
}

// AnalyzeWorkflow analyzes a GitHub Actions workflow for problematic GitHub Actions Expressions.
func AnalyzeWorkflow(workflow *gha.Workflow, matcher ExprMatcher) []Violation {
	violations := make([]Violation, 0)
	if workflow == nil {
		return violations
	}

	for id, job := range workflow.Jobs {
		for _, violation := range analyzeJob(id, &job, matcher) {
			violation.source = workflow
			violations = append(violations, violation)
		}
	}

	return violations
}

func analyzeJob(id string, job *gha.Job, matcher ExprMatcher) []Violation {
	name := job.Name
	if name == "" {
		name = id
	}

	violations := analyzeSteps(job.Steps, matcher)
	out := make([]Violation, 0, len(violations))
	for _, violation := range violations {
		if matrixSafe(violation.Problem, job.Strategy.Matrix, matcher) {
			continue
		}

		violation.jobKey = id
		violation.JobId = name
		out = append(out, violation)
	}

	return out
}

func analyzeSteps(steps []gha.Step, matcher ExprMatcher) []Violation {
	violations := make([][]Violation, len(steps))
	for i, step := range steps {
		violations[i] = analyzeStep(i, &step, matcher)
	}

	return flatten(violations)
}

func analyzeStep(id int, step *gha.Step, matcher ExprMatcher) []Violation {
	name := step.Name
	if step.Name == "" {
		name = fmt.Sprintf("#%d", id)
	}

	rules := make([]rule, 0)
	if uses := step.Uses; uses.Name != "" {
		actionName := strings.ToLower(uses.Name)
		if rs, ok := actionRules[actionName]; ok {
			for _, r := range rs {
				if r.appliesTo(&uses) {
					rules = append(rules, r.rule)
				}
			}
		}
	} else {
		for _, r := range stepRules {
			if r.appliesTo(step) {
				rules = append(rules, r.rule)
			}
		}
	}

	violations := make([][]Violation, len(rules))
	for i, rule := range rules {
		vs := analyzeString(rule.extractFrom(step), matcher)
		for i := range vs {
			vs[i].RuleId = rule.id
			vs[i].StepId = name
			vs[i].stepIndex = id
		}

		violations[i] = vs
	}

	return flatten(violations)
}

func analyzeString(s string, matcher ExprMatcher) []Violation {
	b := []byte(s)
	if len(matcher.FindAll(stripSafe(b))) == 0 {
		return nil
	}

	violations := make([]Violation, 0)
	if matches := matcher.FindAll(b); matches != nil {
		for _, problem := range matches {
			violations = append(violations, Violation{
				Problem: string(problem),
			})
		}
	}

	return violations
}

func flatten[T any](s [][]T) []T {
	size := 0
	for _, s := range s {
		size += len(s)
	}

	offset := 0
	out := make([]T, size)
	for _, s := range s {
		for i, e := range s {
			out[offset+i] = e
		}

		offset += len(s)
	}

	return out
}
