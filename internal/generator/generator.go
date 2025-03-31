package generator

import (
	"fmt"
	"html"
	"log/slog"
	"strings"
	"time"

	"github.com/krzko/jiragitfluence/pkg/models"
)

// Generator handles the transformation of data into Confluence-friendly formats
type Generator struct {
	logger *slog.Logger
}

// Format represents the output format
type Format string

const (
	// TableFormat represents a table layout
	TableFormat Format = "table"
	// KanbanFormat represents a kanban board layout
	KanbanFormat Format = "kanban"
	// CustomFormat represents a custom layout
	CustomFormat Format = "custom"
	// GanttFormat represents a Gantt chart timeline layout
	GanttFormat Format = "gantt"
	// RoadmapFormat represents a roadmap layout for planning future work
	RoadmapFormat Format = "roadmap"
)

// GroupBy represents how to group the data
type GroupBy string

const (
	// StatusGroup groups by status
	StatusGroup GroupBy = "status"
	// AssigneeGroup groups by assignee
	AssigneeGroup GroupBy = "assignee"
	// LabelGroup groups by label
	LabelGroup GroupBy = "label"
)

// RoadmapView represents the type of roadmap view
type RoadmapView string

const (
	// TimelineView shows a timeline-based roadmap
	TimelineView RoadmapView = "timeline"
	// StrategicView shows a high-level strategic roadmap
	StrategicView RoadmapView = "strategic"
	// ReleaseView shows a release-based roadmap
	ReleaseView RoadmapView = "release"
	// EpicGanttView shows a Gantt-style roadmap organized by epics
	EpicGanttView RoadmapView = "epicgantt"
)

// Options represents the generator options
type Options struct {
	Format              Format
	GroupBy             GroupBy
	IncludeMetadata     bool
	VersionLabel        string
	
	// Roadmap specific options
	RoadmapTimeframe    string      // e.g., "Q1-Q4 2025", "6months", "1year"
	RoadmapGrouping     string      // How to group items in roadmap (e.g., "epic", "theme", "team")
	RoadmapView         RoadmapView // Type of roadmap view
	IncludeDependencies bool        // Whether to show dependencies between roadmap items
}

// NewGenerator creates a new generator
func NewGenerator(logger *slog.Logger) *Generator {
	return &Generator{
		logger: logger,
	}
}

// Generate transforms the aggregated data into Confluence markup
func (g *Generator) Generate(data *models.AggregatedData, opts Options) (string, error) {
	g.logger.Info("Generating content", "format", opts.Format, "groupBy", opts.GroupBy)

	var content strings.Builder

	// Add header with version info
	g.addHeader(&content, data, opts)

	// Generate content based on format
	switch opts.Format {
	case TableFormat:
		g.generateTableFormat(&content, data, opts)
	case KanbanFormat:
		g.generateKanbanFormat(&content, data, opts)
	case CustomFormat:
		g.generateCustomFormat(&content, data, opts)
	case GanttFormat:
		g.generateGanttFormat(&content, data, opts)
	case RoadmapFormat:
		g.generateRoadmapFormat(&content, data, opts)
	default:
		return "", fmt.Errorf("unsupported format: %s", opts.Format)
	}

	// Add footer with metadata if requested
	if opts.IncludeMetadata {
		g.addFooter(&content, data)
	}

	return content.String(), nil
}

// addHeader adds a header to the content with Confluence-specific styling
func (g *Generator) addHeader(content *strings.Builder, data *models.AggregatedData, opts Options) {
	content.WriteString("<h1>Project Status Dashboard</h1>\n\n")

	if opts.VersionLabel != "" {
		content.WriteString(fmt.Sprintf("<p><strong>Version:</strong> %s</p>\n", escapeHTML(opts.VersionLabel)))
	}

	content.WriteString(fmt.Sprintf("<p><strong>Generated:</strong> %s</p>\n\n", time.Now().Format(time.RFC1123)))

	// Add summary in an info panel
	content.WriteString("<ac:structured-macro ac:name=\"info\">\n")
	content.WriteString("<ac:rich-text-body>\n")
	content.WriteString("<p><strong>Summary</strong></p>\n")
	content.WriteString("<ul>\n")
	content.WriteString(fmt.Sprintf("<li>Jira Issues: %d</li>\n", len(data.JiraIssues)))
	content.WriteString(fmt.Sprintf("<li>GitHub Issues: %d</li>\n", len(data.GitHubIssues)))
	content.WriteString(fmt.Sprintf("<li>GitHub Pull Requests: %d</li>\n", len(data.GitHubPRs)))
	content.WriteString("</ul>\n")
	content.WriteString("</ac:rich-text-body>\n")
	content.WriteString("</ac:structured-macro>\n\n")
}

