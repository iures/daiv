package plugin

import (
	"os"

	"github.com/iures/daiv-github-activity/plugin/contexts" // Import contexts package
	plug "github.com/iures/daivplug"
	"github.com/iures/daivplug/types"
)

// GitHubPlugin implements the Plugin interface
type GitHubPlugin struct{
	// Configuration
	token    string
	username string
}

// New creates a new instance of the plugin
func New() *GitHubPlugin {
	return &GitHubPlugin{}
}

// Name returns the unique identifier for this plugin
func (p *GitHubPlugin) Name() string {
	return "daiv-github-activity"
}

// Manifest returns the plugin manifest
func (p *GitHubPlugin) Manifest() *plug.PluginManifest {
	return &plug.PluginManifest{
		ConfigKeys: []plug.ConfigKey{
			{
				Type:        0, // ConfigTypeString
				Key:         "daiv-github-activity.token",
				Name:        "GitHub API Token",
				Description: "Personal access token with repo scope",
				Required:    true,
				Secret:      true,
				EnvVar:      "GITHUB_TOKEN",
			},
			{
				Type:        0, // ConfigTypeString
				Key:         "daiv-github-activity.username",
				Name:        "GitHub Username",
				Description: "Your GitHub username",
				Required:    true,
				EnvVar:      "GITHUB_USERNAME",
			},
		},
	}
}

// Initialize sets up the plugin with its configuration
func (p *GitHubPlugin) Initialize(settings map[string]interface{}) error {
	// Process configuration settings
	if token, ok := settings["daiv-github-activity.token"].(string); ok {
		p.token = token
		// Also set environment variable for contexts to use
		os.Setenv("GITHUB_TOKEN", token)
	}
	
	if username, ok := settings["daiv-github-activity.username"].(string); ok {
		p.username = username
		// Also set environment variable for contexts to use
		os.Setenv("GITHUB_USERNAME", username)
	}
	
	return nil
}

// Shutdown performs cleanup when the plugin is being disabled/removed
func (p *GitHubPlugin) Shutdown() error {
	// Clean up any resources
	return nil
}

// GetStandupContext implements the StandupPlugin interface
func (p *GitHubPlugin) GetStandupContext(timeRange types.TimeRange) (types.StandupContext, error) {
	// Delegate to the standup context implementation
	return contexts.GetStandupContext(p.Name(), timeRange)
} 
