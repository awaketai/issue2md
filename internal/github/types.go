package github

import "time"

// Reactions 表示一条内容的 Reactions 统计
// 字段名与 GitHub API 返回的 Reaction 类型一一对应
type Reactions struct {
	PlusOne  int // 👍
	MinusOne int // 👎
	Laugh    int // 😄
	Hooray   int // 🎉
	Confused int // 😕
	Heart    int // ❤️
	Rocket   int // 🚀
	Eyes     int // 👀
}

// Comment 表示一条评论（普通评论或 Review Comment）
type Comment struct {
	Author    string    // 评论者用户名
	CreatedAt time.Time // 评论时间 (ISO 8601)
	Body      string    // 评论内容（原始 Markdown）
	Reactions Reactions // Reactions 统计

	// 以下字段仅 PR Review Comment 使用，普通评论为零值
	FilePath string // e.g. "internal/auth/sso.go"
	Line     int    // e.g. 42
}

// IssueData 表示从 GitHub 获取的完整 Issue/PR/Discussion 数据
// 这是在所有模块间流转的统一数据模型
type IssueData struct {
	// --- 元数据 ---
	Title     string    // 标题
	State     string    // "open", "closed", "merged"
	Author    string    // 创建者用户名
	CreatedAt time.Time // 创建时间 (ISO 8601)
	Labels    []string  // 标签列表
	Assignees []string  // 指派人列表
	Milestone string    // 里程碑名称，空字符串表示无 milestone
	LinkedPRs []int     // 关联的 PR 编号列表（仅 Issue 有效）

	// --- 正文 ---
	Body      string    // 正文内容（原始 Markdown）
	Reactions Reactions // 正文的 Reactions 统计

	// --- 评论 ---
	Comments []Comment // 所有评论，按时间线排序
}
