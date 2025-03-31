package generator

import (
	"fmt"
	"time"

	"github.com/krzko/jiragitfluence/pkg/models"
)

// RoadmapItem represents an item on the roadmap (could be a Jira issue or GitHub issue)
type RoadmapItem struct {
	Type        string              // "jira" or "github-issue"
	JiraIssue   *models.JiraIssue   // Only set if Type is "jira"
	GitHubIssue *models.GitHubIssue // Only set if Type is "github-issue"
	Status      string              // Roadmap status (Planned, In Progress, At Risk, Blocked, Completed)
	StartDate   *time.Time          // Planned start date
	EndDate     *time.Time          // Planned end date
}

// Theme represents a strategic theme with initiatives
type Theme struct {
	Name        string       // Theme name
	Initiatives []Initiative // Initiatives under this theme
}

// Initiative represents a strategic initiative
type Initiative struct {
	Name         string // Initiative name
	Status       string // Status (Planned, In Progress, At Risk, Blocked, Completed)
	StartQuarter int    // Index of the start quarter
	EndQuarter   int    // Index of the end quarter
}

// Milestone represents a release or milestone
type Milestone struct {
	Name       string        // Milestone name
	TargetDate time.Time     // Target date for the milestone
	Status     string        // Status (Planned, In Progress, At Risk, Blocked, Completed)
	Items      []RoadmapItem // Items included in this milestone
}

// groupRoadmapItems groups roadmap items based on the grouping option
func groupRoadmapItems(data *models.AggregatedData, grouping string) map[string][]RoadmapItem {
	result := make(map[string][]RoadmapItem)
	
	// Convert data to RoadmapItems
	var allItems []RoadmapItem
	
	// Add Jira issues
	for _, issue := range data.JiraIssues {
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
		
		allItems = append(allItems, item)
	}
	
	// Add GitHub issues
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
		
		allItems = append(allItems, item)
	}
	
	// We're not including PRs in the roadmap as per requirements
	
	// Group items based on the grouping option
	switch grouping {
	case "epic":
		// Group by epic
		for _, item := range allItems {
			epicKey := "No Epic"
			
			if item.Type == "jira" && item.JiraIssue.EpicLink != "" {
				epicKey = item.JiraIssue.EpicLink
			}
			
			result[epicKey] = append(result[epicKey], item)
		}
		
	case "theme":
		// Group by theme
		for _, item := range allItems {
			theme := "No Theme"
			
			switch item.Type {
			case "jira":
				if item.JiraIssue.Theme != "" {
					theme = item.JiraIssue.Theme
				} else if item.JiraIssue.IssueType != "" {
					// Use issue type as theme if no theme is specified
					theme = item.JiraIssue.IssueType
				}
			case "github-issue":
				if item.GitHubIssue.Theme != "" {
					theme = item.GitHubIssue.Theme
				} else if len(item.GitHubIssue.Labels) > 0 {
					// Use first label as theme if no theme is specified
					theme = item.GitHubIssue.Labels[0]
				}
			// We're not including PRs in the roadmap as per requirements
			}
			
			result[theme] = append(result[theme], item)
		}
		
	case "team":
		// Group by team
		for _, item := range allItems {
			team := "No Team"
			
			switch item.Type {
			case "jira":
				if item.JiraIssue.Team != "" {
					team = item.JiraIssue.Team
				}
			}
			
			result[team] = append(result[team], item)
		}
		
	case "quarter":
		// Group by quarter
		for _, item := range allItems {
			quarter := "No Quarter"
			
			// Derive quarter from start date if not explicitly set
			if item.StartDate != nil {
				year := item.StartDate.Year()
				month := item.StartDate.Month()
				quarterNum := (int(month) - 1) / 3 + 1
				derived := fmt.Sprintf("Q%d %d", quarterNum, year)
				
				switch item.Type {
				case "jira":
					if item.JiraIssue.Quarter != "" {
						quarter = item.JiraIssue.Quarter
					} else {
						quarter = derived
					}
				case "github-issue":
					if item.GitHubIssue.Quarter != "" {
						quarter = item.GitHubIssue.Quarter
					} else {
						quarter = derived
					}
				}
			}
			
			result[quarter] = append(result[quarter], item)
		}
		
	default:
		// Default to no grouping, just put everything in one group
		result["All Items"] = allItems
	}
	
	return result
}

