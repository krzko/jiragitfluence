package github

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/krzko/jiragitfluence/internal/config"
	"github.com/krzko/jiragitfluence/pkg/models"
	"golang.org/x/oauth2"
)

// Client handles interactions with the GitHub API
type Client struct {
	client *github.Client
	logger *slog.Logger
}

// NewClient creates a new GitHub client
func NewClient(cfg config.GitHubConfig, logger *slog.Logger) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.Token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return &Client{
		client: client,
		logger: logger,
	}
}

// FetchIssuesAndPRs fetches issues and pull requests from GitHub
func (c *Client) FetchIssuesAndPRs(repos []string, labels []string, contentFilter string, creator string) ([]models.GitHubIssue, []models.GitHubPR, error) {
	var issues []models.GitHubIssue
	var prs []models.GitHubPR

	for _, repoPath := range repos {
		// Parse owner and repo from the repo path
		parts := strings.Split(repoPath, "/")
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("invalid repository path: %s, expected format: owner/repo", repoPath)
		}
		owner, repo := parts[0], parts[1]

		// Fetch issues
		repoIssues, err := c.fetchIssues(owner, repo, labels, contentFilter, creator)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch issues for %s/%s: %w", owner, repo, err)
		}
		issues = append(issues, repoIssues...)

		// Fetch pull requests
		repoPRs, err := c.fetchPullRequests(owner, repo, labels, contentFilter, creator)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch pull requests for %s/%s: %w", owner, repo, err)
		}
		prs = append(prs, repoPRs...)
	}

	c.logger.Info("Fetched GitHub data", "issues", len(issues), "prs", len(prs))
	return issues, prs, nil
}

// fetchIssues fetches issues from a GitHub repository
func (c *Client) fetchIssues(owner, repo string, labels []string, contentFilter string, creator string) ([]models.GitHubIssue, error) {
	var allIssues []models.GitHubIssue
	ctx := context.Background()

	opts := &github.IssueListByRepoOptions{
		State:     "all",
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	if len(labels) > 0 {
		opts.Labels = labels
	}

	// Add creator filter if specified
	if creator != "" {
		opts.Creator = creator
	}

	for {
		issues, resp, err := c.client.Issues.ListByRepo(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}

		for _, issue := range issues {
			// Skip pull requests (they are also returned by the Issues API)
			if issue.PullRequestLinks != nil {
				continue
			}

			// Apply content filter if specified
			if contentFilter != "" {
				if !containsContent(issue, contentFilter) {
					continue
				}
			}

			ghIssue := convertGitHubIssue(issue, owner, repo)
			allIssues = append(allIssues, ghIssue)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	c.logger.Info("Fetched GitHub issues", "repo", fmt.Sprintf("%s/%s", owner, repo), "count", len(allIssues))
	return allIssues, nil
}

// fetchPullRequests fetches pull requests from a GitHub repository
func (c *Client) fetchPullRequests(owner, repo string, labels []string, contentFilter string, creator string) ([]models.GitHubPR, error) {
	var allPRs []models.GitHubPR
	ctx := context.Background()

	opts := &github.PullRequestListOptions{
		State:     "all",
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	// Note: The GitHub API doesn't support filtering PRs by creator directly in the list options
	// We'll filter them manually after fetching

	for {
		prs, resp, err := c.client.PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return nil, err
		}

		for _, pr := range prs {
			// Filter by labels if specified
			if len(labels) > 0 && !hasMatchingLabels(pr.Labels, labels) {
				continue
			}

			// Apply content filter if specified
			if contentFilter != "" {
				if !containsPRContent(pr, contentFilter) {
					continue
				}
			}

			// Apply creator filter if specified
			if creator != "" && pr.User != nil && pr.User.GetLogin() != creator {
				continue
			}

			ghPR := convertGitHubPR(pr, owner, repo)
			allPRs = append(allPRs, ghPR)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	c.logger.Info("Fetched GitHub pull requests", "repo", fmt.Sprintf("%s/%s", owner, repo), "count", len(allPRs))
	return allPRs, nil
}

// hasMatchingLabels checks if any of the PR labels match the filter labels
func hasMatchingLabels(prLabels []*github.Label, filterLabels []string) bool {
	if len(filterLabels) == 0 {
		return true
	}

	for _, prLabel := range prLabels {
		for _, filterLabel := range filterLabels {
			if prLabel.GetName() == filterLabel {
				return true
			}
		}
	}
	return false
}

// convertGitHubIssue converts a GitHub issue to our model
func convertGitHubIssue(issue *github.Issue, owner, repo string) models.GitHubIssue {
	var labels []string
	for _, label := range issue.Labels {
		labels = append(labels, label.GetName())
	}

	var assignees []string
	for _, assignee := range issue.Assignees {
		assignees = append(assignees, assignee.GetLogin())
	}

	return models.GitHubIssue{
		Title:       issue.GetTitle(),
		Number:      issue.GetNumber(),
		State:       issue.GetState(),
		Labels:      labels,
		Assignees:   assignees,
		CreatedDate: issue.GetCreatedAt().Time,
		UpdatedDate: issue.GetUpdatedAt().Time,
		URL:         issue.GetHTMLURL(),
		Repository:  fmt.Sprintf("%s/%s", owner, repo),
	}
}

// convertGitHubPR converts a GitHub pull request to our model
// containsContent checks if an issue contains the specified text in its title, body, or comments
func containsContent(issue *github.Issue, text string) bool {
	// Check title
	if issue.GetTitle() != "" && strings.Contains(strings.ToLower(issue.GetTitle()), strings.ToLower(text)) {
		return true
	}

	// Check body
	if issue.GetBody() != "" && strings.Contains(strings.ToLower(issue.GetBody()), strings.ToLower(text)) {
		return true
	}

	// Note: Checking comments would require additional API calls
	// This implementation just checks title and body for simplicity
	return false
}

// containsPRContent checks if a pull request contains the specified text in its title, body, or comments
func containsPRContent(pr *github.PullRequest, text string) bool {
	// Check title
	if pr.GetTitle() != "" && strings.Contains(strings.ToLower(pr.GetTitle()), strings.ToLower(text)) {
		return true
	}

	// Check body
	if pr.GetBody() != "" && strings.Contains(strings.ToLower(pr.GetBody()), strings.ToLower(text)) {
		return true
	}

	// Note: Checking comments would require additional API calls
	// This implementation just checks title and body for simplicity
	return false
}

func convertGitHubPR(pr *github.PullRequest, owner, repo string) models.GitHubPR {
	var labels []string
	for _, label := range pr.Labels {
		labels = append(labels, label.GetName())
	}

	var assignees []string
	for _, assignee := range pr.Assignees {
		assignees = append(assignees, assignee.GetLogin())
	}

	// Determine merge status
	mergeStatus := "unknown"
	if pr.Merged != nil && *pr.Merged {
		mergeStatus = "merged"
	} else if pr.MergeableState != nil {
		mergeStatus = *pr.MergeableState
	}

	return models.GitHubPR{
		Title:       pr.GetTitle(),
		Number:      pr.GetNumber(),
		State:       pr.GetState(),
		Labels:      labels,
		Assignees:   assignees,
		CreatedDate: pr.GetCreatedAt().Time,
		UpdatedDate: pr.GetUpdatedAt().Time,
		URL:         pr.GetHTMLURL(),
		Repository:  fmt.Sprintf("%s/%s", owner, repo),
		IsDraft:     pr.GetDraft(),
		MergeStatus: mergeStatus,
	}
}
