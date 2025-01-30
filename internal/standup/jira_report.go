package standup

import (
	"bakuri/internal/jira"
	"fmt"
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
	fmt.Fprintf(&report, "Found %d issues\n\n", len(r.Issues))
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
	for _, issue := range r.Issues {
		fmt.Fprintln(report, "\n--------------------------------")
		r.renderIssueDetails(report, issue)
		r.renderComments(report, issue)
		r.renderChangelog(report, issue)
	}
}

func (r *JiraReport) renderIssueDetails(report *strings.Builder, issue goJira.Issue) {
	fmt.Fprintf(report, "Issue: %s\n", issue.Key)
	fmt.Fprintf(report, "  - Summary: %s\n", issue.Fields.Summary)
	fmt.Fprintf(report, "  - Status: %s\n", issue.Fields.Status.Name)
}

func (r *JiraReport) renderComments(report *strings.Builder, issue goJira.Issue) {
	if issue.Fields.Comments == nil {
		return
	}

	fmt.Fprintln(report, "  - Comments:")
	for _, comment := range issue.Fields.Comments.Comments {
		fmt.Fprintf(report, "     %v - %v: \n       %s\n\n",
			comment.Created,
			comment.Author.DisplayName,
			comment.Body,
		)
	}
}

func (r *JiraReport) renderChangelog(report *strings.Builder, issue goJira.Issue) {
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

		fmt.Fprintln(report, "  - Change Log:")
		for _, item := range history.Items {
			fmt.Fprintf(report, "    Created: %s\n", history.Created)
			fmt.Fprintf(report, "    Author: %s\n", history.Author.DisplayName)
			fmt.Fprintf(report, "    Field: %s\n", item.Field)
			fmt.Fprintf(report, "    From: %s\n", item.FromString)
			fmt.Fprintf(report, "    To: %s\n", item.ToString)
		}
	}
}

