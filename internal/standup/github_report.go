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
	// Try to get token from gh cli
	token, err := getGhCliToken()
	if err != nil {
		fmt.Printf("Error getting GitHub token: %v\n", err)
		return nil, err
	}

	client := github.NewClient(nil).WithAuthToken(token)

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

	// Get yesterday's date boundaries
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	startOfYesterday := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	user, _, err := g.client.Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("error fetching user: %v", err)
	}
	myUsername := user.GetLogin()

	// Fetch events performed by the user
	events, _, err := g.client.Activity.ListEventsPerformedByUser(ctx, myUsername, false, &github.ListOptions{PerPage: 100})
	if err != nil {
		return "", fmt.Errorf("error fetching events for user %s: %v", myUsername, err)
	}

	report.WriteString(fmt.Sprintf("\nGitHub Activity for %s (from %s to %s):\n", myUsername, startOfYesterday.Format("2006-01-02"), startOfToday.Format("2006-01-02")))
	eventFound := false

	for _, event := range events {
		if event.CreatedAt.After(startOfYesterday) && event.CreatedAt.Before(startOfToday) {
			eventFound = true
			// Switch based on event type to display desired information
			switch event.GetType() {
			case "PushEvent":
				// Display commit information from PushEvent, if available
				if payload, ok := event.Payload().(*github.PushEvent); ok {
					repoName := event.GetRepo().GetName()

					for _, commit := range payload.Commits {
						report.WriteString(fmt.Sprintf("- PushEvent in %s: %s (Message: %s)\n",
							repoName,
							commit.GetSHA()[:7],
							commit.GetMessage()))
					}
				} else {
					report.WriteString(fmt.Sprintf("- PushEvent in %s at %s\n", event.GetRepo().GetName(), event.GetCreatedAt().Format(time.RFC1123)))
				}
			case "PullRequestEvent":
				// Display pull request information from PullRequestEvent, if available
				if payload, ok := event.Payload().(*github.PullRequestEvent); ok {
					repoName := event.GetRepo().GetName()
					action := payload.GetAction()
					pr := payload.GetPullRequest()
					report.WriteString(fmt.Sprintf("- PullRequestEvent in %s: PR #%d %s - %s\n",
						repoName, pr.GetNumber(), action, pr.GetTitle()))
				} else {
					report.WriteString(fmt.Sprintf("- PullRequestEvent in %s at %s\n", event.GetRepo().GetName(), event.GetCreatedAt().Format(time.RFC1123)))
				}
			default:
				// For other event types, just print a generic line with event type and repo
				report.WriteString(fmt.Sprintf("- %s in %s at %s\n",
					event.GetType(),
					event.GetRepo().GetName(),
					event.GetCreatedAt().Format(time.RFC1123)))
			}
		}
	}

	if !eventFound {
		report.WriteString("\nNo activity found for yesterday.\n")
	}

	return report.String(), nil
}
