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

var listPluginsCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all installed plugins",
	Long:    `List all plugins installed in daiv, including both built-in and external plugins.

Example:
  daiv plugin list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		registry := plugin.GetRegistry()
		
		// Get all registered plugins
		builtInPlugins := []string{}
		for name := range registry.Plugins {
			builtInPlugins = append(builtInPlugins, name)
		}
		
		// Get plugins directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		
		pluginsDir := filepath.Join(homeDir, ".daiv", "plugins")
		
		// Check if external plugins directory exists
		externalPlugins := []string{}
		if _, err := os.Stat(pluginsDir); !os.IsNotExist(err) {
			// Read directory
			entries, err := os.ReadDir(pluginsDir)
			if err != nil {
				return fmt.Errorf("failed to read plugins directory: %w", err)
			}
			
			// Collect plugin filenames
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				
				filename := entry.Name()
				if ext := filepath.Ext(filename); ext == ".so" || ext == ".dll" {
					// Remove extension from filename
					pluginName := strings.TrimSuffix(filename, ext)
					externalPlugins = append(externalPlugins, pluginName)
				}
			}
		}
		
		// Print built-in plugins
		fmt.Println("Built-in plugins:")
		if len(builtInPlugins) == 0 {
			fmt.Println("  No built-in plugins registered")
		} else {
			for _, name := range builtInPlugins {
				fmt.Printf("  - %s\n", name)
			}
		}
		
		// Print external plugins
		fmt.Println("\nExternal plugins:")
		if len(externalPlugins) == 0 {
			fmt.Println("  No external plugins installed")
		} else {
			for _, name := range externalPlugins {
				fmt.Printf("  - %s\n", name)
			}
		}
		
		return nil
	},
}

func init() {
	pluginCmd.AddCommand(listPluginsCmd)
} 
