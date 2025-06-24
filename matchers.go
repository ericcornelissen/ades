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

package ades

import (
	"bytes"
	"regexp"
)

// ExprMatcher is the interface for types that can find GitHub Actions Expressions in strings.
type ExprMatcher interface {
	// FindAll is the function that returns all relevant GitHub Actions Expressions in the provided
	// input.
	FindAll([]byte) [][]byte
}

var (
	// AllMatcher is an ExprMatcher that will find all GitHub Actions Expressions in strings.
	AllMatcher allExprMatcher

	// ConservativeMatcher is an ExprMatcher that will conservatively find GitHub Workflow
	// Expressions in strings that are known to be controllable by attackers.
	ConservativeMatcher conservativeExprMatcher
)

var allExprRegExp = regexp.MustCompile(`\${{.*?}}`)

type allExprMatcher struct{}

func (m allExprMatcher) FindAll(v []byte) [][]byte {
	if allExprRegExp.Find(stripSafe(v)) == nil {
		return nil
	}

	return allExprRegExp.FindAll(v, len(v))
}

var conservativeExps = []*regexp.Regexp{
	regexp.MustCompile(`github\.event\.comment\.body`),
	regexp.MustCompile(`github\.event\.commits\[\d+\]\.author\.email`),
	regexp.MustCompile(`github\.event\.commits\[\d+\]\.author\.name`),
	regexp.MustCompile(`github\.event\.commits\[\d+\]\.message`),
	regexp.MustCompile(`github\.event\.discussion\.body`),
	regexp.MustCompile(`github\.event\.discussion\.title`),
	regexp.MustCompile(`github\.event\.head_commit\.author\.email`),
	regexp.MustCompile(`github\.event\.head_commit\.author\.name`),
	regexp.MustCompile(`github\.event\.head_commit\.committer\.email`),
	regexp.MustCompile(`github\.event\.head_commit\.message`),
	regexp.MustCompile(`github\.event\.issue\.body`),
	regexp.MustCompile(`github\.event\.issue\.title`),
	regexp.MustCompile(`github\.event\.pages\[\d+\]\.page_name`),
	regexp.MustCompile(`github\.event\.pull_request\.body`),
	regexp.MustCompile(`github\.event\.pull_request\.head\.label`),
	regexp.MustCompile(`github\.event\.pull_request\.head\.ref`),
	regexp.MustCompile(`github\.event\.pull_request\.head\.repo\.default_branch`),
	regexp.MustCompile(`github\.event\.pull_request\.title`),
	regexp.MustCompile(`github\.event\.review\.body`),
	regexp.MustCompile(`github\.event\.review_comment\.body`),
	regexp.MustCompile(`github\.event\.workflow_run\.head_branch`),
	regexp.MustCompile(`github\.event\.workflow_run\.head_commit\.author\.email`),
	regexp.MustCompile(`github\.event\.workflow_run\.head_commit\.author\.name`),
	regexp.MustCompile(`github\.event\.workflow_run\.head_commit\.message`),
	regexp.MustCompile(`github\.event\.workflow_run\.pull_requests\[\d+\]\.head\.ref`),
	regexp.MustCompile(`github\.head_ref`),
}

type conservativeExprMatcher struct{}

func (m conservativeExprMatcher) FindAll(v []byte) [][]byte {
	if allExprRegExp.Find(stripSafe(v)) == nil {
		return nil
	}

	all := allExprRegExp.FindAll(v, len(v))
	conservative := make([][]byte, 0, len(all))
	for _, candidate := range all {
		for _, exp := range conservativeExps {
			if exp.Match(candidate) {
				conservative = append(conservative, candidate)
				break
			}
		}
	}

	return conservative
}

var (
	boundary = `[\s,|&!=()<>]`
	leading  = `(?P<leading>\${{(.*?` + boundary + `|))`
	trailing = `(?P<trailing>(` + boundary + `.*?|)}})`

	LiteralInExprRegExp      = regexp.MustCompile(leading + `(true|false|null|-?\d+(\.\d+)?|0x[0-9A-Fa-f]+|-?\d+\.\d+e-?\d+|'[^']+')` + trailing)
	SafeContextInExprRegExp  = regexp.MustCompile(leading + `(github\.(action_status|event_name|event\.number|event\.workflow_run\.id|ref_protected|ref_type|retention_days|run_attempt|run_id|run_number|secret_source|sha|workflow_sha|workspace)|job\.status|jobs\.[a-z-_]+\.result|steps\.[a-z-_]+\.(conclusion|outcome)|runner\.(arch|debug|environment|os)|strategy\.(fail-fast|job-index|job-total|max-parallel)|needs\.[a-z-_]+\.result)` + trailing)
	SafeFunctionInExprRegExp = regexp.MustCompile(leading + `((always|cancelled|contains|endsWith|failure|hashFiles|success|startsWith)\(([^,]*,)*[^,)]*\)|(format|fromJSON|join|toJSON)\([\s,]*\))` + trailing)

	EmptyExprRegExp = regexp.MustCompile(`\${{` + boundary + `*}}`)
)

func stripSafe(v []byte) []byte {
	exps := []regexp.Regexp{
		*LiteralInExprRegExp,
		*SafeContextInExprRegExp,
		*SafeFunctionInExprRegExp,
	}

	var r []byte
	for !bytes.Equal(v, r) {
		r = v
		for _, exp := range exps {
			v = exp.ReplaceAll(v, []byte("$leading$trailing"))
		}
	}

	return EmptyExprRegExp.ReplaceAll(v, nil)
}
