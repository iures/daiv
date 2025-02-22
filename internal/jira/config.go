package jira

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
)

type JiraConfig struct {
	Username string
	Token    string
	URL      string
	Project  string
}

func (c *JiraConfig) IsConfigured() bool {
	return c.Username != "" && c.Token != "" && c.URL != "" && c.Project != ""
} 

func GetJiraConfig() (*JiraConfig, error) {
	config := JiraConfig{
		Username: viper.GetString("jira.username"),
		Token:    viper.GetString("jira.token"),
		URL:      viper.GetString("jira.url"),
		Project:  viper.GetString("jira.project"),
	}
	return &config, nil
}

func newInput(key, title string, value *string) huh.Field {
	return huh.NewInput().
		Key(key).
		Title(title).
		Value(value).
		Validate(func(s string) error {
			if s == "" {
				return fmt.Errorf("%s is required", key)
			}
			return nil
		})
}

func ConfigPrompt(config *JiraConfig) ([]huh.Field, error) {
	var inputs []huh.Field
	
	if config.Username == "" {
		inputs = append(inputs, newInput("jira.username", "Jira Username", &config.Username))
	}

	if config.Token == "" {
		inputs = append(inputs, newInput("jira.token", "Jira API Token", &config.Token))
	}

	if config.URL == "" {
		inputs = append(inputs, newInput("jira.url", "Jira URL", &config.URL))
	}

	if config.Project == "" {
		inputs = append(inputs, newInput("jira.project", "Jira Project Key", &config.Project))
	}

	if len(inputs) > 0 {
		form := huh.NewForm(huh.NewGroup(inputs...))
		if err := form.Run(); err != nil {
			return nil, err
		}
	}

	return inputs, nil
}
