# 技术实现方案 (Plan)

**Spec**: 001-core-functionality/spec.md v1.0
**Date**: 2026-03-03
**Status**: Draft

---

## 1. 技术上下文总结

### 1.1 技术选型

| 层面 | 选型 | 理由 |
|------|------|------|
| **语言** | Go >= 1.24（当前 go.mod 为 1.22，需升级） | 项目基础语言 |
| **Web 框架** | `net/http`（标准库） | 宪法第一条：标准库优先 |
| **GitHub REST API 客户端** | `google/go-github/v68` | Issue 和 PR 数据获取，社区主流客户端 |
| **GitHub GraphQL API v4** | `shurcooL/githubv4` + `golang.org/x/oauth2` | Discussion 数据仅 GraphQL API 提供，REST API 不支持 |
| **Markdown 终端渲染** | `charmbracelet/glamour`（可选，仅 Web 预览场景） | 初期 CLI 直接输出纯 Markdown 文本，不需要渲染库 |
| **数据存储** | 无 | 所有数据通过 API 实时获取 |

### 1.2 API 策略说明

- **Issue / PR**：使用 `google/go-github` 调用 REST API v3。REST API 对 Issue 和 PR 提供完整的数据模型（元数据、评论、Review Comments、Reactions），且 `go-github` 库已封装好分页逻辑。
- **Discussion**：使用 `shurcooL/githubv4` 调用 GraphQL API v4。GitHub Discussion 不在 REST API 中暴露，只能通过 GraphQL 访问。GraphQL 同时能一次性获取嵌套回复，减少请求次数。

### 1.3 依赖清单

```
google/go-github/v68          # REST API 客户端（Issue, PR）
shurcooL/githubv4              # GraphQL API v4 客户端（Discussion）
golang.org/x/oauth2            # OAuth2 token 传递（graphql 客户端需要）
```

仅此三个外部依赖（及其传递依赖），遵循宪法第一条最小依赖原则。

---

## 2. 合宪性审查

逐条对照 `constitution.md` 检查本方案的合规性。

### 第一条：简单性原则 (Simplicity First)

| 条款 | 审查结果 | 说明 |
|------|----------|------|
| **1.1 YAGNI** | ✅ 合规 | 本方案仅实现 spec.md 中明确要求的功能：URL 解析、三种资源抓取、Markdown 转换、CLI 参数。不包含缓存、数据库、用户系统等未要求功能 |
| **1.2 标准库优先** | ✅ 合规 | Web 服务使用 `net/http`；CLI 参数解析使用 `flag` 标准库；仅在标准库无法覆盖时（GitHub API 客户端）才引入外部依赖 |
| **1.3 反过度工程** | ✅ 合规 | 各包暴露具体函数而非接口（`Parse()` 函数、`FetchIssue()` 方法）；仅在 `github` 包中使用 `Client` 结构体封装状态，其余包均为纯函数 |

### 第二条：测试先行铁律 (Test-First Imperative)

| 条款 | 审查结果 | 说明 |
|------|----------|------|
| **2.1 TDD 循环** | ✅ 方案约束 | 每个包的实现必须遵循 Red-Green-Refactor。本方案第 6 节定义了各包的测试策略和首个测试用例 |
| **2.2 表格驱动** | ✅ 方案约束 | URL 解析（8 个验收标准）和 Markdown 渲染天然适合表格驱动测试 |
| **2.3 拒绝 Mocks** | ✅ 合规 | `parser` 和 `converter` 为纯函数，无需 Mock；`github` 包测试使用 `httptest.Server` 模拟真实 HTTP 响应，而非 Mock 接口 |

### 第三条：明确性原则 (Clarity and Explicitness)

| 条款 | 审查结果 | 说明 |
|------|----------|------|
| **3.1 错误处理** | ✅ 方案约束 | 所有函数返回 `error`；错误传递使用 `fmt.Errorf("...: %w", err)` 包装；spec 中定义了明确的退出码（1=用户输入错误，2=运行时错误） |
| **3.2 无全局变量** | ✅ 合规 | `Config` 通过 `Load()` 返回值传递；`Client` 通过 `NewClient()` 构造；`io.Writer` 通过参数注入 |

