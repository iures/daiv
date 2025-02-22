package cmd

import (
	"daiv/internal/llm"
	"daiv/internal/standup"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh/spinner"
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
		// if err := validateConfig(); err != nil {
		// 	slog.Error(err.Error())
		// 	os.Exit(1)
		// }

		// Create channels for each content type and errors
		githubChan := make(chan string)
		jiraChan := make(chan string)
		worklogChan := make(chan string)
		errChan := make(chan error, 3) // Buffer for potential errors

		// Launch goroutines for each content type
		go func() {
			content, err := getGithubContent()
			if err != nil {
				errChan <- fmt.Errorf("github error: %v", err)
				githubChan <- ""
				return
			}
			githubChan <- content
		}()

		go func() {
			content := getProjectManagementContent()
			jiraChan <- content
		}()

		go func() {
			content, err := getWorklogContent()
			if err != nil {
				errChan <- fmt.Errorf("worklog error: %v", err)
				worklogChan <- ""
				return
			}
			worklogChan <- content
		}()

		// Collect results
		githubContent := <-githubChan
		jiraContent := <-jiraChan
		worklogContent := <-worklogChan

		// Check for errors
		select {
		case err := <-errChan:
			fmt.Printf("Error generating report: %v\n", err)
			os.Exit(1)
		default:
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

			%s `,
			jiraContent,
			githubContent,
			worklogContent,
		)

		if viper.GetBool("prompt") {
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

			fmt.Println(finalReport)
		}
	},
}

// Helper functions that work with the spinner
func getGithubContent() (string, error) {
	var content string
	var err error

	githubReport := standup.NewGitHubReport()
	action := func() {
		content, err = githubReport.Render()
	}

	spinner.New().Title("Generating GitHub report...").Action(action).Run()
	return content, err
}

func getProjectManagementContent() string {
	jiraReport := standup.NewJiraReport()
	if jiraReport == nil {
		return ""
	}

	content := ""
	action := func() {
		content, _ = jiraReport.Render()
	}

	spinner.New().Title("Generating Jira report...").Action(action).Run()
	return content
}

func getWorklogContent() (string, error) {
	var content string
	var err error

	worklogReport := standup.NewWorklogReport()
	action := func() {
		content, err = worklogReport.Render()
	}

	spinner.New().Title("Generating worklog report...").Action(action).Run()
	return content, err
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

	// Add standup-specific flags
	standupCmd.Flags().String("from-time", time.Now().AddDate(0, 0, -1).Truncate(24*time.Hour).Format(time.RFC3339), "Start time for the report (RFC3339 format)")
	standupCmd.Flags().String("to-time", time.Now().Truncate(24*time.Hour).Format(time.RFC3339), "End time for the report (RFC3339 format)")
	standupCmd.Flags().Bool("no-progress", false, "Disable progress bar")
	standupCmd.Flags().Bool("prompt", false, "Show the prompt instead of generating the report")

	// Bind time flags to viper
	viper.BindPFlag("fromTime", standupCmd.Flags().Lookup("from-time"))
	viper.BindPFlag("toTime", standupCmd.Flags().Lookup("to-time"))
	viper.BindPFlag("no-progress", standupCmd.Flags().Lookup("no-progress"))
	viper.BindPFlag("prompt", standupCmd.Flags().Lookup("prompt"))
}
