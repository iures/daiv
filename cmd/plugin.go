package cmd

import (
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage daiv plugins",
	Long: `Manage daiv plugins - install, create, list, and uninstall plugins.

These commands help you work with daiv plugins:
  - install: Install a plugin from a GitHub repository or URL
  - create: Generate a new empty plugin template
  - list: List all installed plugins
  - uninstall: Remove an installed plugin

Example:
  daiv plugin install username/my-plugin
  daiv plugin create my-new-plugin
  daiv plugin list
  daiv plugin uninstall my-plugin`,
}

func init() {
	rootCmd.AddCommand(pluginCmd)
} 
