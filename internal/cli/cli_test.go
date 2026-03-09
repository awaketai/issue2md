package cli

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

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
			name: "Case 1: URL only",
			args: []string{"https://github.com/golang/go/issues/12345"},
			want: Args{URL: "https://github.com/golang/go/issues/12345"},
		},
		{
			name: "Case 2: URL with -o flag",
			args: []string{"https://github.com/golang/go/issues/12345", "-o", "output.md"},
			want: Args{URL: "https://github.com/golang/go/issues/12345", Output: "output.md"},
		},
		{
			name: "Case 3: URL with --output flag",
			args: []string{"https://github.com/golang/go/issues/12345", "--output", "output.md"},
			want: Args{URL: "https://github.com/golang/go/issues/12345", Output: "output.md"},
		},
		{
			name: "Case 4: URL with --with-reactions",
			args: []string{"https://github.com/golang/go/issues/12345", "--with-reactions"},
			want: Args{URL: "https://github.com/golang/go/issues/12345", WithReactions: true},
		},
		{
			name: "Case 5: URL with --verbose",
			args: []string{"https://github.com/golang/go/issues/12345", "--verbose"},
			want: Args{URL: "https://github.com/golang/go/issues/12345", Verbose: true},
		},
		{
			name: "Case 6: all flags combined",
			args: []string{"https://github.com/golang/go/issues/12345", "--with-reactions", "-o", "out.md", "--verbose"},
			want: Args{URL: "https://github.com/golang/go/issues/12345", Output: "out.md", WithReactions: true, Verbose: true},
		},
		{
			name:    "Case 7: no arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "Case 8: unknown flag",
			args:    []string{"--unknown-flag"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
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
			name:     "Case 1: directory + Issue",
			output:   tmpDir,
			resource: parser.Resource{Owner: "golang", Repo: "go", Type: parser.TypeIssue, Number: 12345},
			want:     filepath.Join(tmpDir, "golang_go_issue_12345.md"),
		},
		{
			name:     "Case 2: directory + PR",
			output:   tmpDir,
			resource: parser.Resource{Owner: "golang", Repo: "go", Type: parser.TypePR, Number: 67890},
			want:     filepath.Join(tmpDir, "golang_go_pr_67890.md"),
		},
		{
			name:     "Case 3: directory + Discussion",
			output:   tmpDir,
			resource: parser.Resource{Owner: "golang", Repo: "go", Type: parser.TypeDiscussion, Number: 111},
			want:     filepath.Join(tmpDir, "golang_go_discussion_111.md"),
		},
		{
			name:     "Case 4: explicit file path",
			output:   "output.md",
			resource: parser.Resource{Owner: "golang", Repo: "go", Type: parser.TypeIssue, Number: 12345},
			want:     "output.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveOutputPath(tt.output, tt.resource)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name         string
		args         Args
		wantErr      bool
		wantInput    bool   // expect *InputError
		wantContains string // substring expected in stdout
	}{
		{
			name:      "invalid URL returns InputError",
			args:      Args{URL: "not-a-url"},
			wantErr:   true,
			wantInput: true,
		},
		{
			name:      "non-github URL returns InputError",
			args:      Args{URL: "https://gitlab.com/group/project/-/issues/42"},
			wantErr:   true,
			wantInput: true,
		},
		{
			name:      "unsupported resource type returns InputError",
			args:      Args{URL: "https://github.com/golang/go/wiki/Home"},
			wantErr:   true,
			wantInput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := Run(tt.args, &stdout, &stderr)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
				if tt.wantInput {
					var inputErr *InputError
					if !errors.As(err, &inputErr) {
						t.Errorf("expected *InputError, got %T: %v", err, err)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantContains != "" {
				if !strings.Contains(stdout.String(), tt.wantContains) {
					t.Errorf("stdout missing %q\ngot: %s", tt.wantContains, stdout.String())
				}
			}
		})
	}
}
