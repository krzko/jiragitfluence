package commands

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/krzko/jiragitfluence/internal/config"
	"github.com/krzko/jiragitfluence/internal/github"
	"github.com/krzko/jiragitfluence/internal/jira"
	"github.com/krzko/jiragitfluence/pkg/models"
	"github.com/urfave/cli/v2"
)

// FetchCommand handles the fetch command which combines both Jira and GitHub fetching
func FetchCommand(ctx *cli.Context) error {
	logger := slog.Default()
	
	// Set log level if verbose flag is set
	if ctx.Bool("verbose") {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		slog.SetDefault(logger)
	}

	// Load configuration
	cfg, err := config.LoadConfig(ctx.String("config"))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Get command line arguments
	jiraProjects := ctx.StringSlice("jira-projects")
	jiraJQL := ctx.String("jira-jql")
	githubRepos := ctx.StringSlice("github-repos")
	githubLabels := ctx.StringSlice("github-labels")
	githubContentFilter := ctx.String("github-content-filter")
	githubCreator := ctx.String("github-creator")
	outputPath := ctx.String("output")

	logger.Info("Starting combined fetch operation",
		"jira-projects", jiraProjects,
		"jira-jql", jiraJQL,
		"github-repos", githubRepos,
		"github-labels", githubLabels,
		"github-content-filter", githubContentFilter,
		"github-creator", githubCreator,
		"output", outputPath)

	// Initialize aggregated data
	data := &models.AggregatedData{
		Metadata: models.Metadata{
			FetchTime:          time.Now(),
			JiraProjects:       jiraProjects,
			GitHubRepos:        githubRepos,
			JiraJQL:            jiraJQL,
			GitHubLabels:       githubLabels,
			GitHubContentFilter: githubContentFilter,
			GitHubCreator:      githubCreator,
		},
	}

	// Fetch Jira issues if projects are specified
	if len(jiraProjects) > 0 {
		jiraClient, err := jira.NewClient(cfg.Jira, logger)
		if err != nil {
			return fmt.Errorf("failed to create Jira client: %w", err)
		}

		jiraIssues, err := jiraClient.FetchIssues(jiraProjects, jiraJQL)
		if err != nil {
			return fmt.Errorf("failed to fetch Jira issues: %w", err)
		}
		data.JiraIssues = jiraIssues
		logger.Info("Fetched Jira issues", 
			"count", len(jiraIssues), 
			"projects", jiraProjects, 
			"jql", jiraJQL)
	}

	// Fetch GitHub issues and PRs if repos are specified
	if len(githubRepos) > 0 {
		githubClient := github.NewClient(cfg.GitHub, logger)
		githubIssues, githubPRs, err := githubClient.FetchIssuesAndPRs(githubRepos, githubLabels, githubContentFilter, githubCreator)
		if err != nil {
			return fmt.Errorf("failed to fetch GitHub data: %w", err)
		}
		data.GitHubIssues = githubIssues
		data.GitHubPRs = githubPRs
		logger.Info("Fetched GitHub data", 
			"issues", len(githubIssues), 
			"prs", len(githubPRs), 
			"repos", githubRepos, 
			"labels", githubLabels, 
			"content_filter", githubContentFilter, 
			"creator", githubCreator)
	}

	// Save aggregated data to file
	if err := SaveAggregatedData(data, outputPath); err != nil {
		return fmt.Errorf("failed to save aggregated data: %w", err)
	}
	logger.Info("Completed combined fetch operation", 
		"jira_issues", len(data.JiraIssues), 
		"github_issues", len(data.GitHubIssues), 
		"github_prs", len(data.GitHubPRs), 
		"total_items", len(data.JiraIssues) + len(data.GitHubIssues) + len(data.GitHubPRs), 
		"path", outputPath)

	return nil
}


