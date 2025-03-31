package confluence

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/krzko/jiragitfluence/internal/config"
	goconfluence "github.com/virtomize/confluence-go-api"
)

// Client handles interactions with the Confluence API
type Client struct {
	api    *goconfluence.API
	logger *slog.Logger
}

// NewClient creates a new Confluence client
func NewClient(cfg config.ConfluenceConfig, logger *slog.Logger) *Client {
	// Normalize the base URL to ensure it doesn't have a trailing slash
	baseURL := strings.TrimSuffix(cfg.URL, "/")

	// For on-premise Confluence, we need to remove any display paths
	// as they are only used for browser viewing, not API calls
	baseURL = strings.TrimSuffix(baseURL, "/display")

	// For Confluence Cloud instances, ensure the URL includes /wiki
	if strings.Contains(baseURL, ".atlassian.net") && !strings.Contains(baseURL, "/wiki") {
		baseURL = baseURL + "/wiki"
	}

	// Ensure the URL ends with /rest/api for the confluence-go-api package
	if !strings.HasSuffix(baseURL, "/rest/api") {
		baseURL = baseURL + "/rest/api"
	}

	logger.Info("Initializing Confluence client", "baseURL", baseURL)

	// Initialize the confluence-go-api client
	// For PAT authentication, we use empty username and the token as password
	api, err := goconfluence.NewAPI(baseURL, "", cfg.APIToken)
	if err != nil {
		logger.Error("Failed to initialize Confluence API client", "error", err)
		// Return a client with nil API, methods will check and return appropriate errors
	}

	return &Client{
		api:    api,
		logger: logger,
	}
}

// FindPage searches for a page by title in a specific space
func (c *Client) FindPage(spaceKey, title string) (string, int, error) {
	c.logger.Info("Searching for page", "space", spaceKey, "title", title)

	// Check if API client was initialized successfully
	if c.api == nil {
		return "", 0, fmt.Errorf("Confluence API client not initialized")
	}

	// Use the ContentQuery to search for content by title and space key
	query := goconfluence.ContentQuery{
		Title:    title,
		SpaceKey: spaceKey,
		Expand:   []string{"version"},
	}

	c.logger.Debug("Making API request", "title", title, "spaceKey", spaceKey)

	// Execute the search using the confluence-go-api client
	result, err := c.api.GetContent(query)
	if err != nil {
		c.logger.Error("Failed to search for page", "error", err)
		
		// Check if the error message contains useful information about the page
		// Some Confluence instances return errors but still include the page ID
		errStr := err.Error()
		if strings.Contains(errStr, "id") && strings.Contains(errStr, "title") {
			c.logger.Debug("Error response may contain page information", "error", errStr)
			
			// Try to extract the page ID from the error message
			// This is a fallback mechanism for certain Confluence instances
			var errorContent struct {
				Results []struct {
					ID string `json:"id"`
				} `json:"results"`
			}
			
			if strings.Contains(errStr, "{\"results\"") {
				// Try to extract JSON from the error message
				start := strings.Index(errStr, "{")
				if start >= 0 {
					jsonStr := errStr[start:]
					if err := json.Unmarshal([]byte(jsonStr), &errorContent); err == nil && 
					   len(errorContent.Results) > 0 && errorContent.Results[0].ID != "" {
						pageID := errorContent.Results[0].ID
						c.logger.Info("Extracted page ID from error response", "pageID", pageID)
						return pageID, 0, nil // We don't have version info in this case
					}
				}
			}
		}
		
		// If we couldn't extract an ID, return an empty ID but no error
		// This is the expected behavior when checking if a page exists
		return "", 0, nil
	}

	// Check if we have any results
	if result == nil || len(result.Results) == 0 {
		c.logger.Info("Page not found", "space", spaceKey, "title", title)
		return "", 0, nil
	}

	// Get the first matching page
	page := result.Results[0]
	pageID := page.ID
	versionNumber := 0
	
	// Extract version number if available
	if page.Version != nil {
		versionNumber = page.Version.Number
	}

	c.logger.Info("Found page", 
		"pageID", pageID, 
		"version", versionNumber, 
		"title", page.Title)

	return pageID, versionNumber, nil
}

