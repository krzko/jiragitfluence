package generator

import (
	"fmt"
	"strings"

	"github.com/krzko/jiragitfluence/pkg/models"
)

// generateTimelineView generates a timeline-based roadmap view
func (g *Generator) generateTimelineView(content *strings.Builder, data *models.AggregatedData, quarters []string, opts Options) {
	// Add a more professional header
	content.WriteString("<h2>Roadmap</h2>\n")

	// Add timeframe information if provided
	if opts.RoadmapTimeframe != "" {
		content.WriteString(fmt.Sprintf("<p>This roadmap shows planned work for the period <strong>%s</strong>.</p>\n", opts.RoadmapTimeframe))
	} else {
		content.WriteString("<p>This roadmap shows planned and in-progress work items.</p>\n")
	}

	content.WriteString("<h3>Timeline View</h3>\n")
	content.WriteString("<p>This view shows work items arranged by their planned start and end dates.</p>\n")

	// If no quarters are provided, generate them based on current date
	if len(quarters) == 0 {
		// Generate quarters for the next year (4 quarters)
		quarters = generateQuartersFromNow(4)
	}

	// Create a table for the timeline with improved styling
	content.WriteString("<table class=\"confluenceTable\" style=\"width: 100%; border-collapse: collapse; margin-top: 20px;\">\n")
	content.WriteString("<thead>\n")
	content.WriteString("<tr>\n")
	content.WriteString("<th class=\"confluenceTh\" style=\"padding: 10px; text-align: left; border: 1px solid #ddd; background-color: #f2f2f2; width: 25%;\">Item</th>\n")

	// Add quarter columns
	for _, quarter := range quarters {
		content.WriteString(fmt.Sprintf("<th class=\"confluenceTh\" style=\"padding: 10px; text-align: center; border: 1px solid #ddd; background-color: #f2f2f2;\">%s</th>\n", quarter))
	}

	content.WriteString("</tr>\n")
	content.WriteString("</thead>\n")
	content.WriteString("<tbody>\n")

	// Group items based on the grouping option
	groupedItems := groupRoadmapItems(data, opts.RoadmapGrouping)

	// Add a summary section showing counts by status
	content.WriteString("<div style=\"margin-bottom: 20px;\">\n")
	content.WriteString("<h4>Status Summary</h4>\n")
	content.WriteString("<table class=\"confluenceTable\" style=\"width: 50%; border-collapse: collapse; margin-bottom: 15px;\">\n")
	content.WriteString("<thead>\n<tr>\n<th class=\"confluenceTh\" style=\"padding: 8px; text-align: left; border: 1px solid #ddd; background-color: #f2f2f2;\">Status</th>\n<th class=\"confluenceTh\" style=\"padding: 8px; text-align: center; border: 1px solid #ddd; background-color: #f2f2f2;\">Count</th>\n</tr>\n</thead>\n")
	content.WriteString("<tbody>\n")

	// Count items by status
	statusCounts := make(map[string]int)
	for _, items := range groupedItems {
		for _, item := range items {
			statusCounts[item.Status]++
		}
	}

	// Display counts
	for status, count := range statusCounts {
		statusColor := getRoadmapStatusColor(status)
		textColor := "#000000" // Default to black text

		// Use white text for darker backgrounds
		if status == "Blocked" || status == "At Risk" {
			textColor = "#FFFFFF"
		}

		content.WriteString(fmt.Sprintf("<tr>\n<td class=\"confluenceTd\" style=\"padding: 8px; border: 1px solid #ddd; background-color:%s; color:%s; font-weight:bold;\">%s</td>\n<td class=\"confluenceTd\" style=\"padding: 8px; border: 1px solid #ddd; text-align: center;\">%d</td>\n</tr>\n",
			statusColor, textColor, status, count))
	}

	content.WriteString("</tbody>\n</table>\n</div>\n")

	// Add rows for each group
	for group, items := range groupedItems {
		// Add a group header row with improved styling
		content.WriteString("<tr>\n")
		content.WriteString(fmt.Sprintf("<td class=\"confluenceTd\" colspan=\"%d\" style=\"padding: 10px; background-color:#e9f0f7; font-weight:bold; border: 1px solid #ddd; border-bottom: 2px solid #4a6785;\">%s</td>\n",
			len(quarters)+1, escapeHTML(group)))
		content.WriteString("</tr>\n")

		// Add rows for each item in the group
		for _, item := range items {
			content.WriteString("<tr>\n")

			// Item details cell with improved styling
			content.WriteString("<td class=\"confluenceTd\" style=\"vertical-align:top; padding: 10px; border: 1px solid #ddd;\">\n")

			switch item.Type {
			case "jira":
				jiraIssue := item.JiraIssue
				content.WriteString(fmt.Sprintf("<strong><a href=\"%s\" style=\"text-decoration: none;\">%s</a></strong>: %s<br/>\n",
					jiraIssue.URL, jiraIssue.Key, escapeHTML(jiraIssue.Summary)))
				// Get status color for better visual indication
				statusColor := getRoadmapStatusColor(item.Status)
				textColor := "#000000" // Default to black text

				// Use white text for darker backgrounds
				if item.Status == "Blocked" || item.Status == "At Risk" {
					textColor = "#FFFFFF"
				}

				content.WriteString(fmt.Sprintf("<div style=\"margin-top: 5px;\">\n"))
				content.WriteString(fmt.Sprintf("<span style=\"display: inline-block; padding: 3px 8px; border-radius: 3px; background-color:%s; color:%s; font-size: 12px; margin-right: 5px;\">%s</span>\n",
					statusColor, textColor, item.Status))
				content.WriteString(fmt.Sprintf("<small>Assignee: %s</small>\n",
					escapeHTML(jiraIssue.Assignee)))
				content.WriteString("</div>\n")

			case "github-issue":
				githubIssue := item.GitHubIssue
				content.WriteString(fmt.Sprintf("<a href=\"%s\">%s #%d</a>: %s<br/>\n",
					githubIssue.URL, escapeHTML(githubIssue.Repository), githubIssue.Number, escapeHTML(githubIssue.Title)))
				// Get status color for better visual indication
				statusColor := getRoadmapStatusColor(item.Status)
				textColor := "#000000" // Default to black text

				// Use white text for darker backgrounds
				if item.Status == "Blocked" || item.Status == "At Risk" {
					textColor = "#FFFFFF"
				}

				content.WriteString(fmt.Sprintf("<div style=\"margin-top: 5px;\">\n"))
				content.WriteString(fmt.Sprintf("<span style=\"display: inline-block; padding: 3px 8px; border-radius: 3px; background-color:%s; color:%s; font-size: 12px; margin-right: 5px;\">%s</span>\n",
					statusColor, textColor, item.Status))
				content.WriteString(fmt.Sprintf("<small>Assignees: %s</small>\n",
					strings.Join(githubIssue.Assignees, ", ")))
				content.WriteString("</div>\n")

				// We're not including PRs in the roadmap as per requirements
			}

			content.WriteString("</td>\n")

			// Generate timeline cells for each quarter
			startQuarter, endQuarter := getItemQuarters(item, quarters)

			for i := range quarters {
				if i >= startQuarter && i <= endQuarter {
					// This quarter is part of the item's timeline
					color := getRoadmapStatusColor(item.Status)
					textColor := "#000000" // Default to black text

					// Use white text for darker backgrounds
					if item.Status == "Blocked" || item.Status == "At Risk" {
						textColor = "#FFFFFF"
					}

					// Determine if this is the start, middle, or end of the timeline
					var cellContent string
					var borderStyle string

					if i == startQuarter && i == endQuarter {
						// Single quarter item
						cellContent = "<span style=\"font-weight: bold;\">●</span>"
						borderStyle = "border: 1px solid #ddd; border-radius: 4px;"
					} else if i == startQuarter {
						// Start of timeline
						cellContent = "<span style=\"font-weight: bold;\">▶</span>"
						borderStyle = "border: 1px solid #ddd; border-left: 3px solid " + color + "; border-top: 1px solid " + color + "; border-bottom: 1px solid " + color + ";"
					} else if i == endQuarter {
						// End of timeline
						cellContent = "<span style=\"font-weight: bold;\">◀</span>"
						borderStyle = "border: 1px solid #ddd; border-right: 3px solid " + color + "; border-top: 1px solid " + color + "; border-bottom: 1px solid " + color + ";"
					} else {
						// Middle of timeline
						cellContent = "<span style=\"font-weight: bold;\">━</span>"
						borderStyle = "border: 1px solid #ddd; border-top: 1px solid " + color + "; border-bottom: 1px solid " + color + ";"
					}

					content.WriteString(fmt.Sprintf("<td class=\"confluenceTd\" style=\"text-align: center; padding: 10px; background-color:%s; color:%s; %s\">%s</td>\n",
						color, textColor, borderStyle, cellContent))
				} else {
					// Empty cell
					content.WriteString("<td class=\"confluenceTd\" style=\"border: 1px solid #ddd; padding: 10px;\"></td>\n")
				}
			}

			content.WriteString("</tr>\n")
		}
	}

	content.WriteString("</tbody>\n")
	content.WriteString("</table>\n")

	// Add a legend for the roadmap
	content.WriteString("<div style=\"margin-top: 20px; margin-bottom: 20px;\">\n")
	content.WriteString("<h4>Legend</h4>\n")
	content.WriteString("<table style=\"width: auto; border-collapse: collapse; margin-bottom: 15px;\">\n")
	content.WriteString("<tr>\n")

	// Add legend items for each status
	statuses := []string{"Planned", "In Progress", "At Risk", "Blocked", "Completed"}
	for _, status := range statuses {
		statusColor := getRoadmapStatusColor(status)
		textColor := "#000000"
		if status == "Blocked" || status == "At Risk" {
			textColor = "#FFFFFF"
		}
		content.WriteString(fmt.Sprintf("<td style=\"padding: 5px 10px; margin-right: 10px; background-color:%s; color:%s; border-radius: 3px; font-size: 12px;\">%s</td>\n",
			statusColor, textColor, status))
		content.WriteString("<td style=\"padding-right: 15px;\"></td>\n")
	}

	content.WriteString("</tr>\n")
	content.WriteString("</table>\n")

	// Add timeline symbols legend
	content.WriteString("<table style=\"width: auto; border-collapse: collapse; margin-bottom: 15px;\">\n")
	content.WriteString("<tr>\n")
	content.WriteString("<td style=\"padding: 5px 10px; font-weight: bold;\">●</td>\n")
	content.WriteString("<td style=\"padding-right: 15px;\">Single quarter item</td>\n")

	content.WriteString("<td style=\"padding: 5px 10px; font-weight: bold;\">▶</td>\n")
	content.WriteString("<td style=\"padding-right: 15px;\">Start of multi-quarter item</td>\n")

	content.WriteString("<td style=\"padding: 5px 10px; font-weight: bold;\">━</td>\n")
	content.WriteString("<td style=\"padding-right: 15px;\">Middle of timeline</td>\n")

	content.WriteString("<td style=\"padding: 5px 10px; font-weight: bold;\">◀</td>\n")
	content.WriteString("<td style=\"padding-right: 15px;\">End of multi-quarter item</td>\n")
	content.WriteString("</tr>\n")
	content.WriteString("</table>\n")
	content.WriteString("</div>\n")

	// Add dependencies visualization if requested
	if opts.IncludeDependencies {
		g.addDependenciesVisualization(content, data)
	}
}

