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

//go:build wasm

package main

import (
	"syscall/js"

	"github.com/ericcornelissen/ades"
)

var window = js.Global().Get("window")

func fail(description string, err error) {
	window.Call("showError", description, err.Error())
}

func succeed(result []any) {
	window.Call("showResult", js.ValueOf(result))
}

func encodeErrorAsMap(violation *ades.Violation) map[string]any {
	obj := make(map[string]any, 4)
	obj["job"] = violation.JobId
	obj["step"] = violation.StepId
	obj["problem"] = violation.Problem
	obj["ruleId"] = violation.RuleId
	return obj
}

func lintWorkflow(source string) ([]ades.Violation, error) {
	workflow, err := ades.ParseWorkflow([]byte(source))
	if err != nil {
		return nil, err
	}

	return ades.AnalyzeWorkflow(&workflow, ades.AllMatcher), nil
}

func lintManifest(source string) ([]ades.Violation, error) {
	manifest, err := ades.ParseManifest([]byte(source))
	if err != nil {
		return nil, err
	}

	return ades.AnalyzeManifest(&manifest, ades.AllMatcher), nil
}

func analyze(source string) {
	violations, err := lintWorkflow(source)
	if err != nil || len(violations) == 0 {
		violations, err = lintManifest(source)
	}

	if err != nil {
		fail("Parsing failure", err)
		return
	}

	result := make([]any, 0, len(violations))
	for _, violation := range violations {
		result = append(result, encodeErrorAsMap(&violation))
	}

	succeed(result)
	return
}

func runAdes(_this js.Value, args []js.Value) any {
	source := args[0].String()
	analyze(source)

	return nil
}

func main() {
	window.Set("ades", js.FuncOf(runAdes))

	initialSource := window.Call("getSource").String()
	analyze(initialSource)

	// Keep the program alive
	select {}
}
