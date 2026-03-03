# 任务列表 (Task Breakdown)

**基于**: plan.md, spec.md v1.0, constitution.md v1.0
**Date**: 2026-03-03

> **图例说明**
> - `[P]` = 可与同阶段内其他 `[P]` 任务并行执行
> - `depends: T-XX` = 必须在指定任务完成后才能开始
> - 每个任务仅涉及 **一个主要文件** 的创建或修改
> - 严格遵循 TDD：测试任务（RED）始终排在实现任务（GREEN）之前

---

## Phase 1: Foundation（基础设施与数据结构）

本阶段目标：搭建项目骨架，定义所有核心数据结构，确保 `go build ./...` 和 `go test ./...` 能通过。

### T-01 [P] 升级 go.mod 并安装依赖

- **文件**: `go.mod`
- **操作**:
  1. 将 `go 1.22.0` 升级为 `go 1.24`
  2. `go get google/go-github/v68`
  3. `go get shurcooL/githubv4`
  4. `go get golang.org/x/oauth2`
  5. `go mod tidy`
- **验收**: `go build ./...` 无报错

### T-02 [P] 创建 Makefile

- **文件**: `Makefile`
- **操作**: 创建包含 `build`、`test`、`clean`、`web` 四个 target 的 Makefile（内容见 plan.md 第 8 节）
- **验收**: `make test` 可执行（即使当前无测试文件，退出码为 0）

### T-03 [P] 创建项目目录结构

- **文件**: 多个空目录 + `.gitkeep`
- **操作**: 创建以下目录：
  - `cmd/issue2md/`
  - `cmd/issue2mdweb/`
  - `internal/parser/`
  - `internal/github/`
  - `internal/converter/`
  - `internal/cli/`
  - `internal/config/`
  - `web/templates/`
  - `web/static/`
- **验收**: 目录结构与 plan.md 3.1 一致

### T-04 定义 `parser.Resource` 和 `ResourceType`

- **文件**: `internal/parser/parser.go`
- **depends**: T-03
- **操作**: 创建 `parser.go`，定义 `ResourceType` 常量（`TypeIssue`, `TypePR`, `TypeDiscussion`）和 `Resource` 结构体。声明 `Parse` 函数签名（函数体暂时 `return Resource{}, nil`，桩实现）
- **验收**: `go build ./internal/parser` 通过

### T-05 定义 `github` 包核心数据结构

- **文件**: `internal/github/types.go`
- **depends**: T-03
- **操作**: 创建 `types.go`，定义 `Reactions`、`Comment`、`IssueData` 三个结构体（字段完全按照 plan.md 第 4 节）
- **验收**: `go build ./internal/github` 通过

### T-06 定义 `github.Client` 构造函数

- **文件**: `internal/github/client.go`
- **depends**: T-01, T-05
- **操作**: 创建 `client.go`，定义 `Client` 结构体（未导出字段：`restClient *gogithub.Client`、`graphqlClient *githubv4.Client`、`verbose io.Writer`）和 `NewClient(token string, verbose io.Writer) *Client` 构造函数。声明 `FetchIssue`、`FetchPR`、`FetchDiscussion` 方法签名（桩实现 `return IssueData{}, nil`）
- **验收**: `go build ./internal/github` 通过

### T-07 定义 `converter.Options` 和 `ToMarkdown` 签名

- **文件**: `internal/converter/converter.go`
- **depends**: T-04, T-05
- **操作**: 创建 `converter.go`，定义 `Options` 结构体和 `ToMarkdown` 函数签名（桩实现返回空字符串）
- **验收**: `go build ./internal/converter` 通过

### T-08 定义 `config.Config` 和 `Load` 函数

- **文件**: `internal/config/config.go`
- **depends**: T-03
- **操作**: 创建 `config.go`，定义 `Config` 结构体和 `Load() Config` 函数（桩实现返回零值）
- **验收**: `go build ./internal/config` 通过

### T-09 定义 `cli.Args`、`InputError` 和函数签名