// addFooter adds a footer with metadata using Confluence styling
func (g *Generator) addFooter(content *strings.Builder, data *models.AggregatedData) {
	content.WriteString("<h2>Metadata</h2>\n")
	content.WriteString("<ac:structured-macro ac:name=\"expand\">\n")
	content.WriteString("<ac:parameter ac:name=\"title\">Click to view metadata</ac:parameter>\n")
	content.WriteString("<ac:rich-text-body>\n")
	content.WriteString("<table class=\"confluenceTable\">\n")
	content.WriteString("<tr>")
	content.WriteString("<th class=\"confluenceTh\" style=\"background-color:#f4f5f7;text-align:center\">Property</th>")
	content.WriteString("<th class=\"confluenceTh\" style=\"background-color:#f4f5f7;text-align:center\">Value</th>")
	content.WriteString("</tr>\n")
	content.WriteString(fmt.Sprintf("<tr><td class=\"confluenceTd\">Fetch Time</td><td class=\"confluenceTd\">%s</td></tr>\n", data.Metadata.FetchTime.Format(time.RFC1123)))
	content.WriteString(fmt.Sprintf("<tr><td class=\"confluenceTd\">Jira Projects</td><td class=\"confluenceTd\">%s</td></tr>\n", escapeHTML(strings.Join(data.Metadata.JiraProjects, ", "))))
	content.WriteString(fmt.Sprintf("<tr><td class=\"confluenceTd\">GitHub Repositories</td><td class=\"confluenceTd\">%s</td></tr>\n", escapeHTML(strings.Join(data.Metadata.GitHubRepos, ", "))))

	if data.Metadata.JiraJQL != "" {
		content.WriteString(fmt.Sprintf("<tr><td class=\"confluenceTd\">Jira JQL</td><td class=\"confluenceTd\">%s</td></tr>\n", escapeHTML(data.Metadata.JiraJQL)))
	}

	if len(data.Metadata.GitHubLabels) > 0 {
		content.WriteString(fmt.Sprintf("<tr><td class=\"confluenceTd\">GitHub Labels</td><td class=\"confluenceTd\">%s</td></tr>\n", escapeHTML(strings.Join(data.Metadata.GitHubLabels, ", "))))
	}

	content.WriteString("</table>\n")
	content.WriteString("</ac:rich-text-body>\n")
	content.WriteString("</ac:structured-macro>\n")

	// Add a footer note
	content.WriteString("<hr style=\"border-top: 1px solid #ddd; margin: 20px 0;\" />\n")
	content.WriteString("<p><em>Generated by jiragitfluence - Automated Project Status Reporter</em></p>\n")
}

