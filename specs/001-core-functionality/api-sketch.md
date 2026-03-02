# API Sketch: 核心包接口设计

**Date**: 2026-03-02
**基于**: spec.md v1.0, constitution.md v1.0

本文档描述 `internal/` 下各包对外暴露的主要类型和函数签名，作为后续开发的参考。

> **设计原则**：遵循宪法第一条（简单性）和第三条（明确性）。
> 优先使用具体类型和函数，仅在确实需要多态时才引入接口（宪法 1.3）。
> 所有错误通过返回值显式传递，使用 `fmt.Errorf("...: %w", err)` 包装（宪法 3.1）。

---

## 1. `internal/parser` — URL 解析

职责：将原始 URL 字符串解析为结构化数据，判断资源类型。纯函数，无外部依赖。

```go
package parser

// ResourceType 表示 URL 对应的资源类型
type ResourceType string

const (
    TypeIssue      ResourceType = "issue"
    TypePR         ResourceType = "pr"
    TypeDiscussion ResourceType = "discussion"
)

// Resource 表示从 URL 中解析出的结构化信息
type Resource struct {
    Owner  string
    Repo   string
    Type   ResourceType
    Number int
}

// Parse 解析 GitHub URL，返回 Resource。
// 如果 URL 无效、host 不是 github.com、或路径结构无法识别，返回错误。
func Parse(rawURL string) (Resource, error)
```

---

## 2. `internal/github` — GitHub API 数据获取

职责：调用 GitHub REST API，获取 Issue / PR / Discussion 的完整数据，处理分页。
返回与平台无关的统一数据模型，供 converter 包消费。

```go
package github

import (
    "context"
    "io"
    "time"
)

// Reactions 表示一条内容的 Reactions 统计
type Reactions struct {
    PlusOne    int
    MinusOne   int
    Laugh      int
    Hooray     int
    Confused   int
    Heart      int
    Rocket     int
    Eyes       int
}

// Comment 表示一条评论（普通评论或 Review Comment）
type Comment struct {
    Author    string
    CreatedAt time.Time
    Body      string
    Reactions Reactions

    // 以下字段仅 PR Review Comment 使用，普通评论为零值
    FilePath string // e.g. "internal/auth/sso.go"
    Line     int    // e.g. 42
}

// IssueData 表示从 GitHub 获取的完整 Issue/PR/Discussion 数据
type IssueData struct {
    // 元数据
    Title     string
    State     string // "open", "closed", "merged"
    Author    string
    CreatedAt time.Time
    Labels    []string
    Assignees []string
    Milestone string   // 空字符串表示无 milestone
    LinkedPRs []int    // 仅 Issue 有效，关联的 PR 编号列表

    // 内容
    Body      string
    Reactions Reactions

    // 评论（按时间线排序）
    Comments []Comment
}

// Client 封装 GitHub API 的访问
type Client struct {
    // 未导出字段：httpClient, token, verbose logger 等
}

// NewClient 创建 GitHub API 客户端。
// token 为空字符串时以未认证模式访问。
// verbose 为 stderr 的 logger writer，传 nil 则不输出调试日志。
func NewClient(token string, verbose io.Writer) *Client

// FetchIssue 获取 Issue 的完整数据（元数据 + 正文 + 全部评论）。
func (c *Client) FetchIssue(ctx context.Context, owner, repo string, number int) (IssueData, error)

// FetchPR 获取 PR 的完整数据（元数据 + 描述 + 普通评论 + Review Comments，按时间线合并排序）。
func (c *Client) FetchPR(ctx context.Context, owner, repo string, number int) (IssueData, error)

// FetchDiscussion 获取 Discussion 的完整数据（元数据 + 正文 + 所有嵌套回复平铺）。
func (c *Client) FetchDiscussion(ctx context.Context, owner, repo string, number int) (IssueData, error)
```

---

## 3. `internal/converter` — Markdown 渲染

职责：接受 `github.IssueData`，输出格式化的 Markdown 字符串。纯函数，无 I/O 操作。

```go
package converter

import (
    "github.com/awaketai/issue2md/internal/github"
    "github.com/awaketai/issue2md/internal/parser"
)

// Options 控制 Markdown 输出的可选行为
type Options struct {
    WithReactions bool // 是否在正文和评论下方显示 Reactions 统计
}

// ToMarkdown 将 IssueData 转换为格式化的 Markdown 字符串。
// resource 提供类型和编号信息，用于生成标题（如 "[Issue] Title (#123)"）。
func ToMarkdown(data github.IssueData, resource parser.Resource, opts Options) string
```

---

## 4. `internal/config` — 配置管理

职责：从环境变量读取配置。遵循宪法 3.2（无全局变量），配置通过返回值传递。

```go
package config

// Config 保存从环境变量中读取的配置
type Config struct {
    GitHubToken string // 来自 GITHUB_TOKEN，可能为空
}

// Load 从环境变量加载配置。
func Load() Config
```

---

## 5. `internal/cli` — 命令行接口

职责：解析命令行参数，组装各层调用，处理 I/O（stdout / stderr / 文件写入）。

```go
package cli

import (
    "io"
)

// Args 表示解析后的命令行参数
type Args struct {
    URL           string
    Output        string // 文件路径或空字符串（表示 stdout）
    WithReactions bool
    Verbose       bool
}

// ParseArgs 从 os.Args 解析命令行参数。
// 参数无效时返回错误。
func ParseArgs(args []string) (Args, error)

// Run 是 CLI 的主入口，组装 parser → github.Client → converter 的调用链。
// stdout 和 stderr 通过参数注入，便于测试。
func Run(args Args, stdout io.Writer, stderr io.Writer) error
```

---

## 6. `cmd/issue2md/main.go` — CLI 入口

极简入口，仅负责调用 `cli.Run` 并设置退出码。

```go
package main

import (
    "fmt"
    "os"

    "github.com/awaketai/issue2md/internal/cli"
)

func main() {
    args, err := cli.ParseArgs(os.Args[1:])
    if err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
    if err := cli.Run(args, os.Stdout, os.Stderr); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(2)
    }
}
```

---

## 包依赖关系

```
cmd/issue2md
  └── internal/cli
        ├── internal/parser      (纯函数，无依赖)
        ├── internal/github      (依赖 net/http)
        ├── internal/converter   (依赖 parser, github 的类型定义)
        └── internal/config      (依赖 os 环境变量)
```

依赖方向为**单向**：`cli` → `parser` / `github` / `converter` / `config`。
`converter` 依赖 `parser` 和 `github` 的**类型定义**（非调用），保持包间松耦合。
