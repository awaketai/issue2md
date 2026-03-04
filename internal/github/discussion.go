package github

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/shurcooL/githubv4"
)

// discussionQuery 定义 GraphQL 查询结构体（shurcooL/githubv4 风格）。
type discussionQuery struct {
	Repository struct {
		Discussion struct {
			Title       string
			Body        string
			CreatedAt   time.Time
			StateReason string
			Author      struct {
				Login string
			}
			Labels struct {
				Nodes []struct {
					Name string
				}
			} `graphql:"labels(first: 100)"`
			ReactionGroups []struct {
				Content  string
				Reactors struct {
					TotalCount int
				}
			}
			Comments struct {
				Nodes []discussionCommentNode
			} `graphql:"comments(first: 100)"`
		} `graphql:"discussion(number: $number)"`
	} `graphql:"repository(owner: $owner, name: $name)"`
}

type discussionCommentNode struct {
	Author struct {
		Login string
	}
	Body      string
	CreatedAt time.Time
	ReactionGroups []struct {
		Content  string
		Reactors struct {
			TotalCount int
		}
	}
	Replies struct {
		Nodes []discussionReplyNode
	} `graphql:"replies(first: 100)"`
}

type discussionReplyNode struct {
	Author struct {
		Login string
	}
	Body      string
	CreatedAt time.Time
	ReactionGroups []struct {
		Content  string
		Reactors struct {
			TotalCount int
		}
	}
}

// FetchDiscussion 获取 Discussion 完整数据。嵌套回复按时间线平铺。
func (c *Client) FetchDiscussion(ctx context.Context, owner, repo string, number int) (IssueData, error) {
	c.logf("fetching discussion %s/%s#%d", owner, repo, number)

	var q discussionQuery
	variables := map[string]any{
		"owner":  githubv4.String(owner),
		"name":   githubv4.String(repo),
		"number": githubv4.Int(number),
	}

	if err := c.graphqlClient.Query(ctx, &q, variables); err != nil {
		return IssueData{}, fmt.Errorf("fetching discussion %s/%s#%d: %w", owner, repo, number, err)
	}

	d := q.Repository.Discussion

	// Map state
	state := strings.ToLower(d.StateReason)
	if state == "" {
		state = "open"
	}

	// Map labels
	labels := make([]string, 0, len(d.Labels.Nodes))
	for _, l := range d.Labels.Nodes {
		labels = append(labels, l.Name)
	}

	// Map reactions
	reactions := mapGraphQLReactions(d.ReactionGroups)

	// Flatten comments + replies into a single timeline
	var comments []Comment
	for _, node := range d.Comments.Nodes {
		comments = append(comments, Comment{
			Author:    node.Author.Login,
			CreatedAt: node.CreatedAt,
			Body:      node.Body,
			Reactions: mapGraphQLReactions(node.ReactionGroups),
		})
		for _, reply := range node.Replies.Nodes {
			comments = append(comments, Comment{
				Author:    reply.Author.Login,
				CreatedAt: reply.CreatedAt,
				Body:      reply.Body,
				Reactions: mapGraphQLReactions(reply.ReactionGroups),
			})
		}
	}

	// Sort by time
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	if comments == nil {
		comments = []Comment{}
	}

	data := IssueData{
		Title:     d.Title,
		State:     state,
		Author:    d.Author.Login,
		CreatedAt: d.CreatedAt,
		Labels:    labels,
		Assignees: []string{}, // Discussion 无 Assignees
		Body:      d.Body,
		Reactions: reactions,
		Comments:  comments,
	}

	c.logf("fetched discussion %s/%s#%d: %d comments", owner, repo, number, len(comments))
	return data, nil
}

// mapGraphQLReactions 将 GraphQL reactionGroups 映射为 Reactions。
func mapGraphQLReactions(groups []struct {
	Content  string
	Reactors struct {
		TotalCount int
	}
}) Reactions {
	var r Reactions
	for _, g := range groups {
		count := g.Reactors.TotalCount
		switch g.Content {
		case "THUMBS_UP":
			r.PlusOne = count
		case "THUMBS_DOWN":
			r.MinusOne = count
		case "LAUGH":
			r.Laugh = count
		case "HOORAY":
			r.Hooray = count
		case "CONFUSED":
			r.Confused = count
		case "HEART":
			r.Heart = count
		case "ROCKET":
			r.Rocket = count
		case "EYES":
			r.Eyes = count
		}
	}
	return r
}
