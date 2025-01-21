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
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Workflow is a (simplified) representation of a GitHub Actions workflow.
type Workflow struct {
	Jobs map[string]WorkflowJob `yaml:"jobs"`
}

// WorkflowJob is a (simplified) representation of a workflow job.
type WorkflowJob struct {
	Name  string    `yaml:"name"`
	Steps []JobStep `yaml:"steps"`
}

// JobStep is a (simplified) representation of a workflow job step object.
type JobStep struct {
	With        map[string]string `yaml:"with"`
	Env         map[string]string `yaml:"env"`
	Name        string            `yaml:"name"`
	Run         string            `yaml:"run"`
	Shell       string            `yaml:"shell"`
	Uses        string            `yaml:"uses"`
	UsesComment string            `yaml:"-"`
}

func (step *JobStep) UnmarshalYAML(node *yaml.Node) error {
	for i := range node.Content {
		if i%2 == 1 {
			continue
		}

		key := node.Content[i].Value
		value := node.Content[i+1]

		var err error
		switch key {
		case "env":
			err = value.Decode(&step.Env)
		case "name":
			step.Name = value.Value
		case "run":
			step.Run = value.Value
		case "shell":
			step.Shell = value.Value
		case "uses":
			step.Uses = value.Value
			step.UsesComment = strings.TrimLeft(value.LineComment, "# ")
		case "with":
			err = value.Decode(&step.With)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// ParseWorkflow parses a GitHub Actions workflow file into a Workflow struct.
func ParseWorkflow(data []byte) (Workflow, error) {
	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return workflow, fmt.Errorf("could not parse workflow: %v", err)
	}

	return workflow, nil
}

// Manifest is a (simplified) representation of a GitHub Actions Action manifest.
type Manifest struct {
	Runs ManifestRuns `yaml:"runs"`
}

// ManifestRuns is a (simplified) representation of an Action manifest's `runs:` object.
type ManifestRuns struct {
	Using string    `yaml:"using"`
	Steps []JobStep `yaml:"steps"`
}

// ParseManifest parses a GitHub Actions Action manifest file into a Manifest struct.
func ParseManifest(data []byte) (Manifest, error) {
	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return manifest, fmt.Errorf("could not parse manifest: %v", err)
	}

	return manifest, nil
}

// StepUses is a structured representation of a workflow job step `uses:` value.
type StepUses struct {
	// Name is the name of the Action that is used. Typically <owner>/<repository>.
	Name string

	// Ref is the git reference used for the Action. Typically a tag ref, branch ref, or commit SHA.
	Ref string

	// Annotation is the comment after the `uses:` value, if any.
	Annotation string
}

// ParseUses parses a Github Actions workflow job step's `uses:` value.
func ParseUses(step *JobStep) (StepUses, error) {
	var uses StepUses

	i := strings.LastIndex(step.Uses, "@")
	if i <= 0 || i >= len(step.Uses)-1 {
		return uses, errors.New("step has no or invalid `uses` value")
	}

	uses.Name = step.Uses[:i]
	uses.Ref = step.Uses[i+1:]
	uses.Annotation = step.UsesComment
	return uses, nil
}