// extractThemes extracts themes and initiatives from the data
func extractThemes(data *models.AggregatedData) []Theme {
	// Map to collect initiatives by theme
	themeMap := make(map[string]map[string]Initiative)
	
	// Process Jira issues
	for _, issue := range data.JiraIssues {
		if issue.Theme != "" && issue.Initiative != "" && issue.PlannedStartDate != nil && issue.PlannedEndDate != nil {
			// Create theme entry if it doesn't exist
			if _, exists := themeMap[issue.Theme]; !exists {
				themeMap[issue.Theme] = make(map[string]Initiative)
			}
			
			// Create or update initiative
			initiative := Initiative{
				Name:   issue.Initiative,
				Status: issue.RoadmapStatus,
			}
			
			if initiative.Status == "" {
				initiative.Status = deriveRoadmapStatus(issue.Status)
			}
			
			themeMap[issue.Theme][issue.Initiative] = initiative
		}
	}
	
	// Process GitHub issues
	for _, issue := range data.GitHubIssues {
		if issue.Theme != "" && issue.Initiative != "" && issue.PlannedStartDate != nil && issue.PlannedEndDate != nil {
			// Create theme entry if it doesn't exist
			if _, exists := themeMap[issue.Theme]; !exists {
				themeMap[issue.Theme] = make(map[string]Initiative)
			}
			
			// Create or update initiative
			initiative := Initiative{
				Name:   issue.Initiative,
				Status: issue.RoadmapStatus,
			}
			
			if initiative.Status == "" {
				initiative.Status = deriveRoadmapStatusFromGitHub(issue.State)
			}
			
			themeMap[issue.Theme][issue.Initiative] = initiative
		}
	}
	
	// Convert map to slice
	var themes []Theme
	for themeName, initiatives := range themeMap {
		theme := Theme{
			Name: themeName,
		}
		
		for _, initiative := range initiatives {
			theme.Initiatives = append(theme.Initiatives, initiative)
		}
		
		themes = append(themes, theme)
	}
	
	return themes
}

// extractMilestones extracts milestones from the data
func extractMilestones(data *models.AggregatedData) []Milestone {
	// Map to collect items by milestone
	milestoneMap := make(map[string]Milestone)
	
	// Process Jira issues
	for _, issue := range data.JiraIssues {
		if issue.Milestone != "" && issue.PlannedEndDate != nil {
			// Create milestone entry if it doesn't exist
			if _, exists := milestoneMap[issue.Milestone]; !exists {
				milestoneMap[issue.Milestone] = Milestone{
					Name:       issue.Milestone,
					TargetDate: *issue.PlannedEndDate,
					Status:     "Planned",
				}
			}
			
			// Add issue to milestone
			milestone := milestoneMap[issue.Milestone]
			
			item := RoadmapItem{
				Type:      "jira",
				JiraIssue: &issue,
				Status:    issue.RoadmapStatus,
			}
			
			if item.Status == "" {
				item.Status = deriveRoadmapStatus(issue.Status)
			}
			
			milestone.Items = append(milestone.Items, item)
			
			// Update milestone status based on item statuses
			milestone.Status = aggregateStatus(milestone.Status, item.Status)
			
			milestoneMap[issue.Milestone] = milestone
		}
	}
	
	// Process GitHub issues
	for _, issue := range data.GitHubIssues {
		if issue.Milestone != "" && issue.PlannedEndDate != nil {
			// Create milestone entry if it doesn't exist
			if _, exists := milestoneMap[issue.Milestone]; !exists {
				milestoneMap[issue.Milestone] = Milestone{
					Name:       issue.Milestone,
					TargetDate: *issue.PlannedEndDate,
					Status:     "Planned",
				}
			}
			
			// Add issue to milestone
			milestone := milestoneMap[issue.Milestone]
			
			item := RoadmapItem{
				Type:        "github-issue",
				GitHubIssue: &issue,
				Status:      issue.RoadmapStatus,
			}
			
			if item.Status == "" {
				item.Status = deriveRoadmapStatusFromGitHub(issue.State)
			}
			
			milestone.Items = append(milestone.Items, item)
			
			// Update milestone status based on item statuses
			milestone.Status = aggregateStatus(milestone.Status, item.Status)
			
			milestoneMap[issue.Milestone] = milestone
		}
	}
	
	// Convert map to slice
	var milestones []Milestone
	for _, milestone := range milestoneMap {
		milestones = append(milestones, milestone)
	}
	
	return milestones
}

