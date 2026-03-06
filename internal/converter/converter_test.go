package converter

import (
	"strings"
	"testing"
	"time"

	"github.com/awaketai/issue2md/internal/github"
	"github.com/awaketai/issue2md/internal/parser"
)

func TestToMarkdown(t *testing.T) {
	// 固定时间用于测试
	created := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commentTime1 := time.Date(2024, 1, 16, 8, 0, 0, 0, time.UTC)
	commentTime2 := time.Date(2024, 1, 17, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		data     github.IssueData
		resource parser.Resource
		opts     Options
		// 每个 want 元素都必须出现在输出中
		want []string
		// 每个 notWant 元素都不能出现在输出中
		notWant []string
	}{
		// === T-24: Issue 渲染测试 ===
		{
			name: "Case 1: basic Issue rendering",
			data: github.IssueData{
				Title:     "Fix login timeout",
				State:     "open",
				Author:    "octocat",
				CreatedAt: created,
				Body:      "Login times out after 30 seconds.",
			},
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypeIssue,
				Number: 12345,
			},
			opts: Options{WithReactions: false},
			want: []string{
				"# [Issue] Fix login timeout (#12345)",
				"| **State** | open |",
				"| **Author** | @octocat |",
				"| **Created** | 2024-01-15T10:30:00Z |",
				"---",
				"Login times out after 30 seconds.",
			},
			notWant: []string{
				"## Comments",
			},
		},
		{
			name: "Case 2: Issue with Labels/Assignees/Milestone",
			data: github.IssueData{
				Title:     "Bug report",
				State:     "closed",
				Author:    "dev1",
				CreatedAt: created,
				Labels:    []string{"bug", "priority/high"},
				Assignees: []string{"dev1", "dev2"},
				Milestone: "v2.0",
				Body:      "Some body.",
			},
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypeIssue,
				Number: 100,
			},
			opts: Options{},
			want: []string{
				"| **Labels** | bug, priority/high |",
				"| **Assignees** | @dev1, @dev2 |",
				"| **Milestone** | v2.0 |",
			},
		},
		{
			name: "Case 3: Issue with LinkedPRs",
			data: github.IssueData{
				Title:     "Feature request",
				State:     "open",
				Author:    "user1",
				CreatedAt: created,
				LinkedPRs: []int{456, 789},
				Body:      "Body text.",
			},
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypeIssue,
				Number: 200,
			},
			opts: Options{},
			want: []string{
				"| **Linked PRs** | #456, #789 |",
			},
		},
		{
			name: "Case 4: Issue without Milestone — no Milestone row",
			data: github.IssueData{
				Title:     "Simple issue",
				State:     "open",
				Author:    "user1",
				CreatedAt: created,
				Body:      "Body.",
			},
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypeIssue,
				Number: 300,
			},
			opts: Options{},
			notWant: []string{
				"Milestone",
				"Linked PRs",
			},
		},
		{
			name: "Case 5: Issue with comments",
			data: github.IssueData{
				Title:     "Issue with comments",
				State:     "open",
				Author:    "octocat",
				CreatedAt: created,
				Body:      "Main body.",
				Comments: []github.Comment{
					{
						Author:    "user1",
						CreatedAt: commentTime1,
						Body:      "First comment.",
					},
					{
						Author:    "user2",
						CreatedAt: commentTime2,
						Body:      "Second comment.",
					},
				},
			},
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypeIssue,
				Number: 400,
			},
			opts: Options{},
			want: []string{
				"## Comments (2)",
				"### @user1 — 2024-01-16T08:00:00Z",
				"First comment.",
				"### @user2 — 2024-01-17T14:30:00Z",
				"Second comment.",
			},
		},

		// === T-25: Reactions 渲染测试 ===
		{
			name: "Case 6: WithReactions=false — no reactions shown",
			data: github.IssueData{
				Title:     "No reactions shown",
				State:     "open",
				Author:    "user1",
				CreatedAt: created,
				Body:      "Body.",
				Reactions: github.Reactions{PlusOne: 5, Heart: 2},
				Comments: []github.Comment{
					{
						Author:    "commenter",
						CreatedAt: commentTime1,
						Body:      "Comment.",
						Reactions: github.Reactions{PlusOne: 3},
					},
				},
			},
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypeIssue,
				Number: 500,
			},
			opts: Options{WithReactions: false},
			notWant: []string{
				"👍",
				"❤️",
			},
		},
		{
			name: "Case 7: WithReactions=true — reactions displayed",
			data: github.IssueData{
				Title:     "With reactions",
				State:     "open",
				Author:    "user1",
				CreatedAt: created,
				Body:      "Body.",
				Reactions: github.Reactions{PlusOne: 5, Heart: 2},
				Comments: []github.Comment{
					{
						Author:    "commenter",
						CreatedAt: commentTime1,
						Body:      "Comment.",
						Reactions: github.Reactions{PlusOne: 3},
					},
				},
			},
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypeIssue,
				Number: 600,
			},
			opts: Options{WithReactions: true},
			want: []string{
				"> 👍 5 | ❤️ 2",
				"> 👍 3",
			},
		},
		{
			name: "Case 8: zero-count reactions not shown",
			data: github.IssueData{
				Title:     "Selective reactions",
				State:     "open",
				Author:    "user1",
				CreatedAt: created,
				Body:      "Body.",
				Reactions: github.Reactions{PlusOne: 5, Heart: 2},
			},
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypeIssue,
				Number: 700,
			},
			opts: Options{WithReactions: true},
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

		// === T-26: PR 和 Discussion 渲染测试 ===
		{
			name: "Case 9: PR rendering — title prefix is [PR]",
			data: github.IssueData{
				Title:     "Add retry logic",
				State:     "merged",
				Author:    "user2",
				CreatedAt: created,
				Labels:    []string{"enhancement"},
				Body:      "PR body.",
			},
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypePR,
				Number: 456,
			},
			opts: Options{},
			want: []string{
				"# [PR] Add retry logic (#456)",
				"| **State** | merged |",
			},
			notWant: []string{
				"Linked PRs",
			},
		},
		{
			name: "Case 10: PR Review Comment — shows file path and line",
			data: github.IssueData{
				Title:     "PR with review",
				State:     "open",
				Author:    "user2",
				CreatedAt: created,
				Body:      "Body.",
				Comments: []github.Comment{
					{
						Author:    "reviewer",
						CreatedAt: commentTime1,
						Body:      "Should we make this configurable?",
						FilePath:  "internal/auth/sso.go",
						Line:      42,
					},
				},
			},
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypePR,
				Number: 457,
			},
			opts: Options{},
			want: []string{
				"### @reviewer — 2024-01-16T08:00:00Z `internal/auth/sso.go#L42`",
			},
		},
		{
			name: "Case 11: Discussion rendering — title prefix is [Discussion]",
			data: github.IssueData{
				Title:     "RFC: New auth API",
				State:     "open",
				Author:    "architect",
				CreatedAt: created,
				Labels:    []string{"rfc"},
				Body:      "Discussion body.",
			},
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypeDiscussion,
				Number: 789,
			},
			opts: Options{},
			want: []string{
				"# [Discussion] RFC: New auth API (#789)",
			},
		},
		{
			name: "Case 12: empty comments — no Comments section",
			data: github.IssueData{
				Title:     "No comments",
				State:     "open",
				Author:    "user1",
				CreatedAt: created,
				Body:      "Body.",
				Comments:  nil,
			},
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypeIssue,
				Number: 800,
			},
			opts: Options{},
			notWant: []string{
				"## Comments",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToMarkdown(tt.data, tt.resource, tt.opts)

			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Errorf("output missing expected substring:\n  want: %q\n  got:\n%s", w, got)
				}
			}

			for _, nw := range tt.notWant {
				if strings.Contains(got, nw) {
					t.Errorf("output contains unexpected substring:\n  notWant: %q\n  got:\n%s", nw, got)
				}
			}
		})
	}
}
