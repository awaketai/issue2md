package github

import (
	"context"
	"fmt"
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

// newGraphQLClientForURL 创建指向自定义 URL 的 GraphQL 客户端（用于测试）。
func newGraphQLClientForURL(url string) *githubv4.Client {
	return githubv4.NewEnterpriseClient(url, nil)
}

// NewTestClient 创建一个指向自定义 URL 的测试客户端（仅 REST API）。
// 用于 httptest 集成测试。
func NewTestClient(baseURL string, verbose io.Writer) *Client {
	restClient := gogithub.NewClient(nil)
	restClient.BaseURL, _ = restClient.BaseURL.Parse(baseURL + "/")

	return &Client{
		restClient: restClient,
		verbose:    verbose,
	}
}

// logf 向 verbose writer 输出调试日志（如果非 nil）。
func (c *Client) logf(format string, args ...any) {
	if c.verbose != nil {
		fmt.Fprintf(c.verbose, format+"\n", args...)
	}
}
