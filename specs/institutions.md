# institution

经常需要在本地归档或者引GitHub上的Issue和Pull Request。手动复制粘贴太麻烦了，我想要做一个命令行工具，给它一个URL，它就能自动帮我把那个Issue
或PR的所有内容（标题、正文、评论等）转换成一个格式漂亮的Markdown文件保存下来

目前处在模糊意图阶段

+ “所有内容”具体指什么？Reactions要不要？用户链接要不要？
+ 命令行参数应该怎么设计？需要支持哪些Flag？
+ 除了CLI，未来是否考虑支持web界面？

## 角色设定与目标对齐

你好，现在的任务是：我们要从零开始设计并实现 `issue2md` 工具。
你现在不仅是自身的Go语言工程师，更是一位经验丰富的产品经理。我有一个初步的想法，需要你通过向我提问，帮助我澄清需求、挖掘边缘场景，最终目标是共创一份
高质量的 `spec.md` 
我的初步想法是：**做一个命令行工具，输入一个Github Issue/PR/Discussion的URL，它就能自动将其转换为Markdown文件。**
请开始你的提问

## 第一轮提问

Q1: "内容"的精确范围                                                                    
                                                                                          
一个 GitHub Issue 包含很多元素，请告诉我哪些需要抓取：

┌────────────────────────────┬────────────┐
│            元素             │ 是否需要？ │
├────────────────────────────┼────────────┤
│ 标题 + 正文                 │ ?          │
├────────────────────────────┼────────────┤
│ 所有评论（含评论者、时间）      │ ?          │
├────────────────────────────┼────────────┤
│ Labels                     │ ?          │
├────────────────────────────┼────────────┤
│ Assignees                  │ ?          │
├────────────────────────────┼────────────┤
│ Milestone                  │ ?          │
├────────────────────────────┼────────────┤
│ Reactions（点赞/表情）       │ ?          │
├────────────────────────────┼────────────┤
│ Issue 关联的 PR 链接         │ ?          │
├────────────────────────────┼────────────┤
│ 正文/评论中的图片             │ ?          │
└────────────────────────────┴────────────┘

特别是图片——你是只保留原始图片 URL，还是希望工具把图片也下载到本地？

Q2: PR 和 Discussion 的额外内容

- 对于 PR：是否需要抓取 code review
comments（代码行级别的评论）？还是只要普通评论就够了？
- 对于 Discussion：Discussion 有"回复"的树形结构，你期望怎么展平到 Markdown 里？

Q3: 认证与私有仓库

- 工具是否需要支持私有仓库？如果需要，你倾向于用什么方式认证——GITHUB_TOKEN 环境变量、gh
auth 集成、还是命令行参数传入 token？

Q4: 输出

- 输出文件名你有偏好吗？比如 {owner}_{repo}_issue_{number}.md，还是用户自己指定？
- 输出默认写到当前目录，还是需要 -o 参数指定目录？
- 是否需要支持输出到 stdout（方便管道操作），比如 issue2md <url> | pbcopy？

---
先回答这一轮，我再根据你的答案继续深挖 CLI 设计、错误处理、和未来扩展性的问题。

Q1：

标题 + 正文 是需要的

我们只关注PR的描述和Review评论，因此评论是需要的，包含评论者，时间。

Labels / Assignees / Milestone 也是需要的。

Reactions 请作为一个可选项（通过Flag控制）支持，如果开启，在主楼和评论下方显示Reactions统计。

Issue 关联的 PR链接也是需要的

正文/评论中的图片，直接保留原始的图片链接既可，不需要下载到本地

Q2：

- 对于 PR：需要抓取 comments 和普通评论
- 对于 Discussion：需要抓取回复的树形结构，展平到Markdown里

Q3：
认证只需要支持公有仓库，认证方式仅支持通过环境变量 GITHUB_TOKEN 传入 Personal Access Token，这是 CLI 工具最标准、最简单的做法。

Q4：

