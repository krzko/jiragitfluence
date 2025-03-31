package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Jira       JiraConfig       `yaml:"jira"`
	GitHub     GitHubConfig     `yaml:"github"`
	Confluence ConfluenceConfig `yaml:"confluence"`
}

// JiraConfig holds Jira API configuration
type JiraConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	APIToken string `yaml:"api_token"`
}

// GitHubConfig holds GitHub API configuration
type GitHubConfig struct {
	Token string `yaml:"token"`
}

// ConfluenceConfig holds Confluence API configuration
type ConfluenceConfig struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	APIToken string `yaml:"api_token"`
}

// LoadConfig loads configuration from a YAML file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Default config
	config := &Config{}

	// Load from file if it exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("error parsing config file: %w", err)
		}
	}

	// Override with environment variables
	overrideFromEnv(config)

	return config, nil
}

// overrideFromEnv overrides config values with environment variables
func overrideFromEnv(config *Config) {
	// Jira config
	if val := os.Getenv("JIRA_URL"); val != "" {
		config.Jira.URL = val
	}
	if val := os.Getenv("JIRA_USERNAME"); val != "" {
		config.Jira.Username = val
	}
	if val := os.Getenv("JIRA_API_TOKEN"); val != "" {
		config.Jira.APIToken = val
	}

	// GitHub config
	if val := os.Getenv("GITHUB_TOKEN"); val != "" {
		config.GitHub.Token = val
	}

	// Confluence config
	if val := os.Getenv("CONFLUENCE_URL"); val != "" {
		config.Confluence.URL = val
	}
	if val := os.Getenv("CONFLUENCE_USERNAME"); val != "" {
		config.Confluence.Username = val
	}
	if val := os.Getenv("CONFLUENCE_API_TOKEN"); val != "" {
		config.Confluence.APIToken = val
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	var missingFields []string

	// Validate Jira config
	if c.Jira.URL == "" {
		missingFields = append(missingFields, "jira.url")
	}
	if c.Jira.Username == "" {
		missingFields = append(missingFields, "jira.username")
	}
	if c.Jira.APIToken == "" {
		missingFields = append(missingFields, "jira.api_token")
	}

	// Validate GitHub config
	if c.GitHub.Token == "" {
		missingFields = append(missingFields, "github.token")
	}

	// Validate Confluence config
	if c.Confluence.URL == "" {
		missingFields = append(missingFields, "confluence.url")
	}
	if c.Confluence.Username == "" {
		missingFields = append(missingFields, "confluence.username")
	}
	if c.Confluence.APIToken == "" {
		missingFields = append(missingFields, "confluence.api_token")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing required configuration: %s", strings.Join(missingFields, ", "))
	}

	return nil
}
