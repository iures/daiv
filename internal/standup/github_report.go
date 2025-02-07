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
	username string
}

func NewGitHubReport() *GitHubReport {
	username := viper.GetString("github.username")
	token, err := getGhCliToken()
	if err != nil {
		fmt.Printf("Error getting GitHub token: %v\n", err)
		os.Exit(1)
	}

	authToken := github.BasicAuthTransport{ Username: username, Password: token, }
	client := github.NewClient(authToken.Client())
	if err != nil {
		fmt.Printf("Error creating GitHub client: %v\n", err)
		os.Exit(1)
	}

	return &GitHubReport{
		client: client,
		org:    viper.GetString("github.organization"),
		repos:  viper.GetStringSlice("github.repositories"),
		username: username,
	}
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
			if pr.User.GetLogin() == g.username &&
				pr.GetUpdatedAt().Year() == yesterday.Year() &&
				pr.GetUpdatedAt().Month() == yesterday.Month() &&
				pr.GetUpdatedAt().Day() >= yesterday.Day() {

				report.WriteString(
					fmt.Sprintf("- PR #%d: %s (Status: %s)\n",
						pr.GetNumber(),
						pr.GetTitle(),
						pr.GetState(),
					),
				)

				prCommits, _, err := g.client.PullRequests.ListCommits(ctx, g.org, repo, pr.GetNumber(), nil)
				if err != nil {
					return "", fmt.Errorf("error fetching commits for PR #%d in %s/%s: %v", pr.GetNumber(), g.org, repo, err)
				}

				var relevantCommits []*github.RepositoryCommit
				for _, commit := range prCommits {
					if commit.Author != nil && commit.Author.GetLogin() == g.username && commit.GetCommit().GetCommitter().GetDate().After(yesterday) {
						relevantCommits = append(relevantCommits, commit)
					}
				}

				if len(relevantCommits) > 0 {
					report.WriteString("  Commits:\n")
					for _, commit := range relevantCommits {
						report.WriteString(fmt.Sprintf("    - %s - %s: %s\n",
							commit.GetSHA()[:7],
							commit.GetCommit().GetCommitter().GetDate().Format("2006-01-02 15:04:05"),
							commit.GetCommit().GetMessage(),
						))
					}
				}
			}
		}
	}

	return report.String(), nil
}
