/*
Copyright © 2025 Iure Sales
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var uninstallPluginCmd = &cobra.Command{
	Use:     "uninstall [plugin-name]",
	Aliases: []string{"remove", "rm"},
	Short:   "Remove an installed plugin",
	Long:    `Remove an installed plugin from daiv.
This command removes external plugins that were previously installed.

Example:
  daiv plugin uninstall my-plugin`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		
		// Get plugins directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		
		pluginsDir := filepath.Join(homeDir, ".daiv", "plugins")
		
		// Check if external plugins directory exists
		if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
			return fmt.Errorf("plugins directory does not exist")
		}
		
		// Check for plugin file with .so extension
		soPath := filepath.Join(pluginsDir, pluginName+".so")
		if _, err := os.Stat(soPath); err == nil {
			// Remove the .so file
			if err := os.Remove(soPath); err != nil {
				return fmt.Errorf("failed to remove plugin: %w", err)
			}
			fmt.Printf("Successfully removed plugin: %s\n", pluginName)
			return nil
		}
		
		// Check for plugin file with .dll extension
		dllPath := filepath.Join(pluginsDir, pluginName+".dll")
		if _, err := os.Stat(dllPath); err == nil {
			// Remove the .dll file
			if err := os.Remove(dllPath); err != nil {
				return fmt.Errorf("failed to remove plugin: %w", err)
			}
			fmt.Printf("Successfully removed plugin: %s\n", pluginName)
			return nil
		}
		
		return fmt.Errorf("plugin '%s' not found", pluginName)
	},
}

func init() {
	pluginCmd.AddCommand(uninstallPluginCmd)
} 
