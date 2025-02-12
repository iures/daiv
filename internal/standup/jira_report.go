package standup

import (
	"daiv/internal/jira"
	"daiv/internal/utils"
	"fmt"
	"slices"
	"strings"
	"time"

	goJira "github.com/andygrunwald/go-jira"
)

type JiraReport struct {
	Issues         []goJira.Issue
	FromTime       time.Time
	ToTime         time.Time
}

func NewJiraReport() *JiraReport {
	return &JiraReport{
		Issues:   []goJira.Issue{},
		FromTime: time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour),
		ToTime:   time.Now().Truncate(24 * time.Hour),
	}
}

func (r *JiraReport) Render() (string, error) {
	if err := r.fetchIssues(); err != nil {
		return "", fmt.Errorf("failed to fetch issues: %w", err)
	}

	if len(r.Issues) == 0 {
		return "No issues found", nil
	}

	var report strings.Builder
	r.renderIssues(&report)
	return report.String(), nil
}

func (r *JiraReport) fetchIssues() error {
	config := jira.GetJiraConfig()
	client, err := jira.NewJiraClient()
	if err != nil {
		return err
	}

	searchString := fmt.Sprintf(`
		assignee = currentUser()
		AND project = %s
		AND status != Closed
		AND sprint IN openSprints()
		AND updated > -1d
	`, config.Project)

	opt := &goJira.SearchOptions{
		MaxResults: 100,
		Expand:     "changelog",
		Fields:     []string{"summary", "description", "status", "changelog", "comment"},
	}

	issues, _, err := client.Issue.Search(searchString, opt)

	if err != nil {
		return err
	}

	r.Issues = issues
	return nil
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

func (r *JiraReport) renderChangelog(report *strings.Builder, issue goJira.Issue) {
	if issue.Changelog == nil || len(issue.Changelog.Histories) == 0 {
		return
	}

	fmt.Fprintln(report, "## Change Log:")

	slices.SortFunc(issue.Changelog.Histories, func(a, b goJira.ChangelogHistory) int {
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

	for _, history := range issue.Changelog.Histories {
		layout := "2006-01-02T15:04:05.000-0700"
		createdTime, err := time.Parse(layout, history.Created)
		if err != nil {
			fmt.Fprintf(report, "Failed to parse created time: %v\n", err)
			continue
		}

		if createdTime.Before(r.FromTime) {
			continue
		}

		for _, item := range history.Items {
			fmt.Fprintf(
				report, "%s - %s changed: `%s` from: `%s` to: `%s`\n\n",
				createdTime.Format("2006-01-02 15:04:05"),
				history.Author.DisplayName,
				item.Field,
				item.FromString,
				item.ToString,
			)
		}
	}
}

