package github

import (
	"context"
	"daiv/internal/plugin"
	"daiv/internal/utils"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	externalGithub "github.com/google/go-github/v68/github"
)

func (gc *GithubClient) GetStandupContext(timeRange plugin.TimeRange) (string, error) {
	var report strings.Builder

	for _, repo := range gc.settings.Repos {
		slog.Info("Getting standup context for repo", "repo", repo)

		repoHasContent := false
		repoSection := &strings.Builder{}
    fmt.Fprintf(repoSection, "\n# Repository: %s\n", repo)

		authoredPullRequestCommits, err := gc.renderAuthoredPullRequestCommits(repo, timeRange)
		if err != nil {
			return "", fmt.Errorf("error rendering authored pull request commits for %s/%s: %v", gc.settings.Org, repo, err)
		}

    fmt.Fprintln(repoSection, authoredPullRequestCommits)

		reviewedPullRequestCommits, err := gc.renderReviewedPullRequestCommits(repo, timeRange)
		if err != nil {
			return "", fmt.Errorf("error rendering reviewed pull request commits for %s/%s: %v", gc.settings.Org, repo, err)
		}

    fmt.Fprintln(repoSection, reviewedPullRequestCommits)

		issuesReviewed, err := gc.searchReviewedPullRequests(repo, timeRange)
		if err != nil {
			return "", fmt.Errorf("error searching reviewed PRs for %s/%s: %v", gc.settings.Org, repo, err)
		}

		if len(issuesReviewed) > 0 {
			repoHasContent = true
			repoSection.WriteString("## Reviewed Pull Requests:\n")
			
			var hasReviewsInPeriod bool
			for _, issue := range issuesReviewed {
				reviewReport, err := gc.renderReviews(repo, issue)
				if err != nil {
					return "", fmt.Errorf("error fetching reviews for PR #%d in %s/%s: %v", issue.GetNumber(), gc.settings.Org, repo, err)
				}
				if reviewReport != "" {
					hasReviewsInPeriod = true
					fmt.Fprintln(repoSection, formatPullRequestFromIssue(issue))
					repoSection.WriteString(reviewReport)

					reviewCommentReport, err := gc.renderPrComments(repo, issue.GetNumber())
					if err != nil {
						return "", fmt.Errorf("error fetching comments for PR #%d in %s/%s: %v", issue.GetNumber(), gc.settings.Org, repo, err)
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

func (gc *GithubClient) renderAuthoredPullRequestCommits(repo string, timeRange plugin.TimeRange) (string, error) {
	issues, err := gc.searchPullRequests(repo, timeRange)
	if err != nil {
		return "", err
	}

	var report strings.Builder

	if len(issues) > 0 {
		report.WriteString("## Authored Pull Requests:\n")
		for _, issue := range issues {
			report.WriteString(formatPullRequestFromIssue(issue))

			commitsReport, err := gc.renderCommits(repo, issue.GetNumber())
			if err != nil {
				return "", fmt.Errorf("error fetching commits for PR #%d in %s/%s: %v", issue.GetNumber(), gc.settings.Org, repo, err)
			}
			report.WriteString(commitsReport)
		}
	}

	return report.String(), nil
}

func (gc *GithubClient) renderReviewedPullRequestCommits(repo string, timeRange plugin.TimeRange) (string, error) {
	issues, err := gc.searchPullRequests(repo, timeRange)
	if err != nil {
		return "", err
	}

	var report strings.Builder

	for _, issue := range issues {
		report.WriteString(formatPullRequestFromIssue(issue))

		commitsReport, err := gc.renderCommits(repo, issue.GetNumber())
		if err != nil {
			return "", fmt.Errorf("error fetching commits for PR #%d in %s/%s: %v", issue.GetNumber(), gc.settings.Org, repo, err)
		}
		report.WriteString(commitsReport)
	}

	return report.String(), nil
}

func (gc *GithubClient) searchPullRequests(repo string, timeRange plugin.TimeRange) ([]*externalGithub.Issue, error) {
	ctx := context.Background()

	query := fmt.Sprintf(
		"is:pr author:%s repo:%s/%s base:%s updated:%s..%s",
		gc.settings.Username,
		gc.settings.Org,
		repo,
		"master",
		timeRange.Start.Format("2006-01-02"),
		timeRange.End.Format("2006-01-02"),
	)

  fmt.Println(query)

	searchOptions := &externalGithub.SearchOptions{
		ListOptions: externalGithub.ListOptions{PerPage: 100},
	}
	result, _, err := gc.client.Search.Issues(ctx, query, searchOptions)
	if err != nil {
		return nil, err
	}
	return result.Issues, nil
}

func (gc *GithubClient) searchReviewedPullRequests(repo string, timeRange plugin.TimeRange) ([]*externalGithub.Issue, error) {
	ctx := context.Background()

	query := fmt.Sprintf(
		"is:pr -author:%s reviewed-by:%s repo:%s/%s base:%s updated:%s..%s",
		gc.settings.Username,
		gc.settings.Username,
		gc.settings.Org,
		repo,
		"master",
		timeRange.Start.Format("2006-01-02"),
		timeRange.End.Format("2006-01-02"),
	)

	searchOptions := &externalGithub.SearchOptions{
		Sort: "updated",
		Order: "desc",
		ListOptions: externalGithub.ListOptions{PerPage: 100},
	}

	result, _, err := gc.client.Search.Issues(ctx, query, searchOptions)
	if err != nil {
		return nil, err
	}

	return result.Issues, nil
}

func (gc *GithubClient) renderCommits(repo string, prNumber int) (string, error) {
	ctx := context.Background()

	prCommits, _, err := gc.client.PullRequests.ListCommits(ctx, gc.settings.Org, repo, prNumber, nil)
	if err != nil {
		return "", err
	}

	slices.SortFunc(prCommits, func(a, b *externalGithub.RepositoryCommit) int {
		return a.GetCommit().GetCommitter().GetDate().Compare(b.GetCommit().GetCommitter().GetDate().Time)
	})

	var commitReport strings.Builder
	relevantCommits := filterRelevantCommits(prCommits, gc.settings.Username)
	if len(relevantCommits) > 0 {
		commitReport.WriteString("## Commits:\n\n")
		for _, commit := range relevantCommits {
			commitReport.WriteString(formatCommit(commit))
		}
	}

	return commitReport.String(), nil
}

func (gc *GithubClient) renderPrComments(repo string, prNumber int) (string, error) {
	ctx := context.Background()

	comments, _, err := gc.client.PullRequests.ListComments(ctx, gc.settings.Org, repo, prNumber, nil)
	if err != nil {
		return "", err
	}

	var commentReport strings.Builder
	relevantComments := filterRelevantPRComments(comments, gc.settings.Username)
	if len(relevantComments) > 0 {
		commentReport.WriteString("## Commits:\n\n")
		for _, comment := range relevantComments {
			commentReport.WriteString(formatComment(comment))
		}
	}

	return commentReport.String(), nil
}

func filterRelevantPRComments(comments []*externalGithub.PullRequestComment, username string) []*externalGithub.PullRequestComment {
	var relevant []*externalGithub.PullRequestComment
	for _, comment := range comments {
		if comment.User != nil && comment.User.GetLogin() == username &&
			utils.IsDateTimeInThreshold(comment.GetCreatedAt().Time) {
			relevant = append(relevant, comment)
		}
	}
	return relevant
}

func filterRelevantCommits(commits []*externalGithub.RepositoryCommit, username string) []*externalGithub.RepositoryCommit {
	var relevant []*externalGithub.RepositoryCommit
	for _, commit := range commits {
		if commit.Author != nil && commit.Author.GetLogin() == username &&
			utils.IsDateTimeInThreshold(commit.GetCommit().GetCommitter().GetDate().Time) {
			relevant = append(relevant, commit)
		}
	}
	return relevant
}

func formatPullRequestFromIssue(issue *externalGithub.Issue) string {
	return fmt.Sprintf("# PR #%d: %s (Status: %s)\n", issue.GetNumber(), issue.GetTitle(), issue.GetState())
}

func formatCommit(commit *externalGithub.RepositoryCommit) string {
	return fmt.Sprintf(
		"%s - %s: \n```\n%s\n```\n\n",
		commit.GetSHA()[:7],
		commit.GetCommit().GetCommitter().GetDate().Format("2006-01-02 15:04:05"),
		commit.GetCommit().GetMessage(),
	)
}

func formatComment(comment *externalGithub.PullRequestComment) string {
	return fmt.Sprintf(
		"%s - %s: \n```\n%s\n```\n\n",
		comment.CreatedAt.Time.String(),
		comment.User.GetLogin(),
		*comment.Body,
	)
}

func (gc *GithubClient) renderReviews(repo string, issue *externalGithub.Issue) (string, error) {
	ctx := context.Background()

	reviews, _, err := gc.client.PullRequests.ListReviews(ctx, gc.settings.Org, repo, issue.GetNumber(), nil)
	if err != nil {
		return "", err
	}

	var reviewReport strings.Builder
	var relevantReviews []*externalGithub.PullRequestReview

	// First collect all relevant reviews
	for _, review := range reviews {
		if review.User != nil && review.User.GetLogin() == gc.settings.Username  {
			if review.GetSubmittedAt().IsZero() || !utils.IsDateTimeInThreshold(review.GetSubmittedAt().Time) {
				continue
			}
			relevantReviews = append(relevantReviews, review)
		}
	}

	if len(relevantReviews) > 0 {
		// Sort reviews by submission date
		slices.SortFunc(relevantReviews, func(a, b *externalGithub.PullRequestReview) int {
			return a.GetSubmittedAt().Compare(b.GetSubmittedAt().Time)
		})

		// Write PR header once
		reviewReport.WriteString(fmt.Sprintf("### PR #%d: %s\n", issue.GetNumber(), issue.GetTitle()))
		
		// Write all reviews for this PR
		for _, review := range relevantReviews {
			reviewReport.WriteString(formatPullRequestReview(review))
		}
	}

	if reviewReport.Len() > 0 {
		return reviewReport.String(), nil
	}
	return "", nil
}

func formatPullRequestReview(review *externalGithub.PullRequestReview) string {
	report := fmt.Sprintf("- State: %s, Submitted: %s\n",
		review.GetState(),
		review.GetSubmittedAt().Format("2006-01-02 15:04:05"))

	if body := review.GetBody(); body != "" {
		report += fmt.Sprintf("  Comment:\n```\n%s\n```\n", body)
	}
	return report
}