// generateStrategicView generates a strategic roadmap view focusing on themes and initiatives
func (g *Generator) generateStrategicView(content *strings.Builder, data *models.AggregatedData, quarters []string, opts Options) {
	// Use opts to customize the view based on user preferences
	// Add dependencies visualization if requested
	if opts.IncludeDependencies {
		content.WriteString("<p><em>Dependencies will be shown at the end of the view.</em></p>\n")
	}
	content.WriteString("<h3>Strategic View</h3>\n")
	content.WriteString("<p>This view shows work items grouped by strategic themes and initiatives.</p>\n")

	// Extract themes from the data
	themes := extractThemes(data)

	// Create a table for the strategic view
	content.WriteString("<table class=\"confluenceTable\">\n")
	content.WriteString("<thead>\n")
	content.WriteString("<tr>\n")
	content.WriteString("<th class=\"confluenceTh\">Theme/Initiative</th>\n")

	// Add quarter columns
	for _, quarter := range quarters {
		content.WriteString(fmt.Sprintf("<th class=\"confluenceTh\">%s</th>\n", quarter))
	}

	content.WriteString("</tr>\n")
	content.WriteString("</thead>\n")
	content.WriteString("<tbody>\n")

	// Add rows for each theme and its initiatives
	for _, theme := range themes {
		// Add a theme header row
		content.WriteString("<tr>\n")
		content.WriteString(fmt.Sprintf("<td class=\"confluenceTd\" colspan=\"%d\" style=\"background-color:#f4f5f7; font-weight:bold;\">%s</td>\n",
			len(quarters)+1, escapeHTML(theme.Name)))
		content.WriteString("</tr>\n")

		// Add rows for each initiative under this theme
		for _, initiative := range theme.Initiatives {
			content.WriteString("<tr>\n")
			content.WriteString(fmt.Sprintf("<td class=\"confluenceTd\">%s</td>\n", escapeHTML(initiative.Name)))

			// Generate timeline cells for each quarter
			for i := range quarters {
				if i >= initiative.StartQuarter && i <= initiative.EndQuarter {
					// This quarter is part of the initiative's timeline
					color := getRoadmapStatusColor(initiative.Status)
					content.WriteString(fmt.Sprintf("<td class=\"confluenceTd\" style=\"background-color:%s;\">•</td>\n", color))
				} else {
					// Empty cell
					content.WriteString("<td class=\"confluenceTd\"></td>\n")
				}
			}

			content.WriteString("</tr>\n")
		}
	}

	content.WriteString("</tbody>\n")
	content.WriteString("</table>\n")
}

