# JiraGitFluence Usage Guide

JiraGitFluence is a command-line tool that aggregates data from Jira and GitHub, generates structured reports, and publishes them to Confluence. This document provides detailed usage instructions and examples.

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Commands](#commands)
  - [fetch](#fetch)
  - [generate](#generate)
  - [publish](#publish)
- [Use Cases](#use-cases)
- [Examples](#examples)

## Installation

```bash
# Clone the repository
git clone https://github.com/krzko/jiragitfluence.git
cd jiragitfluence

# Install dependencies
make install

# Build the application
make build
```

## Configuration

JiraGitFluence requires configuration for API access to Jira, GitHub, and Confluence. You can configure it using:

1. A YAML configuration file (default: `config.yaml`)
2. Environment variables

### Configuration File

Create a `config.yaml` file in the root directory (see `config.example.yaml` for reference):

```yaml
jira:
  url: "https://your-company.atlassian.net"
  username: "your-email@example.com"
  api_token: "your-jira-api-token"

github:
  token: "your-github-personal-access-token"

confluence:
  url: "https://your-company.atlassian.net/wiki"
  username: "your-email@example.com"
  api_token: "your-confluence-api-token"
```

### Environment Variables

Alternatively, you can use environment variables:

```bash
# Jira
export JIRA_URL="https://your-company.atlassian.net"
export JIRA_USERNAME="your-email@example.com"
export JIRA_API_TOKEN="your-jira-api-token"

# GitHub
export GITHUB_TOKEN="your-github-personal-access-token"

# Confluence
export CONFLUENCE_URL="https://your-company.atlassian.net/wiki"
export CONFLUENCE_USERNAME="your-email@example.com"
export CONFLUENCE_API_TOKEN="your-confluence-api-token"
```

## Commands

JiraGitFluence operates in three main steps:

1. Retrieve data from Jira and GitHub (using one of the fetch commands)
2. `generate`: Transform data into a structured format
3. `publish`: Upload the generated content to Confluence

The tool provides five main commands:

### fetch

The `fetch` command retrieves issues from Jira and issues/PRs from GitHub, then saves them to a JSON file.

```
jiragitfluence fetch [options]
```

#### Options

| Flag | Alias | Description | Required | Default |
|------|-------|-------------|----------|---------|
| `--jira-projects` | `-j` | Jira projects to query (e.g., 'Foo', 'Bar') | Yes | - |
| `--jira-jql` | `-q` | Advanced filtering in Jira using JQL | No | - |
| `--github-repos` | `-g` | GitHub repositories to scan (e.g., 'foo/qax-infra') | Yes | - |
| `--github-labels` | `-l` | Only fetch GitHub issues/PRs with these labels (comma-separated for multiple labels) | No | - |
| `--github-content-filter` | `-f` | Filter GitHub issues/PRs by text content in titles and descriptions | No | - |
| `--github-creator` | `-u` | Filter GitHub issues/PRs by creator username | No | - |
| `--output` | `-o` | Path to save the raw aggregated data | No | `aggregated_data.json` |
| `--config` | `-c` | Path to config file | No | `config.yaml` |
| `--verbose` | `-v` | Enable verbose logging | No | `false` |

### generate

The `generate` command transforms the fetched data into a structured format suitable for Confluence. It can read data from either a combined file (created by `fetch`) or separate files (created by `fetch-jira` and `fetch-github`).

```
jiragitfluence generate [options]
```

#### Examples

```bash
# Generate from a combined data file
jiragitfluence generate --input "aggregated_data.json" --format "table"

# Generate from separate Jira and GitHub data files
jiragitfluence generate --jira-input "jira_data.json" --github-input "github_data.json" --format "kanban"

# Generate from Jira data only
jiragitfluence generate --jira-input "jira_data.json" --format "table"

# Generate from GitHub data only
jiragitfluence generate --github-input "github_data.json" --format "table"
```

#### Options

| Flag | Alias | Description | Required | Default |
|------|-------|-------------|----------|---------|
| `--input` | `-i` | Input file created by the fetch command (combined data) | No | - |
| `--jira-input` | `-ji` | Input file created by the fetch-jira command | No | - |
| `--github-input` | `-gi` | Input file created by the fetch-github command | No | - |
| `--format` | `-f` | Presentation format (table, kanban, custom) | No | `table` |
| `--output` | `-o` | Path to save the generated Confluence markup | No | `confluence_output.html` |
| `--group-by` | `-g` | How to cluster or group issues (status, assignee, label) | No | `status` |
| `--include-metadata` | `-m` | Include metadata like creation timestamps | No | `false` |
| `--version-label` | `-v` | Tag to embed in the final content | No | - |
| `--verbose` | `-v` | Enable verbose logging | No | `false` |

### publish

The `publish` command uploads the generated content to Confluence.

```
jiragitfluence publish [options]
```

#### Options

| Flag | Alias | Description | Required | Default |
|------|-------|-------------|----------|---------|
| `--space` | `-s` | Confluence space key | Yes | - |
| `--title` | `-t` | Page title | Yes | - |
| `--parent` | `-p` | Parent page ID or title | No | - |
| `--content-file` | `-c` | Generated file from the generate command | Yes | - |
| `--version-comment` | `-v` | Comment for Confluence's version control | No | - |
| `--archive-old-versions` | `-a` | Automatically archive older versions | No | `false` |
| `--config` | `-c` | Path to config file | No | `config.yaml` |
| `--verbose` | `-v` | Enable verbose logging | No | `false` |

## Use Cases

### 1. Weekly Project Status Report

Generate a weekly status report for upper management showing all active work across multiple projects and repositories.

```bash
# Fetch data from multiple Jira projects and GitHub repositories
jiragitfluence fetch \
  --jira-projects "Foo,Bar" \
  --jira-jql "status != Done AND updated >= -7d" \
  --github-repos "foo/qax-infra,foo/qax-data"" \
  --github-labels "approved,pitch" \
  --github-content-filter "qax" \
  --output "weekly_data.json"

# Generate a table format report grouped by status
jiragitfluence generate \
  --input "weekly_data.json" \
  --format "table" \
  --group-by "status" \
  --include-metadata \
  --version-label "Week $(date +%V), $(date +%Y)" \
  --output "weekly_report.html"

# Publish to Confluence
jiragitfluence publish \
  --space "PROJ" \
  --title "Weekly Status Report - $(date +%Y-%m-%d)" \
  --parent "Team Dashboard" \
  --content-file "weekly_report.html" \
  --version-comment "Weekly update $(date +%Y-%m-%d)"
```

### 2. Separate Data Collection and Combined Reporting

Collect Jira and GitHub data separately and then combine them for reporting:

```bash
# Fetch Jira data with specific JQL
jiragitfluence fetch-jira \
  --jira-projects "Foo,Bar" \
  --jira-jql "status != Done AND updated >= -14d" \
  --output "jira_sprint_data.json"

# Fetch GitHub data with specific filters
jiragitfluence fetch-github \
  --github-repos "foo/qax-infra,foo/qax-data"" \
  --github-labels "enhancement,bug" \
  --github-creator "krzko" \
  --github-content-filter "api" \
  --output "github_api_issues.json"

# Generate a combined report from both data sources
jiragitfluence generate \
  --jira-input "jira_sprint_data.json" \
  --github-input "github_api_issues.json" \
  --format "kanban" \
  --group-by "status" \
  --include-metadata \
  --version-label "Sprint Report $(date +%Y-%m-%d)" \
  --output "combined_sprint_report.html"

# Publish to Confluence
jiragitfluence publish \
  --space "TEAM" \
  --title "Combined Sprint Report - $(date +%Y-%m-%d)" \
  --parent "Sprint Dashboard" \
  --content-file "combined_sprint_report.html" \
  --version-comment "Sprint update with API focus"
```

This approach allows you to collect data with different filters and criteria, then combine them into a single report.

### 3. Team Sprint Board

Create a Kanban-style board showing the current sprint's work for a specific team.

```bash
# Fetch data for a specific team's sprint
jiragitfluence fetch \
  --jira-projects "Bar" \
  --jira-jql "sprint in openSprints() AND team = 'Platform'" \
  --github-repos "foo/qax-platform" \
  --output "sprint_data.json"

# Generate a kanban format report
jiragitfluence generate \
  --input "sprint_data.json" \
  --format "kanban" \
  --group-by "assignee" \
  --output "sprint_board.html"

# Publish to Confluence
jiragitfluence publish \
  --space "TEAM" \
  --title "Platform Team - Current Sprint" \
  --content-file "sprint_board.html" \
  --archive-old-versions
```

### 3. Release Notes

Generate release notes for a specific version by aggregating all completed work.

```bash
# Fetch data for a specific version
jiragitfluence fetch \
  --jira-projects "Foo" \
  --jira-jql "fixVersion = '2.1.0' AND status = Done" \
  --github-repos "foo/qax-service" \
  --github-labels "release/2.1.0" \
  --output "release_data.json"

# Generate a custom format report
jiragitfluence generate \
  --input "release_data.json" \
  --format "custom" \
  --output "release_notes.html"

# Publish to Confluence
jiragitfluence publish \
  --space "DOCS" \
  --title "Release Notes v2.1.0" \
  --parent "Product Documentation" \
  --content-file "release_notes.html" \
  --version-comment "Initial release notes for v2.1.0"
```

## Examples

### Multiple GitHub Labels

You can specify multiple GitHub labels in several ways:

```bash
# Method 1: Comma-separated in a single flag
jiragitfluence fetch \
  --jira-projects "AR" \
  --github-repos "foo/qax" \
  --github-labels "approved,pitch,bug" \
  --output "filtered_data.json"

# Method 2: Multiple flag instances
jiragitfluence fetch \
  --jira-projects "AR" \
  --github-repos "foo/qax" \
  --github-labels "approved" --github-labels "pitch" --github-labels "bug" \
  --output "filtered_data.json"

# Method 3: Using the short alias
jiragitfluence fetch \
  -j "AR" \
  -g "foo/qax" \
  -l "approved" -l "pitch" -l "bug" \
  -o "filtered_data.json"
```

### Content Filtering

You can filter GitHub issues and pull requests by their content (title, body):

```bash
# Find all issues related to security
jiragitfluence fetch \
  --jira-projects "AR" \
  --github-repos "foo/qax" \
  --github-content-filter "security" \
  --output "security_issues.json"

# Combine content filtering with labels
jiragitfluence fetch \
  --jira-projects "AR" \
  --github-repos "foo/qax" \
  --github-labels "bug" \
  --github-content-filter "qax" \
  --output "qax_bugs.json"

# Using the short alias for content filtering
jiragitfluence fetch \
  -j "AR" \
  -g "foo/qax" \
  -f "performance" \
  -o "performance_issues.json"

# Advanced filtering with multiple labels and content filter
jiragitfluence fetch \
  --jira-projects "AR,Foo" \
  --github-repos "foo/qax,foo/qax-infra" \
  --github-labels "enhancement,feature" \
  --github-content-filter "api" \
  --output "api_features.json"
```

The content filter performs a case-insensitive search in both the title and body of issues and pull requests. This gives you more flexibility in finding relevant GitHub issues beyond just label filtering.

### Creator Filtering

You can filter GitHub issues and pull requests by their creator (the user who opened them):

```bash
# Find all issues created by a specific user
jiragitfluence fetch \
  --jira-projects "AR" \
  --github-repos "foo/qax" \
  --github-creator "krzko" \
  --output "foo_issues.json"

# Combine creator filtering with labels and content
jiragitfluence fetch \
  --jira-projects "AR" \
  --github-repos "foo/qax" \
  --github-labels "enhancement" \
  --github-creator "krzko" \
  --github-content-filter "api" \
  --output "krzko_api_enhancements.json"

# Using the short alias for creator filtering
jiragitfluence fetch \
  -j "AR" \
  -g "foo/qax" \
  -u "krzko" \
  -o "foo_issues.json"
```

The creator filter uses the GitHub username exactly as it appears on GitHub. This is useful for finding all issues or pull requests created by specific team members.

### fetch-jira

The `fetch-jira` command retrieves issues from Jira only, then saves them to a JSON file.

```
jiragitfluence fetch-jira [options]
```

#### Options

| Flag | Alias | Description | Required | Default |
|------|-------|-------------|----------|----------|
| `--jira-projects` | `-j` | Jira projects to query (e.g., 'Foo', 'Bar') | Yes | - |
| `--jira-jql` | `-q` | Advanced filtering in Jira using JQL | No | - |
| `--output` | `-o` | Path to save the raw aggregated data | No | `jira_data.json` |
| `--config` | `-c` | Path to config file | No | `config.yaml` |
| `--verbose` | `-v` | Enable verbose logging | No | `false` |

### fetch-github

The `fetch-github` command retrieves issues and PRs from GitHub only, then saves them to a JSON file.

```
jiragitfluence fetch-github [options]
```

#### Options

| Flag | Alias | Description | Required | Default |
|------|-------|-------------|----------|----------|
| `--github-repos` | `-g` | GitHub repositories to scan (e.g., 'foo/qax-infra') | Yes | - |
| `--github-labels` | `-l` | Only fetch GitHub issues/PRs with these labels (comma-separated for multiple labels) | No | - |
| `--github-content-filter` | `-f` | Filter GitHub issues/PRs by text content in titles and descriptions | No | - |
| `--github-creator` | `-u` | Filter GitHub issues/PRs by creator username | No | - |
| `--output` | `-o` | Path to save the raw aggregated data | No | `github_data.json` |
| `--config` | `-c` | Path to config file | No | `config.yaml` |
| `--verbose` | `-v` | Enable verbose logging | No | `false` |

### Basic Workflow

The most common workflow follows these steps:

```bash
# Step 1: Fetch data (choose one of the following approaches)

# Option A: Combined fetch from both Jira and GitHub
jiragitfluence fetch --jira-projects "Project1" --github-repos "org/repo1"

# Option B: Fetch from Jira only
jiragitfluence fetch-jira --jira-projects "Project1"

# Option C: Fetch from GitHub only
jiragitfluence fetch-github --github-repos "org/repo1"

# Step 2: Generate report (choose one of the following approaches)

# Option A: Using combined data from fetch command
jiragitfluence generate --input "aggregated_data.json" --format "table"

# Option B: Using separate data files from fetch-jira and fetch-github commands
jiragitfluence generate --jira-input "jira_data.json" --github-input "github_data.json" --format "table"

# Option C: Generate a roadmap view
jiragitfluence generate --input "aggregated_data.json" --format "roadmap" --roadmap-view "epicgantt" --roadmap-timeframe "6months" --output "roadmap_output.html"

# Step 3: Publish to Confluence
jiragitfluence publish --space "TEAM" --title "Status Report" --content-file "confluence_output.html"
```

### Roadmap Format

The roadmap format provides a visual representation of your project timeline, allowing you to see upcoming work organized in different ways. It's particularly useful for planning and tracking progress across multiple issues and epics.

```bash
# Generate a timeline roadmap view
jiragitfluence generate \
  --input "weekly_data.json" \
  --format "roadmap" \
  --roadmap-view "timeline" \
  --roadmap-timeframe "6months" \
  --output "roadmap_output.html"

# Generate an epic-based Gantt roadmap view
jiragitfluence generate \
  --input "weekly_data.json" \
  --format "roadmap" \
  --roadmap-view "epicgantt" \
  --roadmap-timeframe "Q1-Q4 2025" \
  --roadmap-include-dependencies \
  --output "roadmap_output.html"

# Generate a strategic roadmap view grouped by theme
jiragitfluence generate \
  --input "weekly_data.json" \
  --format "roadmap" \
  --roadmap-view "strategic" \
  --roadmap-grouping "theme" \
  --output "roadmap_output.html"
```

#### Roadmap Options

| Flag | Description | Default | Example Values |
|------|-------------|---------|----------------|
| `--roadmap-timeframe` | Timeframe for roadmap | `6months` | `Q1-Q4 2025`, `6months`, `1year` |
| `--roadmap-grouping` | How to group items | `theme` | `epic`, `theme`, `team`, `quarter` |
| `--roadmap-view` | Type of roadmap view | `timeline` | `timeline`, `strategic`, `release`, `epicgantt` |
| `--roadmap-include-dependencies` | Show dependencies between items | `false` | - |

#### Roadmap View Types

- **timeline**: Shows issues along a timeline with quarters or months
- **strategic**: Higher-level view focused on themes and initiatives
- **release**: Organizes items by planned release versions
- **epicgantt**: Gantt-style view organized by epics, showing issues in swimlanes

### Using Make Targets

The project includes make targets for convenience:

```bash
# Run fetch command
make fetch

# Run generate command
make generate

# Run publish command
make publish
```

For development with hot-reload:

```bash
# Run with hot-reload using air
make dev
```
