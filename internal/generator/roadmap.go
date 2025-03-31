package generator

import (
	"fmt"
	"strings"
	"time"

	"github.com/krzko/jiragitfluence/pkg/models"
)

// generateRoadmapFormat generates a roadmap format for planning future work
func (g *Generator) generateRoadmapFormat(content *strings.Builder, data *models.AggregatedData, opts Options) {
	content.WriteString("<h2>Roadmap</h2>\n")

	// Parse the timeframe option to determine the date range
	startDate, endDate := parseTimeframe(opts.RoadmapTimeframe)
	
	// If no specific timeframe is provided, default to a 12-month view starting from current month
	if startDate.IsZero() {
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(1, 0, 0)
	}

	// Generate quarters between start and end date
	quarters := generateQuarters(startDate, endDate)

	// Add a description of the roadmap
	content.WriteString("<p>This roadmap shows planned work for the period ")
	content.WriteString(fmt.Sprintf("<strong>%s</strong> to <strong>%s</strong>.</p>\n", 
		startDate.Format("January 2006"), 
		endDate.Format("January 2006")))

	// Generate the appropriate view based on the option
	switch opts.RoadmapView {
	case TimelineView:
		g.generateTimelineView(content, data, quarters, opts)
	case StrategicView:
		g.generateStrategicView(content, data, quarters, opts)
	case ReleaseView:
		g.generateReleaseView(content, data, quarters, opts)
	case EpicGanttView:
		g.generateEpicGanttView(content, data, quarters, opts)
	default:
		// Default to timeline view
		g.generateTimelineView(content, data, quarters, opts)
	}

	// Add a legend
	g.addRoadmapLegend(content)
}

// parseTimeframe parses the timeframe string and returns start and end dates
func parseTimeframe(timeframe string) (time.Time, time.Time) {
	now := time.Now()
	startDate := time.Time{}
	endDate := time.Time{}

	// Handle empty timeframe
	if timeframe == "" {
		return startDate, endDate
	}

	// Handle relative timeframes like "6months", "1year", etc.
	if strings.HasSuffix(timeframe, "months") {
		months := 0
		fmt.Sscanf(timeframe, "%dmonths", &months)
		if months > 0 {
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			endDate = startDate.AddDate(0, months, 0)
			return startDate, endDate
		}
	}

	if strings.HasSuffix(timeframe, "year") || strings.HasSuffix(timeframe, "years") {
		years := 0
		fmt.Sscanf(timeframe, "%dyear", &years)
		if years == 0 {
			fmt.Sscanf(timeframe, "%dyears", &years)
		}
		if years > 0 {
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			endDate = startDate.AddDate(years, 0, 0)
			return startDate, endDate
		}
	}

	// Handle quarter ranges like "Q1-Q4 2025"
	if strings.Contains(timeframe, "-") && strings.Contains(timeframe, "Q") {
		var startQ, endQ, year int
		parts := strings.Split(timeframe, "-")
		if len(parts) == 2 {
			// Extract year from the second part
			yearStr := strings.TrimSpace(strings.TrimPrefix(parts[1], "Q1"))
			yearStr = strings.TrimSpace(strings.TrimPrefix(yearStr, "Q2"))
			yearStr = strings.TrimSpace(strings.TrimPrefix(yearStr, "Q3"))
			yearStr = strings.TrimSpace(strings.TrimPrefix(yearStr, "Q4"))
			fmt.Sscanf(yearStr, "%d", &year)

			// Extract quarters
			fmt.Sscanf(strings.TrimSpace(parts[0]), "Q%d", &startQ)
			fmt.Sscanf(strings.TrimSpace(parts[1]), "Q%d", &endQ)

			if startQ >= 1 && startQ <= 4 && endQ >= 1 && endQ <= 4 && year > 0 {
				startMonth := (startQ-1)*3 + 1
				endMonth := endQ * 3
				
				startDate = time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, now.Location())
				endDate = time.Date(year, time.Month(endMonth), 1, 0, 0, 0, 0, now.Location())
				endDate = endDate.AddDate(0, 1, 0) // End of the last month
				return startDate, endDate
			}
		}
	}

	// Default to current year if parsing fails
	return startDate, endDate
}

// generateQuarters generates a list of quarters between start and end dates
func generateQuarters(start, end time.Time) []string {
	var quarters []string
	
	current := start
	for current.Before(end) {
		quarter := fmt.Sprintf("Q%d %d", (int(current.Month())-1)/3+1, current.Year())
		
		// Check if this quarter is already in the list
		found := false
		for _, q := range quarters {
			if q == quarter {
				found = true
				break
			}
		}
		
		if !found {
			quarters = append(quarters, quarter)
		}
		
		// Move to next month
		current = current.AddDate(0, 1, 0)
	}
	
	return quarters
}

// addRoadmapLegend adds a legend for the roadmap
func (g *Generator) addRoadmapLegend(content *strings.Builder) {
	content.WriteString("<ac:structured-macro ac:name=\"info\">\n")
	content.WriteString("<ac:rich-text-body>\n")
	content.WriteString("<p><strong>Legend</strong></p>\n")
	content.WriteString("<ul>\n")
	content.WriteString("<li><span style=\"display:inline-block; width:20px; height:10px; background-color:#0052CC; margin-right:5px;\"></span>Planned</li>\n")
	content.WriteString("<li><span style=\"display:inline-block; width:20px; height:10px; background-color:#36B37E; margin-right:5px;\"></span>In Progress</li>\n")
	content.WriteString("<li><span style=\"display:inline-block; width:20px; height:10px; background-color:#FF8B00; margin-right:5px;\"></span>At Risk</li>\n")
	content.WriteString("<li><span style=\"display:inline-block; width:20px; height:10px; background-color:#FF5630; margin-right:5px;\"></span>Blocked</li>\n")
	content.WriteString("<li><span style=\"display:inline-block; width:20px; height:10px; background-color:#6554C0; margin-right:5px;\"></span>Completed</li>\n")
	content.WriteString("</ul>\n")
	content.WriteString("<p><small>This roadmap shows planned work items with their expected timeframes.</small></p>\n")
	content.WriteString("</ac:rich-text-body>\n")
	content.WriteString("</ac:structured-macro>\n")
}