---

## 3. 项目结构细化

### 3.1 目录结构

```
issue2md/
├── cmd/
│   ├── issue2md/
│   │   └── main.go              # CLI 入口，极简
│   └── issue2mdweb/
│       └── main.go              # Web 入口（未来版本）
├── internal/
│   ├── parser/
│   │   ├── parser.go            # URL 解析逻辑
│   │   └── parser_test.go       # 表格驱动测试
│   ├── github/
│   │   ├── client.go            # Client 构造、认证
│   │   ├── issue.go             # FetchIssue 实现
│   │   ├── pr.go                # FetchPR 实现
│   │   ├── discussion.go        # FetchDiscussion 实现（GraphQL）
│   │   ├── types.go             # IssueData, Comment, Reactions 等数据结构
│   │   └── github_test.go       # 使用 httptest 的集成测试
│   ├── converter/
│   │   ├── converter.go         # ToMarkdown 实现
│   │   └── converter_test.go    # 表格驱动测试
│   ├── cli/
│   │   ├── args.go              # ParseArgs 命令行参数解析
│   │   ├── run.go               # Run 主流程编排
│   │   └── cli_test.go          # CLI 集成测试
│   └── config/
│       ├── config.go            # 环境变量读取
│       └── config_test.go
├── web/
│   ├── templates/               # HTML 模板（未来版本）
│   └── static/                  # 静态资源（未来版本）
├── specs/
│   └── 001-core-functionality/
│       ├── spec.md
│       ├── api-sketch.md
│       └── plan.md              # 本文档
├── Makefile
├── go.mod
├── go.sum
├── constitution.md
├── CLAUDE.md
└── README.md
```

### 3.2 包职责与依赖关系

```
cmd/issue2md/main.go
  └── internal/cli
        ├── internal/parser      # 纯函数，零依赖
        ├── internal/config      # 仅依赖 os 标准库
        ├── internal/github      # 依赖 google/go-github, shurcooL/githubv4
        └── internal/converter   # 依赖 parser.Resource 和 github.IssueData 的类型定义
```

**依赖规则（不可违反）：**

1. 依赖方向**单向**：`cli` → 其他包，其他包之间不得相互调用
2. `converter` 仅依赖 `parser` 和 `github` 的**类型定义**（struct），不调用其函数
3. `parser` 和 `config` 为叶子包，不依赖任何 `internal/` 内的包
4. `github` 为叶子包，不依赖任何 `internal/` 内的包（其类型被其他包引用）

### 3.3 各包详细职责

#### `internal/parser`

- **输入**: 原始 URL 字符串
- **输出**: `Resource` 结构体 或 `error`
- **职责**: 使用 `net/url` 标准库解析 URL，校验 host、路径结构，提取 owner/repo/type/number
- **特性**: 纯函数，无副作用，无 I/O

#### `internal/github`

- **输入**: `context.Context`、owner、repo、number
- **输出**: `IssueData` 结构体 或 `error`
- **职责**:
  - Issue/PR：通过 `google/go-github` 调用 REST API，自动处理分页
  - Discussion：通过 `shurcooL/githubv4` 调用 GraphQL API
  - 统一将平台数据映射为 `IssueData`
  - 在 verbose 模式下将调试信息写入注入的 `io.Writer`

#### `internal/converter`

- **输入**: `github.IssueData`、`parser.Resource`、`Options`
- **输出**: Markdown 格式字符串
- **职责**: 纯渲染逻辑，使用 `strings.Builder` 拼接 Markdown，遵循 spec.md 第 5 节的输出格式

#### `internal/cli`

- **输入**: `os.Args`、`io.Writer`（stdout/stderr）
- **职责**: 使用 `flag` 标准库解析参数，加载配置，组装 parser → github → converter 调用链，处理文件输出

#### `internal/config`

- **输入**: 环境变量
- **输出**: `Config` 结构体
- **职责**: 从 `GITHUB_TOKEN` 读取 token，返回 `Config` 值

---

## 4. 核心数据结构

所有核心结构体定义在 `internal/github/types.go` 和 `internal/parser/parser.go` 中。