// generateTableFormat generates a table format with proper Confluence Storage Format
func (g *Generator) generateTableFormat(content *strings.Builder, data *models.AggregatedData, _ Options) {
	// Jira Issues
	if len(data.JiraIssues) > 0 {
		content.WriteString("<h2>Jira Issues</h2>\n")
		content.WriteString("<table>\n")
		content.WriteString("<tbody>\n")
		
		// Header row
		content.WriteString("<tr>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Key</th>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Summary</th>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Status</th>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Assignee</th>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Priority</th>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Updated</th>\n")
		content.WriteString("</tr>\n")

		// Data rows
		for i, issue := range data.JiraIssues {
			// Alternate row colors for better readability
			rowStyle := ""
			if i%2 == 0 {
				rowStyle = " background-color: #f8f9fa;"
			}
			
			content.WriteString("<tr>\n")
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("<a href=\"%s\">%s</a>\n", issue.URL, escapeHTML(issue.Key)))
			content.WriteString("</td>\n")
			
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("%s\n", escapeHTML(issue.Summary)))
			content.WriteString("</td>\n")
			
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("%s\n", getStatusStyle(issue.Status)))
			content.WriteString("</td>\n")
			
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("%s\n", escapeHTML(issue.Assignee)))
			content.WriteString("</td>\n")
			
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("%s\n", escapeHTML(issue.Priority)))
			content.WriteString("</td>\n")
			
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("%s\n", issue.UpdatedDate.Format("2006-01-02")))
			content.WriteString("</td>\n")
			content.WriteString("</tr>\n")
		}

		content.WriteString("</tbody>\n")
		content.WriteString("</table>\n\n")
	}

	// GitHub Issues
	if len(data.GitHubIssues) > 0 {
		content.WriteString("<h2>GitHub Issues</h2>\n")
		content.WriteString("<table>\n")
		content.WriteString("<tbody>\n")
		
		// Header row
		content.WriteString("<tr>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Repository</th>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Number</th>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Title</th>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">State</th>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Assignees</th>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Updated</th>\n")
		content.WriteString("</tr>\n")

		// Data rows
		for i, issue := range data.GitHubIssues {
			// Alternate row colors for better readability
			rowStyle := ""
			if i%2 == 0 {
				rowStyle = " background-color: #f8f9fa;"
			}
			
			content.WriteString("<tr>\n")
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("%s\n", escapeHTML(issue.Repository)))
			content.WriteString("</td>\n")
			
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("<a href=\"%s\">#%d</a>\n", issue.URL, issue.Number))
			content.WriteString("</td>\n")
			
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("%s\n", escapeHTML(issue.Title)))
			content.WriteString("</td>\n")
			
			// Style the state similar to Jira status
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			statusStyle := getStatusStyle(issue.State)
			content.WriteString(fmt.Sprintf("%s\n", statusStyle))
			content.WriteString("</td>\n")
			
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("%s\n", escapeHTML(strings.Join(issue.Assignees, ", "))))
			content.WriteString("</td>\n")
			
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("%s\n", issue.UpdatedDate.Format("2006-01-02")))
			content.WriteString("</td>\n")
			content.WriteString("</tr>\n")
		}

		content.WriteString("</tbody>\n")
		content.WriteString("</table>\n\n")
	}

	// GitHub Pull Requests
	if len(data.GitHubPRs) > 0 {
		content.WriteString("<h2>GitHub Pull Requests</h2>\n")
		content.WriteString("<table>\n")
		content.WriteString("<tbody>\n")
		
		// Header row
		content.WriteString("<tr>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Repository</th>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Number</th>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Title</th>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">State</th>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Assignees</th>\n")
		content.WriteString("<th style=\"background-color: #f4f5f7; text-align: center; border: 1px solid #c1c7d0;\">Updated</th>\n")
		content.WriteString("</tr>\n")

		// Data rows
		for i, pr := range data.GitHubPRs {
			// Alternate row colors for better readability
			rowStyle := ""
			if i%2 == 0 {
				rowStyle = " background-color: #f8f9fa;"
			}
			
			content.WriteString("<tr>\n")
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("%s\n", escapeHTML(pr.Repository)))
			content.WriteString("</td>\n")
			
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("<a href=\"%s\">#%d</a>\n", pr.URL, pr.Number))
			content.WriteString("</td>\n")
			
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("%s\n", escapeHTML(pr.Title)))
			content.WriteString("</td>\n")
			
			// Add draft status if applicable and style the state
			state := pr.State
			if pr.IsDraft {
				state += " (Draft)"
			}
			
			// Style the PR state
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			statusStyle := getStatusStyle(state)
			content.WriteString(fmt.Sprintf("%s\n", statusStyle))
			content.WriteString("</td>\n")
			
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("%s\n", escapeHTML(strings.Join(pr.Assignees, ", "))))
			content.WriteString("</td>\n")
			
			content.WriteString(fmt.Sprintf("<td style=\"border: 1px solid #c1c7d0;%s\">\n", rowStyle))
			content.WriteString(fmt.Sprintf("%s\n", pr.UpdatedDate.Format("2006-01-02")))
			content.WriteString("</td>\n")
			content.WriteString("</tr>\n")
		}

		content.WriteString("</tbody>\n")
		content.WriteString("</table>\n\n")
	}
}

