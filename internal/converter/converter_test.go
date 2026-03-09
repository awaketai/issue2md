package converter

import (
	"strings"
	"testing"
	"time"

	"github.com/awaketai/issue2md/internal/github"
	"github.com/awaketai/issue2md/internal/parser"
)

func TestToMarkdown(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commentTime1 := time.Date(2024, 1, 16, 8, 0, 0, 0, time.UTC)
	commentTime2 := time.Date(2024, 1, 17, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		data     github.IssueData
		resource parser.Resource
		opts     Options
		want     []string // substrings that must appear in output
		notWant  []string // substrings that must NOT appear in output
	}{
		// --- T-24: Issue rendering ---
		{
			name: "Case 1: basic Issue",
			data: github.IssueData{
				Title:     "Fix login timeout",
				State:     "open",
				Author:    "octocat",
				CreatedAt: baseTime,
				Body:      "The login page times out after 30 seconds.",
			},
			resource: parser.Resource{Owner: "golang", Repo: "go", Type: parser.TypeIssue, Number: 12345},
			opts:     Options{},
			want: []string{
				"# [Issue] Fix login timeout (#12345)",
				"| **State** | open |",
				"| **Author** | @octocat |",
				"| **Created** | 2024-01-15T10:30:00Z |",
				"---",
				"The login page times out after 30 seconds.",
			},
		},
		{
			name: "Case 2: Issue with Labels/Assignees/Milestone",
			data: github.IssueData{
				Title:     "Bug report",
				State:     "open",
				Author:    "dev",
				CreatedAt: baseTime,
				Labels:    []string{"bug", "priority/high"},
				Assignees: []string{"dev1", "dev2"},
				Milestone: "v2.0",
				Body:      "Some body.",
			},
			resource: parser.Resource{Owner: "o", Repo: "r", Type: parser.TypeIssue, Number: 1},
			opts:     Options{},
			want: []string{
				"| **Labels** | bug, priority/high |",
				"| **Assignees** | @dev1, @dev2 |",
				"| **Milestone** | v2.0 |",
			},
		},
		{
			name: "Case 3: Issue with LinkedPRs",
			data: github.IssueData{
				Title:     "Issue with PRs",
				State:     "open",
				Author:    "dev",
				CreatedAt: baseTime,
				LinkedPRs: []int{456, 789},
				Body:      "Body.",
			},
			resource: parser.Resource{Owner: "o", Repo: "r", Type: parser.TypeIssue, Number: 1},
			opts:     Options{},
			want: []string{
				"| **Linked PRs** | #456, #789 |",
			},
		},
		{
			name: "Case 4: Issue without Milestone",
			data: github.IssueData{
				Title:     "No milestone",
				State:     "open",
				Author:    "dev",
				CreatedAt: baseTime,
				Body:      "Body.",
			},
			resource: parser.Resource{Owner: "o", Repo: "r", Type: parser.TypeIssue, Number: 1},
			opts:     Options{},
			notWant: []string{
				"Milestone",
			},
		},
		{
			name: "Case 5: Issue with comments",
			data: github.IssueData{
				Title:     "With comments",
				State:     "open",
				Author:    "dev",
				CreatedAt: baseTime,
				Body:      "Body.",
				Comments: []github.Comment{
					{Author: "user1", CreatedAt: commentTime1, Body: "First comment."},
					{Author: "user2", CreatedAt: commentTime2, Body: "Second comment."},
				},
			},
			resource: parser.Resource{Owner: "o", Repo: "r", Type: parser.TypeIssue, Number: 1},
			opts:     Options{},
			want: []string{
				"## Comments (2)",
				"### @user1 — 2024-01-16T08:00:00Z",
				"First comment.",
				"### @user2 — 2024-01-17T14:30:00Z",
				"Second comment.",
			},
		},

		// --- T-25: Reactions rendering ---
		{
			name: "Case 6: WithReactions=false hides reactions",
			data: github.IssueData{
				Title:     "No reactions shown",
				State:     "open",
				Author:    "dev",
				CreatedAt: baseTime,
				Body:      "Body.",
				Reactions: github.Reactions{PlusOne: 5, Heart: 2},
			},
			resource: parser.Resource{Owner: "o", Repo: "r", Type: parser.TypeIssue, Number: 1},
			opts:     Options{WithReactions: false},
			notWant: []string{
				"👍",
				"❤️",
			},
		},
		{
			name: "Case 7: WithReactions=true shows reactions",
			data: github.IssueData{
				Title:     "With reactions",
				State:     "open",
				Author:    "dev",
				CreatedAt: baseTime,
				Body:      "Body.",
				Reactions: github.Reactions{PlusOne: 5, Heart: 2},
				Comments: []github.Comment{
					{
						Author:    "user1",
						CreatedAt: commentTime1,
						Body:      "Nice!",
						Reactions: github.Reactions{PlusOne: 3},
					},
				},
			},
			resource: parser.Resource{Owner: "o", Repo: "r", Type: parser.TypeIssue, Number: 1},
			opts:     Options{WithReactions: true},
			want: []string{
				"> 👍 5 | ❤️ 2",
				"> 👍 3",
			},
		},
		{
			name: "Case 8: zero-count reactions are hidden",
			data: github.IssueData{
				Title:     "Partial reactions",
				State:     "open",
				Author:    "dev",
				CreatedAt: baseTime,
				Body:      "Body.",
				Reactions: github.Reactions{PlusOne: 5, Heart: 2},
			},
			resource: parser.Resource{Owner: "o", Repo: "r", Type: parser.TypeIssue, Number: 1},
			opts:     Options{WithReactions: true},
			want: []string{
				"> 👍 5 | ❤️ 2",
			},
			notWant: []string{
				"👎",
				"😄",
				"🎉",
				"😕",
				"🚀",
				"👀",
			},
		},

		// --- T-26: PR and Discussion rendering ---
		{
			name: "Case 9: PR rendering",
			data: github.IssueData{
				Title:     "Add retry logic",
				State:     "merged",
				Author:    "user2",
				CreatedAt: baseTime,
				Body:      "This PR adds retry.",
			},
			resource: parser.Resource{Owner: "o", Repo: "r", Type: parser.TypePR, Number: 456},
			opts:     Options{},
			want: []string{
				"# [PR] Add retry logic (#456)",
			},
			notWant: []string{
				"Linked PRs",
			},
		},
		{
			name: "Case 10: PR Review Comment with file path and line",
			data: github.IssueData{
				Title:     "PR with review",
				State:     "open",
				Author:    "dev",
				CreatedAt: baseTime,
				Body:      "Body.",
				Comments: []github.Comment{
					{
						Author:    "reviewer",
						CreatedAt: commentTime1,
						Body:      "Should this be configurable?",
						FilePath:  "internal/auth/sso.go",
						Line:      42,
					},
				},
			},
			resource: parser.Resource{Owner: "o", Repo: "r", Type: parser.TypePR, Number: 1},
			opts:     Options{},
			want: []string{
				"### @reviewer — 2024-01-16T08:00:00Z `internal/auth/sso.go#L42`",
			},
		},
		{
			name: "Case 11: Discussion rendering",
			data: github.IssueData{
				Title:     "RFC: New auth design",
				State:     "open",
				Author:    "architect",
				CreatedAt: baseTime,
				Body:      "Proposal here.",
			},
			resource: parser.Resource{Owner: "o", Repo: "r", Type: parser.TypeDiscussion, Number: 789},
			opts:     Options{},
			want: []string{
				"# [Discussion] RFC: New auth design (#789)",
			},
		},
		{
			name: "Case 12: empty comments list omits Comments section",
			data: github.IssueData{
				Title:     "No comments",
				State:     "open",
				Author:    "dev",
				CreatedAt: baseTime,
				Body:      "Body.",
				Comments:  nil,
			},
			resource: parser.Resource{Owner: "o", Repo: "r", Type: parser.TypeIssue, Number: 1},
			opts:     Options{},
			notWant: []string{
				"## Comments",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToMarkdown(tt.data, tt.resource, tt.opts)

			for _, s := range tt.want {
				if !strings.Contains(got, s) {
					t.Errorf("output missing expected substring %q\n\ngot:\n%s", s, got)
				}
			}

			for _, s := range tt.notWant {
				if strings.Contains(got, s) {
					t.Errorf("output should not contain %q\n\ngot:\n%s", s, got)
				}
			}
		})
	}
}
