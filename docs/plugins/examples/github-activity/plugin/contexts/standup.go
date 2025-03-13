package contexts

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v45/github"
	"github.com/iures/daivplug/types"
	"golang.org/x/oauth2"
)

// GetStandupContext generates the standup context for the plugin
func GetStandupContext(pluginName string, timeRange types.TimeRange) (types.StandupContext, error) {
	// Get GitHub token and username from environment variables
	token := os.Getenv("GITHUB_TOKEN")
	username := os.Getenv("GITHUB_USERNAME")
	
	// Create GitHub client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	
	// Fetch user events
	events, _, err := client.Activity.ListEventsPerformedByUser(
		context.Background(), 
		username, 
		false, 
		&github.ListOptions{PerPage: 30},
	)
	if err != nil {
		return types.StandupContext{}, err
	}
	
	// Process and format events
	var content strings.Builder
	content.WriteString("## Recent GitHub Activity\n\n")
	
	// Group events by repository
	eventsByRepo := make(map[string][]github.Event)
	for _, event := range events {
		eventTime := event.GetCreatedAt()
		if eventTime.After(timeRange.Start) && eventTime.Before(timeRange.End) {
			repoName := event.GetRepo().GetName()
			eventsByRepo[repoName] = append(eventsByRepo[repoName], *event)
		}
	}
	
	// Format events by repository
	for repo, repoEvents := range eventsByRepo {
		content.WriteString(fmt.Sprintf("### %s\n\n", repo))
		
		for _, event := range repoEvents {
			eventType := formatEventType(event.GetType())
			eventTime := event.GetCreatedAt().Format("15:04")
			
			content.WriteString(fmt.Sprintf("- %s at %s\n", eventType, eventTime))
		}
		
		content.WriteString("\n")
	}
	
	// If no events were found, add a message
	if len(eventsByRepo) == 0 {
		content.WriteString("No GitHub activity found for the specified time range.\n")
	}
	
	return types.StandupContext{
		PluginName: pluginName,
		Content:    content.String(),
	}, nil
}

// formatEventType converts GitHub event types to more readable descriptions
func formatEventType(eventType string) string {
	switch eventType {
	case "PushEvent":
		return "Pushed commits"
	case "PullRequestEvent":
		return "Worked on a pull request"
	case "IssuesEvent":
		return "Worked on an issue"
	case "IssueCommentEvent":
		return "Commented on an issue"
	case "CreateEvent":
		return "Created a branch or tag"
	case "DeleteEvent":
		return "Deleted a branch or tag"
	case "WatchEvent":
		return "Starred a repository"
	case "ForkEvent":
		return "Forked a repository"
	default:
		return eventType
	}
} 
