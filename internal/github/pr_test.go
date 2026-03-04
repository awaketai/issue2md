package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// serveIssueForPR registers a handler that serves the Issue representation of a PR (for reactions).
func serveIssueForPR(mux *http.ServeMux, path string, reactions map[string]any) {
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		issue := map[string]any{
			"reactions": reactions,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(issue)
	})
}

func TestFetchPR(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mux *http.ServeMux)
		owner   string
		repo    string
		number  int
		want    IssueData
		wantErr bool
	}{
		{
			name: "normal PR with merged state",
			setup: func(mux *http.ServeMux) {
				mux.HandleFunc("/repos/golang/go/pulls/456", func(w http.ResponseWriter, r *http.Request) {
					pr := map[string]any{
						"title":      "Add feature X",
						"state":      "closed",
						"merged":     true,
						"body":       "PR body content",
						"created_at": "2024-03-10T09:00:00Z",
						"user":       map[string]any{"login": "alice"},
						"labels":     []map[string]any{{"name": "enhancement"}},
						"assignees":  []map[string]any{{"login": "bob"}},
						"milestone":  map[string]any{"title": "v2.0"},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(pr)
				})
				// Issue endpoint for reactions
				serveIssueForPR(mux, "/repos/golang/go/issues/456", map[string]any{
					"+1": 10, "-1": 0, "laugh": 0, "hooray": 3,
					"confused": 0, "heart": 0, "rocket": 5, "eyes": 0,
				})
				mux.HandleFunc("/repos/golang/go/issues/456/comments", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode([]map[string]any{})
				})
				mux.HandleFunc("/repos/golang/go/pulls/456/comments", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode([]map[string]any{})
				})
			},
			owner:  "golang",
			repo:   "go",
			number: 456,
			want: IssueData{
				Title:     "Add feature X",
				State:     "merged",
				Author:    "alice",
				CreatedAt: time.Date(2024, 3, 10, 9, 0, 0, 0, time.UTC),
				Labels:    []string{"enhancement"},
				Assignees: []string{"bob"},
				Milestone: "v2.0",
				Body:      "PR body content",
				Reactions: Reactions{PlusOne: 10, Hooray: 3, Rocket: 5},
				Comments:  []Comment{},
			},
		},
		{
			name: "PR with mixed issue comments and review comments sorted by time",
			setup: func(mux *http.ServeMux) {
				mux.HandleFunc("/repos/golang/go/pulls/789", func(w http.ResponseWriter, r *http.Request) {
					pr := map[string]any{
						"title":      "Refactor auth",
						"state":      "open",
						"merged":     false,
						"body":       "refactor body",
						"created_at": "2024-05-01T00:00:00Z",
						"user":       map[string]any{"login": "dev"},
						"labels":     []any{},
						"assignees":  []any{},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(pr)
				})
				serveIssueForPR(mux, "/repos/golang/go/issues/789", map[string]any{
					"+1": 0, "-1": 0, "laugh": 0, "hooray": 0,
					"confused": 0, "heart": 0, "rocket": 0, "eyes": 0,
				})
				mux.HandleFunc("/repos/golang/go/issues/789/comments", func(w http.ResponseWriter, r *http.Request) {
					comments := []map[string]any{
						{
							"user":       map[string]any{"login": "reviewer1"},
							"created_at": "2024-05-02T10:00:00Z",
							"body":       "Looks good overall",
							"reactions": map[string]any{
								"+1": 1, "-1": 0, "laugh": 0, "hooray": 0,
								"confused": 0, "heart": 0, "rocket": 0, "eyes": 0,
							},
						},
						{
							"user":       map[string]any{"login": "reviewer3"},
							"created_at": "2024-05-04T08:00:00Z",
							"body":       "LGTM",
							"reactions": map[string]any{
								"+1": 0, "-1": 0, "laugh": 0, "hooray": 0,
								"confused": 0, "heart": 0, "rocket": 0, "eyes": 0,
							},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(comments)
				})
				mux.HandleFunc("/repos/golang/go/pulls/789/comments", func(w http.ResponseWriter, r *http.Request) {
					comments := []map[string]any{
						{
							"user":       map[string]any{"login": "reviewer2"},
							"created_at": "2024-05-03T14:00:00Z",
							"body":       "Nit: rename this variable",
							"path":       "internal/auth/sso.go",
							"line":       42,
							"reactions": map[string]any{
								"+1": 0, "-1": 0, "laugh": 0, "hooray": 0,
								"confused": 0, "heart": 0, "rocket": 0, "eyes": 0,
							},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(comments)
				})
			},
			owner:  "golang",
			repo:   "go",
			number: 789,
			want: IssueData{
				Title:     "Refactor auth",
				State:     "open",
				Author:    "dev",
				CreatedAt: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
				Labels:    []string{},
				Assignees: []string{},
				Body:      "refactor body",
				Comments: []Comment{
					{
						Author:    "reviewer1",
						CreatedAt: time.Date(2024, 5, 2, 10, 0, 0, 0, time.UTC),
						Body:      "Looks good overall",
						Reactions: Reactions{PlusOne: 1},
					},
					{
						Author:    "reviewer2",
						CreatedAt: time.Date(2024, 5, 3, 14, 0, 0, 0, time.UTC),
						Body:      "Nit: rename this variable",
						FilePath:  "internal/auth/sso.go",
						Line:      42,
					},
					{
						Author:    "reviewer3",
						CreatedAt: time.Date(2024, 5, 4, 8, 0, 0, 0, time.UTC),
						Body:      "LGTM",
					},
				},
			},
		},
		{
			name: "PR with only review comments and no regular comments",
			setup: func(mux *http.ServeMux) {
				mux.HandleFunc("/repos/golang/go/pulls/500", func(w http.ResponseWriter, r *http.Request) {
					pr := map[string]any{
						"title":      "Small fix",
						"state":      "open",
						"merged":     false,
						"body":       "fix body",
						"created_at": "2024-07-01T00:00:00Z",
						"user":       map[string]any{"login": "dev"},
						"labels":     []any{},
						"assignees":  []any{},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(pr)
				})
				serveIssueForPR(mux, "/repos/golang/go/issues/500", map[string]any{
					"+1": 0, "-1": 0, "laugh": 0, "hooray": 0,
					"confused": 0, "heart": 0, "rocket": 0, "eyes": 0,
				})
				mux.HandleFunc("/repos/golang/go/issues/500/comments", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode([]map[string]any{})
				})
				mux.HandleFunc("/repos/golang/go/pulls/500/comments", func(w http.ResponseWriter, r *http.Request) {
					comments := []map[string]any{
						{
							"user":       map[string]any{"login": "reviewer"},
							"created_at": "2024-07-02T00:00:00Z",
							"body":       "Please fix this",
							"path":       "main.go",
							"line":       10,
							"reactions": map[string]any{
								"+1": 0, "-1": 0, "laugh": 0, "hooray": 0,
								"confused": 0, "heart": 0, "rocket": 0, "eyes": 0,
							},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(comments)
				})
			},
			owner:  "golang",
			repo:   "go",
			number: 500,
			want: IssueData{
				Title:     "Small fix",
				State:     "open",
				Author:    "dev",
				CreatedAt: time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
				Labels:    []string{},
				Assignees: []string{},
				Body:      "fix body",
				Comments: []Comment{
					{
						Author:    "reviewer",
						CreatedAt: time.Date(2024, 7, 2, 0, 0, 0, 0, time.UTC),
						Body:      "Please fix this",
						FilePath:  "main.go",
						Line:      10,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			tt.setup(mux)
			server := httptest.NewServer(mux)
			defer server.Close()

			client := newTestClient(server.URL)
			got, err := client.FetchPR(context.Background(), tt.owner, tt.repo, tt.number)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("FetchPR() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("FetchPR() unexpected error: %v", err)
			}

			assertIssueDataEqual(t, got, tt.want)
		})
	}
}
