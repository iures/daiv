package standup

import (
	"bakuri/internal/utils"
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
	client   *github.Client
	org      string
	repos    []string
	username string
}

func NewGitHubReport() *GitHubReport {
	username := viper.GetString("github.username")
	token, err := getGhCliToken()
	if err != nil {
		fmt.Printf("Error getting GitHub token: %v\n", err)
		os.Exit(1)
	}

	authToken := github.BasicAuthTransport{
		Username: username,
		Password: token,
	}
	client := github.NewClient(authToken.Client())

	return &GitHubReport{
		client:   client,
		org:      viper.GetString("github.organization"),
		repos:    viper.GetStringSlice("github.repositories"),
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

	threshold := time.Now().AddDate(0, 0, -1)

	for _, repo := range g.repos {
		report.WriteString(fmt.Sprintf("\nRepository: %s\n", repo))

		prs, err := g.fetchPullRequests(ctx, repo)
		if err != nil {
			return "", fmt.Errorf("error fetching PRs for %s/%s: %v", g.org, repo, err)
		}

		for _, pr := range prs {
			if g.shouldReportPullRequest(pr, threshold) {
				report.WriteString(formatPullRequest(pr))

				commitsReport, err := g.renderCommits(ctx, repo, pr, threshold)
				if err != nil {
					return "", fmt.Errorf("error fetching commits for PR #%d in %s/%s: %v", pr.GetNumber(), g.org, repo, err)
				}
				report.WriteString(commitsReport)
			}
		}
	}

	return report.String(), nil
}

// fetchPullRequests retrieves open pull requests for a given repository.
func (g *GitHubReport) fetchPullRequests(ctx context.Context, repo string) ([]*github.PullRequest, error) {
	opts := &github.PullRequestListOptions{
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	prs, _, err := g.client.PullRequests.List(ctx, g.org, repo, opts)
	return prs, err
}

func (g *GitHubReport) shouldReportPullRequest(pr *github.PullRequest, threshold time.Time) bool {
	if pr.User.GetLogin() != g.username {
		return false
	}
	return utils.IsDateOnOrAfter(pr.GetUpdatedAt().Time, threshold)
}

func (g *GitHubReport) renderCommits(ctx context.Context, repo string, pr *github.PullRequest, threshold time.Time) (string, error) {
	prCommits, _, err := g.client.PullRequests.ListCommits(ctx, g.org, repo, pr.GetNumber(), nil)
	if err != nil {
		return "", err
	}

	var commitReport strings.Builder
	relevantCommits := filterRelevantCommits(prCommits, g.username, threshold)
	if len(relevantCommits) > 0 {
		commitReport.WriteString("  Commits:\n")
		for _, commit := range relevantCommits {
			commitReport.WriteString(formatCommit(commit))
		}
	}

	return commitReport.String(), nil
}

func filterRelevantCommits(commits []*github.RepositoryCommit, username string, threshold time.Time) []*github.RepositoryCommit {
	var relevant []*github.RepositoryCommit
	for _, commit := range commits {
		if commit.Author != nil && commit.Author.GetLogin() == username && commit.GetCommit().GetCommitter().GetDate().After(threshold) {
			relevant = append(relevant, commit)
		}
	}
	return relevant
}

func formatPullRequest(pr *github.PullRequest) string {
	return fmt.Sprintf("- PR #%d: %s (Status: %s)\n", pr.GetNumber(), pr.GetTitle(), pr.GetState())
}

func formatCommit(commit *github.RepositoryCommit) string {
	return fmt.Sprintf("    - %s - %s: %s\n",
		commit.GetSHA()[:7],
		commit.GetCommit().GetCommitter().GetDate().Format("2006-01-02 15:04:05"),
		commit.GetCommit().GetMessage(),
	)
}
