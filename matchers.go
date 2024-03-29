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
	"regexp"
)

// ExprMatcher is the interface for types that can find GitHub Workflow Expressions in strings.
type ExprMatcher interface {
	// FindAll is the function that returns all relevant GitHub Workflow Expressions in the provided
	// input.
	FindAll([]byte) [][]byte
}

var (
	// AllMatcher is an ExprMatcher that will find all GitHub Workflow Expressions in strings.
	AllMatcher allExprMatcher

	// ConservativeMatcher is an ExprMatcher that will conservatively find GitHub Workflow
	// Expressions in strings that are known to be controllable by attackers.
	ConservativeMatcher conservativeExprMatcher
)

var allExprRegExp = regexp.MustCompile(`\$\{\{.*?\}\}`)

type allExprMatcher struct{}

func (m allExprMatcher) FindAll(v []byte) [][]byte {
	return allExprRegExp.FindAll(v, len(v))
}

var conservativeExprRegExp = regexp.MustCompile(`\$\{\{\s*(github\.event\.issue\.title|github\.event\.issue\.body|github\.event\.discussion\.title|github\.event\.discussion\.body|github\.event\.comment\.body|github\.event\.review\.body|github\.event\.review_comment\.body|github\.event\.pages\[\d+\]\.page_name|github\.event\.commits\[\d+\]\.message|github\.event\.commits\[\d+\]\.author\.email|github\.event\.commits\[\d+\]\.author\.name|github\.event\.head_commit\.message|github\.event\.head_commit\.author\.email|github\.event\.head_commit\.author\.name|github\.event\.head_commit\.committer\.email|github\.event\.workflow_run\.head_branch|github\.event\.workflow_run\.head_commit\.message|github\.event\.workflow_run\.head_commit\.author\.email|github\.event\.workflow_run\.head_commit\.author\.name|github\.event\.pull_request\.title|github\.event\.pull_request\.body|github\.event\.pull_request\.head\.label|github\.event\.pull_request\.head\.repo\.default_branch|github\.head_ref|github\.event\.pull_request\.head\.ref|github\.event\.workflow_run\.pull_requests\[\d+\]\.head\.ref)\s*\}\}`)

type conservativeExprMatcher struct{}

func (m conservativeExprMatcher) FindAll(v []byte) [][]byte {
	return conservativeExprRegExp.FindAll(v, len(v))
}
