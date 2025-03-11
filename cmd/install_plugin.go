/*
Copyright © 2025 Iure Sales
*/
package cmd

import (
	"daiv/internal/plugin"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var installPluginCmd = &cobra.Command{
	Use:     "install [github-repo or url or local-file] [version]",
	Aliases: []string{"add"},
	Short:   "Install a plugin from a GitHub repository, URL, or local file",
	Long:    `Install a plugin from a GitHub repository, direct URL, or local file.

For GitHub repositories:
  daiv plugin install username/repo-name [version]

For direct URLs:
  daiv plugin install https://example.com/path/to/plugin.so

For local files:
  daiv plugin install /path/to/plugin.so
  daiv plugin install ./plugin.so

Example:
  daiv plugin install username/daiv-worklog-plugin
  daiv plugin install username/daiv-worklog-plugin v1.0.0
  daiv plugin install https://example.com/plugins/worklog-plugin.so
  daiv plugin install ./my-plugin.so`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		version := ""
		if len(args) > 1 {
			version = args[1]
		}

		// Get plugins directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		pluginsDir := filepath.Join(homeDir, ".daiv", "plugins")
		manager, err := plugin.NewPluginManager(pluginsDir)
		if err != nil {
			return fmt.Errorf("failed to create plugin manager: %w", err)
		}

		// Check if source is a URL
		if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
			// Install from direct URL
			fmt.Printf("Installing plugin from URL: %s\n", source)
			if err := manager.InstallFromURL(source); err != nil {
				return fmt.Errorf("failed to install plugin from URL: %w", err)
			}
			fmt.Printf("Successfully installed plugin from %s\n", source)
			return nil
		}
		
		// Check if source is a local file
		fileInfo, err := os.Stat(source)
		if err == nil && !fileInfo.IsDir() {
			// Install from local file
			fmt.Printf("Installing plugin from local file: %s\n", source)
			if err := manager.InstallFromLocalFile(source); err != nil {
				return fmt.Errorf("failed to install plugin from local file: %w", err)
			}
			fmt.Printf("Successfully installed plugin from %s\n", source)
			return nil
		}
		
		// If we get here, assume it's a GitHub repository
		fmt.Printf("Installing plugin from GitHub: %s", source)
		if version != "" {
			fmt.Printf(" (version: %s)", version)
		}
		fmt.Println()

		if err := manager.InstallFromGitHub(source, version); err != nil {
			return fmt.Errorf("failed to install plugin from GitHub: %w", err)
		}

		if version != "" {
			fmt.Printf("Successfully installed plugin from github.com/%s at version %s\n", source, version)
		} else {
			fmt.Printf("Successfully installed plugin from github.com/%s\n", source)
		}

		return nil
	},
}

func init() {
	pluginCmd.AddCommand(installPluginCmd)
} 
