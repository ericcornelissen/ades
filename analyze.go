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
	"fmt"
	"regexp"
	"strings"
)

type violationKind uint8

const (
	expressionInRunScript violationKind = iota
	expressionInActionsGithubScript
)

type violation struct {
	jobId   string
	stepId  string
	problem string
	kind    violationKind
}

var r = regexp.MustCompile(`\$\{\{.*?\}\}`)

func analyzeManifest(manifest *Manifest) (violations []violation) {
	if manifest.Runs.Using == "composite" {
		violations = analyzeSteps(manifest.Runs.Steps)
	}

	return violations
}

func analyzeWorkflow(workflow *Workflow) (violations []violation) {
	for id, job := range workflow.Jobs {
		job := job
		violations = append(violations, analyzeJob(id, &job)...)
	}

	return violations
}

func analyzeJob(id string, job *WorkflowJob) (violations []violation) {
	name := job.Name
	if name == "" {
		name = id
	}

	for _, v := range analyzeSteps(job.Steps) {
		v.jobId = name
		violations = append(violations, v)
	}

	return violations
}

func analyzeSteps(steps []JobStep) (violations []violation) {
	for i, step := range steps {
		step := step
		violations = append(violations, analyzeStep(i, &step)...)
	}

	return violations
}

func analyzeStep(id int, step *JobStep) (violations []violation) {
	name := step.Name
	if step.Name == "" {
		name = fmt.Sprintf("#%d", id)
	}

	script, kind := extractScript(step)
	for _, v := range analyzeScript(script) {
		v.kind = kind
		v.stepId = name
		violations = append(violations, v)
	}

	return violations
}

func analyzeScript(script string) (violations []violation) {
	if matches := r.FindAll([]byte(script), len(script)); matches != nil {
		for _, problem := range matches {
			violations = append(violations, violation{
				problem: string(problem),
			})
		}
	}

	return violations
}

func extractScript(step *JobStep) (script string, kind violationKind) {
	switch {
	case isRunStep(step):
		return step.Run, expressionInRunScript
	case isActionsGitHubScriptStep(step):
		return step.With.Script, expressionInActionsGithubScript
	default:
		return script, kind
	}
}

func isRunStep(step *JobStep) bool {
	return len(step.Run) > 0
}

func isActionsGitHubScriptStep(step *JobStep) bool {
	return strings.HasPrefix(step.Uses, "actions/github-script@")
}
