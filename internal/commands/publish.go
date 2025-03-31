package commands

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/krzko/jiragitfluence/internal/config"
	"github.com/krzko/jiragitfluence/internal/confluence"
	"github.com/urfave/cli/v2"
)

// PublishCommand handles the publish command
func PublishCommand(ctx *cli.Context) error {
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
	spaceKey := ctx.String("space")
	title := ctx.String("title")
	parentTitle := ctx.String("parent")
	contentFilePath := ctx.String("content-file")
	versionComment := ctx.String("version-comment")
	archiveOldVersions := ctx.Bool("archive-old-versions")

	logger.Info("Starting publish operation",
		"space", spaceKey,
		"title", title,
		"contentFile", contentFilePath)

	// Read content file
	content, err := os.ReadFile(contentFilePath)
	if err != nil {
		return fmt.Errorf("failed to read content file: %w", err)
	}

	// Create Confluence client
	confluenceClient := confluence.NewClient(cfg.Confluence, logger)

	// Parent page is required for publishing
	if parentTitle == "" {
		return fmt.Errorf("parent page title is required for publishing")
	}

	// Find the parent page ID by title
	logger.Info("Finding parent page", "space", spaceKey, "parentTitle", parentTitle)
	parentID, _, err := confluenceClient.FindPage(spaceKey, parentTitle)
	// Only return an error if we got an error AND no parentID
	if err != nil && parentID == "" {
		return fmt.Errorf("failed to find parent page: %w", err)
	}
	if parentID == "" {
		return fmt.Errorf("parent page with title '%s' not found in space '%s'", parentTitle, spaceKey)
	}
	logger.Info("Found parent page", "parentID", parentID, "parentTitle", parentTitle)

	// Check if the target page exists
	pageID, version, err := confluenceClient.FindPage(spaceKey, title)
	// Only return an error if we got an error AND no pageID
	if err != nil && pageID == "" {
		return fmt.Errorf("failed to check if page exists: %w", err)
	}

	var newPageID string

	// Update or create page
	if pageID != "" {
		logger.Info("Updating existing page", "pageID", pageID, "version", version)
		
		// Archive old versions if requested
		if archiveOldVersions {
			if err := confluenceClient.ArchiveOldVersions(pageID); err != nil {
				logger.Warn("Failed to archive old versions", "error", err)
				// Continue anyway
			}
		}
		
		// Update the page
		if err := confluenceClient.UpdatePage(pageID, spaceKey, title, string(content), version, versionComment); err != nil {
			return fmt.Errorf("failed to update page: %w", err)
		}
		
		newPageID = pageID
	} else {
		logger.Info("Creating new page", "space", spaceKey, "title", title)
		
		// Create the page
		var err error
		newPageID, err = confluenceClient.CreatePage(spaceKey, title, string(content), parentID)
		if err != nil {
			return fmt.Errorf("failed to create page: %w", err)
		}
	}

	logger.Info("Successfully published to Confluence", "pageID", newPageID)
	return nil
}