### 4.1 `parser.Resource`

```go
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
```

### 4.2 `github.Reactions`

```go
package github

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
```

### 4.3 `github.Comment`

```go
package github

import "time"

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
```

### 4.4 `github.IssueData`

```go
package github

import "time"

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
    Comments []Comment  // 所有评论，按时间线排序
}
```

### 4.5 `converter.Options`

```go
package converter

// Options 控制 Markdown 输出的可选行为
type Options struct {
    WithReactions bool // 是否显示 Reactions 统计
}
```

### 4.6 `cli.Args`

```go
package cli

// Args 表示解析后的命令行参数
type Args struct {
    URL           string // GitHub Issue/PR/Discussion URL（必填）
    Output        string // 输出文件路径，空字符串表示 stdout
    WithReactions bool   // 是否显示 Reactions
    Verbose       bool   // 是否输出调试日志到 stderr
}
```

### 4.7 `config.Config`

```go
package config

// Config 保存从环境变量中读取的配置
type Config struct {
    GitHubToken string // 来自 GITHUB_TOKEN，可能为空
}
```

---

## 5. 接口设计

遵循宪法第一条 1.3（反过度工程），本项目**不预先定义 Interface**。各包暴露具体类型和函数。

> 原则：Go 的惯例是"接受接口，返回具体类型"。仅在需要多态替换时才定义接口。
> 当前阶段，每个角色只有一种实现，定义接口属于过度抽象。

### 5.1 各包导出的关键函数签名

#### `internal/parser`

```go
// Parse 解析 GitHub URL，返回 Resource。
// 错误场景：URL 格式无效、host 不是 github.com、路径无法识别。
func Parse(rawURL string) (Resource, error)
```

#### `internal/github`

```go
// NewClient 创建 GitHub API 客户端。
// token 为空字符串时以未认证模式访问。
// verbose 传入 io.Writer 输出调试日志，传 nil 则静默。
func NewClient(token string, verbose io.Writer) *Client

// FetchIssue 获取 Issue 完整数据。自动处理分页。
func (c *Client) FetchIssue(ctx context.Context, owner, repo string, number int) (IssueData, error)

// FetchPR 获取 PR 完整数据。普通评论与 Review Comments 按时间线合并排序。
func (c *Client) FetchPR(ctx context.Context, owner, repo string, number int) (IssueData, error)

// FetchDiscussion 获取 Discussion 完整数据。嵌套回复按时间线平铺。
func (c *Client) FetchDiscussion(ctx context.Context, owner, repo string, number int) (IssueData, error)
```

#### `internal/converter`

```go
// ToMarkdown 将 IssueData 转换为格式化的 Markdown 字符串。
// resource 提供类型和编号信息，用于生成标题前缀（如 "[Issue]"）。
func ToMarkdown(data github.IssueData, resource parser.Resource, opts Options) string
```

#### `internal/cli`

```go
// ParseArgs 解析命令行参数。args 为 os.Args[1:]。
func ParseArgs(args []string) (Args, error)

// Run 是 CLI 主入口，串联所有模块完成转换。
// stdout 和 stderr 通过参数注入，便于测试。
func Run(args Args, stdout io.Writer, stderr io.Writer) error
```

#### `internal/config`

```go
// Load 从环境变量加载配置。
func Load() Config
```

### 5.2 何时引入 Interface

若未来出现以下场景，再考虑提取接口：

- **多平台支持**（如同时支持 GitLab）：此时可从 `github.Client` 提取 `Fetcher` 接口
- **Web 层需要替换数据源测试**：此时可在 `cli.Run` 的参数中使用接口类型

当前不做，遵循 YAGNI。

---

## 6. 测试策略

### 6.1 各包测试方式

