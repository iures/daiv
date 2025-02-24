package github

import (
	"context"
	"daiv/internal/plugin"
	"fmt"
	"os/exec"
	"strings"

	"github.com/google/go-github/v68/github"
)

func NewGithubClient() (*GithubClient, error) {
	config, err := GetGithubConfig()
	if err != nil {
		return nil, err
	}

	token, err := getGhCliToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get gh cli token: %w", err)
	}

	authToken := github.BasicAuthTransport{
		Username: config.Username,
		Password: token,
	}

	client := github.NewClient(authToken.Client())

	githubClient := &GithubClient{
		client: client,
		config: config,
	}

	return githubClient, nil
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
type GithubClient struct {
	client *github.Client
	config *GithubConfig
}

func (gc *GithubClient) GetActivityReport(ctx context.Context, timeRange plugin.TimeRange) (string, error) {
	githubReport := NewGitHubReport()
	return githubReport.Render()
}
