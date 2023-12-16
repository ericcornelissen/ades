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
	"testing"
)

func TestExplainRule(t *testing.T) {
	testCases := []string{
		expressionInRunScriptId,
		expressionInActionsGithubScriptId,
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt, func(t *testing.T) {
			t.Parallel()

			explanation, err := explain(tt)
			if err != nil {
				t.Errorf("Unexpected error occurred for %q: %q", tt, err)
			} else if explanation == "" {
				t.Errorf("Unexpected empty explanation for %q", tt)
			}
		})
	}
}

func TestExplainNonRule(t *testing.T) {
	testCases := []string{
		"ADES000",
		"foobar",
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt, func(t *testing.T) {
			t.Parallel()

			_, err := explain(tt)
			if err == nil {
				t.Errorf("Expected an error for %q but got none", tt)
			}
		})
	}
}
