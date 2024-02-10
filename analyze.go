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

type violation struct {
	jobId   string
	stepId  string
	problem string
	ruleId  string
}

var (
	ghaExpressionRegExp = regexp.MustCompile(`\$\{\{.*?\}\}`)
)

func analyzeManifest(manifest *Manifest) []violation {
	if manifest != nil && manifest.Runs.Using == "composite" {
		return analyzeSteps(manifest.Runs.Steps)
	} else {
		return make([]violation, 0)
	}
}

func analyzeWorkflow(workflow *Workflow) []violation {
	violations := make([]violation, 0)
	if workflow == nil {
		return violations
	}

	for id, job := range workflow.Jobs {
		violations = append(violations, analyzeJob(id, &job)...)
	}

	return violations
}

func analyzeJob(id string, job *WorkflowJob) []violation {
	name := job.Name
	if name == "" {
		name = id
	}

	violations := make([]violation, 0)
	for _, v := range analyzeSteps(job.Steps) {
		v.jobId = name
		violations = append(violations, v)
	}

	return violations
}

func analyzeSteps(steps []JobStep) []violation {
	violations := make([]violation, 0)
	for i, step := range steps {
		violations = append(violations, analyzeStep(i, &step)...)
	}

	return violations
}

func analyzeStep(id int, step *JobStep) []violation {
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

	violations := make([]violation, 0)
	for _, rule := range rules {
		for _, violation := range analyzeString(rule.extractFrom(step)) {
			violation.ruleId = rule.id
			violation.stepId = name
			violations = append(violations, violation)
		}
	}

	return violations
}

func analyzeString(s string) []violation {
	violations := make([]violation, 0)
	if matches := ghaExpressionRegExp.FindAll([]byte(s), len(s)); matches != nil {
		for _, problem := range matches {
			violations = append(violations, violation{
				problem: string(problem),
			})
		}
	}

	return violations
}
