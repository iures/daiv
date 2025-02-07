package cmd

import (
	"bakuri/internal/llm"
	"bakuri/internal/standup"
	"fmt"
	"os"

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
			fmt.Println(err)
			os.Exit(1)
		}

		bar := progressbar.Default(
			-1,
			"Generating Jira report",
		)

		jiraReport := standup.NewJiraReport()
		jiraContent, err := jiraReport.Render()
		if err != nil {
			fmt.Printf("Error generating Jira report: %v\n", err)
			os.Exit(1)
		}

		bar.Describe("Generating GitHub report")

		githubReport := standup.NewGitHubReport()
		githubContent, err := githubReport.Render()
		if err != nil {
			fmt.Printf("Error generating GitHub report: %v\n", err)
			os.Exit(1)
		}

		prompt := fmt.Sprintf( `
Generate a standup report for the current day based on. 
Just respond with the report and nothing else.
It should follow the following format:
## Yesterday:
- xxx
- yyy

## Today:
- xxx
- yyy

No blockers

Context:

## Jira Activity:
%s

GitHub Activity:
%s
			`,
			jiraContent,
			githubContent,
		)

		bar.Describe("Generating final report")

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

		bar.Clear()

		fmt.Println(finalReport)
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

	return nil
}

func init() {
	rootCmd.AddCommand(standupCmd)

	// Set default values
	viper.SetDefault("jira.url", "https://ltvco.atlassian.net")
	
	// Add flags that can override config file settings
	standupCmd.Flags().String("jira-username", "", "Jira username (email)")
	standupCmd.Flags().String("jira-token", "", "Jira API token")
	standupCmd.Flags().String("jira-url", "", "Jira instance URL")
	standupCmd.Flags().String("jira-project", "", "Jira project ID")
	
	// Add GitHub-specific flags
	standupCmd.Flags().String("github-organization", "", "GitHub organization name")
	standupCmd.Flags().StringSlice("github-repositories", []string{}, "Comma-separated list of repository names to monitor")

	// Bind flags to viper
	viper.BindPFlag("jira.username", standupCmd.Flags().Lookup("jira-username"))
	viper.BindPFlag("jira.token", standupCmd.Flags().Lookup("jira-token"))
	viper.BindPFlag("jira.url", standupCmd.Flags().Lookup("jira-url"))
	viper.BindPFlag("jira.project", standupCmd.Flags().Lookup("jira-project"))
	
	// Bind GitHub flags
	viper.BindPFlag("github.username", standupCmd.Flags().Lookup("github-username"))
	viper.BindPFlag("github.organization", standupCmd.Flags().Lookup("github-organization"))
	viper.BindPFlag("github.repositories", standupCmd.Flags().Lookup("github-repositories"))

	// Environment variables
	viper.BindEnv("llm.anthropic.apikey", "ANTHROPIC_API_KEY")
	viper.BindEnv("jira.token", "JIRA_API_TOKEN")
	viper.BindEnv("jira.username", "JIRA_USERNAME")
	viper.BindEnv("jira.url", "JIRA_URL")
	viper.BindEnv("jira.project", "JIRA_PROJECT")
	viper.BindEnv("github.organization", "GITHUB_ORG")
	viper.BindEnv("github.repositories", "GITHUB_REPOS") // Comma-separated list in env var
	viper.BindEnv("github.username", "GITHUB_USERNAME")
}
