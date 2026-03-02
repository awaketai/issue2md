# Spec 001: Core Functionality - URL to Markdown Conversion

**Version**: 1.0
**Date**: 2026-03-02
**Status**: Draft

---

## 1. 用户故事

### 1.1 CLI 用户故事

**US-CLI-01**: 作为一名开发者，我想输入一个 GitHub Issue URL，将其完整内容（标题、正文、元数据、评论）转换为 Markdown 输出到终端，以便我可以通过管道重定向进行归档。

**US-CLI-02**: 作为一名开发者，我想输入一个 GitHub PR URL，将其描述和所有评论（包括 Review Comments）按时间线平铺转换为 Markdown，以便我归档完整的代码审查对话。

**US-CLI-03**: 作为一名开发者，我想输入一个 GitHub Discussion URL，将其正文和所有嵌套回复按时间线平铺转换为 Markdown，以便我归档社区讨论。

**US-CLI-04**: 作为一名开发者，我想通过 `--with-reactions` 选项在输出中包含 Reactions 统计，以便了解社区对每条评论的反馈。

**US-CLI-05**: 作为一名开发者，我想通过 `-o` 参数将输出写入指定文件，以便直接归档而无需手动重定向。

**US-CLI-06**: 作为一名开发者，我想通过 `--verbose` 查看 API 请求的调试日志，以便在出问题时排查原因。

### 1.2 Web 用户故事（未来版本）

**US-WEB-01**: 作为一名用户，我想在网页上粘贴一个 Issue/PR/Discussion URL，点击按钮后在页面上预览生成的 Markdown，并可以一键下载为 `.md` 文件。

**US-WEB-02**: 作为一名用户，我想在网页上看到转换进度的实时反馈（如"正在获取评论..."），以便知道工具正在工作。

---

## 2. 功能性需求

### 2.1 URL 解析

工具必须通过解析 URL 的路径结构自动判断**资源类型**（Issue / PR / Discussion），这是核心体验。

#### 2.1.1 支持的 URL 模式

| 资源类型 | URL 模式 |
|----------|----------|
| Issue | `https://github.com/{owner}/{repo}/issues/{number}` |
| PR | `https://github.com/{owner}/{repo}/pull/{number}` |
| Discussion | `https://github.com/{owner}/{repo}/discussions/{number}` |

#### 2.1.2 平台范围

当前版本仅支持 GitHub（`github.com`）。如果用户输入的 URL host 不是 `github.com`，工具应报错退出并提示当前仅支持 GitHub。

### 2.2 CLI 接口

#### 2.2.1 基本用法

```bash
issue2md <url>                           # 输出到 stdout
issue2md <url> -o <filepath>             # 输出到文件
issue2md <url> --with-reactions          # 包含 Reactions 统计
issue2md <url> --verbose                 # 输出调试日志到 stderr
issue2md <url> --with-reactions -o out.md --verbose  # 组合使用
```

#### 2.2.2 参数定义

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `<url>` | positional | 是 | - | GitHub Issue / PR / Discussion 的完整 URL |
| `-o, --output` | string | 否 | stdout | 输出文件路径 |
| `--with-reactions` | bool | 否 | false | 在正文和评论下方显示 Reactions 统计 |
| `--verbose` | bool | 否 | false | 输出调试日志到 stderr |

#### 2.2.3 输出文件名约定

当使用 `-o` 指定一个**目录**时，工具自动按以下规则生成文件名：

```
{owner}_{repo}_issue_{number}.md
{owner}_{repo}_pr_{number}.md
{owner}_{repo}_discussion_{number}.md
```

### 2.3 认证

通过环境变量 `GITHUB_TOKEN` 传入 Personal Access Token。

- 未设置时，以未认证模式访问（仅支持公开仓库，且受更严格的速率限制：60 次/小时）。
- 已设置时，速率限制提升至 5000 次/小时。
- 工具不负责 token 的获取和管理。

### 2.4 抓取内容定义

#### 2.4.1 元数据

| 字段 | 说明 | 是否必须 |
|------|------|----------|
| 标题 | Issue / PR / Discussion 标题 | 是 |
| 状态 | Open / Closed / Merged 等 | 是 |
| 作者 | 创建者用户名 | 是 |
| 创建时间 | ISO 8601 格式 | 是 |
| Labels | 标签列表 | 是 |
| Assignees | 指派人列表 | 是 |
| Milestone | 里程碑名称 | 是（如有） |
| 关联 PR | 与 Issue 关联的 PR 列表 | 是（仅 Issue） |

#### 2.4.2 正文

- 保留原始 Markdown 内容
- 图片保留原始 URL，不下载到本地

#### 2.4.3 评论

所有评论按**时间线顺序**平铺展示，每条评论包含：

- 评论者用户名
- 评论时间（ISO 8601）
- 评论内容（原始 Markdown）
- Reactions 统计（仅在 `--with-reactions` 开启时显示）

