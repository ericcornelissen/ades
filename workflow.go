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

var r = regexp.MustCompile(`\$\{\{.*?\}\}`)

func processManifest(manifest *Manifest) (problems []string) {
	if manifest.Runs.Using == "composite" {
		problems = processSteps(manifest.Runs.Steps)
	}

	return problems
}

func processWorkflow(workflow *Workflow) (problems []string) {
	for id, job := range workflow.Jobs {
		job := job
		problems = append(problems, processJob(id, &job)...)
	}

	return problems
}

func processJob(id string, job *WorkflowJob) (problems []string) {
	name := job.Name
	if name == "" {
		name = id
	}

	for _, problem := range processSteps(job.Steps) {
		problem = fmt.Sprintf("job '%s', %s", name, problem)
		problems = append(problems, problem)
	}

	return problems
}

func processSteps(steps []JobStep) (problems []string) {
	for i, step := range steps {
		step := step
		problems = append(problems, processStep(i, &step)...)
	}

	return problems
}

func processStep(id int, step *JobStep) (problems []string) {
	name := fmt.Sprintf("'%s'", step.Name)
	if step.Name == "" {
		name = fmt.Sprintf("#%d", id)
	}

	if isRunStep(step) {
		for _, problem := range processScript(step.Run) {
			problem := fmt.Sprintf("step %s has '%s' in run", name, problem)
			problems = append(problems, problem)
		}
	} else if isActionsGitHubScriptStep(step) {
		for _, problem := range processScript(step.With.Script) {
			problem := fmt.Sprintf("step %s has '%s' in script", name, problem)
			problems = append(problems, problem)
		}
	}

	return problems
}

func processScript(script string) (problems []string) {
	if matches := r.FindAll([]byte(script), -1); matches != nil {
		for _, problem := range matches {
			problems = append(problems, string(problem))
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
