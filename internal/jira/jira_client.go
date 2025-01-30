package jira

import (
	jira "github.com/andygrunwald/go-jira"
	"github.com/spf13/viper"
)

type JiraConfig struct {
	Username string
	Token    string
	URL      string
	Project  string
}

func GetJiraConfig() JiraConfig {
	return JiraConfig{
		Username: viper.GetString("jira.username"),
		Token:    viper.GetString("jira.token"),
		URL:      viper.GetString("jira.url"),
		Project:  viper.GetString("jira.project"),
	}
}

func NewJiraClient() (*jira.Client, error) {
	config := GetJiraConfig()

	tp := jira.BasicAuthTransport{
		Username: config.Username,
		Password: config.Token,
	}

	client, err := jira.NewClient(tp.Client(), config.URL)
	if err != nil {
		return nil, err
	}

	return client, nil
}