// generateReleaseView generates a release-based roadmap view
func (g *Generator) generateReleaseView(content *strings.Builder, data *models.AggregatedData, quarters []string, opts Options) {
	// Use quarters and opts to customize the view
	timeframe := opts.RoadmapTimeframe

	// Add dependencies note if requested
	if opts.IncludeDependencies {
		content.WriteString("<p><em>Dependencies will be shown at the end of the view.</em></p>\n")
	}

	// Add timeframe information to the view if provided
	if timeframe != "" {
		content.WriteString(fmt.Sprintf("<p><em>Timeframe: %s</em></p>\n", timeframe))
	}
	content.WriteString("<h3>Release View</h3>\n")
	content.WriteString("<p>This view shows work items organized by planned releases or milestones.</p>\n")

	// Extract milestones/releases from the data
	milestones := extractMilestones(data)

	// Create a table for the release view
	content.WriteString("<table class=\"confluenceTable\">\n")
	content.WriteString("<thead>\n")
	content.WriteString("<tr>\n")
	content.WriteString("<th class=\"confluenceTh\">Release/Milestone</th>\n")
	content.WriteString("<th class=\"confluenceTh\">Target Date</th>\n")
	content.WriteString("<th class=\"confluenceTh\">Status</th>\n")
	content.WriteString("<th class=\"confluenceTh\">Key Deliverables</th>\n")
	content.WriteString("</tr>\n")
	content.WriteString("</thead>\n")
	content.WriteString("<tbody>\n")

	// Add rows for each milestone
	for _, milestone := range milestones {
		content.WriteString("<tr>\n")
		content.WriteString(fmt.Sprintf("<td class=\"confluenceTd\">%s</td>\n", escapeHTML(milestone.Name)))
		content.WriteString(fmt.Sprintf("<td class=\"confluenceTd\">%s</td>\n", milestone.TargetDate.Format("Jan 2006")))

		// Status cell with color
		statusColor := getRoadmapStatusColor(milestone.Status)
		content.WriteString(fmt.Sprintf("<td class=\"confluenceTd\" style=\"background-color:%s; color:white;\">%s</td>\n",
			statusColor, milestone.Status))

		// Key deliverables cell
		content.WriteString("<td class=\"confluenceTd\">\n")
		content.WriteString("<ul>\n")
		for _, item := range milestone.Items {
			switch item.Type {
			case "jira":
				jiraIssue := item.JiraIssue
				content.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a>: %s</li>\n",
					jiraIssue.URL, jiraIssue.Key, escapeHTML(jiraIssue.Summary)))

			case "github-issue":
				githubIssue := item.GitHubIssue
				content.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s #%d</a>: %s</li>\n",
					githubIssue.URL, escapeHTML(githubIssue.Repository), githubIssue.Number, escapeHTML(githubIssue.Title)))

				// We're not including PRs in the roadmap as per requirements
			}
		}
		content.WriteString("</ul>\n")
		content.WriteString("</td>\n")

		content.WriteString("</tr>\n")
	}

	content.WriteString("</tbody>\n")
	content.WriteString("</table>\n")
}

