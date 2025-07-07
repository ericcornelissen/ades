// Copyright (C) 2024-2025  Eric Cornelissen
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

//go:build web

package ades_test

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
)

var browserOptions = playwright.BrowserTypeLaunchOptions{
	Headless: playwright.Bool(true),
	Devtools: playwright.Bool(false),
}

func TestWebInitialState(t *testing.T) {
	s, p := setup(t, nil)
	defer s.Close()

	report, err := getReport(p)
	if err != nil {
		t.Fatalf("could not get report: %v", err)
	}

	got, want := report, "Found 1 problem"
	if !strings.Contains(got, want) {
		t.Errorf("unexpected report (got %q, missing %q)", got, want)
	}
}

func TestWebCleanWorkflow(t *testing.T) {
	s, p := setup(t, nil)
	defer s.Close()

	workflow := `name: Example
on: [push]

jobs:
  example:
    name: example
    runs-on: ubuntu-latest
    steps:
    - name: Safe run
      run: echo 'Hello world!'`

	err := setWorkflow(p, workflow)
	if err != nil {
		t.Fatalf("could not set workflow: %v", err)
	}

	report, err := getReport(p)
	if err != nil {
		t.Fatalf("could not get report: %v", err)
	}

	got, want := report, "No problems detected"
	if got != want {
		t.Errorf("unexpected report (got %q, want %q)", got, want)
	}
}

func TestWebConservative(t *testing.T) {
	s, p := setup(t, nil)
	defer s.Close()

	workflow := `name: Example
on: [push]

jobs:
  example:
    name: example
    runs-on: ubuntu-latest
    steps:
    - name: Conservatively safe
      run: echo '${{ matrix.foobar }}'
    - name: Conservatively unsafe
      run: echo '${{ github.event.issue.title }}'`

	err := setWorkflow(p, workflow)
	if err != nil {
		t.Fatalf("could not set workflow: %v", err)
	}

	conservative := p.Locator("#option-conservative")
	if conservative == nil {
		t.Fatal("could not find conservative option")
	}

	err = conservative.Check()
	if err != nil {
		t.Fatal("could not enable conservative option")
	}

	report, err := getReport(p)
	if err != nil {
		t.Fatalf("could not get report: %v", err)
	}

	got, want := report, "Found 1 problem"
	if !strings.Contains(got, want) {
		t.Fatalf("unexpected report (got %q, missing %q)", got, want)
	}

	got, want = report, "${{ github.event.issue.title }}"
	if !strings.Contains(got, want) {
		t.Errorf("unexpected report (got %q, missing %q)", got, want)
	}
}

func TestWebChaos(t *testing.T) {
	scripts := []playwright.Script{
		{Path: playwright.String("./gremlins.min.js")},
		{Content: playwright.String("gremlins.createHorde().unleash()")},
	}

	s, p := setup(t, scripts)
	defer s.Close()

	// Ensure no unexpected errors occur
	p.OnPageError(func(err error) {
		if strings.HasPrefix(err.Error(), "playwright: Cannot read properties of null") {
			// allowed because Gremlins produce these themselves :(
			return
		}

		t.Errorf("an unexpected error occurred on the page: %v", err)
	})

	// Ensure minimum performance
	count, total := 0, float64(0)
	p.OnConsole(func(cm playwright.ConsoleMessage) {
		msg := cm.String()
		if !strings.HasPrefix(msg, "mogwai  fps") {
			return
		}

		parts := strings.Split(msg, " ")
		fps, err := strconv.ParseFloat(parts[len(parts)-1], 64)
		if err != nil {
			t.Fatalf("could not parse %q: %v", msg, err)
		}

		count += 1
		total += fps
	})

	time.Sleep(3 * time.Second)
	if avg := total / float64(count); avg < 45 {
		t.Errorf("performance issue detected, average fps was %f", avg)
	}
}

func setWorkflow(p playwright.Page, workflow string) error {
	input := p.Locator("#workflow-input")
	if input == nil {
		return errors.New("could not find the workflow input")
	}

	err := input.Fill(workflow)
	if err != nil {
		return fmt.Errorf("could not set workflow: %v", err)
	}

	return nil
}

func getReport(p playwright.Page) (string, error) {
	report := p.Locator("#results")
	if report == nil {
		return "", errors.New("could not find the report")
	}

	text, err := report.InnerText()
	if err != nil {
		return "", fmt.Errorf("could not extract the report: %v", err)
	}

	return text, nil
}

func setup(t *testing.T, scripts []playwright.Script) (*http.Server, playwright.Page) {
	t.Helper()

	err := playwright.Install()
	if err != nil {
		t.Fatalf("could not install playwright: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("could not start playwright: %v", err)
	}

	browser, err := pw.Chromium.Launch(browserOptions)
	if err != nil {
		t.Fatalf("could not launch a browser: %v", err)
	}

	page, err := browser.NewPage()
	if err != nil {
		t.Fatalf("could not create a new page: %v", err)
	}

	for _, script := range scripts {
		if err := page.AddInitScript(script); err != nil {
			t.Fatalf("could not add init script %v: %v", script, err)
		}
	}

	port, err := getAvailablePort()
	if err != nil {
		t.Fatalf("no port available: %v", err)
	}

	server := http.Server{
		Addr:    ":" + port,
		Handler: http.FileServer(http.Dir("web")),
	}
	go server.ListenAndServe()

	_, err = page.Goto("localhost:" + port)
	if err != nil {
		t.Fatalf("could not goto: %v", err)
	}

	return &server, page
}

func getAvailablePort() (string, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}

	defer listener.Close()
	addr := listener.Addr().(*net.TCPAddr)
	return strconv.Itoa(addr.Port), nil
}
