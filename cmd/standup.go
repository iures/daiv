package cmd

import (
	"daiv/internal/github"
	"daiv/internal/plugin"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

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

	initFlags()
	registerPlugins()
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

func registerPlugins() {
	registry := plugin.GetRegistry()

	githubPlugin := github.NewGitHubPlugin()

	if err := registry.Register(githubPlugin); err != nil {
		slog.Error("Failed to register GitHub plugin", "error", err)
		os.Exit(1)
	}

	slog.Info("Successfully registered all plugins")
}

func runStandup() error {
	registry := plugin.GetRegistry()

	// Make sure to clean up plugin resources when we're done
	defer func() {
		if err := registry.ShutdownAll(); err != nil {
			slog.Error("Error shutting down plugins", "error", err)
		}
	}()

	now := time.Now()
	timeRange := plugin.TimeRange{
		Start: now.Add(-24 * time.Hour),
		End:   now,
	}

	reporterPlugins := registry.Plugins
	errChan := make(chan error, len(reporterPlugins))
	reportChan := make(chan string, len(reporterPlugins)) // Make buffered channel

	var wg sync.WaitGroup
	for _, reporter := range reporterPlugins {
		wg.Add(1)
		go func(r plugin.Plugin) {
			defer wg.Done()
			if standupPlugin, ok := r.(plugin.StandupPlugin); ok {
				standupContext, err := standupPlugin.GetStandupContext(timeRange)
				if err != nil {
					errChan <- fmt.Errorf("%s: %w", r.Name(), err)

					return
				}
				reportChan <- standupContext
			}
		}(reporter)
	}

	// Start a goroutine to close channels after all workers are done
	go func() {
		wg.Wait()
		close(reportChan)
		close(errChan)
	}()

	// Collect reports and errors
	var reports []string
	for report := range reportChan {
		reports = append(reports, report)
	}

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return formatAndDisplayReports(reports)
}

func formatAndDisplayReports(reports []string) error {
	for _, report := range reports {
		fmt.Println(report)
	}

	return nil
}
