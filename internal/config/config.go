package config

// Config 保存从环境变量中读取的配置
type Config struct {
	GitHubToken string // 来自 GITHUB_TOKEN，可能为空
}

// Load 从环境变量加载配置。
func Load() Config {
	return Config{}
}
