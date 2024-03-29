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
	"fmt"
	"testing"
)

func TestAllMatcher(t *testing.T) {
	type TestCase struct {
		value string
		want  []string
	}

	testCases := []TestCase{
		{
			value: "${{ foo.bar }}",
			want: []string{
				"${{ foo.bar }}",
			},
		},
		{
			value: "${{ input.greeting }}",
			want: []string{
				"${{ input.greeting }}",
			},
		},
		{
			value: "${{ matrix.runtime }}",
			want: []string{
				"${{ matrix.runtime }}",
			},
		},
		{
			value: "${{ vars.command }}",
			want: []string{
				"${{ vars.command }}",
			},
		},
		{
			value: "${{ secrets.value }}",
			want: []string{
				"${{ secrets.value }}",
			},
		},
		{
			value: "${{ github.event.issue.title }}",
			want: []string{
				"${{ github.event.issue.title }}",
			},
		},
		{
			value: "${{ github.event.discussion.body }}",
			want: []string{
				"${{ github.event.discussion.body }}",
			},
		},
		{
			value: "${{ github.event.pages[0].page_name }}",
			want: []string{
				"${{ github.event.pages[0].page_name }}",
			},
		},
		{
			value: "${{ github.event.commits[1].author.email }}",
			want: []string{
				"${{ github.event.commits[1].author.email }}",
			},
		},
		{
			value: "${{ github.head_ref }}",
			want: []string{
				"${{ github.head_ref }}",
			},
		},
		{
			value: "${{ github.event.workflow_run.pull_requests[2].head.ref }}",
			want: []string{
				"${{ github.event.workflow_run.pull_requests[2].head.ref }}",
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.value, func(t *testing.T) {
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

	testCases := []TestCase{
		{
			value: "${{ github.event.issue.title }}",
			want: []string{
				"${{ github.event.issue.title }}",
			},
		},
		{
			value: "${{ github.event.issue.body }}",
			want: []string{
				"${{ github.event.issue.body }}",
			},
		},
		{
			value: "${{ github.event.discussion.title }}",
			want: []string{
				"${{ github.event.discussion.title }}",
			},
		},
		{
			value: "${{ github.event.discussion.body }}",
			want: []string{
				"${{ github.event.discussion.body }}",
			},
		},
		{
			value: "${{ github.event.comment.body }}",
			want: []string{
				"${{ github.event.comment.body }}",
			},
		},
		{
			value: "${{ github.event.review.body }}",
			want: []string{
				"${{ github.event.review.body }}",
			},
		},
		{
			value: "${{ github.event.review_comment.body }}",
			want: []string{
				"${{ github.event.review_comment.body }}",
			},
		},
		{
			value: "${{ github.event.pages[0].page_name }}",
			want: []string{
				"${{ github.event.pages[0].page_name }}",
			},
		},
		{
			value: "${{ github.event.commits[1].message }}",
			want: []string{
				"${{ github.event.commits[1].message }}",
			},
		},
		{
			value: "${{ github.event.commits[2].author.email }}",
			want: []string{
				"${{ github.event.commits[2].author.email }}",
			},
		},
		{
			value: "${{ github.event.commits[3].author.name }}",
			want: []string{
				"${{ github.event.commits[3].author.name }}",
			},
		},
		{
			value: "${{ github.event.head_commit.message }}",
			want: []string{
				"${{ github.event.head_commit.message }}",
			},
		},
		{
			value: "${{ github.event.head_commit.author.email }}",
			want: []string{
				"${{ github.event.head_commit.author.email }}",
			},
		},
		{
			value: "${{ github.event.head_commit.author.name }}",
			want: []string{
				"${{ github.event.head_commit.author.name }}",
			},
		},
		{
			value: "${{ github.event.head_commit.committer.email }}",
			want: []string{
				"${{ github.event.head_commit.committer.email }}",
			},
		},
		{
			value: "${{ github.event.workflow_run.head_branch }}",
			want: []string{
				"${{ github.event.workflow_run.head_branch }}",
			},
		},
		{
			value: "${{ github.event.workflow_run.head_commit.message }}",
			want: []string{
				"${{ github.event.workflow_run.head_commit.message }}",
			},
		},
		{
			value: "${{ github.event.workflow_run.head_commit.author.email }}",
			want: []string{
				"${{ github.event.workflow_run.head_commit.author.email }}",
			},
		},
		{
			value: "${{ github.event.workflow_run.head_commit.author.name }}",
			want: []string{
				"${{ github.event.workflow_run.head_commit.author.name }}",
			},
		},
		{
			value: "${{ github.event.pull_request.title }}",
			want: []string{
				"${{ github.event.pull_request.title }}",
			},
		},
		{
			value: "${{ github.event.pull_request.body }}",
			want: []string{
				"${{ github.event.pull_request.body }}",
			},
		},
		{
			value: "${{ github.event.pull_request.head.label }}",
			want: []string{
				"${{ github.event.pull_request.head.label }}",
			},
		},
		{
			value: "${{ github.event.pull_request.head.repo.default_branch }}",
			want: []string{
				"${{ github.event.pull_request.head.repo.default_branch }}",
			},
		},
		{
			value: "${{ github.head_ref }}",
			want: []string{
				"${{ github.head_ref }}",
			},
		},
		{
			value: "${{ github.event.pull_request.head.ref }}",
			want: []string{
				"${{ github.event.pull_request.head.ref }}",
			},
		},
		{
			value: "${{ github.event.workflow_run.pull_requests[4].head.ref }}",
			want: []string{
				"${{ github.event.workflow_run.pull_requests[4].head.ref }}",
			},
		},
		{
			value: "${{ input.greeting }}",
			want:  []string{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.value, func(t *testing.T) {
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
