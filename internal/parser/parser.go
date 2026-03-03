package parser

type ResourceType string

const (
	TypeIssue      ResourceType = "issue"
	TypePR         ResourceType = "pr"
	TypeDiscussion ResourceType = "discussion"
)

// Resource 表示从 URL 中解析出的结构化信息
type Resource struct {
	Owner  string       // e.g. "golang"
	Repo   string       // e.g. "go"
	Type   ResourceType // e.g. TypeIssue
	Number int          // e.g. 12345
}

// Parse 解析 GitHub URL，返回 Resource。
// 错误场景：URL 格式无效、host 不是 github.com、路径无法识别。
func Parse(rawURL string) (Resource, error) {
	return Resource{}, nil
}