// generateKanbanFormat generates a kanban board format using Confluence Storage Format
func (g *Generator) generateKanbanFormat(content *strings.Builder, data *models.AggregatedData, _ Options) {
	// Define columns
	columns := []string{"To Do", "In Progress", "Review", "Done"}

	content.WriteString("<h2>Kanban Board</h2>\n")
	
	// Use Confluence's layout feature for the kanban board
	content.WriteString("<ac:layout>\n")
	content.WriteString("<ac:layout-section ac:type=\"four_equal\">\n")
	
	// Create columns for the kanban board using layout cells

	// Create layout cells for each column
	for _, column := range columns {
		// Create a layout cell for this column
		content.WriteString("<ac:layout-cell>\n")
		
		// Add column header
		content.WriteString(fmt.Sprintf("<h3 style=\"text-align:center;background-color:#f4f5f7;padding:8px;margin-bottom:10px;border-radius:3px;\">%s</h3>\n", column))

		// Add Jira issues to columns
		for _, issue := range data.JiraIssues {
			if mapStatusToColumn(issue.Status) == column {
				// Generate a unique color based on the issue key
				hash := 0
				for _, c := range issue.Key {
					hash = 31*hash + int(c)
				}
				
				// Generate a color from the hash - use a predefined palette for better visibility
				colors := []string{"#0052CC", "#6554C0", "#00875A", "#FF5630", "#FF8B00", "#36B37E", "#00B8D9", "#4C9AFF", "#172B4D", "#403294"}
				lightColors := []string{"#DEEBFF", "#EAE6FF", "#E3FCEF", "#FFEBE6", "#FFF0B3", "#ABF5D1", "#E6FCFF", "#B3D4FF", "#F4F5F7", "#EAE6FF"}
				colorIndex := hash % len(colors)
				
				// Create a panel for each Jira issue with unique color
				content.WriteString("<ac:structured-macro ac:name=\"panel\">\n")
				content.WriteString("<ac:parameter ac:name=\"borderStyle\">solid</ac:parameter>\n")
				content.WriteString(fmt.Sprintf("<ac:parameter ac:name=\"borderColor\">%s</ac:parameter>\n", colors[colorIndex]))
				content.WriteString("<ac:parameter ac:name=\"borderWidth\">1</ac:parameter>\n")
				content.WriteString(fmt.Sprintf("<ac:parameter ac:name=\"backgroundColor\">%s</ac:parameter>\n", lightColors[colorIndex]))
				content.WriteString("<ac:rich-text-body>\n")
				content.WriteString(fmt.Sprintf("<p><strong><a href=\"%s\">%s</a></strong></p>\n", issue.URL, escapeHTML(issue.Key)))
				content.WriteString(fmt.Sprintf("<p>%s</p>\n", escapeHTML(issue.Summary)))
				if issue.Assignee != "" {
					content.WriteString(fmt.Sprintf("<p><em>Assignee: %s</em></p>\n", escapeHTML(issue.Assignee)))
				}
				content.WriteString("</ac:rich-text-body>\n")
				content.WriteString("</ac:structured-macro>\n")
			}
		}

		// Add GitHub issues to columns
		for _, issue := range data.GitHubIssues {
			if mapGitHubStateToColumn(issue.State) == column {
				// Generate a unique color based on the repository and issue number
				hash := 0
				for _, c := range issue.Repository {
					hash = 31*hash + int(c)
				}
				hash = hash + issue.Number
				
				// Generate a color from the hash - use a predefined palette for better visibility
				colors := []string{"#0052CC", "#6554C0", "#00875A", "#FF5630", "#FF8B00", "#36B37E", "#00B8D9", "#4C9AFF", "#172B4D", "#403294"}
				lightColors := []string{"#DEEBFF", "#EAE6FF", "#E3FCEF", "#FFEBE6", "#FFF0B3", "#ABF5D1", "#E6FCFF", "#B3D4FF", "#F4F5F7", "#EAE6FF"}
				colorIndex := hash % len(colors)
				
				// Create a panel for each GitHub issue with unique color
				content.WriteString("<ac:structured-macro ac:name=\"panel\">\n")
				content.WriteString("<ac:parameter ac:name=\"borderStyle\">solid</ac:parameter>\n")
				content.WriteString(fmt.Sprintf("<ac:parameter ac:name=\"borderColor\">%s</ac:parameter>\n", colors[colorIndex]))
				content.WriteString("<ac:parameter ac:name=\"borderWidth\">1</ac:parameter>\n")
				content.WriteString(fmt.Sprintf("<ac:parameter ac:name=\"backgroundColor\">%s</ac:parameter>\n", lightColors[colorIndex]))
				content.WriteString("<ac:rich-text-body>\n")
				content.WriteString(fmt.Sprintf("<p><strong><a href=\"%s\">%s #%d</a></strong></p>\n", issue.URL, escapeHTML(issue.Repository), issue.Number))
				content.WriteString(fmt.Sprintf("<p>%s</p>\n", escapeHTML(issue.Title)))
				if len(issue.Assignees) > 0 {
					content.WriteString(fmt.Sprintf("<p><em>Assignee: %s</em></p>\n", escapeHTML(strings.Join(issue.Assignees, ", "))))
				}
				content.WriteString("</ac:rich-text-body>\n")
				content.WriteString("</ac:structured-macro>\n")
			}
		}

		// Add GitHub PRs to columns
		for _, pr := range data.GitHubPRs {
			if mapGitHubPRStateToColumn(pr.State, pr.MergeStatus) == column {
				// Generate a unique color based on the repository and PR number
				hash := 0
				for _, c := range pr.Repository {
					hash = 31*hash + int(c)
				}
				hash = hash + pr.Number
				
				// Generate a color from the hash - use a predefined palette for better visibility
				colors := []string{"#0052CC", "#6554C0", "#00875A", "#FF5630", "#FF8B00", "#36B37E", "#00B8D9", "#4C9AFF", "#172B4D", "#403294"}
				lightColors := []string{"#DEEBFF", "#EAE6FF", "#E3FCEF", "#FFEBE6", "#FFF0B3", "#ABF5D1", "#E6FCFF", "#B3D4FF", "#F4F5F7", "#EAE6FF"}
				colorIndex := hash % len(colors)
				
				// Create a panel for each GitHub PR with unique color
				content.WriteString("<ac:structured-macro ac:name=\"panel\">\n")
				content.WriteString("<ac:parameter ac:name=\"borderStyle\">solid</ac:parameter>\n")
				content.WriteString(fmt.Sprintf("<ac:parameter ac:name=\"borderColor\">%s</ac:parameter>\n", colors[colorIndex]))
				content.WriteString("<ac:parameter ac:name=\"borderWidth\">1</ac:parameter>\n")
				content.WriteString(fmt.Sprintf("<ac:parameter ac:name=\"backgroundColor\">%s</ac:parameter>\n", lightColors[colorIndex]))
				content.WriteString("<ac:rich-text-body>\n")
				content.WriteString(fmt.Sprintf("<p><strong><a href=\"%s\">%s #%d</a></strong></p>\n", pr.URL, escapeHTML(pr.Repository), pr.Number))
				content.WriteString(fmt.Sprintf("<p>%s</p>\n", escapeHTML(pr.Title)))
				if len(pr.Assignees) > 0 {
					content.WriteString(fmt.Sprintf("<p><em>Assignee: %s</em></p>\n", escapeHTML(strings.Join(pr.Assignees, ", "))))
				}
				if pr.IsDraft {
					content.WriteString("<p><em>Draft</em></p>\n")
				}
				content.WriteString("</ac:rich-text-body>\n")
				content.WriteString("</ac:structured-macro>\n")
			}
		}

		content.WriteString("</ac:layout-cell>\n")
	}

	content.WriteString("</ac:layout-section>\n")
	content.WriteString("</ac:layout>\n\n")
}

