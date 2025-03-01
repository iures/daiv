package github

import (
	"daiv/internal/plugin"
	"fmt"
	"os/exec"
	"strings"
)

type GitHubPlugin struct {
	client *GithubClient
}

func NewGitHubPlugin() *GitHubPlugin {
	return &GitHubPlugin{
		client: &GithubClient{},
	}
}

func (g *GitHubPlugin) Name() string {
	return "github"
}

func (g *GitHubPlugin) Manifest() *plugin.PluginManifest {
	return &plugin.PluginManifest{
		ConfigKeys: []plugin.ConfigKey{
			{
        Type:        ,
				Key:         "username",
				Name:        "GitHub Username",
				Description: "Your GitHub username",
				Required:    true,
			},
			{
				Type:        ,
				Key:         "organization",
				Name:        "GitHub Organization",
				Description: "The GitHub organization to monitor",
				Required:    true,
			},
			{
				Key:         "repositories",
				Name:        "GitHub Repositories",
				Description: "List of repositories to monitor",
				Required:    true,
			},
		},
	}
}

func (g *GitHubPlugin) Initialize(settings map[string]interface{}) error {
	token, err := getGhCliToken()
	if err != nil {
		return fmt.Errorf("failed to get gh cli token: %w", err)
	}

	g.client.Init(GithubClientSettings{
		Username: settings["username"].(string),
		Token:    token,
		Org:      settings["organization"].(string),
		Repos:    settings["repositories"].([]string),
	})

	return nil
}

func (g *GitHubPlugin) Shutdown() error {
	return nil
}

func (g *GitHubPlugin) GetStandupContext(timeRange plugin.TimeRange) (string, error) {
	return g.client.GetStandupContext(timeRange)
}

func getGhCliToken() (string, error) {
	cmd := exec.Command("gh", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("gh cli error: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to execute gh cli: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}