- 输出文件名按照：{owner}_{repo}_issue_{number}.md
- 输出默认输出到标准输出（stdout），方便用户重定向，同时也支持通过参数（-o）指定输出文件路径

## 第二轮提问：CLI 设计与边缘场景                                                          
   
Q5: CLI 调用方式                                                                        
                
我设想的基本用法是：

# 输出到 stdout
issue2md https://github.com/owner/repo/issues/123

# 输出到文件
issue2md https://github.com/owner/repo/issues/123 -o output.md

# 开启 reactions
issue2md https://github.com/owner/repo/issues/123 --with-reactions

- 这个交互风格你认可吗？还是你更倾向于子命令风格，比如 issue2md fetch <url>？
- 除了 -o 和 --with-reactions，你觉得还需要其他 Flag 吗？比如 --verbose 输出调试信息？

Q6: URL 解析与输入校验

- 如果用户输入了一个非 GitHub URL（比如 GitLab
的链接），工具应该怎么处理？直接报错退出，还是给出友好提示？
- 如果用户输入的 Issue/PR 编号不存在（404），你期望的错误信息是什么样的？

Q7: API 速率限制

- GitHub API 对未认证请求限制 60 次/小时，认证后是 5000 次/小时。一个有几百条评论的大
Issue 可能需要多次分页请求。你希望工具在遇到 rate limit
时怎么做——直接报错、还是等待后自动重试？

Q8: Markdown 输出格式

我需要确认输出的 Markdown 结构。我的初步想法是这样：

# [Issue] Title (#123)

- **State**: Open
- **Labels**: bug, enhancement
- **Assignees**: @user1, @user2
- **Milestone**: v1.0
- **Linked PRs**: #456, #789
- **Created**: 2024-01-01 by @author

---

(正文内容)

---

## Comments (3)

### Comment 1 - @user1 (2024-01-02)

(评论内容)

> 👍 5 | ❤️ 2

---

- 这个结构是否符合你的预期？有需要调整的地方吗？
- 对于 PR 的 review comments（代码行级评论），你希望怎么呈现？我的想法是在文件路径和行号
下面展示评论内容，类似：

### Review Comments

#### `src/main.go` (Line 42)
**@reviewer** (2024-01-03):
> This should handle the error case.

Q9: Discussion 的树形结构

对于 Discussion 的嵌套回复，我打算用缩进引用来表示层级：

### Reply 1 - @user1

内容...

> #### Reply 1.1 - @user2
>
> 嵌套回复内容...

- 这种方式可以接受吗？还是你有更好的展现方式偏好？
- 嵌套层级需要限制吗？（Discussion 理论上可以嵌套很深）

---
回答完这轮后，我就有足够的信息来起草 spec.md 了。

Q5：

- 交互风格按照你的基本想法来实现。
- 同时增加 --verbose Flag 来输出调试信息。

Q6：

- 这个工具应该是支持不同代码仓库的 issue 抓取的，因此只需要按照用户提供的 url抓取既可，同时工具应该解析URL结构来判断是 Issue/PR/Discussion，这是核心体验。
- 如果用户输入的 Issue/PR 编号不存在（404），则提示 Issue/PR 不存在，并给出调用API返回的信息。

Q7：

遇到API速率限制后，进行重试，每次重试的时间间隔需要响应增加，如果重试3次后，还是失败，则提示错误。

Q8：

直接把所有Review Comments和其他Comments一起按时间线平铺展示即可。我们的目标是归档“发生了什么对话”。

Q9：

Discussion 的嵌套回复，按照时间线平铺展示既可。


## 生成文档

1.  在项目根目录下创建一个名为 specs/001-core-functionality/ 的目录。
2.  在该目录下创建 spec.md 文件。内容必须包含：    

*   **用户故事**（含CLI和未来的Web版）    
*   **功能性需求**（涵盖我们讨论的所有细节：URL识别、Flags、Token、Markdown结构）    
*   **非功能性需求**（架构解耦、错误处理）    
*   **验收标准**（列出具体的测试Case）    
*   **输出格式示例**（包含你刚才设计的Markdown结构和Frontmatter）请直接执行工具生成文件。


