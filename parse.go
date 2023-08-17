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

type Workflow struct {
	Jobs map[string]Job `yaml:"jobs"`
}

type Job struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}

func parse(data []byte) (workflow Workflow, err error) {
	if err = yaml.Unmarshal(data, &workflow); err != nil {
		return workflow, fmt.Errorf("could not parse workflow: %v", err)
	}

	return workflow, nil
}