// generateCustomFormat generates a custom format
func (g *Generator) generateCustomFormat(content *strings.Builder, data *models.AggregatedData, _ Options) {
	// This is a placeholder for a custom format
	// In a real implementation, this would be customized based on specific requirements
	content.WriteString("<h2>Custom Format</h2>\n")
	content.WriteString("<p>This is a placeholder for a custom format. The actual implementation would depend on specific requirements.</p>\n")

	// Example: Group by team or component
	content.WriteString("<h3>By Team</h3>\n")
	
	// Get unique teams
	teams := make(map[string]bool)
	for _, issue := range data.JiraIssues {
		if issue.Team != "" {
			teams[issue.Team] = true
		}
	}

	// For each team, list the issues
	for team := range teams {
		content.WriteString(fmt.Sprintf("<h4>%s</h4>\n", team))
		content.WriteString("<ul>\n")
		
		for _, issue := range data.JiraIssues {
			if issue.Team == team {
				content.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a>: %s (%s)</li>\n", 
					issue.URL, issue.Key, issue.Summary, issue.Status))
			}
		}
		
		content.WriteString("</ul>\n")
	}
}

// mapStatusToColumn maps a Jira status to a kanban column
func mapStatusToColumn(status string) string {
	status = strings.ToLower(status)
	
	if strings.Contains(status, "to do") || strings.Contains(status, "backlog") || strings.Contains(status, "open") {
		return "To Do"
	}
	
	if strings.Contains(status, "in progress") || strings.Contains(status, "doing") {
		return "In Progress"
	}
	
	if strings.Contains(status, "review") || strings.Contains(status, "testing") || strings.Contains(status, "qa") {
		return "Review"
	}
	
	if strings.Contains(status, "done") || strings.Contains(status, "closed") || strings.Contains(status, "resolved") {
		return "Done"
	}
	
	// Default to "To Do" if status is unknown
	return "To Do"
}

