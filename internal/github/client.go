package github

import (
	"context"
	"io"

	gogithub "github.com/google/go-github/v68/github"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

// Client 封装 GitHub REST 和 GraphQL API 客户端
type Client struct {
	restClient    *gogithub.Client
	graphqlClient *githubv4.Client
	verbose       io.Writer
}

// NewClient 创建 GitHub API 客户端。
// token 为空字符串时以未认证模式访问。
// verbose 传入 io.Writer 输出调试日志，传 nil 则静默。
func NewClient(token string, verbose io.Writer) *Client {
	var restClient *gogithub.Client
	var graphqlClient *githubv4.Client

	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		httpClient := oauth2.NewClient(context.Background(), ts)
		restClient = gogithub.NewClient(httpClient)
		graphqlClient = githubv4.NewClient(httpClient)
	} else {
		restClient = gogithub.NewClient(nil)
		graphqlClient = githubv4.NewClient(nil)
	}

	return &Client{
		restClient:    restClient,
		graphqlClient: graphqlClient,
		verbose:       verbose,
	}
}

// FetchIssue 获取 Issue 完整数据。自动处理分页。
func (c *Client) FetchIssue(ctx context.Context, owner, repo string, number int) (IssueData, error) {
	return IssueData{}, nil
}

// FetchPR 获取 PR 完整数据。普通评论与 Review Comments 按时间线合并排序。
func (c *Client) FetchPR(ctx context.Context, owner, repo string, number int) (IssueData, error) {
	return IssueData{}, nil
}

// FetchDiscussion 获取 Discussion 完整数据。嵌套回复按时间线平铺。
func (c *Client) FetchDiscussion(ctx context.Context, owner, repo string, number int) (IssueData, error) {
	return IssueData{}, nil
}
