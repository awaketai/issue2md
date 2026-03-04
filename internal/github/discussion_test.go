package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestGraphQLClient creates a Client whose GraphQL endpoint points at the given httptest server.
func newTestGraphQLClient(serverURL string) *Client {
	// shurcooL/githubv4 uses a standard http.Client that POSTs to https://api.github.com/graphql
	// We need to override that. We create a custom client that points at our test server.
	// The githubv4 library accepts an *http.Client in NewEnterpriseClient with a custom URL.
	// We use NewEnterpriseClient to point at our test server.
	graphqlClient := newGraphQLClientForURL(serverURL + "/graphql")

	return &Client{
		graphqlClient: graphqlClient,
		verbose:       nil,
	}
}

func TestFetchDiscussion(t *testing.T) {
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
			name: "normal discussion with metadata",
			setup: func(mux *http.ServeMux) {
				mux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
					resp := map[string]any{
						"data": map[string]any{
							"repository": map[string]any{
								"discussion": map[string]any{
									"title":     "How to use feature Y?",
									"body":      "I need help with Y.",
									"createdAt": "2024-08-01T12:00:00Z",
									"author":    map[string]any{"login": "questioner"},
									"stateReason": "OPEN",
									"labels": map[string]any{
										"nodes": []map[string]any{
											{"name": "question"},
											{"name": "help wanted"},
										},
									},
									"reactionGroups": []map[string]any{
										{"content": "THUMBS_UP", "reactors": map[string]any{"totalCount": 3}},
										{"content": "THUMBS_DOWN", "reactors": map[string]any{"totalCount": 0}},
										{"content": "LAUGH", "reactors": map[string]any{"totalCount": 0}},
										{"content": "HOORAY", "reactors": map[string]any{"totalCount": 0}},
										{"content": "CONFUSED", "reactors": map[string]any{"totalCount": 1}},
										{"content": "HEART", "reactors": map[string]any{"totalCount": 0}},
										{"content": "ROCKET", "reactors": map[string]any{"totalCount": 0}},
										{"content": "EYES", "reactors": map[string]any{"totalCount": 2}},
									},
									"comments": map[string]any{
										"nodes": []map[string]any{},
									},
								},
							},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(resp)
				})
			},
			owner:  "golang",
			repo:   "go",
			number: 111,
			want: IssueData{
				Title:     "How to use feature Y?",
				State:     "open",
				Author:    "questioner",
				CreatedAt: time.Date(2024, 8, 1, 12, 0, 0, 0, time.UTC),
				Labels:    []string{"question", "help wanted"},
				Assignees: []string{},
				Body:      "I need help with Y.",
				Reactions: Reactions{PlusOne: 3, Confused: 1, Eyes: 2},
				Comments:  []Comment{},
			},
		},
		{
			name: "discussion with nested replies flattened by time",
			setup: func(mux *http.ServeMux) {
				mux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
					resp := map[string]any{
						"data": map[string]any{
							"repository": map[string]any{
								"discussion": map[string]any{
									"title":     "Nested replies test",
									"body":      "body",
									"createdAt": "2024-09-01T00:00:00Z",
									"author":    map[string]any{"login": "op"},
									"stateReason": "OPEN",
									"labels": map[string]any{
										"nodes": []any{},
									},
									"reactionGroups": []map[string]any{
										{"content": "THUMBS_UP", "reactors": map[string]any{"totalCount": 0}},
										{"content": "THUMBS_DOWN", "reactors": map[string]any{"totalCount": 0}},
										{"content": "LAUGH", "reactors": map[string]any{"totalCount": 0}},
										{"content": "HOORAY", "reactors": map[string]any{"totalCount": 0}},
										{"content": "CONFUSED", "reactors": map[string]any{"totalCount": 0}},
										{"content": "HEART", "reactors": map[string]any{"totalCount": 0}},
										{"content": "ROCKET", "reactors": map[string]any{"totalCount": 0}},
										{"content": "EYES", "reactors": map[string]any{"totalCount": 0}},
									},
									"comments": map[string]any{
										"nodes": []map[string]any{
											{
												"author":    map[string]any{"login": "commenter1"},
												"body":      "Top-level comment",
												"createdAt": "2024-09-02T10:00:00Z",
												"reactionGroups": []map[string]any{
													{"content": "THUMBS_UP", "reactors": map[string]any{"totalCount": 1}},
													{"content": "THUMBS_DOWN", "reactors": map[string]any{"totalCount": 0}},
													{"content": "LAUGH", "reactors": map[string]any{"totalCount": 0}},
													{"content": "HOORAY", "reactors": map[string]any{"totalCount": 0}},
													{"content": "CONFUSED", "reactors": map[string]any{"totalCount": 0}},
													{"content": "HEART", "reactors": map[string]any{"totalCount": 0}},
													{"content": "ROCKET", "reactors": map[string]any{"totalCount": 0}},
													{"content": "EYES", "reactors": map[string]any{"totalCount": 0}},
												},
												"replies": map[string]any{
													"nodes": []map[string]any{
														{
															"author":    map[string]any{"login": "replier1"},
															"body":      "Reply to top-level",
															"createdAt": "2024-09-02T12:00:00Z",
															"reactionGroups": []map[string]any{
																{"content": "THUMBS_UP", "reactors": map[string]any{"totalCount": 0}},
																{"content": "THUMBS_DOWN", "reactors": map[string]any{"totalCount": 0}},
																{"content": "LAUGH", "reactors": map[string]any{"totalCount": 0}},
																{"content": "HOORAY", "reactors": map[string]any{"totalCount": 0}},
																{"content": "CONFUSED", "reactors": map[string]any{"totalCount": 0}},
																{"content": "HEART", "reactors": map[string]any{"totalCount": 0}},
																{"content": "ROCKET", "reactors": map[string]any{"totalCount": 0}},
																{"content": "EYES", "reactors": map[string]any{"totalCount": 0}},
															},
														},
													},
												},
											},
											{
												"author":    map[string]any{"login": "commenter2"},
												"body":      "Another top-level comment",
												"createdAt": "2024-09-03T08:00:00Z",
												"reactionGroups": []map[string]any{
													{"content": "THUMBS_UP", "reactors": map[string]any{"totalCount": 0}},
													{"content": "THUMBS_DOWN", "reactors": map[string]any{"totalCount": 0}},
													{"content": "LAUGH", "reactors": map[string]any{"totalCount": 0}},
													{"content": "HOORAY", "reactors": map[string]any{"totalCount": 0}},
													{"content": "CONFUSED", "reactors": map[string]any{"totalCount": 0}},
													{"content": "HEART", "reactors": map[string]any{"totalCount": 0}},
													{"content": "ROCKET", "reactors": map[string]any{"totalCount": 0}},
													{"content": "EYES", "reactors": map[string]any{"totalCount": 0}},
												},
												"replies": map[string]any{
													"nodes": []any{},
												},
											},
										},
									},
								},
							},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(resp)
				})
			},
			owner:  "golang",
			repo:   "go",
			number: 222,
			want: IssueData{
				Title:     "Nested replies test",
				State:     "open",
				Author:    "op",
				CreatedAt: time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC),
				Labels:    []string{},
				Assignees: []string{},
				Body:      "body",
				Comments: []Comment{
					{
						Author:    "commenter1",
						CreatedAt: time.Date(2024, 9, 2, 10, 0, 0, 0, time.UTC),
						Body:      "Top-level comment",
						Reactions: Reactions{PlusOne: 1},
					},
					{
						Author:    "replier1",
						CreatedAt: time.Date(2024, 9, 2, 12, 0, 0, 0, time.UTC),
						Body:      "Reply to top-level",
					},
					{
						Author:    "commenter2",
						CreatedAt: time.Date(2024, 9, 3, 8, 0, 0, 0, time.UTC),
						Body:      "Another top-level comment",
					},
				},
			},
		},
		{
			name: "discussion with labels",
			setup: func(mux *http.ServeMux) {
				mux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
					resp := map[string]any{
						"data": map[string]any{
							"repository": map[string]any{
								"discussion": map[string]any{
									"title":     "Labels test",
									"body":      "body",
									"createdAt": "2024-10-01T00:00:00Z",
									"author":    map[string]any{"login": "user"},
									"stateReason": "OPEN",
									"labels": map[string]any{
										"nodes": []map[string]any{
											{"name": "bug"},
											{"name": "docs"},
											{"name": "good first issue"},
										},
									},
									"reactionGroups": []map[string]any{
										{"content": "THUMBS_UP", "reactors": map[string]any{"totalCount": 0}},
										{"content": "THUMBS_DOWN", "reactors": map[string]any{"totalCount": 0}},
										{"content": "LAUGH", "reactors": map[string]any{"totalCount": 0}},
										{"content": "HOORAY", "reactors": map[string]any{"totalCount": 0}},
										{"content": "CONFUSED", "reactors": map[string]any{"totalCount": 0}},
										{"content": "HEART", "reactors": map[string]any{"totalCount": 0}},
										{"content": "ROCKET", "reactors": map[string]any{"totalCount": 0}},
										{"content": "EYES", "reactors": map[string]any{"totalCount": 0}},
									},
									"comments": map[string]any{
										"nodes": []any{},
									},
								},
							},
						},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(resp)
				})
			},
			owner:  "golang",
			repo:   "go",
			number: 333,
			want: IssueData{
				Title:     "Labels test",
				State:     "open",
				Author:    "user",
				CreatedAt: time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC),
				Labels:    []string{"bug", "docs", "good first issue"},
				Assignees: []string{},
				Body:      "body",
				Comments:  []Comment{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			tt.setup(mux)
			server := httptest.NewServer(mux)
			defer server.Close()

			client := newTestGraphQLClient(server.URL)
			got, err := client.FetchDiscussion(context.Background(), tt.owner, tt.repo, tt.number)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("FetchDiscussion() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("FetchDiscussion() unexpected error: %v", err)
			}

			assertIssueDataEqual(t, got, tt.want)
		})
	}
}