## 后续优化

**4. 错误处理：**
* 如果URL无效或资源不存在，直接报错退出，并在stderr输出清晰的错误信息。
* **不需要**复杂的重试机制，保持CLI轻量。
* API限流时，直接透传GitHub API的错误信息给用户即可。


## 生成项目目录结构

非常好。基于这份 `spec.md` 和我们的 `constitution.md`（特别是关于包内聚的原则），请为这个功能设计详细的包结构。

请调用 `Bash` 工具，创建以下目录（如果尚未存在）：

- `cmd/issue2md/` (CLI入口)
- `cmd/issue2mdweb/` (Web入口)
- `internal/`
   ├── github/      # GitHub API 交互
   ├── parser/      # URL 解析与类型识别
   ├── converter/   # 数据转换为 Markdown
   ├── cli/         # 命令行接口
   └── config/      # 配置管理
- `web/templates/` (Web模板)
- `web/static/` (静态资源)

同时，请在 `specs/001-core-functionality/` 下创建一个 `api-sketch.md`，简要描述一下 `internal/converter` 和 `internal/github` 包对外暴露的主要接口（Interface/Function签名），作为后续开发的参考。

## 技术规划
做什么和怎么做

**Prompt1：生成技术方案**

`@specs/001-core-functionality/spec.md`

你现在是`issue2md`项目的首席架构师，你的任务时基于我提供的`spec.md`以及我们已有的`constitution.md`为项目生成一份详细的技术实现方案 `plan.md`。

**技术栈约束（必须遵循）：**
- **语言**：Go（>=1.21.0）。
- **web框架：** 仅使用标准库`net/http`，不引入Gin或Echo等外部框架，遵循简单性原则。
- **Github API客户端：** 使用`google/go-github`库，结合GraphQL API v4。
- **Markdown输出处理：** 如有必要，可以引入第三方库，使得Markdown可以有比较好的展示效果。
- **数据存储：** 本项目初期不需要数据库，所有数据通过API实时获取。

**方案内容要求（必须包含）：**

1.**技术上下文总结：** 明确上述技术选型
2.**合宪性审查：** 逐条对照`constitution.md`的原则，检查并确认技术方案符合所有条款（特别是包内聚，错误处理，TDD）。
3.**项目结构细化：** 明确 `cmd/`，`internal/github`，`internal/converter`等包的具体职责和依赖关系。
4.**核心数据结构：** 定义在模块间流转的核心Go `struct`，例如`IssueData`，必须包含Spec中要求的所有字段（Titile，Author，Reactions等）。
5.**接口设计：** 定义`internal`包对外暴漏的关键Interface。

将文档保存到 `./specs/001-core-functionality/plan.md` 中。

## 生成任务列表

方案非常完美

现在，请扮演技术组长。请仔细阅读 `./specs/001-core-functionality/spec.md` 和 `/specs/001-core-functionality/plan.md`。
你的目标是将 `plan.md` 中描述的技术方案，分解成一个详尽的、原子化的、有依赖关系的、可被AI直接执行的任务列表。

**关键要求：**

1.**任务粒度：** 每个任务应该只涉及一个主要文件的修改或创建一个新文档。不要出现“实现所有功能”这种大任务。
2.**TDD强制：** 根据 `consititution.md` 的“测试先行铁律”，必须先生成测试任务，然后生成实现任务。
3.**并行标记：** 对于没有依赖关系的任务，请标记 `[P]`。
4.**阶段划分：** 即便 `plan.md` 中包含了粗略的阶段划分，也要以下面的为准：
  * **Phase 1：Foundation**（数据结构定义）
  * **Phase 2：GIthubFetcher**（API交互逻辑，TDD）
  * **Phase 3：Markdown Converter**（转换逻辑，TDD）
  * **Phase 4：CLI Asssembly**（命令行入口集成）
    
完成后，将生成的任务列表写入 `./specs/001-core-functionality/task.md`
