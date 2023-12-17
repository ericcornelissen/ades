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
	Name string   `yaml:"name"`
	Run  string   `yaml:"run"`
	Uses string   `yaml:"uses"`
	With StepWith `yaml:"with"`
}

// StepWith is a (simplified) representation of a job step's `with:` object.
type StepWith struct {
	Script string `yaml:"script"`
	Tag    string `yaml:"tag"`
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
