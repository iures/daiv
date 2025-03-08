package jira

import (
	"daiv/internal/plugin"
	"daiv/internal/utils"
	"fmt"
	"slices"
	"strings"
	"time"

	goJira "github.com/andygrunwald/go-jira"
)

type JiraReport struct {
	Issues    []goJira.Issue
	TimeRange plugin.TimeRange
	User      *goJira.User
}

func NewJiraReport() *JiraReport {
	return &JiraReport{
		Issues:    []goJira.Issue{},
		TimeRange: plugin.TimeRange{},
	}
}

func (r *JiraReport) Render() (string, error) {
	if len(r.Issues) == 0 {
		return "", nil
	}

	var report strings.Builder
	r.renderIssues(&report)
	return report.String(), nil
}

func (r *JiraReport) renderIssues(report *strings.Builder) {
	for i, issue := range r.Issues {
		if i > 0 {
			fmt.Fprintln(report, "\n--------------------------------\n")
		}
		r.renderIssueDetails(report, issue)
		r.renderComments(report, issue)
		r.renderChangelog(report, issue)
	}
}

func (r *JiraReport) renderIssueDetails(report *strings.Builder, issue goJira.Issue) {
	fmt.Fprintf(report, "## Jira Issue: [%s] - %s\n\n", issue.Key, issue.Fields.Status.Name)
	fmt.Fprintf(report, "%s\n\n", issue.Fields.Summary)
}

func (r *JiraReport) renderComments(report *strings.Builder, issue goJira.Issue) {
	if issue.Fields.Comments == nil {
		return
	}

	fmt.Fprintln(report, "## Comments:")

	slices.SortFunc(issue.Fields.Comments.Comments, func(a, b *goJira.Comment) int {
		aTime, err := time.Parse("2006-01-02T15:04:05.000-0700", a.Created)
		if err != nil {
			return 1
		}

		bTime, err := time.Parse("2006-01-02T15:04:05.000-0700", b.Created)
		if err != nil {
			return -1
		}

		return aTime.Compare(bTime)
	})

	for _, comment := range issue.Fields.Comments.Comments {
		createdTime, err := time.Parse("2006-01-02T15:04:05.000-0700", comment.Created)
		if err != nil {
			fmt.Fprintf(report, "Failed to parse created time: %v\n", err)
			continue
		}

		if utils.IsDateTimeInThreshold(createdTime) {
			fmt.Fprintf(
				report,
				"%v - %v: \n```\n%s\n```\n\n",
				createdTime.Format("2006-01-02 15:04:05"),
				comment.Author.DisplayName,
				comment.Body,
			)
		}
	}
}

func filter[T any](slice []T, condition func(T) bool) []T {
	filtered := []T{}
	for _, item := range slice {
		if condition(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (r *JiraReport) renderChangelog(report *strings.Builder, issue goJira.Issue) {
	relevantHistories := filter(issue.Changelog.Histories, func(history goJira.ChangelogHistory) bool {
		createdTime, err := time.Parse("2006-01-02T15:04:05.000-0700", history.Created)
		if err != nil {
			fmt.Fprintf(report, "Failed to parse created time for changelog history: %v\n", err)
			return false
		}

		return r.TimeRange.IsInRange(createdTime) && history.Author.AccountID == r.User.AccountID
	})

	if len(relevantHistories) == 0 {
		return
	}

	fmt.Fprintln(report, "## Change Log:")

	slices.SortFunc(relevantHistories, func(a, b goJira.ChangelogHistory) int {
		aTime, err := time.Parse("2006-01-02T15:04:05.000-0700", a.Created)
		if err != nil {
			return 1
		}

		bTime, err := time.Parse("2006-01-02T15:04:05.000-0700", b.Created)
		if err != nil {
			return -1
		}

		return aTime.Compare(bTime)
	})

	for _, history := range relevantHistories {
		for _, item := range history.Items {
			r.renderChangelogItem(report, history, item)
		}
	}
}

func (r *JiraReport) renderChangelogItem(report *strings.Builder, history goJira.ChangelogHistory, item goJira.ChangelogItems) {
	createdTime, err := time.Parse("2006-01-02T15:04:05.000-0700", history.Created)
	if err != nil {
		fmt.Fprintf(report, "Failed to parse created time: %v\n", err)
		return
	}

	fmt.Fprintf(
		report, "%s - %s changed: `%s` from: `%s` to: `%s`\n\n",
		createdTime.Format("2006-01-02 15:04:05"),
		history.Author.DisplayName,
		item.Field,
		item.FromString,
		item.ToString,
	)
}
