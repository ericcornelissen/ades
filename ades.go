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

type options struct {
	conservative bool
}

func parseOptions(opts js.Value) options {
	return options{
		conservative: opts.Get("conservative").Bool(),
	}
}

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

func lintWorkflow(source string, opts *options) ([]ades.Violation, error) {
	workflow, err := ades.ParseWorkflow([]byte(source))
	if err != nil {
		return nil, err
	}

	var matcher ades.ExprMatcher = ades.AllMatcher
	if opts.conservative {
		matcher = ades.ConservativeMatcher
	}

	return ades.AnalyzeWorkflow(&workflow, matcher), nil
}

func lintManifest(source string, opts *options) ([]ades.Violation, error) {
	manifest, err := ades.ParseManifest([]byte(source))
	if err != nil {
		return nil, err
	}

	var matcher ades.ExprMatcher = ades.AllMatcher
	if opts.conservative {
		matcher = ades.ConservativeMatcher
	}

	return ades.AnalyzeManifest(&manifest, matcher), nil
}

func analyze(source string, opts *options) {
	violations, err := lintWorkflow(source, opts)
	if err != nil || len(violations) == 0 {
		violations, err = lintManifest(source, opts)
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
	opts := parseOptions(args[1])
	analyze(source, &opts)

	return nil
}

func main() {
	window.Set("ades", js.FuncOf(runAdes))

	initialSource := window.Call("getSource").String()
	rawOpts := window.Call("getOptions")
	opts := parseOptions(rawOpts)
	analyze(initialSource, &opts)

	// Keep the program alive
	select {}
}
