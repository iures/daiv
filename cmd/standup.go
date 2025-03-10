package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"daiv/internal/llm"
	"daiv/internal/plugin"

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

		runStandup()
	},
}

func validateConfig() error {
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

	initFlags()
}

func initFlags() {
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

func runStandup() error {
	registry := plugin.GetRegistry()

	defer func() {
		if err := registry.ShutdownAll(); err != nil {
			slog.Error("Error shutting down plugins", "error", err)
		}
	}()

	timeRange := plugin.TimeRange{
		Start: viper.GetTime("fromTime"),
		End:   viper.GetTime("toTime"),
	}

	standupContextPlugins := registry.GetStandupPlugins()
	errChan := make(chan error, len(standupContextPlugins))
	reportChan := make(chan string, len(standupContextPlugins)) // Make buffered channel

	var wg sync.WaitGroup
	for _, reporter := range standupContextPlugins {
		wg.Add(1)
		go func(r plugin.StandupPlugin) {
			defer wg.Done()
			standupContext, err := r.GetStandupContext(timeRange)
			if err != nil {
				errChan <- fmt.Errorf("%s: %w", r.Name(), err)
				return
			}

			reportChan <- standupContext.String()
		}(reporter)
	}

	// Start a goroutine to close channels after all workers are done
	go func() {
		wg.Wait()
		close(reportChan)
		close(errChan)
	}()

	var standupContexts []string
	for report := range reportChan {
		if report != "" {	
			standupContexts = append(standupContexts, report)
		}
	}

	for err := range errChan {
		if err != nil {
			slog.Error("Error getting standup context", "error", err)
			return err
		}
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

Here is the context for the report:
%s`,
		strings.Join(standupContexts, "\n\n"),
	)

	if viper.GetBool("prompt") {
		fmt.Println(prompt)
		os.Exit(0)
	}

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

	return nil
}
