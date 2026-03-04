package github

import (
	"context"
	"fmt"

	gogithub "github.com/google/go-github/v68/github"
)

// FetchIssue 获取 Issue 完整数据。自动处理分页。
func (c *Client) FetchIssue(ctx context.Context, owner, repo string, number int) (IssueData, error) {
	c.logf("fetching issue %s/%s#%d", owner, repo, number)

	issue, _, err := c.restClient.Issues.Get(ctx, owner, repo, number)
	if err != nil {
		return IssueData{}, fmt.Errorf("fetching issue %s/%s#%d: %w", owner, repo, number, err)
	}

	data := mapIssueToData(issue)

	// Fetch all comments with pagination
	comments, err := c.fetchIssueComments(ctx, owner, repo, number)
	if err != nil {
		return IssueData{}, fmt.Errorf("fetching comments for %s/%s#%d: %w", owner, repo, number, err)
	}
	data.Comments = comments

	c.logf("fetched issue %s/%s#%d: %d comments", owner, repo, number, len(comments))
	return data, nil
}

// fetchIssueComments 获取 Issue 的所有评论（自动分页）。
func (c *Client) fetchIssueComments(ctx context.Context, owner, repo string, number int) ([]Comment, error) {
	var allComments []Comment
	opts := &gogithub.IssueListCommentsOptions{
		ListOptions: gogithub.ListOptions{PerPage: 100},
	}

	for {
		ghComments, resp, err := c.restClient.Issues.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("listing comments: %w", err)
		}

		for _, gc := range ghComments {
			allComments = append(allComments, mapIssueComment(gc))
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if allComments == nil {
		allComments = []Comment{}
	}
	return allComments, nil
}

// mapIssueToData 将 go-github Issue 映射为 IssueData。
func mapIssueToData(issue *gogithub.Issue) IssueData {
	data := IssueData{
		Title:     issue.GetTitle(),
		State:     issue.GetState(),
		Author:    issue.GetUser().GetLogin(),
		CreatedAt: issue.GetCreatedAt().Time,
		Body:      issue.GetBody(),
	}

	// Labels
	labels := make([]string, 0, len(issue.Labels))
	for _, l := range issue.Labels {
		labels = append(labels, l.GetName())
	}
	data.Labels = labels

	// Assignees
	assignees := make([]string, 0, len(issue.Assignees))
	for _, a := range issue.Assignees {
		assignees = append(assignees, a.GetLogin())
	}
	data.Assignees = assignees

	// Milestone
	if issue.Milestone != nil {
		data.Milestone = issue.Milestone.GetTitle()
	}

	// Reactions
	if issue.Reactions != nil {
		data.Reactions = mapReactions(issue.Reactions)
	}

	return data
}

// mapReactions 将 go-github Reactions 映射为 Reactions。
func mapReactions(r *gogithub.Reactions) Reactions {
	return Reactions{
		PlusOne:  r.GetPlusOne(),
		MinusOne: r.GetMinusOne(),
		Laugh:    r.GetLaugh(),
		Hooray:   r.GetHooray(),
		Confused: r.GetConfused(),
		Heart:    r.GetHeart(),
		Rocket:   r.GetRocket(),
		Eyes:     r.GetEyes(),
	}
}

// mapIssueComment 将 go-github IssueComment 映射为 Comment。
func mapIssueComment(gc *gogithub.IssueComment) Comment {
	c := Comment{
		Author:    gc.GetUser().GetLogin(),
		CreatedAt: gc.GetCreatedAt().Time,
		Body:      gc.GetBody(),
	}
	if gc.Reactions != nil {
		c.Reactions = mapReactions(gc.Reactions)
	}
	return c
}