// generateEpicGanttView generates a Gantt-style roadmap view organized by epics
func (g *Generator) generateEpicGanttView(content *strings.Builder, data *models.AggregatedData, quarters []string, opts Options) {
	// Add header
	content.WriteString("<h3>Epic Roadmap</h3>\n")
	content.WriteString("<p>This view shows work items organized by epics across time periods.</p>\n")

	// If no quarters are provided, generate them based on current date
	if len(quarters) == 0 {
		// Generate quarters for the next year (4 quarters)
		quarters = generateQuartersFromNow(4)
	}

	// Create a map to store issues by epic
	epicMap := make(map[string][]RoadmapItem)

	// Track all epics to maintain order
	allEpics := []string{}
	epicKeyToName := make(map[string]string)

	// First pass: collect all epics
	for _, issue := range data.JiraIssues {
		if issue.IssueType == "Epic" {
			epicKey := issue.Key
			epicName := issue.Summary

			// Only add if not already in the list
			if _, exists := epicKeyToName[epicKey]; !exists {
				allEpics = append(allEpics, epicKey)
				epicKeyToName[epicKey] = epicName
			}
		}
	}

	// Add a special category for issues without epics
	allEpics = append(allEpics, "No Epic")
	epicKeyToName["No Epic"] = "Issues without Epic"

	// Second pass: group issues by epic
	for _, issue := range data.JiraIssues {
		// Skip epics themselves as they'll be headers
		if issue.IssueType == "Epic" {
			continue
		}

		// For issues without planned dates, use created/updated dates as fallbacks
		startDate := issue.PlannedStartDate
		endDate := issue.PlannedEndDate

		// If no planned start date, use created date
		if startDate == nil {
			createdDate := issue.CreatedDate
			startDate = &createdDate
		}

		// If no planned end date, use updated date for completed issues
		// or set it to 2 weeks after start for in-progress issues
		if endDate == nil {
			if issue.Status == "Done" || issue.Status == "Closed" || issue.Status == "Resolved" {
				updatedDate := issue.UpdatedDate
				endDate = &updatedDate
			} else {
				// Set end date to 2 weeks after start date
				twoWeeksLater := startDate.AddDate(0, 0, 14)
				endDate = &twoWeeksLater
			}
		}

		item := RoadmapItem{
			Type:      "jira",
			JiraIssue: &issue,
			Status:    issue.RoadmapStatus,
			StartDate: startDate,
			EndDate:   endDate,
		}

		// If no roadmap status is set, derive it from the issue status
		if item.Status == "" {
			item.Status = deriveRoadmapStatus(issue.Status)
		}

		// Determine which epic this issue belongs to
		epicKey := "No Epic"
		if issue.EpicLink != "" {
			epicKey = issue.EpicLink
		}

		// Add to the appropriate epic group
		epicMap[epicKey] = append(epicMap[epicKey], item)
	}

	// Add GitHub issues to the map (they'll go under "No Epic" unless they have a Jira reference)
	for _, issue := range data.GitHubIssues {
		// For issues without planned dates, use created/updated dates as fallbacks
		startDate := issue.PlannedStartDate
		endDate := issue.PlannedEndDate

		// If no planned start date, use created date
		if startDate == nil {
			createdDate := issue.CreatedDate
			startDate = &createdDate
		}

		// If no planned end date, use updated date for completed issues
		// or set it to 2 weeks after start for in-progress issues
		if endDate == nil {
			if issue.State == "closed" {
				updatedDate := issue.UpdatedDate
				endDate = &updatedDate
			} else {
				// Set end date to 2 weeks after start date
				twoWeeksLater := startDate.AddDate(0, 0, 14)
				endDate = &twoWeeksLater
			}
		}

		item := RoadmapItem{
			Type:        "github-issue",
			GitHubIssue: &issue,
			Status:      issue.RoadmapStatus,
			StartDate:   startDate,
			EndDate:     endDate,
		}

		// If no roadmap status is set, derive it from the issue state
		if item.Status == "" {
			item.Status = deriveRoadmapStatusFromGitHub(issue.State)
		}

		// Check if this GitHub issue references a Jira epic in its title
		// This is a simple check that could be enhanced with regex
		epicKey := "No Epic"
		for key := range epicKeyToName {
			if key != "No Epic" && strings.Contains(issue.Title, key) {
				epicKey = key
				break
			}
		}

		// Add to the appropriate epic group
		epicMap[epicKey] = append(epicMap[epicKey], item)
	}

	// Create the Gantt chart table
	content.WriteString("<table class=\"confluenceTable\" style=\"width: 100%; border-collapse: collapse; margin-top: 20px;\">\n")
	content.WriteString("<thead>\n")
	content.WriteString("<tr>\n")
	content.WriteString("<th class=\"confluenceTh\" style=\"padding: 10px; text-align: left; border: 1px solid #ddd; background-color: #f2f2f2; width: 25%;\">Epic</th>\n")

	// Add month columns
	for _, quarter := range quarters {
		content.WriteString(fmt.Sprintf("<th class=\"confluenceTh\" style=\"padding: 10px; text-align: center; border: 1px solid #ddd; background-color: #f2f2f2;\">%s</th>\n", quarter))
	}

	content.WriteString("</tr>\n")
	content.WriteString("</thead>\n")
	content.WriteString("<tbody>\n")

	// Add rows for each epic
	for _, epicKey := range allEpics {
		// Skip epics with no issues
		if len(epicMap[epicKey]) == 0 && epicKey != "No Epic" {
			continue
		}

		// Add epic header row
		content.WriteString("<tr>\n")
		content.WriteString(fmt.Sprintf("<td class=\"confluenceTd\" style=\"padding: 10px; background-color:#e9f0f7; font-weight:bold; border: 1px solid #ddd; border-bottom: 2px solid #4a6785;\">%s</td>\n",
			escapeHTML(epicKeyToName[epicKey])))

		// Add empty cells for the quarters
		for range quarters {
			content.WriteString("<td class=\"confluenceTd\" style=\"padding: 10px; border: 1px solid #ddd; background-color:#e9f0f7;\"></td>\n")
		}

		content.WriteString("</tr>\n")

		// Add rows for each item in the epic
		for _, item := range epicMap[epicKey] {
			content.WriteString("<tr>\n")

			// Item details cell
			content.WriteString("<td class=\"confluenceTd\" style=\"vertical-align:top; padding: 10px; border: 1px solid #ddd;\">\n")

			switch item.Type {
			case "jira":
				jiraIssue := item.JiraIssue
				content.WriteString(fmt.Sprintf("<strong><a href=\"%s\" style=\"text-decoration: none;\">%s</a></strong>: %s<br/>\n",
					jiraIssue.URL, jiraIssue.Key, escapeHTML(jiraIssue.Summary)))

				// Get status color for better visual indication
				statusColor := getRoadmapStatusColor(item.Status)
				textColor := "#000000" // Default to black text

				// Use white text for darker backgrounds
				if item.Status == "Blocked" || item.Status == "At Risk" {
					textColor = "#FFFFFF"
				}

				content.WriteString(fmt.Sprintf("<div style=\"margin-top: 5px;\">\n"))
				content.WriteString(fmt.Sprintf("<span style=\"display: inline-block; padding: 3px 8px; border-radius: 3px; background-color:%s; color:%s; font-size: 12px; margin-right: 5px;\">%s</span>\n",
					statusColor, textColor, item.Status))
				content.WriteString(fmt.Sprintf("<small>Assignee: %s</small>\n",
					escapeHTML(jiraIssue.Assignee)))
				content.WriteString("</div>\n")

			case "github-issue":
				githubIssue := item.GitHubIssue
				content.WriteString(fmt.Sprintf("<strong><a href=\"%s\" style=\"text-decoration: none;\">%s #%d</a></strong>: %s<br/>\n",
					githubIssue.URL, escapeHTML(githubIssue.Repository), githubIssue.Number, escapeHTML(githubIssue.Title)))

				// Get status color for better visual indication
				statusColor := getRoadmapStatusColor(item.Status)
				textColor := "#000000" // Default to black text

				// Use white text for darker backgrounds
				if item.Status == "Blocked" || item.Status == "At Risk" {
					textColor = "#FFFFFF"
				}

				content.WriteString(fmt.Sprintf("<div style=\"margin-top: 5px;\">\n"))
				content.WriteString(fmt.Sprintf("<span style=\"display: inline-block; padding: 3px 8px; border-radius: 3px; background-color:%s; color:%s; font-size: 12px; margin-right: 5px;\">%s</span>\n",
					statusColor, textColor, item.Status))
				content.WriteString(fmt.Sprintf("<small>Assignees: %s</small>\n",
					strings.Join(githubIssue.Assignees, ", ")))
				content.WriteString("</div>\n")
			}

			content.WriteString("</td>\n")

			// Generate timeline cells for each quarter
			startQuarter, endQuarter := getItemQuarters(item, quarters)

			for i := range quarters {
				if i >= startQuarter && i <= endQuarter {
					// This quarter is part of the item's timeline
					color := getRoadmapStatusColor(item.Status)
					textColor := "#000000" // Default to black text

					// Use white text for darker backgrounds
					if item.Status == "Blocked" || item.Status == "At Risk" {
						textColor = "#FFFFFF"
					}

					// Determine if this is the start, middle, or end of the timeline
					var cellContent string
					var borderStyle string

					if i == startQuarter && i == endQuarter {
						// Single quarter item
						cellContent = "<span style=\"font-weight: bold;\">●</span>"
						borderStyle = "border: 1px solid #ddd; border-radius: 4px;"
					} else if i == startQuarter {
						// Start of timeline
						cellContent = "<span style=\"font-weight: bold;\">▶</span>"
						borderStyle = "border: 1px solid #ddd; border-left: 3px solid " + color + "; border-top: 1px solid " + color + "; border-bottom: 1px solid " + color + ";"
					} else if i == endQuarter {
						// End of timeline
						cellContent = "<span style=\"font-weight: bold;\">◀</span>"
						borderStyle = "border: 1px solid #ddd; border-right: 3px solid " + color + "; border-top: 1px solid " + color + "; border-bottom: 1px solid " + color + ";"
					} else {
						// Middle of timeline
						cellContent = "<span style=\"font-weight: bold;\">━</span>"
						borderStyle = "border: 1px solid #ddd; border-top: 1px solid " + color + "; border-bottom: 1px solid " + color + ";"
					}

					content.WriteString(fmt.Sprintf("<td class=\"confluenceTd\" style=\"text-align: center; padding: 10px; background-color:%s; color:%s; %s\">%s</td>\n",
						color, textColor, borderStyle, cellContent))
				} else {
					// Empty cell
					content.WriteString("<td class=\"confluenceTd\" style=\"border: 1px solid #ddd; padding: 10px;\"></td>\n")
				}
			}

			content.WriteString("</tr>\n")
		}
	}

	content.WriteString("</tbody>\n")
	content.WriteString("</table>\n")

	// Add a legend for the roadmap
	content.WriteString("<div style=\"margin-top: 20px; margin-bottom: 20px;\">\n")
	content.WriteString("<h4>Legend</h4>\n")
	content.WriteString("<table style=\"width: auto; border-collapse: collapse; margin-bottom: 15px;\">\n")
	content.WriteString("<tr>\n")

	// Add legend items for each status
	statuses := []string{"Planned", "In Progress", "At Risk", "Blocked", "Completed"}
	for _, status := range statuses {
		statusColor := getRoadmapStatusColor(status)
		textColor := "#000000"
		if status == "Blocked" || status == "At Risk" {
			textColor = "#FFFFFF"
		}
		content.WriteString(fmt.Sprintf("<td style=\"padding: 5px 10px; margin-right: 10px; background-color:%s; color:%s; border-radius: 3px; font-size: 12px;\">%s</td>\n",
			statusColor, textColor, status))
		content.WriteString("<td style=\"padding-right: 15px;\"></td>\n")
	}

	content.WriteString("</tr>\n")
	content.WriteString("</table>\n")

	// Add timeline symbols legend
	content.WriteString("<table style=\"width: auto; border-collapse: collapse; margin-bottom: 15px;\">\n")
	content.WriteString("<tr>\n")
	content.WriteString("<td style=\"padding: 5px 10px; font-weight: bold;\">●</td>\n")
	content.WriteString("<td style=\"padding-right: 15px;\">Single quarter item</td>\n")

	content.WriteString("<td style=\"padding: 5px 10px; font-weight: bold;\">▶</td>\n")
	content.WriteString("<td style=\"padding-right: 15px;\">Start of multi-quarter item</td>\n")

	content.WriteString("<td style=\"padding: 5px 10px; font-weight: bold;\">━</td>\n")
	content.WriteString("<td style=\"padding-right: 15px;\">Middle of timeline</td>\n")

	content.WriteString("<td style=\"padding: 5px 10px; font-weight: bold;\">◀</td>\n")
	content.WriteString("<td style=\"padding-right: 15px;\">End of multi-quarter item</td>\n")
	content.WriteString("</tr>\n")
	content.WriteString("</table>\n")
	content.WriteString("</div>\n")

	// Add dependencies visualization if requested
	if opts.IncludeDependencies {
		g.addDependenciesVisualization(content, data)
	}
}

