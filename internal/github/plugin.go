package github

import (
	"context"
	"daiv/internal/plugin"
	"fmt"
)

type GitHubPlugin struct {
	client *GithubClient
	config *GithubConfig
}

func NewGitHubPlugin() (*GitHubPlugin, error) {
  client, err := NewGithubClient()
  if err != nil {
    return nil, err
  }

  config, err := GetGithubConfig()
  if err != nil {
    return nil, err
  }

  return &GitHubPlugin {
    client: client,
    config: config,
  }, nil
}

func (g *GitHubPlugin) Name() string {
	return "github"
}

func (g *GitHubPlugin) Manifest() *plugin.PluginManifest {
	return &plugin.PluginManifest{
		ConfigKeys: []plugin.ConfigKey{
			{
				Key:         "github.organization",
				Name:        "GitHub Organization",
				Description: "The GitHub organization to monitor",
				Required:    true,
			},
			{
				Key:         "github.repositories",
				Name:        "GitHub Repositories",
				Description: "List of repositories to monitor",
				Required:    true,
			},
		},
	}
}

func (g *GitHubPlugin) Initialize() error {
	// Initialize the plugin with configuration from viper or other sources
	return nil
}

func (g *GitHubPlugin) Shutdown() error {
	return nil
}

func (g *GitHubPlugin) GenerateReport(ctx context.Context, timeRange plugin.TimeRange) (plugin.Report, error) {
  fmt.Println("GenerateReport")

	content, err := g.client.GetActivityReport(ctx, timeRange)
	if err != nil {
		return plugin.Report{}, err
	}

	return plugin.Report{
		PluginName: g.Name(),
		Content:    content,
		Metadata:   map[string]interface{}{
			"repository_count": len(g.config.Repositories),
		},
	}, nil
} 
