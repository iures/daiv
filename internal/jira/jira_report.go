package jira

import (
	"daiv/internal/plugin"
	"encoding/xml"
	"fmt"
	"slices"
	"time"

	goJira "github.com/andygrunwald/go-jira"
)

// XML structures for proper marshaling
type JiraXMLReport struct {
	XMLName xml.Name   `xml:"jira_report"`
	Issues  []XMLIssue `xml:"issue"`
}

type XMLIssue struct {
	Key      string      `xml:"key"`
	Status   string      `xml:"status"`
	Summary  string      `xml:"summary"`
	Comments XMLComments `xml:"comments"`
	Changelog XMLChangelog `xml:"changelog"`
}

type XMLComments struct {
	Comments []XMLComment `xml:"comment"`
}

type XMLComment struct {
	Timestamp string `xml:"timestamp"`
	Author    string `xml:"author"`
	Content   string `xml:"content"`
}

type XMLChangelog struct {
	Changes []XMLChange `xml:"change"`
}

type XMLChange struct {
	Timestamp string `xml:"timestamp"`
	Author    string `xml:"author"`
	Field     string `xml:"field"`
	From      string `xml:"from"`
	To        string `xml:"to"`
}

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

	xmlReport := JiraXMLReport{
		Issues: make([]XMLIssue, 0, len(r.Issues)),
	}

	for _, issue := range r.Issues {
		xmlIssue := XMLIssue{
			Key:     issue.Key,
			Status:  issue.Fields.Status.Name,
			Summary: issue.Fields.Summary,
		}

		// Process comments
		if issue.Fields.Comments != nil {
			comments := make([]XMLComment, 0)
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
					continue
				}

				if r.TimeRange.IsInRange(createdTime) {
					comments = append(comments, XMLComment{
						Timestamp: createdTime.Format("2006-01-02 15:04:05"),
						Author:    comment.Author.DisplayName,
						Content:   comment.Body,
					})
				}
			}
			xmlIssue.Comments = XMLComments{Comments: comments}
		}

		// Process changelog
		if issue.Changelog != nil {
			relevantHistories := filter(issue.Changelog.Histories, func(history goJira.ChangelogHistory) bool {
				createdTime, err := time.Parse("2006-01-02T15:04:05.000-0700", history.Created)
				if err != nil {
					return false
				}
				return r.TimeRange.IsInRange(createdTime) && history.Author.AccountID == r.User.AccountID
			})

			changes := make([]XMLChange, 0)
			for _, history := range relevantHistories {
				createdTime, err := time.Parse("2006-01-02T15:04:05.000-0700", history.Created)
				if err != nil {
					continue
				}

				for _, item := range history.Items {
					changes = append(changes, XMLChange{
						Timestamp: createdTime.Format("2006-01-02 15:04:05"),
						Author:    history.Author.DisplayName,
						Field:     item.Field,
						From:      item.FromString,
						To:        item.ToString,
					})
				}
			}
			xmlIssue.Changelog = XMLChangelog{Changes: changes}
		}

		xmlReport.Issues = append(xmlReport.Issues, xmlIssue)
	}

	// Marshal to XML with proper indentation
	output, err := xml.MarshalIndent(xmlReport, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal XML: %w", err)
	}

	// Add XML header and return
	return xml.Header + string(output), nil
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
