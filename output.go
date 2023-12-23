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
	"encoding/json"
	"fmt"
	"sort"
	"strings"
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

func printJson(rawViolations map[string]map[string][]violation) string {
	violations := make([]jsonViolation, 0)
	for target, targetViolations := range rawViolations {
		for file, fileViolations := range targetViolations {
			for _, fileViolation := range fileViolations {
				violations = append(violations, jsonViolation{
					Target:  target,
					File:    file,
					Job:     fileViolation.jobId,
					Step:    fileViolation.stepId,
					Problem: fileViolation.problem,
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

func printViolations(violations map[string][]violation, suggestions bool) string {
	var sb strings.Builder
	for file, fileViolations := range violations {
		if cnt := len(fileViolations); cnt > 0 {
			sb.WriteString(fmt.Sprintf("Detected %d violation(s) in %q:", cnt, file))
			sb.WriteRune('\n')
			for _, violation := range fileViolations {
				violation := violation
				sb.WriteString(printViolation(&violation, suggestions))
				sb.WriteRune('\n')
			}
		}
	}

	return sb.String()
}

func printViolation(v *violation, suggestions bool) string {
	var sb strings.Builder
	if v.jobId == "" {
		sb.WriteString(fmt.Sprintf("  step %q has %q", v.stepId, v.problem))
	} else {
		sb.WriteString(fmt.Sprintf("  job %q, step %q has %q", v.jobId, v.stepId, v.problem))
	}

	if suggestions {
		r, _ := findRule(v.ruleId)
		suggestion := r.suggestion(v)

		sb.WriteString(", suggestion:\n")
		sb.WriteString(suggestion)
	} else {
		sb.WriteString(fmt.Sprintf(" (%s)", v.ruleId))
	}

	return sb.String()
}
