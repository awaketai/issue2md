package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		envValue string // GITHUB_TOKEN 的值
		setEnv   bool   // 是否设置环境变量
		want     Config
	}{
		{
			name:     "GITHUB_TOKEN is set",
			envValue: "ghp_test_token_12345",
			setEnv:   true,
			want:     Config{GitHubToken: "ghp_test_token_12345"},
		},
		{
			name:   "GITHUB_TOKEN is not set",
			setEnv: false,
			want:   Config{GitHubToken: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv("GITHUB_TOKEN", tt.envValue)
			}

			got := Load()

			if got != tt.want {
				t.Errorf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
