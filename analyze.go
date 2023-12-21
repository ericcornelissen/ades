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

	"golang.org/x/mod/semver"
)

type violationKind uint8

const (
	expressionInRunScript violationKind = iota
	expressionInActionsGithubScript
	expressionInGitTagAnnotationActionTagInput
	expressionInGitMessageActionShaInput
)

var (
	expressionInRunScriptId                      = "ADES100"
	expressionInActionsGithubScriptId            = "ADES101"
	expressionInGitTagAnnotationActionTagInputId = "ADES200"
	expressionInGitMessageActionShaInputId       = "ADES201"
)

func (kind violationKind) String() string {
	var s string
	switch kind {
	case expressionInRunScript:
		s = expressionInRunScriptId
	case expressionInActionsGithubScript:
		s = expressionInActionsGithubScriptId
	case expressionInGitTagAnnotationActionTagInput:
		s = expressionInGitTagAnnotationActionTagInputId
	case expressionInGitMessageActionShaInput:
		s = expressionInGitMessageActionShaInputId
	}

	return s
}

type violation struct {
	jobId   string
	stepId  string
	problem string
	kind    violationKind
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
		job := job
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
		step := step
		violations = append(violations, analyzeStep(i, &step)...)
	}

	return violations
}

func analyzeStep(id int, step *JobStep) []violation {
	name := step.Name
	if step.Name == "" {
		name = fmt.Sprintf("#%d", id)
	}

	var violations []violation
	var kind violationKind

	uses, _ := ParseUses(step)
	switch {
	case isRunStep(step):
		kind = expressionInRunScript
		violations = analyzeString(step.Run)
	case uses.Name == "actions/github-script":
		kind = expressionInActionsGithubScript
		violations = analyzeString(step.With["script"])
	case uses.Name == "ericcornelissen/git-tag-annotation-action":
		if isBeforeOrAtVersion(uses, "v1.0.0") {
			kind = expressionInGitTagAnnotationActionTagInput
			violations = analyzeString(step.With["tag"])
		}
	case uses.Name == "kceb/git-message-action":
		if isBeforeOrAtVersion(uses, "v1.1.0") {
			kind = expressionInGitMessageActionShaInput
			violations = analyzeString(step.With["sha"])
		}
	}

	for i := range violations {
		violations[i].kind = kind
		violations[i].stepId = name
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

func isRunStep(step *JobStep) bool {
	return len(step.Run) > 0
}

func isBeforeOrAtVersion(uses StepUses, version string) bool {
	// This comparison checks that the `Ref` is a semantic version string, which is currently the
	// only supported type of `Ref`.
	if semver.Canonical(uses.Ref) != uses.Ref {
		return false
	}

	return semver.Compare(version, uses.Ref) >= 0
}
