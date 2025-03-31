package models

import "time"

// AggregatedData represents the combined data from Jira and GitHub
type AggregatedData struct {
	JiraIssues  []JiraIssue   `json:"jiraIssues"`
	GitHubIssues []GitHubIssue `json:"githubIssues"`
	GitHubPRs    []GitHubPR    `json:"githubPRs"`
	Metadata     Metadata      `json:"metadata"`
}

// JiraIssue represents a Jira issue
type JiraIssue struct {
	Key              string       `json:"key"`
	IssueType        string       `json:"issueType"`
	Summary          string       `json:"summary"`
	Status           string       `json:"status"`
	Priority         string       `json:"priority"`
	Assignee         string       `json:"assignee"`
	Reporter         string       `json:"reporter"`
	Team             string       `json:"team"`
	Labels           []string     `json:"labels"`
	EpicLink         string       `json:"epicLink"`
	CreatedDate      time.Time    `json:"createdDate"`
	UpdatedDate      time.Time    `json:"updatedDate"`
	Description      string       `json:"description"`
	FixVersions      []string     `json:"fixVersions"`
	Watchers         []string     `json:"watchers"`
	URL              string       `json:"url"`
	// Roadmap planning fields
	PlannedStartDate *time.Time   `json:"plannedStartDate,omitempty"`
	PlannedEndDate   *time.Time   `json:"plannedEndDate,omitempty"`
	Theme            string       `json:"theme,omitempty"`
	Initiative       string       `json:"initiative,omitempty"`
	Dependencies     []string     `json:"dependencies,omitempty"` // List of issue keys this issue depends on
	PriorityScore    int          `json:"priorityScore,omitempty"` // Numeric priority (1-100)
	RoadmapStatus    string       `json:"roadmapStatus,omitempty"` // Planning status (e.g., "Planned", "In Progress", "Completed")
	Milestone        string       `json:"milestone,omitempty"`
	Quarter          string       `json:"quarter,omitempty"` // Which quarter this is planned for (e.g., "Q1 2025")
}

// GitHubIssue represents a GitHub issue
type GitHubIssue struct {
	Title            string       `json:"title"`
	Number           int          `json:"number"`
	State            string       `json:"state"`
	Labels           []string     `json:"labels"`
	Assignees        []string     `json:"assignees"`
	CreatedDate      time.Time    `json:"createdDate"`
	UpdatedDate      time.Time    `json:"updatedDate"`
	URL              string       `json:"url"`
	Repository       string       `json:"repository"`
	// Roadmap planning fields
	PlannedStartDate *time.Time   `json:"plannedStartDate,omitempty"`
	PlannedEndDate   *time.Time   `json:"plannedEndDate,omitempty"`
	Theme            string       `json:"theme,omitempty"`
	Initiative       string       `json:"initiative,omitempty"`
	Dependencies     []string     `json:"dependencies,omitempty"` // List of issue references this issue depends on
	PriorityScore    int          `json:"priorityScore,omitempty"` // Numeric priority (1-100)
	RoadmapStatus    string       `json:"roadmapStatus,omitempty"` // Planning status
	Milestone        string       `json:"milestone,omitempty"`
	Quarter          string       `json:"quarter,omitempty"` // Which quarter this is planned for
}

// GitHubPR represents a GitHub pull request
type GitHubPR struct {
	Title       string    `json:"title"`
	Number      int       `json:"number"`
	State       string    `json:"state"`
	Labels      []string  `json:"labels"`
	Assignees   []string  `json:"assignees"`
	CreatedDate time.Time `json:"createdDate"`
	UpdatedDate time.Time `json:"updatedDate"`
	URL         string    `json:"url"`
	Repository  string    `json:"repository"`
	IsDraft     bool      `json:"isDraft"`
	MergeStatus string    `json:"mergeStatus"`
}

// Metadata contains information about the data collection
type Metadata struct {
	FetchTime          time.Time `json:"fetchTime"`
	JiraProjects       []string  `json:"jiraProjects"`
	GitHubRepos        []string  `json:"githubRepos"`
	JiraJQL            string    `json:"jiraJql,omitempty"`
	GitHubLabels       []string  `json:"githubLabels,omitempty"`
	GitHubContentFilter string    `json:"githubContentFilter,omitempty"`
	GitHubCreator      string    `json:"githubCreator,omitempty"`
	VersionLabel       string    `json:"versionLabel,omitempty"`
	// Roadmap planning metadata
	RoadmapTimeframe   string    `json:"roadmapTimeframe,omitempty"`
	RoadmapGrouping    string    `json:"roadmapGrouping,omitempty"`
	RoadmapView        string    `json:"roadmapView,omitempty"`
	RoadmapThemes      []string  `json:"roadmapThemes,omitempty"`
	RoadmapQuarters    []string  `json:"roadmapQuarters,omitempty"`
}