| 包 | 测试类型 | 测试工具 | 说明 |
|----|----------|----------|------|
| `parser` | 单元测试（表格驱动） | `testing` | 纯函数，直接输入输出断言。覆盖 spec 4.1 的全部 8 个验收标准 |
| `github` | 集成测试 | `testing` + `net/http/httptest` | 使用 `httptest.NewServer` 返回预制 JSON 响应，验证数据解析和分页逻辑。不 Mock 接口 |
| `converter` | 单元测试（表格驱动） | `testing` | 构造 `IssueData` 输入，断言输出 Markdown 字符串包含预期内容 |
| `cli` | 集成测试 | `testing` | 测试 `ParseArgs` 的参数解析逻辑（表格驱动） |
| `config` | 单元测试 | `testing` + `t.Setenv` | 设置/清除环境变量，验证 `Load()` 返回值 |

### 6.2 TDD 实施顺序

严格遵循宪法第二条，每个包的开发顺序为：

1. **`internal/parser`**（零依赖，第一个实现）
   - 先写 `parser_test.go`，覆盖 AC-URL-01 ~ AC-URL-08
   - 再实现 `parser.go` 使测试通过
2. **`internal/config`**（零依赖）
   - 先写测试验证环境变量读取
   - 再实现
3. **`internal/github`**（依赖外部库）
   - 先写 httptest 测试，预制 JSON 响应
   - 再实现 REST（Issue/PR）和 GraphQL（Discussion）客户端
4. **`internal/converter`**（依赖 parser 和 github 的类型）
   - 先写测试，构造 IssueData，断言 Markdown 输出
   - 再实现渲染逻辑
5. **`internal/cli`**（组装层，最后实现）
   - 先写 ParseArgs 测试
   - 再实现参数解析和 Run 流程
6. **`cmd/issue2md/main.go`**（极简入口，无需单独测试）

---

## 7. 错误处理策略

### 7.1 退出码映射

| 退出码 | 含义 | 触发位置 |
|--------|------|----------|
| 0 | 成功 | `main.go` 正常返回 |
| 1 | 用户输入错误 | `cli.ParseArgs` 返回错误、`parser.Parse` 返回错误 |
| 2 | 运行时错误 | `github.Client.Fetch*` 返回错误、文件写入失败 |

### 7.2 错误包装链

```
main.go 捕获 error → 打印到 stderr → os.Exit(退出码)

cli.Run:
  parser.Parse 失败     → return fmt.Errorf("parsing URL: %w", err)
  github.Fetch* 失败    → return fmt.Errorf("fetching data: %w", err)
  文件写入失败           → return fmt.Errorf("writing output: %w", err)
```

### 7.3 cli.Run 中区分退出码

`cli.Run` 返回的 error 需要区分是用户输入错误（退出码 1）还是运行时错误（退出码 2）。方案：

```go
// internal/cli/run.go

// InputError 表示用户输入导致的错误（退出码 1）
type InputError struct {
    Err error
}

func (e *InputError) Error() string { return e.Err.Error() }
func (e *InputError) Unwrap() error { return e.Err }
```

`main.go` 中通过 `errors.As` 判断：

```go
var inputErr *cli.InputError
if errors.As(err, &inputErr) {
    os.Exit(1)
} else {
    os.Exit(2)
}
```

---

## 8. Makefile

```makefile
.PHONY: build test clean web

build:
	go build -o bin/issue2md ./cmd/issue2md

web:
	go build -o bin/issue2mdweb ./cmd/issue2mdweb

test:
	go test ./...

clean:
	rm -rf bin/
```

---

## 9. 实施路线图

| 阶段 | 内容 | 产出 |
|------|------|------|
| **Phase 1** | `internal/parser` | URL 解析功能，通过 AC-URL-01~08 全部测试 |
| **Phase 2** | `internal/config` | 环境变量配置读取 |
| **Phase 3** | `internal/github` — Issue | REST API 抓取 Issue 数据（含分页、Reactions） |
| **Phase 4** | `internal/converter` | Markdown 渲染，输出格式符合 spec 5.1 |
| **Phase 5** | `internal/cli` + `cmd/issue2md` | CLI 完整功能，端到端可用 |
| **Phase 6** | `internal/github` — PR | REST API 抓取 PR 数据（含 Review Comments） |
| **Phase 7** | `internal/github` — Discussion | GraphQL API 抓取 Discussion 数据 |
| **Phase 8** | 端到端验收 | 覆盖 spec 第 4 节全部验收标准 |
