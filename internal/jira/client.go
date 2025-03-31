package jira

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	jiralib "github.com/andygrunwald/go-jira"
	"github.com/krzko/jiragitfluence/internal/config"
	"github.com/krzko/jiragitfluence/pkg/models"
)

// Client handles interactions with the Jira API
type Client struct {
	client *jiralib.Client
	logger *slog.Logger
}

// Custom transport for Bearer token authentication
type bearerAuthTransport struct {
	http.RoundTripper
	Token string
}

// RoundTrip implements the http.RoundTripper interface
func (t *bearerAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := cloneRequest(req) // Clone the request to avoid modifying the original
	req2.Header.Set("Authorization", "Bearer "+t.Token)
	return t.RoundTripper.RoundTrip(req2)
}

// Client returns an *http.Client with the transport that will add the bearer token to the request
func (t *bearerAuthTransport) Client() *http.Client {
	return &http.Client{Transport: t}
}

// cloneRequest creates a shallow copy of the request along with a deep copy of the Headers
func cloneRequest(req *http.Request) *http.Request {
	r2 := new(http.Request)
	*r2 = *req
	r2.Header = make(http.Header, len(req.Header))
	for k, s := range req.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}

// NewClient creates a new Jira client
func NewClient(cfg config.JiraConfig, logger *slog.Logger) (*Client, error) {
	// Log the Jira URL being used (without credentials)
	logger.Info("Creating Jira client", "url", cfg.URL, "auth_type", "Bearer Token")

	// Create a transport with Bearer token authentication
	tp := &bearerAuthTransport{
		RoundTripper: http.DefaultTransport,
		Token:        cfg.APIToken,
	}

	// Log token length for debugging (don't log the actual token)
	logger.Debug("Using PAT token for authentication", "token_length", len(cfg.APIToken))

	// Create the client with the Bearer auth transport
	client, err := jiralib.NewClient(tp.Client(), cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jira client: %w", err)
	}

	return &Client{
		client: client,
		logger: logger,
	}, nil
}

// FetchIssues fetches issues from Jira based on the provided projects and JQL
func (c *Client) FetchIssues(projects []string, jql string) ([]models.JiraIssue, error) {
	var issues []models.JiraIssue

	// Log the request with detailed information
	baseURL := c.client.GetBaseURL()
	c.logger.Info("Fetching issues from Jira", 
		"projects", projects, 
		"jql", jql, 
		"url", baseURL.String(),
		"auth_method", "Bearer Token")

	// Build JQL query
	query := buildJQLQuery(projects, jql)
	c.logger.Info("Fetching Jira issues", "jql", query)

	// Test authentication first with a simple API call
	c.logger.Debug("Testing Jira authentication")
	myself, _, err := c.client.User.GetSelf()
	if err != nil {
		c.logger.Error("Authentication test failed", "error", err)
		// Try a different endpoint to verify if it's an authentication issue
		c.logger.Debug("Attempting to access projects endpoint")
		_, _, projErr := c.client.Project.GetList()
		if projErr != nil {
			c.logger.Error("Projects endpoint access failed", "error", projErr)
			return nil, fmt.Errorf("authentication failed: %w (projects error: %v)", err, projErr)
		}
	} else {
		c.logger.Info("Authentication successful", "username", myself.DisplayName)
	}

	// Define constants for pagination
	const (
		maxResultsPerPage = 100 // Reduced from 1000 to avoid potential performance issues
		maxPages = 50          // Safety limit to prevent infinite loops
	)

	// Initialize search options
	options := &jiralib.SearchOptions{
		MaxResults: maxResultsPerPage,
		StartAt:    0,
		Fields:     []string{
			"summary", "status", "priority", "assignee", "reporter", "labels", 
			"created", "updated", "description", "fixVersions", "watches", "issuetype",
			"epic", "parent", "duedate", "customfield_10000", "customfield_10001", "customfield_10002", 
			"customfield_10003", "customfield_10004", "customfield_10005", "customfield_10006", 
			"customfield_10007", "customfield_10008", "customfield_10009", "customfield_10010",
		},
	}

	// Implement pagination
	totalFetched := 0
	page := 1

	for {
		c.logger.Debug("Executing Jira search", 
			"query", query, 
			"page", page, 
			"start_at", options.StartAt, 
			"max_results", options.MaxResults)

		// Execute search
		jiraIssues, resp, err := c.client.Issue.Search(query, options)
		if err != nil {
			statusCode := -1
			if resp != nil {
				statusCode = resp.StatusCode
			}
			c.logger.Error("Failed to search Jira issues", 
				"error", err, 
				"status_code", statusCode,
				"query", query,
				"page", page)
			return nil, fmt.Errorf("failed to search Jira issues (status: %d): %w", statusCode, err)
		}

		// Convert and append issues
		pageCount := len(jiraIssues)
		totalFetched += pageCount

		c.logger.Info("Fetched page of Jira issues", 
			"page", page, 
			"count", pageCount, 
			"total_so_far", totalFetched)

		// Convert Jira issues to our model
		for _, issue := range jiraIssues {
			jiraIssue := convertJiraIssue(issue)
			issues = append(issues, jiraIssue)
		}

		// Check if we've reached the end of results
		if pageCount < options.MaxResults {
			c.logger.Debug("Reached end of results", "total_fetched", totalFetched)
			break
		}

		// Safety check to prevent infinite loops
		if page >= maxPages {
			c.logger.Warn("Reached maximum page limit", 
				"max_pages", maxPages, 
				"total_fetched", totalFetched)
			break
		}

		// Move to next page
		options.StartAt += pageCount
		page++
	}

	c.logger.Info("Completed fetching Jira issues", 
		"total_count", len(issues), 
		"pages_fetched", page)

	return issues, nil
}

