// Copyright (C) 2025  Eric Cornelissen
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
	"strings"

	"github.com/ericcornelissen/go-gha-models"
)

var matrixExprRegExp = regexp.MustCompile(`matrix\.[a-z._-]+`)

// Check if the expression is safe under the given matrix.
func matrixSafe(expr string, matrix gha.Matrix, matcher ExprMatcher) bool {
	matrixExprs := matrixExprRegExp.FindAllString(expr, len(expr))
	for _, expr := range matrixExprs {
		values := getMatrixValues(expr, matrix.Matrix)
		for _, include := range matrix.Include {
			values = append(values, getMatrixValues(expr, include)...)
		}

		if len(values) == 0 {
			return false
		}

		for _, value := range values {
			if violations := analyzeString(value, matcher); len(violations) > 0 {
				return false
			}
		}
	}

	for _, match := range matrixExprs {
		expr = strings.ReplaceAll(expr, match, "")
	}
	if violations := analyzeString(expr, matcher); len(violations) != 0 {
		return false
	}

	return true
}

// Get all possible values of the given matrix expression from the given matrix.
func getMatrixValues(expr string, matrix map[string]any) []string {
	expr, _ = strings.CutPrefix(expr, "matrix.")

	parts := strings.Split(expr, ".")
	head, tail := parts[0], parts[1:]

	if value, ok := matrix[head]; ok {
		switch tmp := value.(type) {
		case map[string]any:
			expr := strings.Join(tail, ".")
			return getMatrixValues(expr, tmp)
		case []any:
			var values []string
			for _, tmp := range tmp {
				switch tmp := tmp.(type) {
				case string:
					values = append(values, tmp)
				case map[string]any:
					expr := strings.Join(tail, ".")
					values = append(values, getMatrixValues(expr, tmp)...)
				}
			}
			return values
		case any:
			value, _ := tmp.(string)
			return []string{value}
		}
	}

	return nil
}
