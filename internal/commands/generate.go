package commands

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/krzko/jiragitfluence/internal/generator"
	"github.com/krzko/jiragitfluence/pkg/models"
	"github.com/urfave/cli/v2"
)

// GenerateCommand handles the generate command
func GenerateCommand(ctx *cli.Context) error {
	logger := slog.Default()
	
	// Set log level if verbose flag is set
	if ctx.Bool("verbose") {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		slog.SetDefault(logger)
	}

	// Get command line arguments
	inputPath := ctx.String("input")
	jiraInputPath := ctx.String("jira-input")
	githubInputPath := ctx.String("github-input")
	format := ctx.String("format")
	outputPath := ctx.String("output")
	groupBy := ctx.String("group-by")
	includeMetadata := ctx.Bool("include-metadata")
	versionLabel := ctx.String("version-label")
	
	// Roadmap specific options
	roadmapTimeframe := ctx.String("roadmap-timeframe")
	roadmapGrouping := ctx.String("roadmap-grouping")
	roadmapView := ctx.String("roadmap-view")
	includeDependencies := ctx.Bool("roadmap-include-dependencies")

	logger.Info("Starting generate operation",
		"input", inputPath,
		"jira-input", jiraInputPath,
		"github-input", githubInputPath,
		"format", format,
		"output", outputPath,
		"roadmap-timeframe", roadmapTimeframe,
		"roadmap-grouping", roadmapGrouping,
		"roadmap-view", roadmapView)

	// Initialize aggregated data with empty arrays to prevent nil pointer panics
	data := &models.AggregatedData{
		JiraIssues:   []models.JiraIssue{},
		GitHubIssues: []models.GitHubIssue{},
		GitHubPRs:    []models.GitHubPR{},
	}

	// Track if we've loaded any data
	dataLoaded := false

	// If main input is provided, use it
	if inputPath != "" {
		loadedData, err := loadAggregatedData(inputPath)
		if err != nil {
			return fmt.Errorf("failed to load aggregated data from %s: %w", inputPath, err)
		}
		// Replace our empty data with the loaded data
		data = loadedData
		dataLoaded = true
		logger.Info("Loaded combined data", "path", inputPath)
	} 

	// Load Jira data if provided
	if jiraInputPath != "" {
		jiraData, err := loadAggregatedData(jiraInputPath)
		if err != nil {
			return fmt.Errorf("failed to load Jira data from %s: %w", jiraInputPath, err)
		}

		// If we haven't loaded any data yet, use the Jira data's metadata
		if !dataLoaded {
			// Copy metadata
			data.Metadata.JiraProjects = jiraData.Metadata.JiraProjects
			data.Metadata.JiraJQL = jiraData.Metadata.JiraJQL
			data.Metadata.FetchTime = jiraData.Metadata.FetchTime
		} else if inputPath == "" {
			// We're combining with GitHub data, merge metadata
			data.Metadata.JiraProjects = jiraData.Metadata.JiraProjects
			data.Metadata.JiraJQL = jiraData.Metadata.JiraJQL
			// Only update fetch time if it's newer or not set
			if data.Metadata.FetchTime.IsZero() || jiraData.Metadata.FetchTime.After(data.Metadata.FetchTime) {
				data.Metadata.FetchTime = jiraData.Metadata.FetchTime
			}
		}

		// Always copy the Jira issues
		data.JiraIssues = jiraData.JiraIssues
		dataLoaded = true
		logger.Info("Loaded Jira data", "path", jiraInputPath, "issues", len(jiraData.JiraIssues))
	}

	// Load GitHub data if provided
	if githubInputPath != "" {
		githubData, err := loadAggregatedData(githubInputPath)
		if err != nil {
			return fmt.Errorf("failed to load GitHub data from %s: %w", githubInputPath, err)
		}

		// If we haven't loaded any data yet, use the GitHub data's metadata
		if !dataLoaded {
			// Copy metadata
			data.Metadata.GitHubRepos = githubData.Metadata.GitHubRepos
			data.Metadata.GitHubLabels = githubData.Metadata.GitHubLabels
			data.Metadata.GitHubContentFilter = githubData.Metadata.GitHubContentFilter
			data.Metadata.GitHubCreator = githubData.Metadata.GitHubCreator
			data.Metadata.FetchTime = githubData.Metadata.FetchTime
		} else if inputPath == "" {
			// We're combining with Jira data, merge metadata
			data.Metadata.GitHubRepos = githubData.Metadata.GitHubRepos
			data.Metadata.GitHubLabels = githubData.Metadata.GitHubLabels
			data.Metadata.GitHubContentFilter = githubData.Metadata.GitHubContentFilter
			data.Metadata.GitHubCreator = githubData.Metadata.GitHubCreator
			// Only update fetch time if it's newer or not set
			if data.Metadata.FetchTime.IsZero() || githubData.Metadata.FetchTime.After(data.Metadata.FetchTime) {
				data.Metadata.FetchTime = githubData.Metadata.FetchTime
			}
		}

		// Always copy the GitHub issues and PRs
		data.GitHubIssues = githubData.GitHubIssues
		data.GitHubPRs = githubData.GitHubPRs
		dataLoaded = true
		logger.Info("Loaded GitHub data", "path", githubInputPath, 
			"issues", len(githubData.GitHubIssues), 
			"prs", len(githubData.GitHubPRs))
	}

	// Check if we loaded any data
	if !dataLoaded {
		return fmt.Errorf("no input files specified, please provide either --input, --jira-input, or --github-input")
	}

	// Log counts of issues and PRs being processed
	jiraCount := len(data.JiraIssues)
	githubIssueCount := len(data.GitHubIssues)
	githubPRCount := len(data.GitHubPRs)
	totalCount := jiraCount + githubIssueCount + githubPRCount

	logger.Info("Processing data for generation",
		"total_items", totalCount,
		"jira_issues", jiraCount,
		"github_issues", githubIssueCount,
		"github_prs", githubPRCount)

	// Create generator
	gen := generator.NewGenerator(logger)

	// Set generator options
	opts := generator.Options{
		Format:              generator.Format(format),
		GroupBy:             generator.GroupBy(groupBy),
		IncludeMetadata:     includeMetadata,
		VersionLabel:        versionLabel,
		
		// Roadmap specific options
		RoadmapTimeframe:    roadmapTimeframe,
		RoadmapGrouping:     roadmapGrouping,
		RoadmapView:         generator.RoadmapView(roadmapView),
		IncludeDependencies: includeDependencies,
	}

	// Generate content
	content, err := gen.Generate(data, opts)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	// Save generated content to file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	logger.Info("Saved generated content", "path", outputPath)

	return nil
}

// loadAggregatedData loads the aggregated data from a JSON file
func loadAggregatedData(inputPath string) (*models.AggregatedData, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var aggregatedData models.AggregatedData
	if err := json.Unmarshal(data, &aggregatedData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return &aggregatedData, nil
}
