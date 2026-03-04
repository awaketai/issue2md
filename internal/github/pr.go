package github

import (
	"context"
	"fmt"
	"sort"

	gogithub "github.com/google/go-github/v68/github"
)

// FetchPR 获取 PR 完整数据。普通评论与 Review Comments 按时间线合并排序。
func (c *Client) FetchPR(ctx context.Context, owner, repo string, number int) (IssueData, error) {
	c.logf("fetching PR %s/%s#%d", owner, repo, number)

	pr, _, err := c.restClient.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return IssueData{}, fmt.Errorf("fetching PR %s/%s#%d: %w", owner, repo, number, err)
	}

	// PR body reactions 通过 Issues.Get 获取（PullRequest 结构体不含 Reactions）
	issue, _, err := c.restClient.Issues.Get(ctx, owner, repo, number)
	if err != nil {
		return IssueData{}, fmt.Errorf("fetching issue data for PR %s/%s#%d: %w", owner, repo, number, err)
	}

	data := mapPRToData(pr)
	if issue.Reactions != nil {
		data.Reactions = mapReactions(issue.Reactions)
	}

	// Fetch regular issue comments
	issueComments, err := c.fetchIssueComments(ctx, owner, repo, number)
	if err != nil {
		return IssueData{}, fmt.Errorf("fetching issue comments for PR %s/%s#%d: %w", owner, repo, number, err)
	}

	// Fetch PR review comments
	reviewComments, err := c.fetchPRReviewComments(ctx, owner, repo, number)
	if err != nil {
		return IssueData{}, fmt.Errorf("fetching review comments for PR %s/%s#%d: %w", owner, repo, number, err)
	}

	// Merge and sort by CreatedAt
	allComments := append(issueComments, reviewComments...)
	sort.Slice(allComments, func(i, j int) bool {
		return allComments[i].CreatedAt.Before(allComments[j].CreatedAt)
	})

	if allComments == nil {
		allComments = []Comment{}
	}
	data.Comments = allComments

	c.logf("fetched PR %s/%s#%d: %d comments", owner, repo, number, len(allComments))
	return data, nil
}

// fetchPRReviewComments 获取 PR 的所有 Review Comments（自动分页）。
func (c *Client) fetchPRReviewComments(ctx context.Context, owner, repo string, number int) ([]Comment, error) {
	var allComments []Comment
	opts := &gogithub.PullRequestListCommentsOptions{
		ListOptions: gogithub.ListOptions{PerPage: 100},
	}

	for {
		ghComments, resp, err := c.restClient.PullRequests.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("listing review comments: %w", err)
		}

		for _, gc := range ghComments {
			allComments = append(allComments, mapPRReviewComment(gc))
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allComments, nil
}

// mapPRToData 将 go-github PullRequest 映射为 IssueData。
func mapPRToData(pr *gogithub.PullRequest) IssueData {
	state := pr.GetState()
	if pr.GetMerged() {
		state = "merged"
	}

	data := IssueData{
		Title:     pr.GetTitle(),
		State:     state,
		Author:    pr.GetUser().GetLogin(),
		CreatedAt: pr.GetCreatedAt().Time,
		Body:      pr.GetBody(),
	}

	// Labels
	labels := make([]string, 0, len(pr.Labels))
	for _, l := range pr.Labels {
		labels = append(labels, l.GetName())
	}
	data.Labels = labels

	// Assignees
	assignees := make([]string, 0, len(pr.Assignees))
	for _, a := range pr.Assignees {
		assignees = append(assignees, a.GetLogin())
	}
	data.Assignees = assignees

	// Milestone
	if pr.Milestone != nil {
		data.Milestone = pr.Milestone.GetTitle()
	}

	return data
}

// mapPRReviewComment 将 go-github PullRequestComment 映射为 Comment。
func mapPRReviewComment(gc *gogithub.PullRequestComment) Comment {
	c := Comment{
		Author:    gc.GetUser().GetLogin(),
		CreatedAt: gc.GetCreatedAt().Time,
		Body:      gc.GetBody(),
		FilePath:  gc.GetPath(),
		Line:      gc.GetLine(),
	}
	if gc.Reactions != nil {
		c.Reactions = mapReactions(gc.Reactions)
	}
	return c
}
