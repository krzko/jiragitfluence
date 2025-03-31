package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/krzko/jiragitfluence/internal/commands"
	"github.com/urfave/cli/v2"
)

// Build information. Populated at build-time using -ldflags:
//
//	go build -ldflags "-X main.version=1.0.0 -X main.commit=abc123 -X main.date=2025-04-01"
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Run executes the CLI application
func Run() error {
	// Set custom version printer that includes commit and build date
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("jiragitfluence version: %s commit: %s built: %s\n", version, commit, date)
	}

	app := &cli.App{
		Name:        "jiragitfluence",
		Usage:       "Sync data between Jira, GitHub, and Confluence",
		Version:     version,
		Compiled:    time.Now(),
		Description: "A tool to aggregate data from Jira and GitHub and publish it to Confluence",
		Authors: []*cli.Author{
			{
				Name:  "Kristof Kowalski",
				Email: "k@ko.wal.ski",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "fetch",
				Usage: "Fetch data from both Jira and GitHub (combined operation)",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:     "jira-projects",
						Aliases:  []string{"j"},
						Usage:    "Jira projects to query (e.g., 'Foo', 'Bar')",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "jira-jql",
						Aliases: []string{"q"},
						Usage:   "Advanced filtering in Jira using JQL",
					},
					&cli.StringSliceFlag{
						Name:     "github-repos",
						Aliases:  []string{"g"},
						Usage:    "GitHub repositories to scan (e.g., 'foo/qax-infra')",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:    "github-labels",
						Aliases: []string{"l"},
						Usage:   "Only fetch GitHub issues/PRs with these labels",
					},
					&cli.StringFlag{
						Name:    "github-content-filter",
						Aliases: []string{"f"},
						Usage:   "Filter GitHub issues/PRs by text content",
					},
					&cli.StringFlag{
						Name:    "github-creator",
						Aliases: []string{"u"},
						Usage:   "Filter GitHub issues/PRs by creator username",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Path to save the raw aggregated data",
						Value:   "aggregated_data.json",
					},
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Path to config file",
						Value:   "config.yaml",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose logging",
					},
				},
				Action: commands.FetchCommand,
			},
			{
				Name:  "fetch-jira",
				Usage: "Fetch data from Jira only",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:     "jira-projects",
						Aliases:  []string{"j"},
						Usage:    "Jira projects to query (e.g., 'Foo', 'Bar')",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "jira-jql",
						Aliases: []string{"q"},
						Usage:   "Advanced filtering in Jira using JQL",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Path to save the raw aggregated data",
						Value:   "jira_data.json",
					},
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Path to config file",
						Value:   "config.yaml",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose logging",
					},
				},
				Action: commands.FetchJiraCommand,
			},
			{
				Name:  "fetch-github",
				Usage: "Fetch data from GitHub only",
				Flags: []cli.Flag{
					&cli.StringSliceFlag{
						Name:     "github-repos",
						Aliases:  []string{"g"},
						Usage:    "GitHub repositories to scan (e.g., 'foo/qax-infra')",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:    "github-labels",
						Aliases: []string{"l"},
						Usage:   "Only fetch GitHub issues/PRs with these labels",
					},
					&cli.StringFlag{
						Name:    "github-content-filter",
						Aliases: []string{"f"},
						Usage:   "Filter GitHub issues/PRs by text content",
					},
					&cli.StringFlag{
						Name:    "github-creator",
						Aliases: []string{"u"},
						Usage:   "Filter GitHub issues/PRs by creator username",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Path to save the raw aggregated data",
						Value:   "github_data.json",
					},
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Path to config file",
						Value:   "config.yaml",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose logging",
					},
				},
				Action: commands.FetchGitHubCommand,
			},
			{
				Name:  "generate",
				Usage: "Generate a structured representation for Confluence",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "input",
						Aliases:  []string{"i"},
						Usage:    "Input file created by the fetch command (combined data)",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "jira-input",
						Aliases:  []string{"ji"},
						Usage:    "Input file created by the fetch-jira command",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "github-input",
						Aliases:  []string{"gi"},
						Usage:    "Input file created by the fetch-github command",
						Required: false,
					},
					&cli.StringFlag{
						Name:    "format",
						Aliases: []string{"f"},
						Usage:   "Presentation format (table, kanban, gantt, roadmap, custom)",
						Value:   "table",
					},
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Usage:   "Path to save the generated Confluence markup",
						Value:   "confluence_output.html",
					},
					&cli.StringFlag{
						Name:    "group-by",
						Aliases: []string{"g"},
						Usage:   "How to cluster or group issues (status, assignee, label)",
						Value:   "status",
					},
					&cli.BoolFlag{
						Name:    "include-metadata",
						Aliases: []string{"m"},
						Usage:   "Include metadata like creation timestamps",
					},
					&cli.StringFlag{
						Name:    "version-label",
						Aliases: []string{"vl"},
						Usage:   "Tag to embed in the final content",
					},
					// Roadmap specific options
					&cli.StringFlag{
						Name:  "roadmap-timeframe",
						Usage: "Timeframe for roadmap (e.g., 'Q1-Q4 2025', '6months', '1year')",
						Value: "6months",
					},
					&cli.StringFlag{
						Name:  "roadmap-grouping",
						Usage: "How to group items in roadmap (epic, theme, team, quarter)",
						Value: "theme",
					},
					&cli.StringFlag{
						Name:  "roadmap-view",
						Usage: "Type of roadmap view (timeline, strategic, release, epicgantt)",
						Value: "timeline",
					},
					&cli.BoolFlag{
						Name:  "roadmap-include-dependencies",
						Usage: "Whether to show dependencies between roadmap items",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose logging",
					},
				},
				Action: commands.GenerateCommand,
			},
			{
				Name:  "publish",
				Usage: "Publish generated content to Confluence",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "space",
						Aliases:  []string{"s"},
						Usage:    "Confluence space key",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "title",
						Aliases:  []string{"t"},
						Usage:    "Page title",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "parent",
						Aliases:  []string{"p"},
						Usage:    "Parent page title (required for creating pages)",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "content-file",
						Aliases:  []string{"c"},
						Usage:    "Generated file from the generate command",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "version-comment",
						Aliases: []string{"vc"},
						Usage:   "Comment for Confluence's version control",
					},
					&cli.BoolFlag{
						Name:    "archive-old-versions",
						Aliases: []string{"a"},
						Usage:   "Automatically archive older versions",
					},
					&cli.StringFlag{
						Name:  "config",
						Usage: "Path to config file",
						Value: "config.yaml",
					},
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Enable verbose logging",
					},
				},
				Action: commands.PublishCommand,
			},
		},
	}

	return app.Run(os.Args)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := Run(); err != nil {
		slog.Error("application error", "error", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
