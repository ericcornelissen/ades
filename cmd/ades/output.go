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
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/ericcornelissen/ades"
)

type jsonOutput struct {
	Violations []jsonViolation `json:"problems"`
}

type jsonViolation struct {
	Target  string `json:"target"`
	File    string `json:"file"`
	Job     string `json:"job"`
	Step    string `json:"step"`
	Problem string `json:"problem"`
}

func printJson(rawViolations map[string]map[string][]ades.Violation) string {
	violations := make([]jsonViolation, 0)
	for target, targetViolations := range rawViolations {
		for file, fileViolations := range targetViolations {
			for _, fileViolation := range fileViolations {
				violations = append(violations, jsonViolation{
					Target:  target,
					File:    file,
					Job:     fileViolation.JobId,
					Step:    fileViolation.StepId,
					Problem: fileViolation.Problem,
				})
			}
		}
	}

	sort.Slice(violations, func(i, j int) bool {
		return violations[i].File < violations[j].File
	})

	jsonBytes, _ := json.Marshal(jsonOutput{Violations: violations})
	return string(jsonBytes)
}

func printProjectViolations(violations map[string][]ades.Violation) string {
	clean := true

	files := slices.Collect(maps.Keys(violations))
	sort.Strings(files)

	var sb strings.Builder
	for _, file := range files {
		fileViolations := violations[file]
		if cnt := len(fileViolations); cnt > 0 {
			clean = false
			sb.WriteString(fmt.Sprintf("Detected %d violation(s) in %q:", cnt, file))
			sb.WriteRune('\n')
			sb.WriteString(printFileViolations(fileViolations))
		}
	}

	if clean {
		return "Ok\n"
	} else {
		return sb.String()
	}
}

func printFileViolations(violations []ades.Violation) string {
	byJob := make(map[string][]ades.Violation, len(violations))
	for _, violation := range violations {
		jobId := violation.JobId
		if _, ok := byJob[jobId]; !ok {
			byJob[jobId] = make([]ades.Violation, 0)
		}

		byJob[jobId] = append(byJob[jobId], violation)
	}

	jobs := slices.Collect(maps.Keys(byJob))
	sort.Strings(jobs)

	var sb strings.Builder
	for _, job := range jobs {
		violations := byJob[job]

		if job != "" {
			sb.WriteString("  ")
			sb.WriteString(fmt.Sprintf("%d in job %q:", len(violations), job))
			sb.WriteRune('\n')
		}

		for _, violation := range violations {
			sb.WriteString("    ")
			sb.WriteString(fmt.Sprintf("step %q contains %q (%s)", violation.StepId, violation.Problem, violation.RuleId))
			sb.WriteRune('\n')
		}
	}

	return sb.String()
}
