package standup

import (
	"context"
	"daiv/internal/utils"
	"fmt"
	"os"
	"os/exec"
	"slices"
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

	for _, repo := range g.repos {
		report.WriteString(fmt.Sprintf("\n\n# Repository: %s\n", repo))

		issues, err := g.searchPullRequests(ctx, repo)
		if err != nil {
			return "", fmt.Errorf("error searching PRs for %s/%s: %v", g.org, repo, err)
		}
		if len(issues) > 0 {
			report.WriteString("## Authored Pull Requests:\n")
			for _, issue := range issues {
				if !utils.IsDateTimeInThreshold(issue.GetUpdatedAt().Time) {
					continue
				}
				report.WriteString(formatPullRequestFromIssue(issue))

				commitsReport, err := g.renderCommits(ctx, repo, issue.GetNumber())
				if err != nil {
					return "", fmt.Errorf("error fetching commits for PR #%d in %s/%s: %v", issue.GetNumber(), g.org, repo, err)
				}
				report.WriteString(commitsReport)
			}
		}

		issuesReviewed, err := g.searchReviewedPullRequests(ctx, repo)
		if err != nil {
			return "", fmt.Errorf("error searching reviewed PRs for %s/%s: %v", g.org, repo, err)
		}
		if len(issuesReviewed) > 0 {
			report.WriteString("## Reviewed Pull Requests:\n")
			for _, issue := range issuesReviewed {
				if !utils.IsDateTimeInThreshold(issue.GetUpdatedAt().Time) {
					continue
				}
				reviewReport, err := g.renderReviews(ctx, repo, issue.GetNumber())
				if err != nil {
					return "", fmt.Errorf("error fetching reviews for PR #%d in %s/%s: %v", issue.GetNumber(), g.org, repo, err)
				}
				if reviewReport != "" {
					report.WriteString(formatPullRequestFromIssue(issue))
					report.WriteString(reviewReport)
				}
			}
		}
	}

	return report.String(), nil
}

func (g *GitHubReport) searchPullRequests(ctx context.Context, repo string) ([]*github.Issue, error) {
	fromTime, err := time.Parse(time.RFC3339, viper.GetString("fromTime"))
	if err != nil {
		return nil, fmt.Errorf("error parsing fromTime: %v", err)
	}

	toTime, err := time.Parse(time.RFC3339, viper.GetString("toTime"))
	if err != nil {
		return nil, fmt.Errorf("error parsing toTime: %v", err)
	}

	query := fmt.Sprintf(
		"is:pr author:%s repo:%s/%s base:%s updated:>=%s updated:<=%s",
		g.username,
		g.org,
		repo,
		"master",
		fromTime.Format("2006-01-02"),
		toTime.Format("2006-01-02"),
	)
	searchOptions := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	result, _, err := g.client.Search.Issues(ctx, query, searchOptions)
	if err != nil {
		return nil, err
	}
	return result.Issues, nil
}

func (g *GitHubReport) searchReviewedPullRequests(ctx context.Context, repo string) ([]*github.Issue, error) {
	fromTime, err := time.Parse(time.RFC3339, viper.GetString("fromTime"))
	if err != nil {
		return nil, fmt.Errorf("error parsing fromTime: %v", err)
	}

	toTime, err := time.Parse(time.RFC3339, viper.GetString("toTime"))
	if err != nil {
		return nil, fmt.Errorf("error parsing toTime: %v", err)
	}

	query := fmt.Sprintf(
		"is:pr reviewed-by:%s repo:%s/%s base:%s updated:>=%s updated:<=%s",
		g.username,
		g.org,
		repo,
		"master",
		fromTime.Format("2006-01-02"),
		toTime.Format("2006-01-02"),
	)
	searchOptions := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	result, _, err := g.client.Search.Issues(ctx, query, searchOptions)
	if err != nil {
		return nil, err
	}
	return result.Issues, nil
}

func (g *GitHubReport) renderCommits(ctx context.Context, repo string, prNumber int) (string, error) {
	prCommits, _, err := g.client.PullRequests.ListCommits(ctx, g.org, repo, prNumber, nil)
	if err != nil {
		return "", err
	}

	slices.SortFunc(prCommits, func(a, b *github.RepositoryCommit) int {
		return a.GetCommit().GetCommitter().GetDate().Time.Compare(b.GetCommit().GetCommitter().GetDate().Time)
	})

	var commitReport strings.Builder
	relevantCommits := filterRelevantCommits(prCommits, g.username)
	if len(relevantCommits) > 0 {
		commitReport.WriteString("## Commits:\n\n")
		for _, commit := range relevantCommits {
			commitReport.WriteString(formatCommit(commit))
		}
	}

	return commitReport.String(), nil
}

func filterRelevantCommits(commits []*github.RepositoryCommit, username string) []*github.RepositoryCommit {
	var relevant []*github.RepositoryCommit
	for _, commit := range commits {
		if commit.Author != nil && commit.Author.GetLogin() == username &&
			utils.IsDateTimeInThreshold(commit.GetCommit().GetCommitter().GetDate().Time) {
			relevant = append(relevant, commit)
		}
	}
	return relevant
}

func formatPullRequestFromIssue(issue *github.Issue) string {
	return fmt.Sprintf("# PR #%d: %s (Status: %s)\n", issue.GetNumber(), issue.GetTitle(), issue.GetState())
}

func formatCommit(commit *github.RepositoryCommit) string {
	return fmt.Sprintf(
		"%s - %s: \n```\n%s\n```\n\n",
		commit.GetSHA()[:7],
		commit.GetCommit().GetCommitter().GetDate().Format("2006-01-02 15:04:05"),
		commit.GetCommit().GetMessage(),
	)
}

func (g *GitHubReport) renderReviews(ctx context.Context, repo string, prNumber int) (string, error) {
	reviews, _, err := g.client.PullRequests.ListReviews(ctx, g.org, repo, prNumber, nil)
	if err != nil {
		return "", err
	}

	var reviewReport strings.Builder
	var userReviewCount int
	for _, review := range reviews {
		if review.User != nil && review.User.GetLogin() == g.username {
			if review.GetSubmittedAt().IsZero() || !utils.IsDateTimeInThreshold(review.GetSubmittedAt().Time) {
				continue
			}
			reviewReport.WriteString(formatReview(review))
			userReviewCount++
		}
	}

	if userReviewCount > 0 {
		return "## Reviews:\n" + reviewReport.String(), nil
	}
	return "", nil
}

func formatReview(review *github.PullRequestReview) string {
	report := fmt.Sprintf("- Review ID: %d, State: %s, Submitted: %s\n",
		review.GetID(),
		review.GetState(),
		review.GetSubmittedAt().Format("2006-01-02 15:04:05"))
	if body := review.GetBody(); body != "" {
		report += fmt.Sprintf("  Comment:\n```\n%s\n```\n", body)
	}
	return report
}
