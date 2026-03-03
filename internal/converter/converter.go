package converter

import (
	"github.com/awaketai/issue2md/internal/github"
	"github.com/awaketai/issue2md/internal/parser"
)

// Options 控制 Markdown 输出的可选行为
type Options struct {
	WithReactions bool // 是否显示 Reactions 统计
}

// ToMarkdown 将 IssueData 转换为格式化的 Markdown 字符串。
// resource 提供类型和编号信息，用于生成标题前缀（如 "[Issue]"）。
func ToMarkdown(data github.IssueData, resource parser.Resource, opts Options) string {
	return ""
}