// getItemQuarters determines the start and end quarters for an item
func getItemQuarters(item RoadmapItem, quarters []string) (int, int) {
	if item.StartDate == nil || item.EndDate == nil {
		return 0, 0
	}
	
	startQuarter := -1
	endQuarter := -1
	
	// Extract year and quarter from item dates
	startYear := item.StartDate.Year()
	startQ := (int(item.StartDate.Month())-1)/3 + 1
	startQStr := "Q" + string(rune('0'+startQ)) + " " + string(rune('0'+startYear/1000)) + string(rune('0'+(startYear/100)%10)) + string(rune('0'+(startYear/10)%10)) + string(rune('0'+startYear%10))
	
	endYear := item.EndDate.Year()
	endQ := (int(item.EndDate.Month())-1)/3 + 1
	endQStr := "Q" + string(rune('0'+endQ)) + " " + string(rune('0'+endYear/1000)) + string(rune('0'+(endYear/100)%10)) + string(rune('0'+(endYear/10)%10)) + string(rune('0'+endYear%10))
	
	// Find matching quarters in the quarters slice
	for i, q := range quarters {
		if q == startQStr && startQuarter == -1 {
			startQuarter = i
		}
		if q == endQStr && endQuarter == -1 {
			endQuarter = i
		}
	}
	
	// If not found, use approximation
	if startQuarter == -1 {
		startQuarter = 0
	}
	if endQuarter == -1 {
		endQuarter = len(quarters) - 1
	}
	
	return startQuarter, endQuarter
}

// deriveRoadmapStatus derives a roadmap status from a Jira issue status
func deriveRoadmapStatus(status string) string {
	switch status {
	case "To Do", "Open", "Backlog":
		return "Planned"
	case "In Progress", "In Review":
		return "In Progress"
	case "Blocked", "Impediment":
		return "Blocked"
	case "Done", "Closed", "Resolved":
		return "Completed"
	default:
		return "Planned"
	}
}

// deriveRoadmapStatusFromGitHub derives a roadmap status from a GitHub issue/PR state
func deriveRoadmapStatusFromGitHub(state string) string {
	switch state {
	case "open":
		return "In Progress"
	case "closed":
		return "Completed"
	default:
		return "Planned"
	}
}

// getRoadmapStatusColor returns a color for a roadmap status
func getRoadmapStatusColor(status string) string {
	switch status {
	case "Planned":
		return "#0052CC" // Blue
	case "In Progress":
		return "#36B37E" // Green
	case "At Risk":
		return "#FF8B00" // Orange
	case "Blocked":
		return "#FF5630" // Red
	case "Completed":
		return "#6554C0" // Purple
	default:
		return "#0052CC" // Default blue
	}
}

// aggregateStatus determines the aggregate status based on multiple item statuses
func aggregateStatus(currentStatus, newStatus string) string {
	// Priority order: Blocked > At Risk > In Progress > Planned > Completed
	statusPriority := map[string]int{
		"Blocked":    5,
		"At Risk":    4,
		"In Progress": 3,
		"Planned":    2,
		"Completed":  1,
	}
	
	currentPriority := statusPriority[currentStatus]
	newPriority := statusPriority[newStatus]
	
	if newPriority > currentPriority {
		return newStatus
	}
	
	return currentStatus
}

// generateQuartersFromNow generates a list of quarter strings starting from the current quarter
func generateQuartersFromNow(count int) []string {
	now := time.Now()
	currentYear := now.Year()
	currentQuarter := (int(now.Month()) - 1) / 3 + 1
	
	quarters := make([]string, count)
	for i := 0; i < count; i++ {
		quarterNum := (currentQuarter + i - 1) % 4 + 1
		yearOffset := (currentQuarter + i - 1) / 4
		year := currentYear + yearOffset
		quarters[i] = fmt.Sprintf("Q%d %d", quarterNum, year)
	}
	
	return quarters
}
