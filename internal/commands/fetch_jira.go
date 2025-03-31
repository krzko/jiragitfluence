package commands

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/krzko/jiragitfluence/internal/config"
	"github.com/krzko/jiragitfluence/internal/jira"
	"github.com/krzko/jiragitfluence/pkg/models"
	"github.com/urfave/cli/v2"
)

// FetchJiraCommand handles the fetch-jira command
func FetchJiraCommand(ctx *cli.Context) error {
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
	outputPath := ctx.String("output")

	logger.Info("Starting Jira fetch operation",
		"jira-projects", jiraProjects,
		"jira-jql", jiraJQL,
		"output", outputPath)

	// Initialize aggregated data
	data := &models.AggregatedData{
		Metadata: models.Metadata{
			FetchTime:    time.Now(),
			JiraProjects: jiraProjects,
			JiraJQL:      jiraJQL,
		},
	}

	// Fetch Jira issues
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

	// Save aggregated data to file
	if err := SaveAggregatedData(data, outputPath); err != nil {
		return fmt.Errorf("failed to save aggregated data: %w", err)
	}
	logger.Info("Completed Jira fetch operation", 
		"jira_issues", len(data.JiraIssues), 
		"total_items", len(data.JiraIssues), 
		"path", outputPath)

	return nil
}


