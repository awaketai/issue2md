package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/awaketai/issue2md/internal/github"
	"github.com/awaketai/issue2md/internal/parser"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    Args
		wantErr bool
	}{
		{
			name: "URL only",
			args: []string{"https://github.com/golang/go/issues/12345"},
			want: Args{
				URL:           "https://github.com/golang/go/issues/12345",
				Output:        "",
				WithReactions: false,
				Verbose:       false,
			},
		},
		{
			name: "URL with -o flag",
			args: []string{"https://github.com/golang/go/issues/12345", "-o", "output.md"},
			want: Args{
				URL:    "https://github.com/golang/go/issues/12345",
				Output: "output.md",
			},
		},
		{
			name: "URL with --output flag",
			args: []string{"https://github.com/golang/go/issues/12345", "--output", "output.md"},
			want: Args{
				URL:    "https://github.com/golang/go/issues/12345",
				Output: "output.md",
			},
		},
		{
			name: "URL with --with-reactions",
			args: []string{"https://github.com/golang/go/issues/12345", "--with-reactions"},
			want: Args{
				URL:           "https://github.com/golang/go/issues/12345",
				WithReactions: true,
			},
		},
		{
			name: "URL with --verbose",
			args: []string{"https://github.com/golang/go/issues/12345", "--verbose"},
			want: Args{
				URL:     "https://github.com/golang/go/issues/12345",
				Verbose: true,
			},
		},
		{
			name: "all flags combined",
			args: []string{
				"https://github.com/golang/go/issues/12345",
				"-o", "out.md",
				"--with-reactions",
				"--verbose",
			},
			want: Args{
				URL:           "https://github.com/golang/go/issues/12345",
				Output:        "out.md",
				WithReactions: true,
				Verbose:       true,
			},
		},
		{
			name:    "no arguments — error",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "unknown flag — error",
			args:    []string{"--unknown-flag", "https://github.com/golang/go/issues/12345"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseArgs(%v) expected error, got nil", tt.args)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseArgs(%v) unexpected error: %v", tt.args, err)
				return
			}
			if got != tt.want {
				t.Errorf("ParseArgs(%v) = %+v, want %+v", tt.args, got, tt.want)
			}
		})
	}
}

func TestResolveOutputPath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		output   string
		resource parser.Resource
		want     string
	}{
		{
			name:   "directory path — Issue auto filename",
			output: tmpDir,
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypeIssue,
				Number: 12345,
			},
			want: filepath.Join(tmpDir, "golang_go_issue_12345.md"),
		},
		{
			name:   "directory path — PR auto filename",
			output: tmpDir,
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypePR,
				Number: 67890,
			},
			want: filepath.Join(tmpDir, "golang_go_pr_67890.md"),
		},
		{
			name:   "directory path — Discussion auto filename",
			output: tmpDir,
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypeDiscussion,
				Number: 111,
			},
			want: filepath.Join(tmpDir, "golang_go_discussion_111.md"),
		},
		{
			name:   "explicit file path — used as-is",
			output: "output.md",
			resource: parser.Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   parser.TypeIssue,
				Number: 12345,
			},
			want: "output.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveOutputPath(tt.output, tt.resource)
			if got != tt.want {
				t.Errorf("resolveOutputPath(%q, %+v) = %q, want %q", tt.output, tt.resource, got, tt.want)
			}
		})
	}
}

// suppress unused import warnings during RED phase
func TestRun(t *testing.T) {
	t.Run("normal Issue URL — stdout contains Markdown", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/repos/golang/go/issues/12345", func(w http.ResponseWriter, r *http.Request) {
			issue := map[string]any{
				"title":      "Fix the bug",
				"state":      "open",
				"body":       "Bug description.",
				"created_at": "2024-01-15T10:30:00Z",
				"user":       map[string]any{"login": "alice"},
				"labels":     []map[string]any{},
				"assignees":  []map[string]any{},
				"reactions":  map[string]any{"+1": 0, "-1": 0, "laugh": 0, "hooray": 0, "confused": 0, "heart": 0, "rocket": 0, "eyes": 0},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(issue)
		})
		mux.HandleFunc("/repos/golang/go/issues/12345/comments", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]any{})
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		client := github.NewTestClient(server.URL, nil)
		var stdout, stderr bytes.Buffer
		args := Args{URL: "https://github.com/golang/go/issues/12345"}

		err := runCore(args, client, &stdout, &stderr)
		if err != nil {
			t.Fatalf("runCore() unexpected error: %v", err)
		}

		out := stdout.String()
		if !strings.Contains(out, "# [Issue] Fix the bug (#12345)") {
			t.Errorf("stdout missing title, got:\n%s", out)
		}
		if !strings.Contains(out, "Bug description.") {
			t.Errorf("stdout missing body, got:\n%s", out)
		}
	})

	t.Run("invalid URL — returns InputError", func(t *testing.T) {
		client := github.NewTestClient("http://unused", nil)
		var stdout, stderr bytes.Buffer
		args := Args{URL: "not-a-valid-url"}

		err := runCore(args, client, &stdout, &stderr)
		if err == nil {
			t.Fatal("runCore() expected error, got nil")
		}
		var inputErr *InputError
		if !errors.As(err, &inputErr) {
			t.Errorf("expected *InputError, got %T: %v", err, err)
		}
	})

	t.Run("API 404 — returns non-InputError", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/repos/golang/go/issues/99999", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"message":"Not Found"}`))
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		client := github.NewTestClient(server.URL, nil)
		var stdout, stderr bytes.Buffer
		args := Args{URL: "https://github.com/golang/go/issues/99999"}

		err := runCore(args, client, &stdout, &stderr)
		if err == nil {
			t.Fatal("runCore() expected error, got nil")
		}
		var inputErr *InputError
		if errors.As(err, &inputErr) {
			t.Errorf("expected non-InputError, got *InputError: %v", err)
		}
	})

	t.Run("verbose mode — stderr contains debug log", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/repos/golang/go/issues/12345", func(w http.ResponseWriter, r *http.Request) {
			issue := map[string]any{
				"title":      "Fix the bug",
				"state":      "open",
				"body":       "Body.",
				"created_at": "2024-01-15T10:30:00Z",
				"user":       map[string]any{"login": "alice"},
				"labels":     []map[string]any{},
				"assignees":  []map[string]any{},
				"reactions":  map[string]any{"+1": 0, "-1": 0, "laugh": 0, "hooray": 0, "confused": 0, "heart": 0, "rocket": 0, "eyes": 0},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(issue)
		})
		mux.HandleFunc("/repos/golang/go/issues/12345/comments", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]any{})
		})
		server := httptest.NewServer(mux)
		defer server.Close()

		var stderr bytes.Buffer
		client := github.NewTestClient(server.URL, &stderr)
		var stdout bytes.Buffer
		args := Args{
			URL:     "https://github.com/golang/go/issues/12345",
			Verbose: true,
		}

		err := runCore(args, client, &stdout, &stderr)
		if err != nil {
			t.Fatalf("runCore() unexpected error: %v", err)
		}

		if stderr.Len() == 0 {
			t.Error("expected verbose output on stderr, got empty")
		}
	})
}
