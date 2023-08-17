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
)

var r = regexp.MustCompile(`\$\{\{.*?\}\}`)

func processWorkflow(workflow *Workflow) (problems []string) {
	for id, job := range workflow.Jobs {
		job := job
		problems = append(problems, processJob(id, &job)...)
	}

	return problems
}

func processJob(id string, job *Job) (problems []string) {
	name := job.Name
	if name == "" {
		name = id
	}

	for i, step := range job.Steps {
		step := step
		for _, problem := range processStep(i, &step) {
			problem = fmt.Sprintf("job '%s', %s", name, problem)
			problems = append(problems, problem)
		}
	}

	return problems
}

func processStep(id int, step *Step) (problems []string) {
	if matches := r.FindAll([]byte(step.Run), -1); matches != nil {
		name := fmt.Sprintf("'%s'", step.Name)
		if step.Name == "" {
			name = fmt.Sprintf("#%d", id)
		}

		for _, match := range matches {
			problem := fmt.Sprintf("step %s has '%s'", name, match)
			problems = append(problems, problem)
		}
	}

	return problems
}