// addDependenciesVisualization adds a visualization of dependencies between items
func (g *Generator) addDependenciesVisualization(content *strings.Builder, data *models.AggregatedData) {
	content.WriteString("<h3>Dependencies</h3>\n")
	content.WriteString("<p>This diagram shows dependencies between work items.</p>\n")

	// Use Confluence's PlantUML macro for dependency visualization
	content.WriteString("<ac:structured-macro ac:name=\"plantuml\">\n")
	content.WriteString("<ac:parameter ac:name=\"atlassian-macro-output-type\">BLOCK</ac:parameter>\n")
	content.WriteString("<ac:plain-text-body><![CDATA[\n")

	// Generate PlantUML diagram
	content.WriteString("@startuml\n")
	content.WriteString("skinparam monochrome true\n")
	content.WriteString("skinparam shadowing false\n")
	content.WriteString("skinparam defaultFontName Arial\n")
	content.WriteString("skinparam defaultFontSize 12\n")

	// Add nodes for Jira issues
	for _, issue := range data.JiraIssues {
		if issue.Dependencies != nil && len(issue.Dependencies) > 0 {
			content.WriteString(fmt.Sprintf("rectangle \"%s\\n%s\" as %s\n",
				issue.Key, escapeHTML(issue.Summary), issue.Key))
		}
	}

	// Add nodes for GitHub issues
	for i, issue := range data.GitHubIssues {
		if issue.Dependencies != nil && len(issue.Dependencies) > 0 {
			nodeID := fmt.Sprintf("GH_ISSUE_%d", i)
			content.WriteString(fmt.Sprintf("rectangle \"%s #%d\\n%s\" as %s\n",
				escapeHTML(issue.Repository), issue.Number, escapeHTML(issue.Title), nodeID))
		}
	}

	// We're not including PRs in the roadmap as per requirements

	// Add dependency relationships
	for _, issue := range data.JiraIssues {
		for _, dep := range issue.Dependencies {
			content.WriteString(fmt.Sprintf("%s --> %s\n", issue.Key, dep))
		}
	}

	content.WriteString("@enduml\n")
	content.WriteString("]]></ac:plain-text-body>\n")
	content.WriteString("</ac:structured-macro>\n")
}
