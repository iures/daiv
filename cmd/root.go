/*
Copyright © 2025 Iure Sales
*/
package cmd

import (
	"daiv/internal/jira"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "daiv",
	Short: "Daiv is a CLI tool to streamline developer workflows",
	Long: `Daiv is a command-line tool designed to streamline developer workflows 
and enhance team communication. It provides various utilities to help developers 
be more productive in their daily tasks.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.daiv.yaml)")
	
	// Jira flags
	rootCmd.PersistentFlags().String("jira-username", "", "Jira username (email)")
	rootCmd.PersistentFlags().String("jira-token", "", "Jira API token")
	rootCmd.PersistentFlags().String("jira-url", "", "Jira instance URL")
	rootCmd.PersistentFlags().String("jira-project", "", "Jira project ID")
	
	// LLM flags
	rootCmd.PersistentFlags().String("llm-anthropic-apikey", "", "Anthropic API Key")
	
	// GitHub flags
	rootCmd.PersistentFlags().String("github-organization", "", "GitHub organization name")
	rootCmd.PersistentFlags().StringSlice("github-repositories", []string{}, "Comma-separated list of repository names to monitor")
	
	// Worklog flags
	rootCmd.PersistentFlags().String("worklog-path", "", "Path to the worklog file")
	
	// Bind flags to viper
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("jira.username", rootCmd.PersistentFlags().Lookup("jira-username"))
	viper.BindPFlag("jira.token", rootCmd.PersistentFlags().Lookup("jira-token"))
	viper.BindPFlag("jira.url", rootCmd.PersistentFlags().Lookup("jira-url"))
	viper.BindPFlag("jira.project", rootCmd.PersistentFlags().Lookup("jira-project"))
	viper.BindPFlag("worklog.path", rootCmd.PersistentFlags().Lookup("worklog-path"))
	viper.BindPFlag("llm.anthropic.apikey", rootCmd.PersistentFlags().Lookup("llm-anthropic-apikey"))
	viper.BindPFlag("github.organization", rootCmd.PersistentFlags().Lookup("github-organization"))
	viper.BindPFlag("github.repositories", rootCmd.PersistentFlags().Lookup("github-repositories"))

	// Set default values
	viper.SetDefault("jira.url", "https://ltvco.atlassian.net")

	// Bind environment variables
	viper.BindEnv("llm.anthropic.apikey", "ANTHROPIC_API_KEY")
	viper.BindEnv("jira.token", "JIRA_API_TOKEN")
	viper.BindEnv("jira.username", "JIRA_USERNAME")
	viper.BindEnv("jira.url", "JIRA_URL")
	viper.BindEnv("jira.project", "JIRA_PROJECT")
	viper.BindEnv("github.organization", "GITHUB_ORG")
	viper.BindEnv("github.repositories", "GITHUB_REPOS")
	viper.BindEnv("github.username", "GITHUB_USERNAME")
	viper.BindEnv("worklog.path", "WORKLOG_PATH")
}

func initConfig() {
	viper.AutomaticEnv()

	if err := loadConfigs(); err != nil {
		fmt.Println("Error loading configs:", err)
		os.Exit(1)
	}

	if err := jira.InitializeJira(); err != nil {
		fmt.Println("Error initializing Jira:", err)
		os.Exit(1)
	}
}

func loadConfigs() error {
	if cfgFile := viper.GetString("config"); cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		viper.AddConfigPath(filepath.Join(home, ".config", "daiv"))
		viper.AddConfigPath(home)
		viper.SetConfigName(".daiv")
		viper.SetConfigType("yaml")

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return fmt.Errorf("error reading config: %w", err)
			} 
		} else {
				fmt.Println("Config file read: ", viper.ConfigFileUsed())
			}

		if err := readCacheConfig(); err != nil {
			return err
		}
	}

	return nil
}

func readCacheConfig() error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}

	daivCacheDir := filepath.Join(cacheDir, "daiv")
	if _, err := os.Stat(daivCacheDir); errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(daivCacheDir, 0755); err != nil {
			return err
		}
	}

	cacheFile := filepath.Join(daivCacheDir, "config.yaml")
	if _, err := os.Stat(cacheFile); errors.Is(err, fs.ErrNotExist) {
		if err := os.WriteFile(cacheFile, []byte{}, 0644); err != nil {
			return err
		}
	}

	viper.SetConfigFile(cacheFile)

	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading cache config: %w", err)
		} 
	} else {
			fmt.Println("Cache config file merged: ", cacheFile)
		}

	return nil
}
