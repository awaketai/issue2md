package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gogithub "github.com/google/go-github/v68/github"
)

// newTestClient creates a Client whose REST API points at the given httptest server.
func newTestClient(serverURL string) *Client {
	restClient := gogithub.NewClient(nil)
	restClient.BaseURL, _ = restClient.BaseURL.Parse(serverURL + "/")

	return &Client{
		restClient: restClient,
		verbose:    nil,
	}
}

func TestFetchIssue(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(mux *http.ServeMux)
		owner      string
		repo       string
		number     int
		want       IssueData
		wantErr    bool
	}{
		{
			name: "normal issue with metadata",
			setup: func(mux *http.ServeMux) {
				// GET /repos/{owner}/{repo}/issues/{number}
				mux.HandleFunc("/repos/golang/go/issues/12345", func(w http.ResponseWriter, r *http.Request) {
					issue := map[string]any{
						"title":      "Fix the bug",
						"state":      "open",
						"body":       "This is the body",
						"created_at": "2024-01-15T10:30:00Z",
						"user":       map[string]any{"login": "alice"},
						"labels":     []map[string]any{{"name": "bug"}, {"name": "priority"}},
						"assignees":  []map[string]any{{"login": "bob"}, {"login": "carol"}},
						"milestone":  map[string]any{"title": "v1.0"},
						"reactions": map[string]any{
							"+1": 5, "-1": 1, "laugh": 2, "hooray": 0,
							"confused": 0, "heart": 3, "rocket": 1, "eyes": 0,
						},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(issue)
				})
				// GET /repos/{owner}/{repo}/issues/{number}/comments
				mux.HandleFunc("/repos/golang/go/issues/12345/comments", func(w http.ResponseWriter, r *http.Request) {
					comments := []map[string]any{}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(comments)
				})
			},
			owner:  "golang",
			repo:   "go",
			number: 12345,
			want: IssueData{
				Title:     "Fix the bug",
				State:     "open",
				Author:    "alice",
				CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Labels:    []string{"bug", "priority"},
				Assignees: []string{"bob", "carol"},
				Milestone: "v1.0",
				Body:      "This is the body",
				Reactions: Reactions{
					PlusOne: 5, MinusOne: 1, Laugh: 2, Heart: 3, Rocket: 1,
				},
				Comments: []Comment{},
			},
		},
		{
			name: "issue with comments sorted by time",
			setup: func(mux *http.ServeMux) {
				mux.HandleFunc("/repos/golang/go/issues/100", func(w http.ResponseWriter, r *http.Request) {
					issue := map[string]any{
						"title":      "Test issue",
						"state":      "closed",
						"body":       "body",
						"created_at": "2024-06-01T00:00:00Z",
						"user":       map[string]any{"login": "author"},
						"labels":     []any{},
						"assignees":  []any{},
						"reactions": map[string]any{
							"+1": 0, "-1": 0, "laugh": 0, "hooray": 0,
							"confused": 0, "heart": 0, "rocket": 0, "eyes": 0,
						},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(issue)
				})
				mux.HandleFunc("/repos/golang/go/issues/100/comments", func(w http.ResponseWriter, r *http.Request) {
					comments := []map[string]any{
						{
							"user":       map[string]any{"login": "user1"},
							"created_at": "2024-06-02T10:00:00Z",
							"body":       "First comment",
							"reactions": map[string]any{
								"+1": 1, "-1": 0, "laugh": 0, "hooray": 0,
								"confused": 0, "heart": 0, "rocket": 0, "eyes": 0,
							},
						},
						{
							"user":       map[string]any{"login": "user2"},
							"created_at": "2024-06-03T12:00:00Z",
							"body":       "Second comment",
							"reactions": map[string]any{
								"+1": 0, "-1": 0, "laugh": 0, "hooray": 0,
								"confused": 0, "heart": 2, "rocket": 0, "eyes": 0,
							},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(comments)
				})
			},
			owner:  "golang",
			repo:   "go",
			number: 100,
			want: IssueData{
				Title:     "Test issue",
				State:     "closed",
				Author:    "author",
				CreatedAt: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				Labels:    []string{},
				Assignees: []string{},
				Body:      "body",
				Comments: []Comment{
					{
						Author:    "user1",
						CreatedAt: time.Date(2024, 6, 2, 10, 0, 0, 0, time.UTC),
						Body:      "First comment",
						Reactions: Reactions{PlusOne: 1},
					},
					{
						Author:    "user2",
						CreatedAt: time.Date(2024, 6, 3, 12, 0, 0, 0, time.UTC),
						Body:      "Second comment",
						Reactions: Reactions{Heart: 2},
					},
				},
			},
		},
		{
			name: "404 not found",
			setup: func(mux *http.ServeMux) {
				mux.HandleFunc("/repos/golang/go/issues/99999", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					fmt.Fprint(w, `{"message":"Not Found"}`)
				})
			},
			owner:   "golang",
			repo:    "go",
			number:  99999,
			wantErr: true,
		},
		{
			name: "paginated comments",
			setup: func(mux *http.ServeMux) {
				mux.HandleFunc("/repos/golang/go/issues/200", func(w http.ResponseWriter, r *http.Request) {
					issue := map[string]any{
						"title":      "Paginated",
						"state":      "open",
						"body":       "body",
						"created_at": "2024-01-01T00:00:00Z",
						"user":       map[string]any{"login": "alice"},
						"labels":     []any{},
						"assignees":  []any{},
						"reactions": map[string]any{
							"+1": 0, "-1": 0, "laugh": 0, "hooray": 0,
							"confused": 0, "heart": 0, "rocket": 0, "eyes": 0,
						},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(issue)
				})
				mux.HandleFunc("/repos/golang/go/issues/200/comments", func(w http.ResponseWriter, r *http.Request) {
					page := r.URL.Query().Get("page")
					if page == "" || page == "1" {
						// Page 1: return one comment + Link header pointing to page 2
						comments := []map[string]any{
							{
								"user":       map[string]any{"login": "page1user"},
								"created_at": "2024-01-02T00:00:00Z",
								"body":       "Comment from page 1",
								"reactions": map[string]any{
									"+1": 0, "-1": 0, "laugh": 0, "hooray": 0,
									"confused": 0, "heart": 0, "rocket": 0, "eyes": 0,
								},
							},
						}
						// Build Link header for next page
						nextURL := fmt.Sprintf("http://%s/repos/golang/go/issues/200/comments?page=2", r.Host)
						w.Header().Set("Link", fmt.Sprintf(`<%s>; rel="next"`, nextURL))
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(comments)
					} else {
						// Page 2: return one comment, no Link header
						comments := []map[string]any{
							{
								"user":       map[string]any{"login": "page2user"},
								"created_at": "2024-01-03T00:00:00Z",
								"body":       "Comment from page 2",
								"reactions": map[string]any{
									"+1": 0, "-1": 0, "laugh": 0, "hooray": 0,
									"confused": 0, "heart": 0, "rocket": 0, "eyes": 0,
								},
							},
						}
						w.Header().Set("Content-Type", "application/json")
						json.NewEncoder(w).Encode(comments)
					}
				})
			},
			owner:  "golang",
			repo:   "go",
			number: 200,
			want: IssueData{
				Title:     "Paginated",
				State:     "open",
				Author:    "alice",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Labels:    []string{},
				Assignees: []string{},
				Body:      "body",
				Comments: []Comment{
					{
						Author:    "page1user",
						CreatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
						Body:      "Comment from page 1",
					},
					{
						Author:    "page2user",
						CreatedAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
						Body:      "Comment from page 2",
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
			got, err := client.FetchIssue(context.Background(), tt.owner, tt.repo, tt.number)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("FetchIssue() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("FetchIssue() unexpected error: %v", err)
			}

			assertIssueDataEqual(t, got, tt.want)
		})
	}
}

// assertIssueDataEqual compares two IssueData values field by field for better error messages.
func assertIssueDataEqual(t *testing.T, got, want IssueData) {
	t.Helper()

	if got.Title != want.Title {
		t.Errorf("Title = %q, want %q", got.Title, want.Title)
	}
	if got.State != want.State {
		t.Errorf("State = %q, want %q", got.State, want.State)
	}
	if got.Author != want.Author {
		t.Errorf("Author = %q, want %q", got.Author, want.Author)
	}
	if !got.CreatedAt.Equal(want.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, want.CreatedAt)
	}
	if got.Body != want.Body {
		t.Errorf("Body = %q, want %q", got.Body, want.Body)
	}
	if got.Milestone != want.Milestone {
		t.Errorf("Milestone = %q, want %q", got.Milestone, want.Milestone)
	}
	if got.Reactions != want.Reactions {
		t.Errorf("Reactions = %+v, want %+v", got.Reactions, want.Reactions)
	}
	assertStringSliceEqual(t, "Labels", got.Labels, want.Labels)
	assertStringSliceEqual(t, "Assignees", got.Assignees, want.Assignees)

	if len(got.Comments) != len(want.Comments) {
		t.Fatalf("Comments count = %d, want %d", len(got.Comments), len(want.Comments))
	}
	for i := range want.Comments {
		if got.Comments[i].Author != want.Comments[i].Author {
			t.Errorf("Comments[%d].Author = %q, want %q", i, got.Comments[i].Author, want.Comments[i].Author)
		}
		if !got.Comments[i].CreatedAt.Equal(want.Comments[i].CreatedAt) {
			t.Errorf("Comments[%d].CreatedAt = %v, want %v", i, got.Comments[i].CreatedAt, want.Comments[i].CreatedAt)
		}
		if got.Comments[i].Body != want.Comments[i].Body {
			t.Errorf("Comments[%d].Body = %q, want %q", i, got.Comments[i].Body, want.Comments[i].Body)
		}
		if got.Comments[i].Reactions != want.Comments[i].Reactions {
			t.Errorf("Comments[%d].Reactions = %+v, want %+v", i, got.Comments[i].Reactions, want.Comments[i].Reactions)
		}
		if got.Comments[i].FilePath != want.Comments[i].FilePath {
			t.Errorf("Comments[%d].FilePath = %q, want %q", i, got.Comments[i].FilePath, want.Comments[i].FilePath)
		}
		if got.Comments[i].Line != want.Comments[i].Line {
			t.Errorf("Comments[%d].Line = %d, want %d", i, got.Comments[i].Line, want.Comments[i].Line)
		}
	}
}

func assertStringSliceEqual(t *testing.T, field string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s length = %d, want %d (got %v, want %v)", field, len(got), len(want), got, want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s[%d] = %q, want %q", field, i, got[i], want[i])
		}
	}
}