// buildJQLQuery constructs a JQL query from the provided projects and additional JQL
func buildJQLQuery(projects []string, additionalJQL string) string {
	var projectQuery string
	if len(projects) == 1 {
		projectQuery = fmt.Sprintf("project = \"%s\"", projects[0])
	} else {
		projectQuery = "project IN ("
		for i, project := range projects {
			if i > 0 {
				projectQuery += ", "
			}
			projectQuery += fmt.Sprintf("\"%s\"", project)
		}
		projectQuery += ")"
	}

	if additionalJQL != "" {
		return fmt.Sprintf("%s AND %s", projectQuery, additionalJQL)
	}
	return projectQuery
}

// convertJiraIssue converts a Jira issue to our model
func convertJiraIssue(issue jiralib.Issue) models.JiraIssue {
	jiraIssue := models.JiraIssue{
		Key:         issue.Key,
		Summary:     issue.Fields.Summary,
		Status:      issue.Fields.Status.Name,
		Description: issue.Fields.Description,
		URL:         fmt.Sprintf("%s/browse/%s", issue.Self[:len(issue.Self)-len(issue.Key)-len("/rest/api/2/issue/")], issue.Key),
	}

	// Set issue type
	if issue.Fields.Type.Name != "" {
		jiraIssue.IssueType = issue.Fields.Type.Name
	}

	// Set priority
	if issue.Fields.Priority != nil {
		jiraIssue.Priority = issue.Fields.Priority.Name
	}

	// Set assignee
	if issue.Fields.Assignee != nil {
		jiraIssue.Assignee = issue.Fields.Assignee.DisplayName
	}

	// Set reporter
	if issue.Fields.Reporter != nil {
		jiraIssue.Reporter = issue.Fields.Reporter.DisplayName
	}

	// Set labels
	jiraIssue.Labels = issue.Fields.Labels

	// Set created and updated dates
	jiraIssue.CreatedDate = time.Time(issue.Fields.Created)
	jiraIssue.UpdatedDate = time.Time(issue.Fields.Updated)

	// Set fix versions
	for _, version := range issue.Fields.FixVersions {
		jiraIssue.FixVersions = append(jiraIssue.FixVersions, version.Name)
	}

	// Set epic link if available
	// This is typically stored in a custom field, commonly customfield_10008
	if epicLink, ok := issue.Fields.Unknowns["customfield_10008"].(string); ok && epicLink != "" {
		jiraIssue.EpicLink = epicLink
	}
	epicLink, ok := issue.Fields.Unknowns["epic"]
	if ok && epicLink != nil {
		if epicStr, ok := epicLink.(string); ok {
			jiraIssue.EpicLink = epicStr
		}
	}

	// Set team if available
	team, ok := issue.Fields.Unknowns["team"]
	if ok && team != nil {
		if teamStr, ok := team.(string); ok {
			jiraIssue.Team = teamStr
		}
	}

	// Extract roadmap-related fields from custom fields
	
	// PlannedStartDate - typically stored in a custom field
	if startDateStr, ok := issue.Fields.Unknowns["customfield_10001"].(string); ok && startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			jiraIssue.PlannedStartDate = &startDate
		}
	}
	
	// PlannedEndDate - can be the due date or a custom field
	if !time.Time(issue.Fields.Duedate).IsZero() {
		dueDate := time.Time(issue.Fields.Duedate)
		jiraIssue.PlannedEndDate = &dueDate
	} else if endDateStr, ok := issue.Fields.Unknowns["customfield_10002"].(string); ok && endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			jiraIssue.PlannedEndDate = &endDate
		}
	}
	
	// Theme - typically stored in a custom field
	if theme, ok := issue.Fields.Unknowns["customfield_10003"].(string); ok {
		jiraIssue.Theme = theme
	}
	
	// Initiative - typically stored in a custom field
	if initiative, ok := issue.Fields.Unknowns["customfield_10004"].(string); ok {
		jiraIssue.Initiative = initiative
	}
	
	// Dependencies - can be extracted from links or a custom field
	if dependencies, ok := issue.Fields.Unknowns["customfield_10005"].([]interface{}); ok {
		for _, dep := range dependencies {
			if depStr, ok := dep.(string); ok {
				jiraIssue.Dependencies = append(jiraIssue.Dependencies, depStr)
			}
		}
	}
	
	// PriorityScore - typically stored in a custom field
	if priorityScore, ok := issue.Fields.Unknowns["customfield_10006"].(float64); ok {
		jiraIssue.PriorityScore = int(priorityScore)
	}
	
	// RoadmapStatus - can be derived from status or stored in a custom field
	if roadmapStatus, ok := issue.Fields.Unknowns["customfield_10007"].(string); ok {
		jiraIssue.RoadmapStatus = roadmapStatus
	} else {
		// Derive from status
		switch issue.Fields.Status.Name {
		case "To Do", "Open", "Backlog":
			jiraIssue.RoadmapStatus = "Planned"
		case "In Progress", "In Review":
			jiraIssue.RoadmapStatus = "In Progress"
		case "Blocked", "Impediment":
			jiraIssue.RoadmapStatus = "Blocked"
		case "Done", "Closed", "Resolved":
			jiraIssue.RoadmapStatus = "Completed"
		default:
			jiraIssue.RoadmapStatus = "Planned"
		}
	}
	
	// Milestone - typically stored in a custom field or can be a fix version
	if milestone, ok := issue.Fields.Unknowns["customfield_10009"].(string); ok {
		jiraIssue.Milestone = milestone
	} else if len(jiraIssue.FixVersions) > 0 {
		jiraIssue.Milestone = jiraIssue.FixVersions[0]
	}
	
	// Quarter - typically stored in a custom field
	if quarter, ok := issue.Fields.Unknowns["customfield_10010"].(string); ok {
		jiraIssue.Quarter = quarter
	} else if jiraIssue.PlannedStartDate != nil {
		// Derive quarter from planned start date
		quarterNum := (int(jiraIssue.PlannedStartDate.Month())-1)/3 + 1
		jiraIssue.Quarter = fmt.Sprintf("Q%d %d", quarterNum, jiraIssue.PlannedStartDate.Year())
	}

	return jiraIssue
}
