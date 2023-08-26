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

type Problem struct {
	jobId   string
	stepId  string
	problem string
}

var r = regexp.MustCompile(`\$\{\{.*?\}\}`)

func analyzeManifest(manifest *Manifest) (problems []Problem) {
	if manifest.Runs.Using == "composite" {
		problems = analyzeSteps(manifest.Runs.Steps)
	}

	return problems
}

func analyzeWorkflow(workflow *Workflow) (problems []Problem) {
	for id, job := range workflow.Jobs {
		job := job
		problems = append(problems, analyzeJob(id, &job)...)
	}

	return problems
}

func analyzeJob(id string, job *WorkflowJob) (problems []Problem) {
	name := job.Name
	if name == "" {
		name = id
	}

	for _, problem := range analyzeSteps(job.Steps) {
		problem.jobId = fmt.Sprintf("'%s'", name)
		problems = append(problems, problem)
	}

	return problems
}

func analyzeSteps(steps []JobStep) (problems []Problem) {
	for i, step := range steps {
		step := step
		problems = append(problems, analyzeStep(i, &step)...)
	}

	return problems
}

func analyzeStep(id int, step *JobStep) (problems []Problem) {
	name := fmt.Sprintf("'%s'", step.Name)
	if step.Name == "" {
		name = fmt.Sprintf("#%d", id)
	}

	var script string
	if isRunStep(step) {
		script = step.Run
	} else if isActionsGitHubScriptStep(step) {
		script = step.With.Script
	} else {
		return nil
	}

	for _, problem := range analyzeScript(script) {
		problem.stepId = name
		problems = append(problems, problem)
	}

	return problems
}

func analyzeScript(script string) (problems []Problem) {
	if matches := r.FindAll([]byte(script), -1); matches != nil {
		for _, problem := range matches {
			problems = append(problems, Problem{
				problem: string(problem),
			})
		}
	}

	return problems
}

func isRunStep(step *JobStep) bool {
	return len(step.Run) > 0
}

func isActionsGitHubScriptStep(step *JobStep) bool {
	return strings.HasPrefix(step.Uses, "actions/github-script@")
}
