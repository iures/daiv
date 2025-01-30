# Bakuri

Bakuri is a command-line tool designed to streamline developer workflows and enhance team communication. It provides various utilities to help developers be more productive in their daily tasks.

## Features

- Generate standup reports automatically (`bakuri standup`)
- More features coming soon...

## Installation

### Using Homebrew

```bash
brew install bakuri

### Using Go

```bash
go install github.com/iures/bakuri
```

## Configuration

Bakuri requires some configuration to access your work tools. Create a config file at:

`~/.config/bakuri/config.yaml` or `~/.bakuri.yaml`

```yaml
# ~/.config/bakuri/config.yaml or ~/.bakuri.yaml

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
  bakuri standup [flags]
```

```log
Flags:
  -h, --help                   help for standup
      --jira-project string    Jira project ID
      --jira-token string      Jira API token
      --jira-url string        Jira instance URL
      --jira-username string   Jira username (email)

Global Flags:
      --config string   config file (default is $HOME/.bakuri.yaml)
```

## Troubleshooting Section:
Common issues and their solutions, such as:

### Configuration file not found
Make sure the config file is in the correct location: `~/.config/bakuri/config.yaml` or `~/.bakuri.yaml`

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
