package main

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	commands := map[string]func() int{
		"ades": ades,
	}

	os.Exit(testscript.RunMain(m, commands))
}

func TestCli(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "test",
	})
}
