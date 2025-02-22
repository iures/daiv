package jira

import (
	jira "github.com/andygrunwald/go-jira"
)

func NewJiraClient() (*jira.Client, error) {
	config, err := GetJiraConfig()
	if err != nil {
		return nil, err
	}

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
