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
				Type:        plugin.ConfigTypeString,
				Key:         "username",
				Name:        "GitHub Username",
				Description: "Your GitHub username",
				Required:    true,
			},
			{
				Type:        plugin.ConfigTypeString,
				Key:         "organization",
				Name:        "GitHub Organization",
				Description: "The GitHub organization to monitor",
				Required:    true,
			},
			{
				Type:        plugin.ConfigTypeMultiline,
				Key:         "repositories",
				Name:        "GitHub Repositories",
				Description: "List of repositories to monitor",
				Required:    true,
			},
		},
	}
}

func (g *GitHubPlugin) Initialize(settings map[string]any) error {
	token, err := getGhCliToken()
	if err != nil {
		return fmt.Errorf("failed to get gh cli token: %w", err)
	}

	repos := settings["repositories"].([]any)
	var reposStr []string
	for _, repo := range repos {
		if str, ok := repo.(string); ok {
			reposStr = append(reposStr, str)
		}
	}
	username, ok := settings["username"].(string)
	if !ok {
		return fmt.Errorf("username is required")
	}
	org, ok := settings["organization"].(string)
	if !ok {
		return fmt.Errorf("organization is required")
	}

	g.client.Init(GithubClientSettings{
		Username: username,
		Token:    token,
		Org:      org,
		Repos:    reposStr,
	})

	return nil
}

func (g *GitHubPlugin) Shutdown() error {
	return nil
}

func (g *GitHubPlugin) GetStandupContext(timeRange plugin.TimeRange) (plugin.StandupContext, error) {
	standupContext, err := g.client.GetStandupContext(timeRange)
	if err != nil {
		return plugin.StandupContext{}, fmt.Errorf("failed to get standup context: %w", err)
	}

	return plugin.StandupContext{
		PluginName: g.Name(),
		Content:    standupContext,
	}, nil
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
