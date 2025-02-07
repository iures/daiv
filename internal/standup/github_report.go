package standup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/go-github/v68/github"
	"github.com/spf13/viper"
)

type GitHubReport struct {
	client *github.Client
	org    string
	repos  []string
}

func NewGitHubReport() *GitHubReport {
	client, err := NewGithubClient()
	if err != nil {
		fmt.Printf("Error creating GitHub client: %v\n", err)
		os.Exit(1)
	}

	return &GitHubReport{
		client: client,
		org:    viper.GetString("github.organization"),
		repos:  viper.GetStringSlice("github.repositories"),
	}
}

func NewGithubClient() (*github.Client, error) {
	token, err := getGhCliToken()
	if err != nil {
		fmt.Printf("Error getting GitHub token: %v\n", err)
		return nil, err
	}

	username := viper.GetString("github.username")

	authToken := github.BasicAuthTransport{ Username: username, Password: token, }
	client := github.NewClient(authToken.Client())
	
	// Test the authentication
	ctx := context.Background()
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return nil, err
	}

	fmt.Printf("User: %v\n", github.Stringify(user))

	return client, nil
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

func (g *GitHubReport) Render() (string, error) {
	ctx := context.Background()
	var report strings.Builder

	yesterday := time.Now().AddDate(0, 0, -1)
	
	for _, repo := range g.repos {
		opts := &github.PullRequestListOptions{
			State: "open",
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}
		
		prs, _, err := g.client.PullRequests.List(ctx, g.org, repo, opts)
		if err != nil {
			return "", fmt.Errorf("error fetching PRs for %s/%s: %v", g.org, repo, err)
		}

		report.WriteString(fmt.Sprintf("\nRepository: %s\n", repo))
		
		for _, pr := range prs {
			if pr.GetUpdatedAt().Year() == yesterday.Year() &&
				pr.GetUpdatedAt().Month() == yesterday.Month() &&
				pr.GetUpdatedAt().Day() == yesterday.Day() {
				
				report.WriteString(fmt.Sprintf("- PR #%d: %s (Status: %s)\n",
					pr.GetNumber(),
					pr.GetTitle(),
					pr.GetState()))
			}
		}

		commits, _, err := g.client.Repositories.ListCommits(ctx, g.org, repo, &github.CommitsListOptions{
			Since: yesterday,
			Until: time.Now(),
			Author: viper.GetString("github.username"),
		})
		if err != nil {
			return "", fmt.Errorf("error fetching commits for %s/%s: %v", g.org, repo, err)
		}

		if len(commits) > 0 {
			report.WriteString("\nCommits:\n")
			for _, commit := range commits {
				report.WriteString(fmt.Sprintf("- %s: %s\n",
					commit.GetSHA()[:7],
					commit.GetCommit().GetMessage()))
			}
		}
	}

	return report.String(), nil
}