- **文件**: `internal/cli/args.go`
- **depends**: T-04, T-05, T-06, T-07, T-08
- **操作**: 创建 `args.go`，定义 `Args` 结构体、`InputError` 错误类型、`ParseArgs` 函数签名（桩实现）
- **验收**: `go build ./internal/cli` 通过

### T-10 创建 `cli.Run` 函数签名

- **文件**: `internal/cli/run.go`
- **depends**: T-09
- **操作**: 创建 `run.go`，定义 `Run(args Args, stdout io.Writer, stderr io.Writer) error` 函数（桩实现返回 nil）
- **验收**: `go build ./internal/cli` 通过

### T-11 创建 `cmd/issue2md/main.go` 入口

- **文件**: `cmd/issue2md/main.go`
- **depends**: T-09, T-10
- **操作**: 创建极简 `main.go`，调用 `cli.ParseArgs` 和 `cli.Run`，根据 `InputError` 区分退出码 1/2（代码见 plan.md 第 7.3 节）
- **验收**: `go build ./cmd/issue2md` 通过

### T-12 Phase 1 门禁检查

- **depends**: T-01 ~ T-11 全部完成
- **操作**: 运行 `make test` 和 `go vet ./...`，确保零错误、零警告
- **验收**: 全部通过，所有包可编译，骨架代码完整

---

## Phase 2: GitHub Fetcher（API 交互逻辑，TDD）

本阶段目标：实现 `internal/parser`、`internal/config`、`internal/github` 三个包的完整功能。

### 2.1 URL Parser（`internal/parser`）

#### T-13 [P] RED: 编写 `parser_test.go` 表格驱动测试

- **文件**: `internal/parser/parser_test.go`
- **depends**: T-04
- **操作**: 编写表格驱动测试，覆盖 spec 4.1 全部 8 个验收标准：
  - `AC-URL-01`: Issue URL → owner=golang, repo=go, type=issue, number=12345
  - `AC-URL-02`: PR URL → type=pr
  - `AC-URL-03`: Discussion URL → type=discussion
  - `AC-URL-04`: 无效 URL → 返回错误
  - `AC-URL-05`: 非 GitHub URL → 返回错误（包含"仅支持 GitHub"提示）
  - `AC-URL-06`: 不支持的路径（wiki）→ 返回错误
  - `AC-URL-07`: 尾部斜杠 → 正常解析
  - `AC-URL-08`: 带查询参数 → 正常解析，忽略查询参数
- **验收**: `go test ./internal/parser` 全部 FAIL（RED 状态）

#### T-14 GREEN: 实现 `parser.Parse` 函数

- **文件**: `internal/parser/parser.go`
- **depends**: T-13
- **操作**: 使用 `net/url` 标准库实现 URL 解析逻辑：
  1. `url.Parse` 解析原始 URL
  2. 校验 scheme 为 `https`
  3. 校验 host 为 `github.com`
  4. `strings.Split` 解析路径：`/{owner}/{repo}/{type}/{number}`
  5. 去除尾部斜杠，忽略查询参数
  6. 映射路径段到 `ResourceType`：`issues`→Issue, `pull`→PR, `discussions`→Discussion
  7. 所有错误使用 `fmt.Errorf("...: %w", err)` 包装
- **验收**: `go test ./internal/parser` 全部 PASS（GREEN 状态）

### 2.2 Config（`internal/config`）

#### T-15 [P] RED: 编写 `config_test.go`

- **文件**: `internal/config/config_test.go`
- **depends**: T-08
- **操作**: 编写表格驱动测试：
  - Case 1: `GITHUB_TOKEN` 已设置 → `Config.GitHubToken` 返回对应值
  - Case 2: `GITHUB_TOKEN` 未设置 → `Config.GitHubToken` 为空字符串
- **验收**: `go test ./internal/config` 全部 FAIL（RED 状态）

#### T-16 GREEN: 实现 `config.Load` 函数

- **文件**: `internal/config/config.go`
- **depends**: T-15
- **操作**: 使用 `os.Getenv("GITHUB_TOKEN")` 读取环境变量
- **验收**: `go test ./internal/config` 全部 PASS

