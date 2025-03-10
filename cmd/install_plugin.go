/*
Copyright © 2025 Iure Sales
*/
package cmd

import (
	"daiv/pkg/plugin"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var installPluginCmd = &cobra.Command{
	Use:     "install [github-repo or url] [version]",
	Aliases: []string{"add"},
	Short:   "Install a plugin from a GitHub repository or URL",
	Long:    `Install a plugin from a GitHub repository or direct URL.

For GitHub repositories:
  daiv plugin install username/repo-name [version]

For direct URLs:
  daiv plugin install https://example.com/path/to/plugin.so

Example:
  daiv plugin install username/daiv-worklog-plugin
  daiv plugin install username/daiv-worklog-plugin v1.0.0
  daiv plugin install https://example.com/plugins/worklog-plugin.so`,
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

		if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
			// Install from direct URL
			fmt.Printf("Installing plugin from URL: %s\n", source)
			if err := manager.InstallFromURL(source); err != nil {
				return fmt.Errorf("failed to install plugin from URL: %w", err)
			}
			fmt.Printf("Successfully installed plugin from %s\n", source)
		} else {
			// Assume it's a GitHub repository
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
		}

		return nil
	},
}

func init() {
	pluginCmd.AddCommand(installPluginCmd)
} 
