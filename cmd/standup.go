package cmd

import (
	"bakuri/internal/standup"
	"fmt"
	"os"

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
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validateConfig(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		jiraReport := standup.NewJiraReport()

		if err := jiraReport.Print(); err != nil {
			fmt.Printf("Error generating report: %v\n", err)
			os.Exit(1)
		}
	},
}

func validateConfig() error {
	required := []string{"jira.username", "jira.token", "jira.url"}
	for _, key := range required {
			if viper.GetString(key) == "" {
					return fmt.Errorf("missing required config: %s", key)
			}
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

	// Bind flags to viper
	viper.BindPFlag("jira.username", standupCmd.Flags().Lookup("jira-username"))
	viper.BindPFlag("jira.token", standupCmd.Flags().Lookup("jira-token"))
	viper.BindPFlag("jira.url", standupCmd.Flags().Lookup("jira-url"))
	viper.BindPFlag("jira.project", standupCmd.Flags().Lookup("jira-project"))

	// Also check for environment variables
	viper.BindEnv("jira.token", "JIRA_API_TOKEN")
	viper.BindEnv("jira.username", "JIRA_USERNAME")
	viper.BindEnv("jira.url", "JIRA_URL")
	viper.BindEnv("jira.project", "JIRA_PROJECT")
}