### 2.3 GitHub Client — Issue（`internal/github`）

#### T-17 RED: 编写 `FetchIssue` httptest 集成测试

- **文件**: `internal/github/issue_test.go`
- **depends**: T-06
- **操作**: 使用 `httptest.NewServer` 模拟 GitHub REST API 响应，编写测试：
  - Case 1: 正常 Issue — 预制 JSON 包含 title, state, author, labels, assignees, milestone, body, reactions。断言映射到 `IssueData` 各字段正确
  - Case 2: Issue 带评论 — 预制评论列表 JSON，断言 `Comments` 按时间排序且字段正确
  - Case 3: 404 Not Found — 预制 404 响应，断言返回错误
  - Case 4: 分页 — 预制带 `Link` header 的分页响应，断言评论全部抓取
- **验收**: `go test ./internal/github -run TestFetchIssue` 全部 FAIL

#### T-18 GREEN: 实现 `FetchIssue` 方法

- **文件**: `internal/github/issue.go`
- **depends**: T-17
- **操作**:
  1. 调用 `go-github` 的 `Issues.Get` 获取 Issue 元数据
  2. 调用 `Issues.ListComments` 获取评论（自动分页通过 `ListOptions` 控制）
  3. 将 `go-github` 的 `*github.Issue` 和 `[]*github.IssueComment` 映射为 `IssueData`
  4. Reactions 映射：从 `issue.Reactions` 读取各字段
  5. verbose 日志输出到 `c.verbose`（如非 nil）
  6. 所有错误使用 `fmt.Errorf` 包装
- **验收**: `go test ./internal/github -run TestFetchIssue` 全部 PASS

### 2.4 GitHub Client — PR（`internal/github`）

#### T-19 RED: 编写 `FetchPR` httptest 集成测试

- **文件**: `internal/github/pr_test.go`
- **depends**: T-06
- **操作**: 使用 `httptest.NewServer` 模拟响应，编写测试：
  - Case 1: 正常 PR — 元数据映射正确，state 为 "merged" 时正确识别
  - Case 2: PR 评论 + Review Comments 混合 — 断言按时间线合并排序，Review Comment 包含 `FilePath` 和 `Line`
  - Case 3: 仅 Review Comments 无普通评论 — 断言 Comments 列表仅包含 Review Comments
- **验收**: `go test ./internal/github -run TestFetchPR` 全部 FAIL

#### T-20 GREEN: 实现 `FetchPR` 方法

- **文件**: `internal/github/pr.go`
- **depends**: T-19
- **操作**:
  1. 调用 `PullRequests.Get` 获取 PR 元数据
  2. 调用 `Issues.ListComments` 获取普通评论
  3. 调用 `PullRequests.ListComments` 获取 Review Comments
  4. 合并两种评论，按 `CreatedAt` 排序
  5. Review Comment 映射：`FilePath` = `comment.Path`, `Line` = `comment.Line`
  6. PR state 特殊处理：如果 `pr.Merged` 为 true，state 设为 "merged"
- **验收**: `go test ./internal/github -run TestFetchPR` 全部 PASS

### 2.5 GitHub Client — Discussion（`internal/github`）

#### T-21 RED: 编写 `FetchDiscussion` 测试

- **文件**: `internal/github/discussion_test.go`
- **depends**: T-06
- **操作**: 使用 `httptest.NewServer` 模拟 GraphQL 端点，编写测试：
  - Case 1: 正常 Discussion — 元数据 + 正文映射正确
  - Case 2: 嵌套回复 — 树形结构平铺为时间线顺序
  - Case 3: Discussion 带 Labels — Labels 列表映射正确
- **验收**: `go test ./internal/github -run TestFetchDiscussion` 全部 FAIL

#### T-22 GREEN: 实现 `FetchDiscussion` 方法

- **文件**: `internal/github/discussion.go`
- **depends**: T-21
- **操作**:
  1. 定义 GraphQL query 结构体（`shurcooL/githubv4` 风格）
  2. Query Discussion 数据：title, body, author, labels, state, comments（含 replies）
  3. 将嵌套的 `comments.nodes[].replies.nodes[]` 平铺为 `[]Comment`，按 `CreatedAt` 排序
  4. 映射 Reactions
  5. Discussion 无 Assignees/Milestone/LinkedPRs，对应字段为零值
