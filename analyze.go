// Copyright (C) 2023-2025  Eric Cornelissen
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

	violations := make([]Violation, 0)
	for _, violation := range analyzeSteps(job.Steps, matcher) {
		if matrixSafe(violation.Problem, job.Strategy.Matrix, matcher) {
			continue
		}

		violation.jobKey = id
		violation.JobId = name
		violations = append(violations, violation)
	}

	return violations
}

func analyzeSteps(steps []gha.Step, matcher ExprMatcher) []Violation {
	violations := make([]Violation, 0)
	for i, step := range steps {
		violations = append(violations, analyzeStep(i, &step, matcher)...)
	}

	return violations
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

	violations := make([]Violation, 0)
	for _, rule := range rules {
		for _, violation := range analyzeString(rule.extractFrom(step), matcher) {
			violation.RuleId = rule.id
			violation.StepId = name
			violation.stepIndex = id
			violations = append(violations, violation)
		}
	}

	return violations
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