**PR 特殊处理**：Review Comments（代码行级评论）与普通评论混合，按时间线统一排列。Review Comment 额外显示文件路径和行号。

**Discussion 特殊处理**：树形嵌套回复按时间线顺序平铺，与普通评论格式一致。

#### 2.4.4 Reactions（可选）

启用 `--with-reactions` 时，在正文和每条评论下方以引用块格式显示：

```
> 👍 5 | 👎 1 | ❤️ 3 | 🎉 2
```

仅显示数量 > 0 的 Reaction 类型。

### 2.5 分页处理

- 自动处理 GitHub API 分页，抓取全部评论，不设上限
- `--verbose` 模式下输出分页进度信息到 stderr

---

## 3. 非功能性需求

### 3.1 架构解耦

工具的核心逻辑必须与 CLI 层解耦，以便未来复用于 Web 服务：

- **URL 解析层**：负责将 URL 解析为资源类型、owner、repo、number 的结构化数据
- **数据获取层**：负责调用 GitHub API 获取数据，返回统一的数据模型
- **Markdown 渲染层**：接受统一数据模型，输出格式化的 Markdown 字符串
- **CLI 层**：仅负责参数解析和 I/O，调用上述各层完成工作

### 3.2 错误处理

所有错误直接报错退出，在 stderr 输出清晰的错误信息。不实现重试机制，保持 CLI 轻量。

#### 3.2.1 错误类型与行为

| 场景 | 行为 |
|------|------|
| URL 格式无效 | 报错退出（退出码 1），提示正确的 URL 格式 |
| 非 GitHub URL | 报错退出（退出码 1），提示当前仅支持 GitHub |
| 无法识别的资源类型 | 报错退出（退出码 1），提示支持的资源类型（Issue / PR / Discussion） |
| 资源不存在 (404) | 报错退出（退出码 2），透传 GitHub API 返回的错误信息 |
| 认证失败 (401/403) | 报错退出（退出码 2），透传 GitHub API 返回的错误信息 |
| 网络错误 | 报错退出（退出码 2），输出底层错误信息 |
| API 速率限制 (429) | 报错退出（退出码 2），透传 GitHub API 返回的错误信息 |

#### 3.2.2 退出码

| 退出码 | 含义 |
|--------|------|
| 0 | 成功 |
| 1 | 用户输入错误（参数错误、URL 无效、资源类型不支持） |
| 2 | 运行时错误（网络错误、API 错误） |

### 3.3 构建与分发

- 编译为单个二进制文件，无外部运行时依赖
- 使用 `Makefile` 管理构建、测试等标准化操作

### 3.4 调试能力

- `--verbose` 模式下，通过 stderr 输出调试信息，包括：
  - API 请求 URL
  - HTTP 响应状态码
  - 分页信息（第几页/共几页）
  - 速率限制剩余配额

---

## 4. 验收标准

### 4.1 URL 解析

| # | 测试用例 | 输入 | 期望结果 |
|---|----------|------|----------|
| AC-URL-01 | 解析 GitHub Issue URL | `https://github.com/golang/go/issues/12345` | owner=golang, repo=go, type=issue, number=12345 |
| AC-URL-02 | 解析 GitHub PR URL | `https://github.com/golang/go/pull/67890` | owner=golang, repo=go, type=pr, number=67890 |
| AC-URL-03 | 解析 GitHub Discussion URL | `https://github.com/golang/go/discussions/111` | owner=golang, repo=go, type=discussion, number=111 |
| AC-URL-04 | 无效 URL | `not-a-url` | 错误：URL 格式无效（退出码 1） |
| AC-URL-05 | 非 GitHub URL | `https://gitlab.com/group/project/-/issues/42` | 错误：当前仅支持 GitHub（退出码 1） |
| AC-URL-06 | 不支持的路径结构 | `https://github.com/golang/go/wiki/Home` | 错误：无法识别资源类型（退出码 1） |
| AC-URL-07 | URL 带尾部斜杠 | `https://github.com/golang/go/issues/12345/` | 正常解析，等同于无尾部斜杠 |
| AC-URL-08 | URL 带查询参数 | `https://github.com/golang/go/issues/12345?foo=bar` | 正常解析，忽略查询参数 |

### 4.2 内容抓取

| # | 测试用例 | 期望结果 |
|---|----------|----------|
| AC-FETCH-01 | 抓取 Issue 基本信息 | 输出包含标题、状态、作者、创建时间、Labels、Assignees、Milestone |
| AC-FETCH-02 | 抓取 Issue 关联 PR | 输出包含关联的 PR 编号和链接 |
| AC-FETCH-03 | 抓取 Issue 评论 | 所有评论按时间线顺序排列，包含评论者、时间、内容 |
| AC-FETCH-04 | 抓取 PR Review Comments | Review Comments 与普通评论混合按时间线排列，Review Comment 显示文件路径和行号 |
| AC-FETCH-05 | 抓取 Discussion 嵌套回复 | 所有嵌套回复按时间线平铺展示 |
| AC-FETCH-06 | Reactions 默认不显示 | 不传 `--with-reactions` 时，输出中不包含 Reactions |
| AC-FETCH-07 | Reactions 开启后显示 | 传 `--with-reactions` 时，正文和评论下方显示 Reactions 统计 |
| AC-FETCH-08 | 图片链接保留 | 正文和评论中的图片保留原始 URL，不被修改 |
| AC-FETCH-09 | 分页评论完整抓取 | 超过单页限制的评论能完整抓取（通过 API 分页） |