- **验收**: `go test ./internal/github -run TestFetchDiscussion` 全部 PASS

### T-23 Phase 2 门禁检查

- **depends**: T-14, T-16, T-18, T-20, T-22
- **操作**: 运行 `make test`，确保 `parser`、`config`、`github` 三个包全部测试通过
- **验收**: `make test` 退出码 0，零失败

---

## Phase 3: Markdown Converter（转换逻辑，TDD）

本阶段目标：实现 `internal/converter` 包，将 `IssueData` 转换为符合 spec 5.1~5.3 格式的 Markdown 字符串。

#### T-24 RED: 编写 `converter_test.go` — Issue 渲染测试

- **文件**: `internal/converter/converter_test.go`
- **depends**: T-07, T-14（需要 `parser.Resource` 和 `github.IssueData` 类型已定义）
- **操作**: 编写表格驱动测试，构造 `IssueData`，断言输出 Markdown：
  - Case 1: 基础 Issue — 断言包含 `# [Issue] Title (#123)`、元数据表格、正文、分隔线
  - Case 2: Issue 带 Labels/Assignees/Milestone — 断言元数据表格行存在
  - Case 3: Issue 带 LinkedPRs — 断言 `Linked PRs` 行显示 `#456, #789`
  - Case 4: Issue 无 Milestone — 断言元数据表格中不含 Milestone 行
  - Case 5: Issue 带评论 — 断言 `## Comments (N)` 标题、每条评论格式正确
- **验收**: `go test ./internal/converter` 全部 FAIL

#### T-25 RED: 追加 `converter_test.go` — Reactions 渲染测试

- **文件**: `internal/converter/converter_test.go`
- **depends**: T-24
- **操作**: 追加测试用例：
  - Case 6: `WithReactions=false` — 断言输出中不含 Reactions 引用块
  - Case 7: `WithReactions=true` — 断言正文和评论下方显示 `> 👍 5 | ❤️ 2` 格式
  - Case 8: Reaction 数量为 0 的类型不显示 — 仅 PlusOne=5, Heart=2 时，输出仅含这两项
- **验收**: 新增测试用例 FAIL

#### T-26 RED: 追加 `converter_test.go` — PR 和 Discussion 渲染测试

- **文件**: `internal/converter/converter_test.go`
- **depends**: T-24
- **操作**: 追加测试用例：
  - Case 9: PR 渲染 — 断言标题为 `# [PR] Title (#456)`，无 LinkedPRs 行
  - Case 10: PR Review Comment — 断言评论标题含 `` `file.go#L42` `` 标记
  - Case 11: Discussion 渲染 — 断言标题为 `# [Discussion] Title (#789)`
  - Case 12: 空评论列表 — 断言不输出 `## Comments` 部分
- **验收**: 新增测试用例 FAIL

#### T-27 GREEN: 实现 `converter.ToMarkdown` — 元数据与正文渲染

- **文件**: `internal/converter/converter.go`
- **depends**: T-24, T-25, T-26
- **操作**: 使用 `strings.Builder` 实现：
  1. 标题行：`# [{Type}] {Title} (#{Number})`（Type 首字母大写映射：issue→Issue, pr→PR, discussion→Discussion）
  2. 元数据表格：`| Field | Value |` 格式，逐行输出 State/Author/Created/Labels/Assignees/Milestone/LinkedPRs
  3. Milestone 为空时跳过该行；LinkedPRs 为空时跳过该行
  4. 正文渲染：分隔线 + 原始 Body
  5. Reactions 渲染（`WithReactions=true` 时）：遍历 Reactions 各字段，数量>0 的以 `> emoji count` 格式输出，用 ` | ` 分隔
  6. 评论渲染：`## Comments (N)` + 每条评论 `### @author — timestamp` + body
  7. PR Review Comment 特殊处理：标题追加 `` `FilePath#L{Line}` ``
