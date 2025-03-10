package jira

import (
	"context"
	"daiv/pkg/plugin"
	"fmt"

	jira "github.com/andygrunwald/go-jira"
)

type JiraClient struct {
	client *jira.Client
	config *JiraConfig
	user   *jira.User
}

func NewJiraClient(config *JiraConfig) (*JiraClient, error) {
	tp := jira.BasicAuthTransport{
		Username: config.Username,
		Password: config.Token,
	}

	client, err := jira.NewClient(tp.Client(), config.URL)
	if err != nil {
		return nil, err
	}

	return &JiraClient{
		client: client,
		config: config,
	}, nil
}

func (j *JiraClient) GetSelf() (*jira.User, error) {
	user, _, err := j.client.User.GetSelf()
	if err != nil {
		return nil, err
	}

	j.user = user

	return user, nil
}

func (j *JiraClient) GetActivityReport(ctx context.Context, timeRange plugin.TimeRange) (string, error) {
	report := NewJiraReport()
	report.TimeRange = timeRange
	report.User = j.user

	issues, err := j.fetchUpdatedIssues(timeRange)
	if err != nil {
		return "", err
	}

	report.Issues = issues
	return report.Render()
}

func (j *JiraClient) fetchUpdatedIssues(timeRange plugin.TimeRange) ([]jira.Issue, error) {
	fromTime := timeRange.Start.Format("2006-01-02")
	toTime := timeRange.End.Format("2006-01-02")

	searchString := fmt.Sprintf(
		`assignee = currentUser() AND project = %s AND status != Closed AND sprint IN openSprints() AND (updatedDate >= %s AND updatedDate < %s)`,
		j.config.Project,
		fromTime,
		toTime,
	)

	opt := &jira.SearchOptions{
		MaxResults: 100,
		Expand:     "changelog",
		Fields:     []string{"summary", "description", "status", "changelog", "comment"},
	}

	issues, _, err := j.client.Issue.Search(searchString, opt)

	if err != nil {
		return nil, err
	}

	return issues, nil
}