// mapGitHubStateToColumn maps a GitHub issue state to a kanban column
func mapGitHubStateToColumn(state string) string {
	if state == "open" {
		return "To Do"
	}
	return "Done"
}

// mapGitHubPRStateToColumn maps a GitHub PR state and merge status to a kanban column
func mapGitHubPRStateToColumn(state, mergeStatus string) string {
	if state == "open" {
		if mergeStatus == "clean" || mergeStatus == "unstable" {
			return "Review"
		}
		return "In Progress"
	}
	if state == "closed" {
		return "Done"
	}
	return "In Progress"
}

// getStatusStyle returns a styled status indicator for Confluence Storage Format
func getStatusStyle(status string) string {
	var color string
	var backgroundColor string
	
	switch strings.ToLower(status) {
	case "done", "closed", "resolved", "complete", "completed", "ready for deployment":
		color = "#FFFFFF" // White text
		backgroundColor = "#36B37E" // Green background
	case "in progress", "review", "reviewing":
		color = "#000000" // Black text
		backgroundColor = "#FFAB00" // Yellow/Amber background
	case "blocked", "impediment":
		color = "#FFFFFF" // White text
		backgroundColor = "#FF5630" // Red background
	default:
		color = "#FFFFFF" // White text
		backgroundColor = "#6554C0" // Purple background for other statuses
	}
	
	// Use Confluence's native span with styling
	return fmt.Sprintf("<span style=\"display:inline-block; padding:2px 5px; background-color:%s; color:%s; border-radius:3px; font-size:11px; text-align:center;\">%s</span>", backgroundColor, color, html.EscapeString(status))
}

// escapeHTML safely escapes HTML content for Confluence
func escapeHTML(content string) string {
	return html.EscapeString(content)
}

