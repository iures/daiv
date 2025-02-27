package jira

import (
	"context"
	"daiv/internal/plugin"
)

type JiraPlugin struct {
	client *JiraClient
	config *JiraConfig
}

func NewJiraPlugin() (*JiraPlugin, error) {
  client, err := NewJiraClient()
  if err != nil {
    return nil, err
  }

  config, err := GetJiraConfig()
  if err != nil {
    return nil, err
  }

  return &JiraPlugin{
    client: client,
    config: config,
  }, nil
}

func (j *JiraPlugin) Manifest() *plugin.PluginManifest {
	return &plugin.PluginManifest{
		ConfigKeys: []plugin.ConfigKey{
			{Key: "username", Name: "Jira Username", Description: "The username for the Jira user", Required: true, Secret: false},
			{Key: "password", Name: "Jira Password", Description: "The password for the Jira user", Required: true, Secret: true},
			{Key: "url", Name: "Jira URL", Description: "The URL for the Jira instance", Required: true, Secret: false},
			{Key: "project", Name: "Jira Project", Description: "The project to generate the report for", Required: true, Secret: false},
		},
	}
}

func (j *JiraPlugin) Name() string {
	return "jira"
}

func (j *JiraPlugin) Initialize() error {
	return nil
}

func (j *JiraPlugin) Shutdown() error {
	return nil
}

func (j *JiraPlugin) GenerateReport(ctx context.Context, timeRange plugin.TimeRange) (plugin.Report, error) {
	content, err := j.client.GetActivityReport(ctx, timeRange)
	if err != nil {
		return plugin.Report{}, err
	}

	return plugin.Report{
		PluginName: j.Name(),
		Content:    content,
		Metadata:   map[string]interface{}{ },
	}, nil
} 