- **验收**: `go test ./internal/converter` 全部 PASS

#### T-28 Phase 3 门禁检查

- **depends**: T-27
- **操作**: 运行 `make test`，确保 `converter` 包全部测试通过
- **验收**: `make test` 退出码 0，零失败

---

## Phase 4: CLI Assembly（命令行入口集成）

本阶段目标：实现 `internal/cli` 包和 `cmd/issue2md/main.go`，使工具端到端可用。

### 4.1 参数解析（`internal/cli`）

#### T-29 RED: 编写 `cli_test.go` — ParseArgs 表格驱动测试

- **文件**: `internal/cli/cli_test.go`
- **depends**: T-09
- **操作**: 编写表格驱动测试：
  - Case 1: 仅 URL — `Args{URL: "https://...", Output: "", WithReactions: false, Verbose: false}`
  - Case 2: URL + `-o output.md` — `Args{Output: "output.md"}`
  - Case 3: URL + `--output output.md`（长参数形式）
  - Case 4: URL + `--with-reactions` — `Args{WithReactions: true}`
  - Case 5: URL + `--verbose` — `Args{Verbose: true}`
  - Case 6: 全部组合 — 所有 Flag 同时使用
  - Case 7: 无参数 — 返回错误
  - Case 8: 未知 Flag — 返回错误
- **验收**: `go test ./internal/cli -run TestParseArgs` 全部 FAIL

#### T-30 GREEN: 实现 `cli.ParseArgs` 函数

- **文件**: `internal/cli/args.go`
- **depends**: T-29
- **操作**: 使用 `flag.NewFlagSet` 实现：
  1. 注册 `-o` / `--output` String Flag
  2. 注册 `--with-reactions` Bool Flag
  3. 注册 `--verbose` Bool Flag
  4. 解析后取第一个位置参数为 URL
  5. URL 为空时返回错误（附带使用帮助信息）
- **验收**: `go test ./internal/cli -run TestParseArgs` 全部 PASS

### 4.2 输出文件名生成

#### T-31 RED: 编写文件名生成测试

- **文件**: `internal/cli/cli_test.go`
- **depends**: T-30
- **操作**: 追加测试用例，测试 `-o` 指定目录时的自动文件名生成逻辑：
  - Case 1: `-o /tmp/` + Issue → `/tmp/golang_go_issue_12345.md`
  - Case 2: `-o /tmp/` + PR → `/tmp/golang_go_pr_67890.md`
  - Case 3: `-o /tmp/` + Discussion → `/tmp/golang_go_discussion_111.md`
  - Case 4: `-o output.md`（非目录）→ 直接使用 `output.md`
- **验收**: 新增测试 FAIL

#### T-32 GREEN: 实现输出路径解析逻辑

- **文件**: `internal/cli/run.go`
- **depends**: T-31
- **操作**: 在 `Run` 函数中实现：
  1. 如果 `Output` 为空，写入 `stdout`
  2. 如果 `Output` 是已存在的目录，按 `{owner}_{repo}_{type}_{number}.md` 规则生成文件名
  3. 否则直接作为文件路径使用
- **验收**: 文件名生成测试 PASS

### 4.3 主流程编排

#### T-33 RED: 编写 `cli.Run` 集成测试

- **文件**: `internal/cli/cli_test.go`
- **depends**: T-30, T-14, T-18, T-27
- **操作**: 编写端到端测试（使用 `httptest.NewServer` 模拟 GitHub API）：
  - Case 1: 正常 Issue URL → stdout 输出包含预期 Markdown 内容
  - Case 2: 无效 URL → 返回 `*InputError`
  - Case 3: API 404 → 返回非 InputError 的 error
  - Case 4: verbose 模式 → stderr 包含调试日志
- **验收**: `go test ./internal/cli -run TestRun` 全部 FAIL

#### T-34 GREEN: 实现 `cli.Run` 主流程