### 4.3 CLI 行为

| # | 测试用例 | 期望结果 |
|---|----------|----------|
| AC-CLI-01 | 默认输出到 stdout | 不指定 `-o` 时，Markdown 输出到 stdout |
| AC-CLI-02 | `-o` 指定文件路径 | Markdown 写入指定文件 |
| AC-CLI-03 | `-o` 指定目录 | 文件按命名规则自动生成在指定目录下 |
| AC-CLI-04 | `--verbose` 调试日志 | 调试信息输出到 stderr，不影响 stdout 内容 |
| AC-CLI-05 | 无参数运行 | 输出使用帮助信息，退出码 1 |

### 4.4 错误处理

| # | 测试用例 | 期望结果 |
|---|----------|----------|
| AC-ERR-01 | 资源不存在 (404) | 报错退出（退出码 2），stderr 输出 GitHub API 返回的错误信息 |
| AC-ERR-02 | 认证失败 (401) | 报错退出（退出码 2），stderr 输出 GitHub API 返回的错误信息 |
| AC-ERR-03 | 速率限制 (429) | 报错退出（退出码 2），stderr 输出 GitHub API 返回的错误信息 |
| AC-ERR-04 | 网络不可达 | 报错退出（退出码 2），stderr 输出网络错误信息 |

---

## 5. 输出格式示例

### 5.1 GitHub Issue 示例

```markdown
# [Issue] Fix login timeout (#12345)

| Field | Value |
|-------|-------|
| **State** | Open |
| **Author** | @octocat |
| **Created** | 2024-01-15T10:30:00Z |
| **Labels** | bug, priority/high |
| **Assignees** | @dev1, @dev2 |
| **Milestone** | v2.0 |
| **Linked PRs** | #456, #789 |

---

When attempting to login with SSO credentials, the request times out
after 30 seconds. This happens consistently on the production server.

Steps to reproduce:
1. Navigate to /login
2. Click "SSO Login"
3. Wait 30 seconds

Expected: Login succeeds within 5 seconds.

> 👍 5 | ❤️ 2

---

## Comments (3)

### @user1 — 2024-01-16T08:00:00Z

I can reproduce this issue. It seems to be related to the DNS
resolution timeout in the SSO provider configuration.

> 👍 2

---

### @user2 — 2024-01-17T14:30:00Z

Fixed in #456. The timeout was caused by a misconfigured retry policy.

---

### @maintainer — 2024-01-18T09:00:00Z

Closing as fixed. Thanks @user2!

---
```

### 5.2 GitHub PR 示例（含 Review Comment）

```markdown
# [PR] Add retry logic for SSO login (#456)

| Field | Value |
|-------|-------|
| **State** | Merged |
| **Author** | @user2 |
| **Created** | 2024-01-17T16:00:00Z |
| **Labels** | enhancement, auth |
| **Assignees** | @user2 |
| **Milestone** | v2.0 |

---

This PR adds exponential backoff retry logic to the SSO login flow
to fix #12345.

Changes:
- Add `retryWithBackoff()` helper function
- Set max retries to 3
- Add unit tests for retry logic

---

## Comments (3)

### @reviewer — 2024-01-18T10:00:00Z

Overall looks good! Left a few comments on the implementation.

---

### @reviewer — 2024-01-18T10:05:00Z `internal/auth/sso.go#L42`

Should we make the max retry count configurable?

---

### @user2 — 2024-01-18T11:00:00Z

@reviewer Good point, I've made it a constant for now. We can make
it configurable if needed later.

---
```

### 5.3 GitHub Discussion 示例

```markdown
# [Discussion] RFC: New authentication API design (#789)

| Field | Value |
|-------|-------|
| **State** | Open |
| **Author** | @architect |
| **Created** | 2024-02-01T12:00:00Z |
| **Labels** | rfc |

---

I'd like to propose a new design for our authentication API.
The current implementation has several issues:

1. Too many round trips
2. No support for MFA
3. Session management is fragile

Proposed solution: Move to JWT-based auth with refresh tokens.

---

## Comments (3)

### @backend-dev — 2024-02-02T09:00:00Z

+1 on JWT. We should also consider adding OAuth2 support
for third-party integrations.

---

### @security-lead — 2024-02-02T10:30:00Z

JWT sounds good but we need to be careful about token
size. I'd suggest keeping the claims minimal.

---

### @architect — 2024-02-02T14:00:00Z

Great feedback. I'll update the RFC with the OAuth2 scope
and minimal claims approach.

---
```