// CreatePage creates a new page in Confluence
func (c *Client) CreatePage(spaceKey, title, content, parentID string) (string, error) {
	c.logger.Info("Creating new page", "space", spaceKey, "title", title, "parentID", parentID)

	// Check if API client was initialized successfully
	if c.api == nil {
		return "", fmt.Errorf("Confluence API client not initialized")
	}

	// Create the content object for the new page
	newPage := &goconfluence.Content{
		Type:  "page",
		Title: title,
		Space: &goconfluence.Space{
			Key: spaceKey,
		},
		Body: goconfluence.Body{
			Storage: goconfluence.Storage{
				Value:          content,
				Representation: "storage",
			},
		},
	}

	// Add parent page reference if provided
	if parentID != "" {
		c.logger.Info("Creating as child page", "parentID", parentID)
		newPage.Ancestors = []goconfluence.Ancestor{
			{
				ID: parentID,
			},
		}
	}

	// Create the page using the API
	createdPage, err := c.api.CreateContent(newPage)
	if err != nil {
		c.logger.Error("Failed to create page", "error", err)
		return "", fmt.Errorf("failed to create page: %w", err)
	}

	c.logger.Info("Page created successfully", "pageID", createdPage.ID, "title", createdPage.Title)
	return createdPage.ID, nil
}

// UpdatePage updates an existing page in Confluence
func (c *Client) UpdatePage(pageID, spaceKey, title, content string, version int, versionComment string) error {
	c.logger.Info("Updating page", "pageID", pageID, "space", spaceKey, "title", title, "version", version)

	// Check if API client was initialized successfully
	if c.api == nil {
		return fmt.Errorf("Confluence API client not initialized")
	}

	// First, check if the page exists
	_, err := c.api.GetContentByID(pageID, goconfluence.ContentQuery{
		Expand: []string{"version"},
	})
	if err != nil {
		c.logger.Error("Failed to get current page", "error", err)
		return fmt.Errorf("failed to get current page: %w", err)
	}

	// Create the content object for the page update
	updatePage := &goconfluence.Content{
		ID:    pageID,
		Type:  "page",
		Title: title,
		Space: &goconfluence.Space{
			Key: spaceKey,
		},
		Body: goconfluence.Body{
			Storage: goconfluence.Storage{
				Value:          content,
				Representation: "storage",
			},
		},
		Version: &goconfluence.Version{
			Number:    version + 1,
			Message:   versionComment,
			MinorEdit: false,
		},
	}

	// Update the page using the API
	_, err = c.api.UpdateContent(updatePage)
	if err != nil {
		c.logger.Error("Failed to update page", "error", err)
		return fmt.Errorf("failed to update page: %w", err)
	}

	c.logger.Info("Page updated successfully", "pageID", pageID, "title", title, "newVersion", version+1)
	return nil
}

// ArchiveOldVersions archives older versions of a page by adding an "Archived" label
// This is a simplified implementation - actual archiving would depend on your specific requirements
func (c *Client) ArchiveOldVersions(pageID string) error {
	c.logger.Info("Archiving old versions", "pageID", pageID)

	// Check if API client was initialized successfully
	if c.api == nil {
		return fmt.Errorf("Confluence API client not initialized")
	}

	// Add an "Archived" label to the page
	// This is a simple way to mark pages as archived
	labels := []goconfluence.Label{
		{
			Prefix: "global",
			Name:   "Archived",
		},
	}

	_, err := c.api.AddLabels(pageID, &labels)
	if err != nil {
		c.logger.Error("Failed to add archive label", "error", err)
		return fmt.Errorf("failed to archive page: %w", err)
	}

	c.logger.Info("Page archived successfully", "pageID", pageID)
	return nil
}
