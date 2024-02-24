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
	"fmt"
	"regexp"
)

type Violation struct {
	JobId   string
	StepId  string
	Problem string
	RuleId  string
}

var (
	ghaExpressionRegExp = regexp.MustCompile(`\$\{\{.*?\}\}`)
)

func AnalyzeManifest(manifest *Manifest) []Violation {
	violations := make([]Violation, 0)
	if manifest == nil {
		return violations
	}

	if manifest.Runs.Using != "composite" {
		return violations
	}

	return analyzeSteps(manifest.Runs.Steps)
}

func AnalyzeWorkflow(workflow *Workflow) []Violation {
	violations := make([]Violation, 0)
	if workflow == nil {
		return violations
	}

	for id, job := range workflow.Jobs {
		violations = append(violations, analyzeJob(id, &job)...)
	}

	return violations
}

func analyzeJob(id string, job *WorkflowJob) []Violation {
	name := job.Name
	if name == "" {
		name = id
	}

	violations := make([]Violation, 0)
	for _, violation := range analyzeSteps(job.Steps) {
		violation.JobId = name
		violations = append(violations, violation)
	}

	return violations
}

func analyzeSteps(steps []JobStep) []Violation {
	violations := make([]Violation, 0)
	for i, step := range steps {
		violations = append(violations, analyzeStep(i, &step)...)
	}

	return violations
}

func analyzeStep(id int, step *JobStep) []Violation {
	name := step.Name
	if step.Name == "" {
		name = fmt.Sprintf("#%d", id)
	}

	rules := make([]rule, 0)
	if uses, err := ParseUses(step); err == nil {
		if rs, ok := actionRules[uses.Name]; ok {
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
		for _, violation := range analyzeString(rule.extractFrom(step)) {
			violation.RuleId = rule.id
			violation.StepId = name
			violations = append(violations, violation)
		}
	}

	return violations
}

func analyzeString(s string) []Violation {
	violations := make([]Violation, 0)
	if matches := ghaExpressionRegExp.FindAll([]byte(s), len(s)); matches != nil {
		for _, problem := range matches {
			violations = append(violations, Violation{
				Problem: string(problem),
			})
		}
	}

	return violations
}
