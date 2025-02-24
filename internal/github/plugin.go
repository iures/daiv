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

  return &GitHubPlugin{
    client: client,
    config: config,
  }, nil
}

func (g *GitHubPlugin) Name() string {
	return "github"
}

func (g *GitHubPlugin) Initialize(config map[string]interface{}) error {
	// Parse config and initialize client
	return nil
}

func (g *GitHubPlugin) Shutdown() error {
	return nil
}

func (g *GitHubPlugin) GenerateReport(ctx context.Context, timeRange plugin.TimeRange) (plugin.Report, error) {
  fmt.Println("GenerateReport")
	// Your existing GitHub report generation logic here
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
