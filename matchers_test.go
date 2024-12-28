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
	"fmt"
	"testing"
)

func TestAllMatcher(t *testing.T) {
	type TestCase struct {
		value string
		want  []string
	}

	testCases := map[string]TestCase{
		"not an expression": {
			value: `'Hello world'`,
			want:  nil,
		},
		"safe expression": {
			value: `${{ true }}`,
			want:  nil,
		},
		"safe use of an unsafe expression": {
			value: `${{ contains(github.event.issue.title, 'foobar') }}`,
			want:  nil,
		},
		"unsafe expression ": {
			value: `${{ foo.bar }}`,
			want: []string{
				`${{ foo.bar }}`,
			},
		},
		"input expression": {
			value: `${{ input.greeting }}`,
			want: []string{
				`${{ input.greeting }}`,
			},
		},
		"matrix expression": {
			value: `${{ matrix.runtime }}`,
			want: []string{
				`${{ matrix.runtime }}`,
			},
		},
		"vars expression": {
			value: `${{ vars.command }}`,
			want: []string{
				`${{ vars.command }}`,
			},
		},
		"secrets expression": {
			value: `${{ secrets.value }}`,
			want: []string{
				`${{ secrets.value }}`,
			},
		},
		"github.event.issue.title": {
			value: `${{ github.event.issue.title }}`,
			want: []string{
				`${{ github.event.issue.title }}`,
			},
		},
		"github.event.discussion.body": {
			value: `${{ github.event.discussion.body }}`,
			want: []string{
				`${{ github.event.discussion.body }}`,
			},
		},
		"github.event.pages[*].page_name": {
			value: `${{ github.event.pages[0].page_name }}`,
			want: []string{
				`${{ github.event.pages[0].page_name }}`,
			},
		},
		"github.event.commits[*].author.email": {
			value: `${{ github.event.commits[1].author.email }}`,
			want: []string{
				`${{ github.event.commits[1].author.email }}`,
			},
		},
		"github.head_ref": {
			value: `${{ github.head_ref }}`,
			want: []string{
				`${{ github.head_ref }}`,
			},
		},
		"github.event.workflow_run.pull_requests[*].head.ref": {
			value: `${{ github.event.workflow_run.pull_requests[2].head.ref }}`,
			want: []string{
				`${{ github.event.workflow_run.pull_requests[2].head.ref }}`,
			},
		},
		"safe & unsafe in one expression": {
			value: `${{ foo.bar || true }}`,
			want: []string{
				`${{ foo.bar || true }}`,
			},
		},
		"unsafe & safe in one expression": {
			value: `${{ false || foo.baz }}`,
			want: []string{
				`${{ false || foo.baz }}`,
			},
		},
		"safe & safe in one expression": {
			value: `echo ${{ false || true }}`,
			want:  nil,
		},
		"unsafe & unsafe in one expression": {
			value: `${{ foo.bar || foo.baz }}`,
			want: []string{
				`${{ foo.bar || foo.baz }}`,
			},
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := fmt.Sprintf("echo '%s'", tt.value)

			matches := AllMatcher.FindAll([]byte(s))
			if got, want := len(matches), len(tt.want); got != want {
				t.Fatalf("Unexpected number of violations (got %d, want %d)", got, want)
			}

			for i, match := range matches {
				if got, want := string(match), tt.want[i]; got != want {
					t.Errorf("Unexpected #%d violation (got %q, want %q)", i, got, want)
				}
			}
		})
	}
}

