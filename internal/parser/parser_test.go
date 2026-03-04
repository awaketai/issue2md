package parser

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Resource
		wantErr bool
		errMsg  string // 错误消息中应包含的关键词（可选）
	}{
		// AC-URL-01: 解析 GitHub Issue URL
		{
			name:  "valid issue URL",
			input: "https://github.com/golang/go/issues/12345",
			want: Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   TypeIssue,
				Number: 12345,
			},
		},
		// AC-URL-02: 解析 GitHub PR URL
		{
			name:  "valid PR URL",
			input: "https://github.com/golang/go/pull/67890",
			want: Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   TypePR,
				Number: 67890,
			},
		},
		// AC-URL-03: 解析 GitHub Discussion URL
		{
			name:  "valid discussion URL",
			input: "https://github.com/golang/go/discussions/111",
			want: Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   TypeDiscussion,
				Number: 111,
			},
		},
		// AC-URL-04: 无效 URL
		{
			name:    "invalid URL format",
			input:   "not-a-url",
			wantErr: true,
		},
		// AC-URL-05: 非 GitHub URL
		{
			name:    "non-GitHub URL",
			input:   "https://gitlab.com/group/project/-/issues/42",
			wantErr: true,
			errMsg:  "github.com",
		},
		// AC-URL-06: 不支持的路径结构
		{
			name:    "unsupported path - wiki",
			input:   "https://github.com/golang/go/wiki/Home",
			wantErr: true,
		},
		// AC-URL-07: URL 带尾部斜杠
		{
			name:  "trailing slash",
			input: "https://github.com/golang/go/issues/12345/",
			want: Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   TypeIssue,
				Number: 12345,
			},
		},
		// AC-URL-08: URL 带查询参数
		{
			name:  "with query parameters",
			input: "https://github.com/golang/go/issues/12345?foo=bar",
			want: Resource{
				Owner:  "golang",
				Repo:   "go",
				Type:   TypeIssue,
				Number: 12345,
			},
		},
		// 额外边界用例：仅仓库主页路径（不支持的路径结构）
		{
			name:    "repo root URL without resource type",
			input:   "https://github.com/golang/go",
			wantErr: true,
		},
		// 额外边界用例：number 非数字
		{
			name:    "non-numeric number",
			input:   "https://github.com/golang/go/issues/abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q) expected error, got nil", tt.input)
				}
				if tt.errMsg != "" {
					if !containsSubstring(err.Error(), tt.errMsg) {
						t.Errorf("Parse(%q) error = %q, want error containing %q", tt.input, err.Error(), tt.errMsg)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tt.input, err)
			}

			if got != tt.want {
				t.Errorf("Parse(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

// containsSubstring 检查 s 中是否包含 substr
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
