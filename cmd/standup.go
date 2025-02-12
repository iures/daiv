package cmd

import (
	"bakuri/internal/llm"
	"bakuri/internal/standup"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// standupCmd represents the standup command
var standupCmd = &cobra.Command{
	Use:   "standup",
	Short: "Generate a standup report for the current day",
	Long: `
		Generate a standup report for the current day based on activity in the watched repositories:
		  - Gathers a list of jira tickets assigned to the current user
		  - Checks for status updates on the tickets that happened yesterday
		  - Gathers GitHub activity from watched repositories
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validateConfig(); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}

		var bar *progressbar.ProgressBar
		if !viper.GetBool("no-progress") {
			bar = progressbar.Default(
				-1,
				"Generating Jira report",
			)
		}

		jiraReport := standup.NewJiraReport()
		jiraContent, err := jiraReport.Render()
		if err != nil {
			fmt.Printf("Error generating Jira report: %v\n", err)
			os.Exit(1)
		}

		if bar != nil {
			bar.Describe("Generating GitHub report")
		}

		githubReport := standup.NewGitHubReport()
		githubContent, err := githubReport.Render()
		if err != nil {
			fmt.Printf("Error generating GitHub report: %v\n", err)
			os.Exit(1)
		}

		if bar != nil {
			bar.Describe("Generating filesystem report")
		}

		worklogReport := standup.NewWorklogReport()
		worklogContent, err := worklogReport.Render()
		if err != nil {
			fmt.Printf("Error generating worklog report: %v\n", err)
			os.Exit(1)
		}

		prompt := fmt.Sprintf(`
Generate a standup report for the current day based on. 
Just respond with the report and nothing else.
Make sure to include the correct Jira ticket number if available. (e.g. [PBR-1234])
It should follow the following format:
## Yesterday:
- xxx
- yyy

## Today:
- xxx
- yyy

No blockers or Blocked by ...

Context:

# Jira Activity:

%s


# GitHub Activity:

%s


# Manual Work Log:

%s

			`,
			jiraContent,
			githubContent,
			worklogContent,
		)

		if bar != nil {
			bar.Describe("Generating final report")
		}

		if viper.GetBool("prompt") {
			if bar != nil {
				bar.Clear()
			}
			fmt.Println(prompt)
		} else {
			llmClient, err := llm.NewClient()
			if err != nil {
				fmt.Printf("Error creating LLM client: %v\n", err)
				os.Exit(1)
			}

			finalReport, err := llmClient.GenerateFromSinglePrompt(prompt)
			if err != nil {
				fmt.Printf("Error generating report: %v\n", err)
				os.Exit(1)
			}

			if bar != nil {
				bar.Clear()
			}
			fmt.Println(finalReport)
		}
	},
}

func validateConfig() error {
	required := []string{
		"jira.username",
		"jira.token",
		"jira.url",
		"github.organization",
	}

	for _, key := range required {
		if viper.GetString(key) == "" {
			return fmt.Errorf("missing required config: %s", key)
		}
	}

	repos := viper.GetStringSlice("github.repositories")
	if len(repos) == 0 {
		return fmt.Errorf("github.repositories must contain at least one repository")
	}

	if path := viper.GetString("worklog.path"); path != "" {
		dir := filepath.Dir(path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("worklog directory does not exist: %s", dir)
		}
	}

	// Validate time formats
	_, err := time.Parse(time.RFC3339, viper.GetString("fromTime"))
	if err != nil {
		return fmt.Errorf("invalid from-time format. Must be RFC3339 format: %v", err)
	}

	_, err = time.Parse(time.RFC3339, viper.GetString("toTime"))
	if err != nil {
		return fmt.Errorf("invalid to-time format. Must be RFC3339 format: %v", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(standupCmd)

	// Set default values
	viper.SetDefault("jira.url", "https://ltvco.atlassian.net")

	// Update default time settings with flags
	standupCmd.Flags().String("from-time", time.Now().AddDate(0, 0, -1).Truncate(24*time.Hour).Format(time.RFC3339), "Start time for the report (RFC3339 format)")
	standupCmd.Flags().String("to-time", time.Now().Truncate(24*time.Hour).Format(time.RFC3339), "End time for the report (RFC3339 format)")

	// Add flags that can override config file settings
	standupCmd.Flags().String("jira-username", "", "Jira username (email)")
	standupCmd.Flags().String("jira-token", "", "Jira API token")
	standupCmd.Flags().String("jira-url", "", "Jira instance URL")
	standupCmd.Flags().String("jira-project", "", "Jira project ID")
	standupCmd.Flags().String("worklog-path", "", "Path to the worklog file")

	standupCmd.Flags().String("llm-anthropic-apikey", "", "Anthropic API Key")

	// Add GitHub-specific flags
	standupCmd.Flags().String("github-organization", "", "GitHub organization name")
	standupCmd.Flags().StringSlice("github-repositories", []string{}, "Comma-separated list of repository names to monitor")

	// Bind flags to viper
	viper.BindPFlag("jira.username", standupCmd.Flags().Lookup("jira-username"))
	viper.BindPFlag("jira.token", standupCmd.Flags().Lookup("jira-token"))
	viper.BindPFlag("jira.url", standupCmd.Flags().Lookup("jira-url"))
	viper.BindPFlag("jira.project", standupCmd.Flags().Lookup("jira-project"))
	viper.BindPFlag("worklog.path", standupCmd.Flags().Lookup("worklog-path"))
	viper.BindPFlag("llm.anthropic.apikey", standupCmd.Flags().Lookup("llm-anthropic-apikey"))

	// Bind GitHub flags
	viper.BindPFlag("github.username", standupCmd.Flags().Lookup("github-username"))
	viper.BindPFlag("github.organization", standupCmd.Flags().Lookup("github-organization"))
	viper.BindPFlag("github.repositories", standupCmd.Flags().Lookup("github-repositories"))

	// Bind time flags to viper
	viper.BindPFlag("fromTime", standupCmd.Flags().Lookup("from-time"))
	viper.BindPFlag("toTime", standupCmd.Flags().Lookup("to-time"))

	// Environment variables
	viper.BindEnv("llm.anthropic.apikey", "ANTHROPIC_API_KEY")
	viper.BindEnv("jira.token", "JIRA_API_TOKEN")
	viper.BindEnv("jira.username", "JIRA_USERNAME")
	viper.BindEnv("jira.url", "JIRA_URL")
	viper.BindEnv("jira.project", "JIRA_PROJECT")
	viper.BindEnv("github.organization", "GITHUB_ORG")
	viper.BindEnv("github.repositories", "GITHUB_REPOS") // Comma-separated list in env var
	viper.BindEnv("github.username", "GITHUB_USERNAME")
	viper.BindEnv("worklog.path", "WORKLOG_PATH")

	// Add the no-progress flag
	standupCmd.Flags().Bool("no-progress", false, "Disable progress bar")
	viper.BindPFlag("no-progress", standupCmd.Flags().Lookup("no-progress"))
}
