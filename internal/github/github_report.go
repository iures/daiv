package github

import (
	"context"
	"daiv/internal/utils"
	"fmt"
	"os"
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

func (g *GitHubReport) Render() (string, error) {
	ctx := context.Background()
	var report strings.Builder

	for _, repo := range g.repos {
		repoHasContent := false
		repoSection := &strings.Builder{}
		repoSection.WriteString(fmt.Sprintf("\n# Repository: %s\n", repo))

		authoredPullRequestCommits, err := g.renderAuthoredPullRequestCommits(ctx, repo)
		if err != nil {
			return "", fmt.Errorf("error rendering authored pull request commits for %s/%s: %v", g.org, repo, err)
		}
		repoSection.WriteString(authoredPullRequestCommits)

		reviewedPullRequestCommits, err := g.renderReviewedPullRequestCommits(ctx, repo)
		if err != nil {
			return "", fmt.Errorf("error rendering reviewed pull request commits for %s/%s: %v", g.org, repo, err)
		}
		repoSection.WriteString(reviewedPullRequestCommits)

		issuesReviewed, err := g.searchReviewedPullRequests(ctx, repo)
		if err != nil {
			return "", fmt.Errorf("error searching reviewed PRs for %s/%s: %v", g.org, repo, err)
		}

		if len(issuesReviewed) > 0 {
			repoHasContent = true
			repoSection.WriteString("## Reviewed Pull Requests:\n")
			
			var hasReviewsInPeriod bool
			for _, issue := range issuesReviewed {
				reviewReport, err := g.renderReviews(ctx, repo, issue)
				if err != nil {
					return "", fmt.Errorf("error fetching reviews for PR #%d in %s/%s: %v", issue.GetNumber(), g.org, repo, err)
				}
				if reviewReport != "" {
					hasReviewsInPeriod = true
					repoSection.WriteString(formatPullRequestFromIssue(issue))
					repoSection.WriteString(reviewReport)

					reviewCommentReport, err := g.renderPrComments(ctx, repo, issue.GetNumber())
					if err != nil {
						return "", fmt.Errorf("error fetching comments for PR #%d in %s/%s: %v", issue.GetNumber(), g.org, repo, err)
					}
					repoSection.WriteString(reviewCommentReport)
				}
			}

			if !hasReviewsInPeriod {
				repoSection.WriteString("No reviews found in the specified time period.\n")
			}
		}

		if repoHasContent {
			report.WriteString(repoSection.String())
		}
	}

	if report.Len() == 0 {
		report.WriteString("\nNo GitHub activity found in the specified time period.\n")
	}

	return report.String(), nil
}

func (g *GitHubReport) renderAuthoredPullRequestCommits(ctx context.Context, repo string) (string, error) {
	issues, err := g.searchPullRequests(ctx, repo)
	if err != nil {
		return "", err
	}

	var report strings.Builder

	if len(issues) > 0 {
		report.WriteString("## Authored Pull Requests:\n")
		for _, issue := range issues {
			report.WriteString(formatPullRequestFromIssue(issue))

			commitsReport, err := g.renderCommits(ctx, repo, issue.GetNumber())
			if err != nil {
				return "", fmt.Errorf("error fetching commits for PR #%d in %s/%s: %v", issue.GetNumber(), g.org, repo, err)
			}
			report.WriteString(commitsReport)
		}
	}

	return report.String(), nil
}

func (g *GitHubReport) renderReviewedPullRequestCommits(ctx context.Context, repo string) (string, error) {
	issues, err := g.searchPullRequests(ctx, repo)
	if err != nil {
		return "", err
	}

	var report strings.Builder

	for _, issue := range issues {
		report.WriteString(formatPullRequestFromIssue(issue))

		commitsReport, err := g.renderCommits(ctx, repo, issue.GetNumber())
		if err != nil {
			return "", fmt.Errorf("error fetching commits for PR #%d in %s/%s: %v", issue.GetNumber(), g.org, repo, err)
		}
		report.WriteString(commitsReport)
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
		"is:pr author:%s repo:%s/%s base:%s updated:%s..%s",
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
		"is:pr reviewed-by:%s repo:%s/%s base:%s updated:%s..%s",
		g.username,
		g.org,
		repo,
		"master",
		fromTime.Format("2006-01-02"),
		toTime.Format("2006-01-02"),
	)
	searchOptions := &github.SearchOptions{
		Sort: "updated",
		Order: "desc",
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
		return a.GetCommit().GetCommitter().GetDate().Compare(b.GetCommit().GetCommitter().GetDate().Time)
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

func (g *GitHubReport) renderPrComments(ctx context.Context, repo string, prNumber int) (string, error) {
	comments, _, err := g.client.PullRequests.ListComments(ctx, g.org, repo, prNumber, nil)
	if err != nil {
		return "", err
	}

	var commentReport strings.Builder
	relevantComments := filterRelevantComments(comments, g.username)
	if len(relevantComments) > 0 {
		commentReport.WriteString("## Commits:\n\n")
		for _, comment := range relevantComments {
			commentReport.WriteString(formatComment(comment))
		}
	}

	return commentReport.String(), nil
}

func filterRelevantComments(comments []*github.PullRequestComment, username string) []*github.PullRequestComment {
	var relevant []*github.PullRequestComment
	for _, comment := range comments {
		if comment.User != nil && comment.User.GetLogin() == username &&
			utils.IsDateTimeInThreshold(comment.GetCreatedAt().Time) {
			relevant = append(relevant, comment)
		}
	}
	return relevant
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

func formatComment(comment *github.PullRequestComment) string {
	return fmt.Sprintf(
		"%s - %s: \n```\n%s\n```\n\n",
		comment.CreatedAt.Time.String(),
		comment.User.GetLogin(),
		*comment.Body,
	)
}

func (g *GitHubReport) renderReviews(ctx context.Context, repo string, issue *github.Issue) (string, error) {
	reviews, _, err := g.client.PullRequests.ListReviews(ctx, g.org, repo, issue.GetNumber(), nil)
	if err != nil {
		return "", err
	}

	var reviewReport strings.Builder
	for _, review := range reviews {
		if review.User != nil && review.User.GetLogin() == g.username {
			if review.GetSubmittedAt().IsZero() || !utils.IsDateTimeInThreshold(review.GetSubmittedAt().Time) {
				continue
			}
			reviewReport.WriteString(formatReview(review, issue))
		}
	}

	if reviewReport.Len() > 0 {
		return "### Reviews:\n" + reviewReport.String(), nil
	}
	return "", nil
}

func formatReview(review *github.PullRequestReview, issue *github.Issue) string {
	report := fmt.Sprintf("- PR: %d - %s\nAuthor: %s, State: %s, Submitted: %s\n",
		issue.GetNumber(),
		issue.GetTitle(),
		*review.GetUser().Login,
		review.GetState(),
		review.GetSubmittedAt().Format("2006-01-02 15:04:05"))

	if body := review.GetBody(); body != "" {
		report += fmt.Sprintf("  Comment:\n```\n%s\n```\n", body)
	}
	return report
}
