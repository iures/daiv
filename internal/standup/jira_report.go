package standup

import (
	"bakuri/internal/jira"
	"fmt"
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

func (r *JiraReport) Print() error {
	if err := r.fetchIssues(); err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	if len(r.Issues) == 0 {
		fmt.Println("No issues found")
		return nil
	}

	fmt.Printf("Found %d issues\n\n", len(r.Issues))
	r.printIssues()
	return nil
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

func (r *JiraReport) printIssues() {
	for _, issue := range r.Issues {
		fmt.Println("\n--------------------------------")
		r.printIssueDetails(issue)
		r.printComments(issue)
		r.printChangelog(issue)
	}
}

func (r *JiraReport) printIssueDetails(issue goJira.Issue) {
	fmt.Printf("Issue: %s\n", issue.Key)
	fmt.Printf("  - Summary: %s\n", issue.Fields.Summary)
	fmt.Printf("  - Status: %s\n", issue.Fields.Status.Name)
}

func (r *JiraReport) printComments(issue goJira.Issue) {
	if issue.Fields.Comments == nil {
		return
	}

	fmt.Println("\tComments:")
	for _, comment := range issue.Fields.Comments.Comments {
		fmt.Printf("    %v - %v: \n %s\n\n",
			comment.Created,
			comment.Author.DisplayName,
			comment.Body,
		)
	}
}

func (r *JiraReport) printChangelog(issue goJira.Issue) {
	for _, history := range issue.Changelog.Histories {
		// parse example: 2024-09-11T17:27:32.642-0400
		layout := "2006-01-02T15:04:05.000-0700"
		createdTime, err := time.Parse(layout, history.Created)
		if err != nil {
			fmt.Printf("Failed to parse created time: %v\n", err)
			continue
		}

		if createdTime.Before(r.FromTime) {
			continue
		}

		fmt.Println("  - Change Log:")
		for _, item := range history.Items {
			fmt.Printf("    Created: %s\n", history.Created)
			fmt.Printf("    Author: %s\n", history.Author.DisplayName)
			fmt.Printf("    Field: %s\n", item.Field)
			fmt.Printf("    From: %s\n", item.FromString)
			fmt.Printf("    To: %s\n", item.ToString)
		}
	}
}