- **文件**: `internal/cli/run.go`
- **depends**: T-33
- **操作**: 实现完整调用链：
  1. `config.Load()` 获取 token
  2. `parser.Parse(args.URL)` 解析 URL（失败时包装为 `InputError`）
  3. `github.NewClient(token, verboseWriter)` 创建客户端
  4. 根据 `resource.Type` 调用对应的 `Fetch*` 方法
  5. `converter.ToMarkdown(data, resource, opts)` 生成 Markdown
  6. 写入 stdout 或文件
  7. 所有错误使用 `fmt.Errorf("...: %w", err)` 包装
- **验收**: `go test ./internal/cli -run TestRun` 全部 PASS

### 4.4 CLI 入口完善

#### T-35 完善 `cmd/issue2md/main.go`

- **文件**: `cmd/issue2md/main.go`
- **depends**: T-34
- **操作**: 确保 `main.go` 正确区分退出码：
  1. `ParseArgs` 失败 → 退出码 1
  2. `Run` 返回 `*InputError` → 退出码 1
  3. `Run` 返回其他 error → 退出码 2
  4. 所有错误信息输出到 stderr
- **验收**: `go build ./cmd/issue2md` 通过，手动测试退出码正确

### 4.5 最终验收

#### T-36 Phase 4 门禁检查

- **depends**: T-35
- **操作**:
  1. `make test` — 全部包测试通过
  2. `make build` — 编译成功
  3. `go vet ./...` — 零警告
- **验收**: 全部通过

#### T-37 端到端手动验收

- **depends**: T-36
- **操作**: 使用编译后的二进制文件，对照 spec 4.1~4.4 验收标准逐项验证：
  1. `./bin/issue2md https://github.com/golang/go/issues/12345` — 输出 Markdown 到 stdout
  2. `./bin/issue2md <url> -o /tmp/` — 自动生成文件名
  3. `./bin/issue2md <url> --with-reactions` — 包含 Reactions
  4. `./bin/issue2md <url> --verbose` — stderr 输出调试日志
  5. `./bin/issue2md` （无参数）— 退出码 1，输出帮助
  6. `./bin/issue2md https://gitlab.com/...` — 退出码 1，提示仅支持 GitHub
  7. `./bin/issue2md <不存在的 issue url>` — 退出码 2，透传 API 错误
- **验收**: 全部验收标准通过

---

## 任务依赖关系总览

```
Phase 1: Foundation
  T-01 [P] ──┐
  T-02 [P]   │
  T-03 [P] ──┤
              ├── T-04 ──┐
              ├── T-05 ──┤
              │          ├── T-06 ──┐
              ├── T-08   │         │
              │          │         │
              ├── T-07 ◄─┤         │
              │          │         │
              └──────────┴── T-09 ◄┘
                              │
                              ├── T-10
                              │    │
                              └────┴── T-11
                                        │
                                       T-12

Phase 2: GitHub Fetcher
  T-13 [P] → T-14          (parser: RED → GREEN)
  T-15 [P] → T-16          (config: RED → GREEN)
  T-17 → T-18              (github/issue: RED → GREEN)
  T-19 → T-20              (github/pr: RED → GREEN)
  T-21 → T-22              (github/discussion: RED → GREEN)
  T-14, T-16, T-18, T-20, T-22 → T-23

Phase 3: Markdown Converter
  T-24 → T-25 [P] ┐
              T-26 [P] ┤
                       └── T-27 → T-28

Phase 4: CLI Assembly
  T-29 → T-30 → T-31 → T-32
                   │
  T-33 ◄───────────┘
    │
  T-34 → T-35 → T-36 → T-37
```

---

## 统计

| 阶段 | 任务数 | 测试任务 (RED) | 实现任务 (GREEN) | 其他 |
|------|--------|---------------|-----------------|------|
| Phase 1: Foundation | 12 | 0 | 0 | 12（骨架/基础设施） |
| Phase 2: GitHub Fetcher | 11 | 5 | 5 | 1（门禁） |
| Phase 3: Markdown Converter | 5 | 3 | 1 | 1（门禁） |
| Phase 4: CLI Assembly | 9 | 3 | 3 | 3（门禁/验收） |
| **合计** | **37** | **11** | **9** | **17** |