func TestConservativeMatcher(t *testing.T) {
	type TestCase struct {
		value string
		want  []string
	}

	testCases := map[string]TestCase{
		"not an expression": {
			value: `'Hello world'`,
			want:  nil,
		},
		"safe expression": {
			value: `${{ true }}`,
			want:  nil,
		},
		"conservatively safe expression": {
			value: `${{ input.greeting }}`,
			want:  nil,
		},
		"safe use of conservatively unsafe expression": {
			value: `${{ contains(github.event.issue.title, 'foobar') }}`,
			want:  nil,
		},
		"github.event.issue.title": {
			value: `${{ github.event.issue.title }}`,
			want: []string{
				`${{ github.event.issue.title }}`,
			},
		},
		"github.event.issue.body": {
			value: `${{ github.event.issue.body }}`,
			want: []string{
				`${{ github.event.issue.body }}`,
			},
		},
		"github.event.discussion.title": {
			value: `${{ github.event.discussion.title }}`,
			want: []string{
				`${{ github.event.discussion.title }}`,
			},
		},
		"github.event.discussion.body": {
			value: `${{ github.event.discussion.body }}`,
			want: []string{
				`${{ github.event.discussion.body }}`,
			},
		},
		"github.event.comment.body": {
			value: `${{ github.event.comment.body }}`,
			want: []string{
				`${{ github.event.comment.body }}`,
			},
		},
		"github.event.review.body": {
			value: `${{ github.event.review.body }}`,
			want: []string{
				`${{ github.event.review.body }}`,
			},
		},
		"github.event.review_comment.body": {
			value: `${{ github.event.review_comment.body }}`,
			want: []string{
				`${{ github.event.review_comment.body }}`,
			},
		},
		"github.event.pages[*].page_name": {
			value: `${{ github.event.pages[0].page_name }}`,
			want: []string{
				`${{ github.event.pages[0].page_name }}`,
			},
		},
		"github.event.commits[*].message": {
			value: `${{ github.event.commits[1].message }}`,
			want: []string{
				`${{ github.event.commits[1].message }}`,
			},
		},
		"github.event.commits[*].author.email": {
			value: `${{ github.event.commits[2].author.email }}`,
			want: []string{
				`${{ github.event.commits[2].author.email }}`,
			},
		},
		"github.event.commits[*].author.name": {
			value: `${{ github.event.commits[3].author.name }}`,
			want: []string{
				`${{ github.event.commits[3].author.name }}`,
			},
		},
		"github.event.head_commit.message": {
			value: `${{ github.event.head_commit.message }}`,
			want: []string{
				`${{ github.event.head_commit.message }}`,
			},
		},
		"github.event.head_commit.author.email": {
			value: `${{ github.event.head_commit.author.email }}`,
			want: []string{
				`${{ github.event.head_commit.author.email }}`,
			},
		},
		"github.event.head_commit.author.name": {
			value: `${{ github.event.head_commit.author.name }}`,
			want: []string{
				`${{ github.event.head_commit.author.name }}`,
			},
		},
		"github.event.head_commit.committer.email": {
			value: `${{ github.event.head_commit.committer.email }}`,
			want: []string{
				`${{ github.event.head_commit.committer.email }}`,
			},
		},
		"github.event.workflow_run.head_branch": {
			value: `${{ github.event.workflow_run.head_branch }}`,
			want: []string{
				`${{ github.event.workflow_run.head_branch }}`,
			},
		},
		"github.event.workflow_run.head_commit.message": {
			value: `${{ github.event.workflow_run.head_commit.message }}`,
			want: []string{
				`${{ github.event.workflow_run.head_commit.message }}`,
			},
		},
		"github.event.workflow_run.head_commit.author.email": {
			value: `${{ github.event.workflow_run.head_commit.author.email }}`,
			want: []string{
				`${{ github.event.workflow_run.head_commit.author.email }}`,
			},
		},
		"github.event.workflow_run.head_commit.author.name": {
			value: `${{ github.event.workflow_run.head_commit.author.name }}`,
			want: []string{
				`${{ github.event.workflow_run.head_commit.author.name }}`,
			},
		},
		"github.event.pull_request.title": {
			value: `${{ github.event.pull_request.title }}`,
			want: []string{
				`${{ github.event.pull_request.title }}`,
			},
		},
		"github.event.pull_request.body": {
			value: `${{ github.event.pull_request.body }}`,
			want: []string{
				`${{ github.event.pull_request.body }}`,
			},
		},
		"github.event.pull_request.head.label": {
			value: `${{ github.event.pull_request.head.label }}`,
			want: []string{
				`${{ github.event.pull_request.head.label }}`,
			},
		},
		"github.event.pull_request.head.repo.default_branch": {
			value: `${{ github.event.pull_request.head.repo.default_branch }}`,
			want: []string{
				`${{ github.event.pull_request.head.repo.default_branch }}`,
			},
		},
		"github.head_ref": {
			value: `${{ github.head_ref }}`,
			want: []string{
				`${{ github.head_ref }}`,
			},
		},
		"github.event.pull_request.head.ref": {
			value: `${{ github.event.pull_request.head.ref }}`,
			want: []string{
				`${{ github.event.pull_request.head.ref }}`,
			},
		},
		"github.event.workflow_run.pull_requests[*].head.ref": {
			value: `${{ github.event.workflow_run.pull_requests[4].head.ref }}`,
			want: []string{
				`${{ github.event.workflow_run.pull_requests[4].head.ref }}`,
			},
		},

		"(conservatively) safe & unsafe in one expression": {
			value: `${{ foo.bar || github.head_ref }}`,
			want: []string{
				`${{ foo.bar || github.head_ref }}`,
			},
		},
		"unsafe & (conservatively) safe in one expression": {
			value: `${{ github.event.pull_request.head.ref || inputs.backup }}`,
			want: []string{
				`${{ github.event.pull_request.head.ref || inputs.backup }}`,
			},
		},
		"(conservatively) safe & (conservatively) safe in one expression": {
			value: `echo ${{ foo.bar || inputs.backup }}`,
			want:  nil,
		},
		"unsafe & unsafe in one expression": {
			value: `${{ github.event.pull_request.head.ref || github.head_ref }}`,
			want: []string{
				`${{ github.event.pull_request.head.ref || github.head_ref }}`,
			},
		},

		"unsafe and unsafe in one expression": {},
		"two, only one is dangerous":          {},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := fmt.Sprintf("echo '%s'", tt.value)

			matches := ConservativeMatcher.FindAll([]byte(s))
			if got, want := len(matches), len(tt.want); got != want {
				t.Fatalf("Unexpected number of violations (got %d, want %d)", got, want)
			}

			for i, match := range matches {
				if got, want := string(match), tt.want[i]; got != want {
					t.Errorf("Unexpected #%d violation (got %q, want %q)", i, got, want)
				}
			}
		})
	}
}

