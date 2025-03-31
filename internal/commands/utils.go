package commands

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/krzko/jiragitfluence/pkg/models"
)

// SaveAggregatedData saves the aggregated data to a JSON file
func SaveAggregatedData(data *models.AggregatedData, outputPath string) error {
	// Log counts of issues and PRs
	jiraCount := len(data.JiraIssues)
	githubIssueCount := len(data.GitHubIssues)
	githubPRCount := len(data.GitHubPRs)
	totalCount := jiraCount + githubIssueCount + githubPRCount

	// Log the counts
	slog.Info(fmt.Sprintf("Saving aggregated data to %s", outputPath),
		"total_items", totalCount,
		"jira_issues", jiraCount,
		"github_issues", githubIssueCount,
		"github_prs", githubPRCount)

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, jsonData, 0644)
}
