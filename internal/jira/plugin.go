package jira

import (
	"context"
	"daiv/internal/plugin"
	"fmt"

	"github.com/andygrunwald/go-jira"
)

type JiraConfig struct {
	Username string
	Token    string
	URL      string
	Project  string
}

type JiraPlugin struct {
	client *JiraClient
	config *JiraConfig
	user   *jira.User
}

func NewJiraPlugin() *JiraPlugin {
	return &JiraPlugin{}
}

func (j *JiraPlugin) Name() string {
	return "jira"
}

func (j *JiraPlugin) Manifest() *plugin.PluginManifest {
	return &plugin.PluginManifest{
		ConfigKeys: []plugin.ConfigKey{
			{
				Type:        plugin.ConfigTypeString,
				Key:         "jira.username",
				Name:        "Jira Username",
				Description: "The username for the Jira user",
				Required:    true,
				Secret:      false,
			},
			{
				Type:        plugin.ConfigTypeString,
				Key:         "jira.token",
				Name:        "Jira API Token",
				Description: "The API token for the Jira user",
				Required:    true,
				EnvVar:      "JIRA_API_TOKEN",
			},
			{
				Type:        plugin.ConfigTypeString,
				Key:         "jira.url",
				Name:        "Jira URL",
				Description: "The URL for the Jira instance",
				Required:    true,
				Secret:      false,
			},
			{
				Type:        plugin.ConfigTypeString,
				Key:         "jira.project",
				Name:        "Jira Project",
				Description: "The project to generate the report for",
				Required:    true,
				Secret:      false,
			},
		},
	}
}

func (j *JiraPlugin) Initialize(settings map[string]any) error {
	config := &JiraConfig{
		Username: settings["jira.username"].(string),
		Token:    settings["jira.token"].(string),
		URL:      settings["jira.url"].(string),
		Project:  settings["jira.project"].(string),
	}

	client, err := NewJiraClient(config)
	if err != nil {
		return fmt.Errorf("failed to create Jira client: %w", err)
	}

	j.client = client
	j.config = config
	j.user, err = j.client.GetSelf()

	if err != nil {
		return fmt.Errorf("failed to get Jira user: %w", err)
	}

	return nil
}

func (j *JiraPlugin) Shutdown() error {
	return nil
}

func (j *JiraPlugin) GetStandupContext(timeRange plugin.TimeRange) (plugin.StandupContext, error) {
	content, err := j.client.GetActivityReport(context.Background(), timeRange)
	if err != nil {
		return plugin.StandupContext{}, fmt.Errorf("failed to get activity report: %w", err)
	}

	return plugin.StandupContext{
		PluginName: j.Name(),
		Content:    content,
	}, nil
} 
