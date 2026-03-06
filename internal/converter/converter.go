package converter

import (
	"fmt"
	"strings"

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
	var b strings.Builder

	// 标题行: # [{Type}] {Title} (#{Number})
	b.WriteString(fmt.Sprintf("# [%s] %s (#%d)\n", typeLabel(resource.Type), data.Title, resource.Number))
	b.WriteString("\n")

	// 元数据表格
	b.WriteString("| Field | Value |\n")
	b.WriteString("|-------|-------|\n")
	b.WriteString(fmt.Sprintf("| **State** | %s |\n", data.State))
	b.WriteString(fmt.Sprintf("| **Author** | @%s |\n", data.Author))
	b.WriteString(fmt.Sprintf("| **Created** | %s |\n", data.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")))

	if len(data.Labels) > 0 {
		b.WriteString(fmt.Sprintf("| **Labels** | %s |\n", strings.Join(data.Labels, ", ")))
	}
	if len(data.Assignees) > 0 {
		assignees := make([]string, len(data.Assignees))
		for i, a := range data.Assignees {
			assignees[i] = "@" + a
		}
		b.WriteString(fmt.Sprintf("| **Assignees** | %s |\n", strings.Join(assignees, ", ")))
	}
	if data.Milestone != "" {
		b.WriteString(fmt.Sprintf("| **Milestone** | %s |\n", data.Milestone))
	}
	if len(data.LinkedPRs) > 0 {
		prs := make([]string, len(data.LinkedPRs))
		for i, pr := range data.LinkedPRs {
			prs[i] = fmt.Sprintf("#%d", pr)
		}
		b.WriteString(fmt.Sprintf("| **Linked PRs** | %s |\n", strings.Join(prs, ", ")))
	}

	// 分隔线 + 正文
	b.WriteString("\n---\n\n")
	b.WriteString(data.Body)
	b.WriteString("\n")

	// 正文 Reactions
	if opts.WithReactions {
		if line := formatReactions(data.Reactions); line != "" {
			b.WriteString("\n")
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	// 评论
	if len(data.Comments) > 0 {
		b.WriteString("\n---\n\n")
		b.WriteString(fmt.Sprintf("## Comments (%d)\n", len(data.Comments)))

		for _, c := range data.Comments {
			b.WriteString("\n")
			// 评论标题: ### @author — timestamp [可选: `file#Lline`]
			header := fmt.Sprintf("### @%s — %s", c.Author, c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"))
			if c.FilePath != "" && c.Line > 0 {
				header += fmt.Sprintf(" `%s#L%d`", c.FilePath, c.Line)
			}
			b.WriteString(header)
			b.WriteString("\n\n")
			b.WriteString(c.Body)
			b.WriteString("\n")

			// 评论 Reactions
			if opts.WithReactions {
				if line := formatReactions(c.Reactions); line != "" {
					b.WriteString("\n")
					b.WriteString(line)
					b.WriteString("\n")
				}
			}

			b.WriteString("\n---\n")
		}
	}

	return b.String()
}

// typeLabel 将 ResourceType 映射为显示标签
func typeLabel(t parser.ResourceType) string {
	switch t {
	case parser.TypeIssue:
		return "Issue"
	case parser.TypePR:
		return "PR"
	case parser.TypeDiscussion:
		return "Discussion"
	default:
		return string(t)
	}
}

// formatReactions 将 Reactions 格式化为引用块字符串。
// 仅包含数量 > 0 的 Reaction 类型。返回空字符串表示无 Reactions。
func formatReactions(r github.Reactions) string {
	var parts []string
	if r.PlusOne > 0 {
		parts = append(parts, fmt.Sprintf("👍 %d", r.PlusOne))
	}
	if r.MinusOne > 0 {
		parts = append(parts, fmt.Sprintf("👎 %d", r.MinusOne))
	}
	if r.Laugh > 0 {
		parts = append(parts, fmt.Sprintf("😄 %d", r.Laugh))
	}
	if r.Hooray > 0 {
		parts = append(parts, fmt.Sprintf("🎉 %d", r.Hooray))
	}
	if r.Confused > 0 {
		parts = append(parts, fmt.Sprintf("😕 %d", r.Confused))
	}
	if r.Heart > 0 {
		parts = append(parts, fmt.Sprintf("❤️ %d", r.Heart))
	}
	if r.Rocket > 0 {
		parts = append(parts, fmt.Sprintf("🚀 %d", r.Rocket))
	}
	if r.Eyes > 0 {
		parts = append(parts, fmt.Sprintf("👀 %d", r.Eyes))
	}
	if len(parts) == 0 {
		return ""
	}
	return "> " + strings.Join(parts, " | ")
}