func TestDefinitelySafe(t *testing.T) {
	type TestCase struct {
		value string
		want  string
	}

	testCases := map[string]TestCase{
		"literal, boolean, true": {
			value: `echo ${{ true }}`,
			want:  `echo `,
		},
		"literal, boolean, false": {
			value: `echo ${{ false }}`,
			want:  `echo `,
		},
		"literal, null": {
			value: `echo ${{ null }}`,
			want:  `echo `,
		},
		"literal, number, positive integer": {
			value: `echo ${{ 42 }}`,
			want:  `echo `,
		},
		"literal, number, negative integer": {
			value: `echo ${{ -36 }}`,
			want:  `echo `,
		},
		"literal, number, float": {
			value: `echo ${{ 3.14 }}`,
			want:  `echo `,
		},
		"literal, number, hexadecimal (uppercase)": {
			value: `echo ${{ 0x2A }}`,
			want:  `echo `,
		},
		"literal, number, hexadecimal (lowercase)": {
			value: `echo ${{ 0x2a }}`,
			want:  `echo `,
		},
		"literal, number, hexadecimal (mixed case)": {
			value: `echo ${{ 0xDeAdBeEf }}`,
			want:  `echo `,
		},
		"literal, number, scientific notation, positive, small": {
			value: `echo ${{ 2.99e-2 }}`,
			want:  `echo `,
		},
		"literal, number, scientific notation, positive, big": {
			value: `echo ${{ 2.99e2 }}`,
			want:  `echo `,
		},
		"literal, number, scientific notation, negative, small": {
			value: `echo ${{ -2.99e-2 }}`,
			want:  `echo `,
		},
		"literal, number, scientific notation, negative, big": {
			value: `echo ${{ -2.99e2 }}`,
			want:  `echo `,
		},
		"literal, string": {
			value: `echo ${{ 'Hello world' }}`,
			want:  `echo `,
		},
		"literal, multiple": {
			value: `echo ${{ true || 42 }}`,
			want:  `echo `,
		},
		"function, always": {
			value: `echo ${{ always() }}`,
			want:  `echo `,
		},
		"function, cancelled": {
			value: `echo 'The job was cancelled: ${{ cancelled() }}'`,
			want:  `echo 'The job was cancelled: '`,
		},
		"function, contains, no variables": {
			value: `echo 'Does the string contain llo: ${{ contains('Hello world', 'llo') }}'`,
			want:  `echo 'Does the string contain llo: '`,
		},
		"function, contains, with variables": {
			value: `echo 'Does ${{ inputs.greeting }} contain llo: ${{ contains(inputs.greeting, 'llo') }}'`,
			want:  `echo 'Does ${{ inputs.greeting }} contain llo: '`,
		},
		"function, endsWith, no variables": {
			value: `echo 'Does the string end with llo: ${{ endsWith('Hello world', 'llo') }}'`,
			want:  `echo 'Does the string end with llo: '`,
		},
		"function, endsWith, with variables": {
			value: `echo 'Does ${{ inputs.greeting }} end with llo: ${{ endsWith(inputs.greeting, 'llo') }}'`,
			want:  `echo 'Does ${{ inputs.greeting }} end with llo: '`,
		},
		"function, failure": {
			value: `echo 'The job failed: ${{ failure() }}'`,
			want:  `echo 'The job failed: '`,
		},
		"function, format, no variables": {
			value: `echo ${{ format('Hello {0}', 'world') }}`,
			want:  `echo `,
		},
		"function, format, with variables": {
			value: `echo ${{ format('Hello {0}', inputs.who) }}`,
			want:  `echo ${{ format(, inputs.who) }}`,
		},
		"function, fromJSON, no variables": {
			value: `obj=${{ fromJSON('["foo", "bar"]') }}`,
			want:  `obj=`,
		},
		"function, fromJSON, with variables": {
			value: `obj=${{ fromJSON(inputs.json) }}`,
			want:  `obj=${{ fromJSON(inputs.json) }}`,
		},
		"function, hashFiles, no variables": {
			value: `echo 'hash: ${{ hashFiles('**/*.go') }}'`,
			want:  `echo 'hash: '`,
		},
		"function, hashFiles, with variables": {
			value: `echo 'hash: ${{ hashFiles(input.files) }}'`,
			want:  `echo 'hash: '`,
		},
		"function, join, no variables": {
			value: `echo 'The elements are: ${{ join(fromJSON('["foo", "bar"]'), ', ') }}'`,
			want:  `echo 'The elements are: '`,
		},
		"function, join, with variables": {
			value: `echo 'The elements are: ${{ join(inputs.list, ', ') }}'`,
			want:  `echo 'The elements are: ${{ join(inputs.list, ) }}'`,
		},
		"function, startsWith, no variables": {
			value: `echo 'Does the string start with llo: ${{ startsWith('Hello world', 'llo') }}'`,
			want:  `echo 'Does the string start with llo: '`,
		},
		"function, startsWith, with variables": {
			value: `echo 'Does ${{ inputs.greeting }} start with llo: ${{ startsWith(inputs.greeting, 'llo') }}'`,
			want:  `echo 'Does ${{ inputs.greeting }} start with llo: '`,
		},
		"function, success": {
			value: `echo 'The job succeeded: ${{ success() }}'`,
			want:  `echo 'The job succeeded: '`,
		},
		"function, toJSON, with variables": {
			value: `json=${{ toJSON(inputs.value) }}`,
			want:  `json=${{ toJSON(inputs.value) }}`,
		},
		"edge case, literals with odd spacing": {
			value: `echo ${{true}}${{false }}${{ null}}${{ (0x2A) }}${{ false||1 }}${{ true&&0 }}${{ 0==false }}${{ 4!=2 }}${{ 3>14 }}${{ 14>=3 }}${{ 2<718 }}${{ 718<=2 }}${{ !true }}`,
			want:  `echo `,
		},
		"edge case, functions with odd spacing": {
			value: `echo ${{success( )}}${{failure() }}${{ always()}}${{ (cancelled()) }}${{ startsWith('foobar', 'foo')&&endsWith('foobar', 'bar') }}${{ startsWith('foobar', 'foo')||endsWith('foobar', 'baz') }}`,
			want:  `echo `,
		},
		"edge case, identifier like a literal": {
			value: `echo ${{ trueish }}`,
			want:  `echo ${{ trueish }}`,
		},
		"edge case, identifier like a function": {
			value: `echo ${{ contained('foo', 'bar') }}`,
			want:  `echo ${{ contained(, ) }}`,
		},
	}

	for name, tt := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := []byte(tt.value)
			if got, want := stripSafe(v), []byte(tt.want); !bytes.Equal(got, want) {
				t.Errorf("Unexpected result (got %q, want %q)", got, want)
			}
		})
	}
}
