/*
Copyright © 2025 Iure Sales
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var createPluginCmd = &cobra.Command{
	Use:     "create [name]",
	Aliases: []string{"new", "generate"},
	Short:   "Generate a new empty plugin template",
	Long:    `Generate a new empty plugin template in the current directory.
This creates a new directory with the basic structure for a daiv plugin.

Example:
  daiv plugin create my-plugin`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		
		// Prompt for GitHub username using huh
		var githubUsername string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("GitHub Username").
					Description("Your GitHub username for repository references").
					Placeholder("username").
					Value(&githubUsername).
					Validate(func(s string) error {
						if s == "" {
							return fmt.Errorf("GitHub username is required")
						}
						return nil
					}),
			),
		)
		
		if err := form.Run(); err != nil {
			return fmt.Errorf("failed to get GitHub username: %w", err)
		}
		
		// Create plugin directory
		if err := os.Mkdir(pluginName, 0755); err != nil {
			return fmt.Errorf("failed to create plugin directory: %w", err)
		}
		
		// Create go.mod file
		goModContent := fmt.Sprintf(`module github.com/%s/%s

go 1.21

require (
	github.com/iures/daiv v0.1.0
)

// For local development, uncomment and update the path to your local daiv repository:
// replace github.com/iures/daiv => /absolute/path/to/local/daiv
`, githubUsername, pluginName)

		if err := os.WriteFile(filepath.Join(pluginName, "go.mod"), []byte(goModContent), 0644); err != nil {
			return fmt.Errorf("failed to create go.mod file: %w", err)
		}
		
		// Create main.go file
		mainContent := fmt.Sprintf(`package main

import (
	"github.com/%s/%s/plugin"
)

// Plugin is exported as a symbol for the daiv plugin system to find
var Plugin plugin.MyPlugin

func main() {} // Required for building as a plugin
`, githubUsername, pluginName)

		if err := os.WriteFile(filepath.Join(pluginName, "main.go"), []byte(mainContent), 0644); err != nil {
			return fmt.Errorf("failed to create main.go file: %w", err)
		}
		
		// Create plugin directory
		pluginDir := filepath.Join(pluginName, "plugin")
		if err := os.Mkdir(pluginDir, 0755); err != nil {
			return fmt.Errorf("failed to create plugin directory: %w", err)
		}
		
		// Create plugin implementation file
		pluginImplContent := fmt.Sprintf(`package plugin

import (
	"github.com/iures/daiv/pkg/plugin"
)

// MyPlugin implements the Plugin interface
type MyPlugin struct{}

// Name returns the unique identifier for this plugin
func (p *MyPlugin) Name() string {
	return "%s"
}

// Manifest returns the plugin manifest
func (p *MyPlugin) Manifest() *plugin.PluginManifest {
	return &plugin.PluginManifest{
		ConfigKeys: []plugin.ConfigKey{
			{
				Type:        plugin.ConfigTypeString,
				Key:         "%s.apikey",
				Name:        "API Key",
				Description: "API key for the service",
				Required:    true,
			},
			// Add more config keys as needed
		},
	}
}

// Initialize sets up the plugin with its configuration
func (p *MyPlugin) Initialize(settings map[string]interface{}) error {
	// Process configuration settings
	// apiKey := settings["%s.apikey"].(string)
	// TODO: Initialize your plugin with the settings
	return nil
}

// Shutdown performs cleanup when the plugin is being disabled/removed
func (p *MyPlugin) Shutdown() error {
	// TODO: Clean up any resources
	return nil
}

// GetStandupContext implements the StandupPlugin interface
func (p *MyPlugin) GetStandupContext(timeRange plugin.TimeRange) (plugin.StandupContext, error) {
	// TODO: Implement your plugin-specific standup context generation
	return plugin.StandupContext{
		PluginName: p.Name(),
		Content:    "Example content from %s plugin",
	}, nil
}
`, pluginName, pluginName, pluginName, pluginName)

		if err := os.WriteFile(filepath.Join(pluginDir, "plugin.go"), []byte(pluginImplContent), 0644); err != nil {
			return fmt.Errorf("failed to create plugin implementation file: %w", err)
		}
		
		// Create README.md
		readmeContent := fmt.Sprintf(`# %s

A plugin for the daiv CLI tool.

## Installation

### From Source

1. Clone the repository:
   ` + "```" + `
   git clone https://github.com/%s/%s.git
   cd %s
   ` + "```" + `

2. Build the plugin:
   ` + "```" + `
   go build -buildmode=plugin -o %s.so
   ` + "```" + `

3. Install the plugin:
   ` + "```" + `
   daiv plugin install /path/to/%s.so
   ` + "```" + `

### From GitHub

` + "```" + `
daiv plugin install %s/%s
` + "```" + `

## Configuration

This plugin requires the following configuration:

- %s.apikey: API key for the service

You can configure these settings when you first run daiv after installing the plugin.

## Usage

After installation, the plugin will be automatically loaded when you start daiv.

## Development

1. Fork this repository
2. Make your changes
3. Build and test locally
4. Submit a pull request
`, 
			strings.Title(pluginName), 
			githubUsername, 
			pluginName, 
			pluginName, 
			pluginName, 
			pluginName, 
			githubUsername,
			pluginName,
			pluginName)

		if err := os.WriteFile(filepath.Join(pluginName, "README.md"), []byte(readmeContent), 0644); err != nil {
			return fmt.Errorf("failed to create README.md file: %w", err)
		}
		
		fmt.Printf("Successfully generated plugin template in ./%s\n", pluginName)
		fmt.Println("\nNext steps:")
		fmt.Println("1. Implement your plugin functionality in 'plugin/plugin.go'")
		fmt.Println("2. Build your plugin with: go build -buildmode=plugin -o " + pluginName + ".so")
		fmt.Println("3. Install your plugin with: daiv plugin install ./" + pluginName + ".so")
		
		return nil
	},
}

func init() {
	pluginCmd.AddCommand(createPluginCmd)
} 
