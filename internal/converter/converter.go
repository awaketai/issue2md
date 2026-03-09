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

	// 标题
	typeLabel := resourceTypeLabel(resource.Type)
	fmt.Fprintf(&b, "# [%s] %s (#%d)\n\n", typeLabel, data.Title, resource.Number)

	// 元数据表格
	b.WriteString("| Field | Value |\n")
	b.WriteString("|-------|-------|\n")
	fmt.Fprintf(&b, "| **State** | %s |\n", data.State)
	fmt.Fprintf(&b, "| **Author** | @%s |\n", data.Author)
	fmt.Fprintf(&b, "| **Created** | %s |\n", data.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"))
	if len(data.Labels) > 0 {
		fmt.Fprintf(&b, "| **Labels** | %s |\n", strings.Join(data.Labels, ", "))
	}
	if len(data.Assignees) > 0 {
		assignees := make([]string, len(data.Assignees))
		for i, a := range data.Assignees {
			assignees[i] = "@" + a
		}
		fmt.Fprintf(&b, "| **Assignees** | %s |\n", strings.Join(assignees, ", "))
	}
	if data.Milestone != "" {
		fmt.Fprintf(&b, "| **Milestone** | %s |\n", data.Milestone)
	}
	if len(data.LinkedPRs) > 0 {
		prs := make([]string, len(data.LinkedPRs))
		for i, n := range data.LinkedPRs {
			prs[i] = fmt.Sprintf("#%d", n)
		}
		fmt.Fprintf(&b, "| **Linked PRs** | %s |\n", strings.Join(prs, ", "))
	}

	// 正文
	b.WriteString("\n---\n\n")
	b.WriteString(data.Body)
	b.WriteString("\n")

	// 正文 Reactions
	if opts.WithReactions {
		if line := formatReactions(data.Reactions); line != "" {
			b.WriteString("\n" + line + "\n")
		}
	}

	// 评论
	if len(data.Comments) > 0 {
		fmt.Fprintf(&b, "\n---\n\n## Comments (%d)\n", len(data.Comments))

		for _, c := range data.Comments {
			b.WriteString("\n")
			header := fmt.Sprintf("### @%s — %s", c.Author, c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"))
			if c.FilePath != "" {
				header += fmt.Sprintf(" `%s#L%d`", c.FilePath, c.Line)
			}
			b.WriteString(header + "\n\n")
			b.WriteString(c.Body)
			b.WriteString("\n")

			if opts.WithReactions {
				if line := formatReactions(c.Reactions); line != "" {
					b.WriteString("\n" + line + "\n")
				}
			}

			b.WriteString("\n---\n")
		}
	}

	return b.String()
}

func resourceTypeLabel(rt parser.ResourceType) string {
	switch rt {
	case parser.TypeIssue:
		return "Issue"
	case parser.TypePR:
		return "PR"
	case parser.TypeDiscussion:
		return "Discussion"
	default:
		return string(rt)
	}
}

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
