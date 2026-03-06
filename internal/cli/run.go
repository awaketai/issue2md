package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/awaketai/issue2md/internal/config"
	"github.com/awaketai/issue2md/internal/converter"
	"github.com/awaketai/issue2md/internal/github"
	"github.com/awaketai/issue2md/internal/parser"
)

// Run 是 CLI 主入口，串联所有模块完成转换。
// stdout 和 stderr 通过参数注入，便于测试。
func Run(args Args, stdout io.Writer, stderr io.Writer) error {
	cfg := config.Load()

	var verboseWriter io.Writer
	if args.Verbose {
		verboseWriter = stderr
	}
	client := github.NewClient(cfg.GitHubToken, verboseWriter)

	return runCore(args, client, stdout, stderr)
}

// runCore 是 Run 的内部实现，接受已构造的 github.Client 便于测试注入。
func runCore(args Args, client *github.Client, stdout io.Writer, stderr io.Writer) error {
	// 1. 解析 URL
	resource, err := parser.Parse(args.URL)
	if err != nil {
		return &InputError{Err: fmt.Errorf("parsing URL: %w", err)}
	}

	// 2. 根据资源类型获取数据
	ctx := context.Background()
	var data github.IssueData
	switch resource.Type {
	case parser.TypeIssue:
		data, err = client.FetchIssue(ctx, resource.Owner, resource.Repo, resource.Number)
	case parser.TypePR:
		data, err = client.FetchPR(ctx, resource.Owner, resource.Repo, resource.Number)
	case parser.TypeDiscussion:
		data, err = client.FetchDiscussion(ctx, resource.Owner, resource.Repo, resource.Number)
	}
	if err != nil {
		return fmt.Errorf("fetching data: %w", err)
	}

	// 3. 转换为 Markdown
	opts := converter.Options{WithReactions: args.WithReactions}
	md := converter.ToMarkdown(data, resource, opts)

	// 4. 输出
	if args.Output == "" {
		if _, err := fmt.Fprint(stdout, md); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		return nil
	}

	outPath := resolveOutputPath(args.Output, resource)
	if err := os.WriteFile(outPath, []byte(md), 0644); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}

// resolveOutputPath 解析输出路径。
// 如果 output 是已存在的目录，按 {owner}_{repo}_{type}_{number}.md 规则生成文件名。
// 否则直接使用 output 作为文件路径。
func resolveOutputPath(output string, resource parser.Resource) string {
	info, err := os.Stat(output)
	if err == nil && info.IsDir() {
		filename := fmt.Sprintf("%s_%s_%s_%d.md",
			resource.Owner,
			resource.Repo,
			string(resource.Type),
			resource.Number,
		)
		return filepath.Join(output, filename)
	}
	return output
}
