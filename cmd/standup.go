package cmd

import (
	"fmt"
	"os"
	"time"

	jira "github.com/andygrunwald/go-jira"
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

		jiraUsername := viper.GetString("jira.username")
		jiraToken := viper.GetString("jira.token")

		tp := jira.BasicAuthTransport{
			Username: jiraUsername,
			Password: jiraToken,
		}

		jiraURL := viper.GetString("jira.url")
		
		client, err := jira.NewClient(tp.Client(), jiraURL)
		if err != nil {
			fmt.Println("Error creating Jira client:", err)
			return
		}
	
		projectId := viper.GetString("jira.project")

		searchString := fmt.Sprintf(`
			assignee = currentUser()
			AND project = %s
			AND status != Closed
			AND sprint IN openSprints()
			AND updated > -1d
		`, projectId)

		issues, err := getAllIssues(client, searchString)
		if err != nil {
			fmt.Println("Error getting issues:", err)
			return
		}

		if len(issues) == 0 {
			fmt.Println("No issues found")
			return
		} else {
			fmt.Printf("Found %d issues\n\n\n", len(issues))
		}

		// Print all issues's description
		for _, issue := range issues {
			fmt.Println("\n\n--------------------------------")
			fmt.Printf("Issue: %s\n", issue.Key)
			fmt.Printf("\tSummary: %s\n", issue.Fields.Summary)
			fmt.Printf("\tStatus: %s\n", issue.Fields.Status.Name)

			yesterday := time.Now().AddDate(0, 0, -1)

			if issue.Fields.Comments != nil {
				fmt.Println("\tComments:")
				for _, comment := range issue.Fields.Comments.Comments {
					fmt.Printf(
						"\t\t%v - %v: \n %s\n\n",
						comment.Created,
						comment.Author.DisplayName,
						comment.Body,
					)
				}
			}

			for _, changelogHistory := range issue.Changelog.Histories {
				if changelogHistory.Created > yesterday.Format(time.RFC3339) {
					for _, changelogItem := range changelogHistory.Items {
						fmt.Println("\tChange Log:")
						fmt.Printf("\t\tCreated: %s\n", changelogHistory.Created)
						fmt.Printf("\t\tAuthor: %s\n", changelogHistory.Author.DisplayName)
						fmt.Printf("\t\tField: %s\n", changelogItem.Field)
						fmt.Printf("\t\tFrom String: %s\n", changelogItem.FromString)
						fmt.Printf("\t\tTo String: %s\n", changelogItem.ToString)
					}
				}
			}
		}
	},
}

func getAllIssues(client *jira.Client, searchString string) ([]jira.Issue, error) {
	opt := &jira.SearchOptions{
		MaxResults: 100,
		Expand:     "changelog",
		Fields:     []string{"summary", "description", "status", "changelog", "comment"},
	}

	issues, resp, err := client.Issue.Search(searchString, opt)

	fmt.Printf("Response: %v\n", resp)

	if err != nil {
		return nil, err
	}

	return issues, nil
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
