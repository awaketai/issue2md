package parser

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

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

// Parse 解析 GitHub URL，返回 Resource。
// 错误场景：URL 格式无效、host 不是 github.com、路径无法识别。
func Parse(rawURL string) (Resource, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return Resource{}, fmt.Errorf("parsing URL: %w", err)
	}

	if u.Scheme != "https" {
		return Resource{}, fmt.Errorf("unsupported scheme %q: only https is supported", u.Scheme)
	}

	if u.Host != "github.com" {
		return Resource{}, fmt.Errorf("unsupported host %q: only github.com is supported", u.Host)
	}

	// 去除尾部斜杠，按 / 切分路径
	path := strings.TrimSuffix(u.Path, "/")
	path = strings.TrimPrefix(path, "/")
	segments := strings.Split(path, "/")

	// 路径格式: {owner}/{repo}/{type}/{number}
	if len(segments) < 4 {
		return Resource{}, fmt.Errorf("invalid path %q: expected /{owner}/{repo}/{type}/{number}", u.Path)
	}

	owner := segments[0]
	repo := segments[1]
	resourceTypeStr := segments[2]
	numberStr := segments[3]

	var resourceType ResourceType
	switch resourceTypeStr {
	case "issues":
		resourceType = TypeIssue
	case "pull":
		resourceType = TypePR
	case "discussions":
		resourceType = TypeDiscussion
	default:
		return Resource{}, fmt.Errorf("unsupported resource type %q: expected issues, pull, or discussions", resourceTypeStr)
	}

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return Resource{}, fmt.Errorf("invalid number %q: %w", numberStr, err)
	}

	return Resource{
		Owner:  owner,
		Repo:   repo,
		Type:   resourceType,
		Number: number,
	}, nil
}
