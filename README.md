# Daiv

Daiv is a command-line tool designed to streamline developer workflows and enhance team communication. It provides various utilities to help developers be more productive in their daily tasks.

## Features

- Generate standup reports automatically (`daiv standup`)
- Generate relevant pull requests report (`daiv relevantPrs`)
- More features coming soon...

## Installation

### Using Go

```bash
go install github.com/iures/daiv
```

## Configuration

Daiv requires some configuration to access your work tools. Create a config file at:

`~/.config/daiv/config.yaml` or `~/.daiv.yaml`

```yaml
# ~/.config/daiv/config.yaml or ~/.daiv.yaml

# Jira Configuration
jira:
  url: "https://your-company.atlassian.net"
  email: "your.email@company.com"
  token: "your-jira-api-token" # or add JIRA_API_TOKEN environment variable
  project: "PROJECT_KEY"  # Your Jira project key

# LLM Configuration (Anthropic)
llm:
  anthropic:
    token: "your-anthropic-api-key" # or add ANTHROPIC_API_KEY environment variable

# Relevant PRs Configuration
relevantPrs:
  repositories:
    - owner: yourOrganization
      repo: yourRepoName
      keywords:
        - keyword1
        - keyword2
```

## Jira
Getting Jira API token:

1. Go to Jira settings
2. Go to "Security"
3. Go to "API tokens"
4. Click "Create API token"
5. Enter a name for the token
6. Click "Create"
7. Copy the token

## Usage

### Standup Report
Generate your daily standup report based on your recent work.

This command will:
- Fetch your recent Jira activities
- Generate a concise summary of your work
- Format it in a standup-friendly format

```bash
  daiv standup [flags]
```

If you want, you can override most configuration parameters with flags.

```log
Flags:
  -h, --help                   help for standup
      --jira-project string    Jira project ID
      --jira-token string      Jira API token
      --jira-url string        Jira instance URL
      --jira-username string   Jira username (email)

Global Flags:
      --config string   config file (default is $HOME/.daiv.yaml)
```

### Relevant PRs Report

Generate a report of pull requests that match your configured keywords across specified repositories. This command will:
- Scan pull requests in the repositories defined under the `relevantPrs` configuration section
- Filter them based on keywords (e.g., adyen, growthbook, telemetry)
- Provide a concise report to help you track the PRs relevant to your work

```bash
daiv relevantPrs [flags]
```

**Flags:**
```
  -h, --help                   help for relevantPrs
      --config string          config file (default is $HOME/.daiv.yaml)
```

## Troubleshooting Section:
Common issues and their solutions, such as:

### Configuration file not found
Make sure the config file is in the correct location: `~/.config/daiv/config.yaml` or `~/.daiv.yaml`

### Authentication errors with Jira
Make sure your Jira credentials are correct and have the necessary permissions.

### API rate limiting
Make sure you have a valid API key and are not exceeding the rate limits.

### LLM integration issues
Make sure your LLM API key is correct and you have the necessary permissions.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT License](LICENSE)
