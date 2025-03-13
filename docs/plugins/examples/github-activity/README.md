# GitHub Activity Plugin Example

This example plugin fetches your recent GitHub activity and includes it in your standup report. It demonstrates how to:

1. Connect to an external API (GitHub)
2. Process and format data for the standup report
3. Handle configuration and authentication

## Features

- Fetches your recent GitHub activity (commits, pull requests, issues, etc.)
- Filters activity by time range
- Formats activity in a readable format for standup reports
- Handles authentication with GitHub API

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/daiv-github-activity.git
cd daiv-github-activity

# Build and install the plugin
make install
```

## Configuration

This plugin requires the following configuration:

- **GitHub API Token**: A personal access token with the `repo` scope
- **GitHub Username**: Your GitHub username

When you first run Daiv after installing this plugin, you'll be prompted to provide these values.

Alternatively, you can set the following environment variables:
- `GITHUB_TOKEN`: Your GitHub API token
- `GITHUB_USERNAME`: Your GitHub username

## Usage

Once installed and configured, the plugin will automatically provide GitHub activity context to your standup reports:

```bash
daiv standup
```

## How It Works

The plugin uses the GitHub API to fetch your recent activity, including:

- Commits you've made
- Pull requests you've opened, reviewed, or commented on
- Issues you've opened, closed, or commented on
- Repositories you've starred or forked

It then filters this activity based on the time range for the standup report and formats it in a readable way.

## Code Structure

- **main.go**: Plugin entry point
- **plugin/plugin.go**: Core plugin implementation
- **plugin/contexts/standup.go**: GitHub activity context provider

## Implementation Details

The main logic is in the `GetStandupContext` function in `plugin/contexts/standup.go`:

```go
func GetStandupContext(pluginName string, timeRange types.TimeRange) (types.StandupContext, error) {
    // Get GitHub token and username from configuration
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
```

## Dependencies

- github.com/google/go-github/v45/github
- golang.org/x/oauth2 
