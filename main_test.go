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
	"os"
	"testing"
	"testing/quick"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

func TestMain(m *testing.M) {
	commands := map[string]func() int{
		"ades": run,
	}

	os.Exit(testscript.RunMain(m, commands))
}

func TestCli(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "test",
	})
}

func TestJsonSchema(t *testing.T) {
	schema, err := jsonschema.Compile("schema.json")
	if err != nil {
		t.Fatalf("schema.json is not a valid JSON Schema: %v", err)
	}

	f := func(output jsonOutput) bool {
		bytes, err := json.Marshal(output)
		if err != nil {
			return false
		}

		var data any
		if err = json.Unmarshal(bytes, &data); err != nil {
			return false
		}

		err = schema.Validate(data)
		return err == nil
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