// generateGanttFormat generates a Gantt chart format showing timeline of issues
func (g *Generator) generateGanttFormat(content *strings.Builder, data *models.AggregatedData, _ Options) {
	// Find the earliest and latest dates to set the chart boundaries
	var earliestDate, latestDate time.Time
	
	// Initialize with the first Jira issue if available
	if len(data.JiraIssues) > 0 {
		earliestDate = data.JiraIssues[0].CreatedDate
		latestDate = data.JiraIssues[0].UpdatedDate
	} else if len(data.GitHubIssues) > 0 {
		earliestDate = data.GitHubIssues[0].CreatedDate
		latestDate = data.GitHubIssues[0].UpdatedDate
	} else if len(data.GitHubPRs) > 0 {
		earliestDate = data.GitHubPRs[0].CreatedDate
		latestDate = data.GitHubPRs[0].UpdatedDate
	} else {
		// No data available
		content.WriteString("<h2>Gantt Chart</h2>\n")
		content.WriteString("<p>No data available to generate a Gantt chart.</p>\n")
		return
	}
	
	// Find the earliest creation date and latest update date across all items
	for _, issue := range data.JiraIssues {
		if issue.CreatedDate.Before(earliestDate) {
			earliestDate = issue.CreatedDate
		}
		if issue.UpdatedDate.After(latestDate) {
			latestDate = issue.UpdatedDate
		}
	}
	
	for _, issue := range data.GitHubIssues {
		if issue.CreatedDate.Before(earliestDate) {
			earliestDate = issue.CreatedDate
		}
		if issue.UpdatedDate.After(latestDate) {
			latestDate = issue.UpdatedDate
		}
	}
	
	for _, pr := range data.GitHubPRs {
		if pr.CreatedDate.Before(earliestDate) {
			earliestDate = pr.CreatedDate
		}
		if pr.UpdatedDate.After(latestDate) {
			latestDate = pr.UpdatedDate
		}
	}
	
	// Normalize dates to the first day of their respective months for better scaling
	earliestDate = time.Date(earliestDate.Year(), earliestDate.Month(), 1, 0, 0, 0, 0, earliestDate.Location())
	// For the end date, go to the first day of the next month
	latestDate = time.Date(latestDate.Year(), latestDate.Month()+1, 1, 0, 0, 0, 0, latestDate.Location())
	
	// Calculate the total duration in months
	monthsDiff := (latestDate.Year()-earliestDate.Year())*12 + int(latestDate.Month()-earliestDate.Month())
	if monthsDiff < 1 {
		monthsDiff = 1 // Ensure at least one month range
	}
	
	content.WriteString("<h2>Gantt Chart</h2>\n")
	
	// Use Confluence's layout feature for the Gantt chart
	content.WriteString("<ac:layout>\n")
	content.WriteString("<ac:layout-section ac:type=\"single\">\n")
	content.WriteString("<ac:layout-cell>\n")
	
	// Create a table for the Gantt chart
	content.WriteString("<table style=\"width:100%; border-collapse:collapse;\">\n")
	content.WriteString("<tbody>\n")
	
	// Header row with month names
	content.WriteString("<tr>\n")
	content.WriteString("<th style=\"background-color:#f4f5f7; text-align:left; padding:8px; border:1px solid #ddd; width:250px;\">Item</th>\n")
	
	// Generate month headers
	currentDate := earliestDate
	for i := 0; i < monthsDiff; i++ {
		monthName := currentDate.Format("Jan 2006")
		content.WriteString(fmt.Sprintf("<th style=\"background-color:#f4f5f7; text-align:center; padding:8px; border:1px solid #ddd;\">%s</th>\n", monthName))
		currentDate = currentDate.AddDate(0, 1, 0)
	}
	
	content.WriteString("</tr>\n")
	
	// Now add rows for each Jira issue
	for _, issue := range data.JiraIssues {
		content.WriteString("<tr>\n")
		
		// Item name cell
		content.WriteString(fmt.Sprintf("<td style=\"text-align:left; padding:8px; border:1px solid #ddd;\">\n"))
		content.WriteString(fmt.Sprintf("<a href=\"%s\">%s</a>: %s\n", issue.URL, escapeHTML(issue.Key), escapeHTML(issue.Summary)))
		content.WriteString("</td>\n")
		
		// Calculate the position of the issue in the timeline
		startMonth := (issue.CreatedDate.Year()-earliestDate.Year())*12 + int(issue.CreatedDate.Month()-earliestDate.Month())
		endMonth := (issue.UpdatedDate.Year()-earliestDate.Year())*12 + int(issue.UpdatedDate.Month()-earliestDate.Month())
		
		// Ensure we don't go out of bounds
		if startMonth < 0 {
			startMonth = 0
		}
		if endMonth >= monthsDiff {
			endMonth = monthsDiff - 1
		}
		if startMonth > endMonth {
			startMonth = endMonth
		}
		
		// Generate a unique color based on the issue key
		hash := 0
		for _, c := range issue.Key {
			hash = 31*hash + int(c)
		}
		
		// Generate a color from the hash - use a predefined palette for better visibility
		colors := []string{"#0052CC", "#6554C0", "#00875A", "#FF5630", "#FF8B00", "#36B37E", "#00B8D9", "#6554C0", "#4C9AFF", "#172B4D"}
		colorIndex := hash % len(colors)
		color := colors[colorIndex]
		
		// Generate cells for each month
		for i := 0; i < monthsDiff; i++ {
			if i >= startMonth && i <= endMonth {
				// This month is part of the issue's timeline with unique color
				content.WriteString(fmt.Sprintf("<td style=\"text-align:center; padding:8px; border:1px solid #ddd; background-color:%s; color:white;\">•</td>\n", color))
			} else {
				// Empty cell
				content.WriteString("<td style=\"text-align:center; padding:8px; border:1px solid #ddd;\"></td>\n")
			}
		}
		
		content.WriteString("</tr>\n")
	}
	
	// Add GitHub issues to the chart
	for _, issue := range data.GitHubIssues {
		content.WriteString("<tr>\n")
		
		// Item name cell
		content.WriteString(fmt.Sprintf("<td style=\"text-align:left; padding:8px; border:1px solid #ddd;\">\n"))
		content.WriteString(fmt.Sprintf("<a href=\"%s\">%s #%d</a>: %s\n", issue.URL, escapeHTML(issue.Repository), issue.Number, escapeHTML(issue.Title)))
		content.WriteString("</td>\n")
		
		// Calculate the position of the issue in the timeline
		startMonth := (issue.CreatedDate.Year()-earliestDate.Year())*12 + int(issue.CreatedDate.Month()-earliestDate.Month())
		endMonth := (issue.UpdatedDate.Year()-earliestDate.Year())*12 + int(issue.UpdatedDate.Month()-earliestDate.Month())
		
		// Ensure we don't go out of bounds
		if startMonth < 0 {
			startMonth = 0
		}
		if endMonth >= monthsDiff {
			endMonth = monthsDiff - 1
		}
		if startMonth > endMonth {
			startMonth = endMonth
		}
		
		// Generate a unique color based on the repository and issue number
		hash := 0
		for _, c := range issue.Repository {
			hash = 31*hash + int(c)
		}
		hash = hash + issue.Number
		
		// Generate a color from the hash - use a predefined palette for better visibility
		colors := []string{"#0052CC", "#6554C0", "#00875A", "#FF5630", "#FF8B00", "#36B37E", "#00B8D9", "#6554C0", "#4C9AFF", "#172B4D"}
		colorIndex := hash % len(colors)
		color := colors[colorIndex]
		
		// Generate cells for each month
		for i := 0; i < monthsDiff; i++ {
			if i >= startMonth && i <= endMonth {
				// This month is part of the issue's timeline with unique color
				content.WriteString(fmt.Sprintf("<td style=\"text-align:center; padding:8px; border:1px solid #ddd; background-color:%s; color:white;\">•</td>\n", color))
			} else {
				// Empty cell
				content.WriteString("<td style=\"text-align:center; padding:8px; border:1px solid #ddd;\"></td>\n")
			}
		}
		
		content.WriteString("</tr>\n")
	}
	
	// Add GitHub PRs to the chart
	for _, pr := range data.GitHubPRs {
		content.WriteString("<tr>\n")
		
		// Item name cell
		content.WriteString(fmt.Sprintf("<td style=\"text-align:left; padding:8px; border:1px solid #ddd;\">\n"))
		content.WriteString(fmt.Sprintf("<a href=\"%s\">%s #%d</a>: %s\n", pr.URL, escapeHTML(pr.Repository), pr.Number, escapeHTML(pr.Title)))
		content.WriteString("</td>\n")
		
		// Calculate the position of the PR in the timeline
		startMonth := (pr.CreatedDate.Year()-earliestDate.Year())*12 + int(pr.CreatedDate.Month()-earliestDate.Month())
		endMonth := (pr.UpdatedDate.Year()-earliestDate.Year())*12 + int(pr.UpdatedDate.Month()-earliestDate.Month())
		
		// Ensure we don't go out of bounds
		if startMonth < 0 {
			startMonth = 0
		}
		if endMonth >= monthsDiff {
			endMonth = monthsDiff - 1
		}
		if startMonth > endMonth {
			startMonth = endMonth
		}
		
		// Generate a unique color based on the repository and PR number
		hash := 0
		for _, c := range pr.Repository {
			hash = 31*hash + int(c)
		}
		hash = hash + pr.Number
		
		// Generate a color from the hash - use a predefined palette for better visibility
		colors := []string{"#0052CC", "#6554C0", "#00875A", "#FF5630", "#FF8B00", "#36B37E", "#00B8D9", "#6554C0", "#4C9AFF", "#172B4D"}
		colorIndex := hash % len(colors)
		color := colors[colorIndex]
		
		// Generate cells for each month
		for i := 0; i < monthsDiff; i++ {
			if i >= startMonth && i <= endMonth {
				// This month is part of the PR's timeline with unique color
				content.WriteString(fmt.Sprintf("<td style=\"text-align:center; padding:8px; border:1px solid #ddd; background-color:%s; color:white;\">•</td>\n", color))
			} else {
				// Empty cell
				content.WriteString("<td style=\"text-align:center; padding:8px; border:1px solid #ddd;\"></td>\n")
			}
		}
		
		content.WriteString("</tr>\n")
	}
	
	// Close the table
	content.WriteString("</tbody>\n")
	content.WriteString("</table>\n")
	
	// Close the layout
	content.WriteString("</ac:layout-cell>\n")
	content.WriteString("</ac:layout-section>\n")
	// Add a legend using Confluence's panel macro
	content.WriteString("<ac:structured-macro ac:name=\"info\">\n")
	content.WriteString("<ac:rich-text-body>\n")
	content.WriteString("<p><strong>Legend</strong></p>\n")
	content.WriteString("<p>Each item has a unique color based on its identifier:</p>\n")
	content.WriteString("<ul>\n")
	content.WriteString("<li>Each row represents a Jira issue, GitHub issue, or GitHub pull request</li>\n")
	content.WriteString("<li>Colored cells indicate months where the item was active</li>\n")
	content.WriteString("<li>The timeline spans from the item's creation date to its last update date</li>\n")
	content.WriteString("</ul>\n")
	content.WriteString("<p><small>The timeline shows when each item was created to when it was last updated.</small></p>\n")
	content.WriteString("</ac:rich-text-body>\n")
	content.WriteString("</ac:structured-macro>\n")
	
	// Close the layout
	content.WriteString("</ac:layout>\n")
}
