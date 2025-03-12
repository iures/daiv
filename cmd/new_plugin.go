package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var createPluginCmd = &cobra.Command{
	Use:     "create [name]",
	Aliases: []string{"new", "generate"},
	Short:   "Generate a new empty plugin template",
	Long:    `Generate a new empty plugin template in the current directory.
This creates a new directory with the basic structure for a daiv plugin.

Example:
  daiv plugin create myplugin
  daiv plugin create myplugin --dir ~/projects/daiv-plugins`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]
		titleCaser := cases.Title(language.English)
		
		// Get the directory from flag or use current directory
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			// Use current directory if not specified
			dir = "."
		}
		
		// Create a valid Go identifier from the plugin name
		goIdent := strings.ReplaceAll(pluginName, "-", "")
		goIdent = titleCaser.String(goIdent)
		
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
		
		// Create full path for the plugin directory
		pluginDir := filepath.Join(dir, pluginName)
		
		// Create plugin directory
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			return fmt.Errorf("failed to create plugin directory: %w", err)
		}
		
		// Create go.mod file
		goModContent := fmt.Sprintf(`module github.com/%s/%s

go 1.21

require github.com/iures/daivplug v0.0.3

// For local development, uncomment and update the path to your local daiv repository:
// replace github.com/iures/daivplug => /absolute/path/to/local/daiv
`, githubUsername, pluginName)

		if err := os.WriteFile(filepath.Join(pluginDir, "go.mod"), []byte(goModContent), 0644); err != nil {
			return fmt.Errorf("failed to create go.mod file: %w", err)
		}
		
		// Create main.go file
		mainContent := fmt.Sprintf(`package main

import (
	plug "github.com/iures/daivplug"
)

// %sPlugin implements the Plugin interface
type %sPlugin struct{
	// Add any fields needed by the plugin here
}

// Plugin is exported as a symbol for the daiv plugin system to find
var Plugin plug.Plugin = &%sPlugin{}

// Name returns the unique identifier for this plugin
func (p *%sPlugin) Name() string {
	return "%s"
}

// Manifest returns the plugin manifest
func (p *%sPlugin) Manifest() *plug.PluginManifest {
	return &plug.PluginManifest{
		ConfigKeys: []plug.ConfigKey{
			{
				Type:        0, // ConfigTypeString
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
func (p *%sPlugin) Initialize(settings map[string]interface{}) error {
	// Process configuration settings
	// apiKey := settings["%s.apikey"].(string)
	// TODO: Initialize your plugin with the settings
	return nil
}

// Shutdown performs cleanup when the plugin is being disabled/removed
func (p *%sPlugin) Shutdown() error {
	// TODO: Clean up any resources
	return nil
}

// GetStandupContext implements the StandupPlugin interface
func (p *%sPlugin) GetStandupContext(timeRange plug.TimeRange) (plug.StandupContext, error) {
	// TODO: Implement your plugin-specific standup context generation
	return plug.StandupContext{
		PluginName: p.Name(),
		Content:    "Example content from %s plugin",
	}, nil
}
`, 
			goIdent, 
			goIdent,
			goIdent,
			goIdent,
			pluginName, 
			goIdent,
			pluginName,
			goIdent,
			pluginName,
			goIdent,
			goIdent,
			pluginName)

		if err := os.WriteFile(filepath.Join(pluginDir, "main.go"), []byte(mainContent), 0644); err != nil {
			return fmt.Errorf("failed to create main.go file: %w", err)
		}
		
		// Create README.md
		readmeContent := fmt.Sprintf(`# %s

A plugin for the daiv CLI tool.

## Installation

### From GitHub

` + "```" + `
daiv plugin install %s/%s
` + "```" + `

### From Source

1. Clone the repository:
   ` + "```" + `
   git clone https://github.com/%s/%s.git
   cd %s
   ` + "```" + `

2. Build the plugin:
   ` + "```" + `
   go build -o out/%s.so -buildmode=plugin
   ` + "```" + `

3. Install the plugin:
   ` + "```" + `
   daiv plugin install ./out/%s.so
   ` + "```" + `

## Configuration

This plugin requires the following configuration:

- %s.apikey: API key for the service

You can configure these settings when you first run daiv after installing the plugin.

## Usage

After installation, the plugin will be automatically loaded when you start daiv.

`, 
			titleCaser.String(strings.ReplaceAll(pluginName, "-", " ")), 
			githubUsername, 
			pluginName, 
			githubUsername,
			pluginName,
			pluginName,
			pluginName,
			pluginName,
			pluginName)

		if err := os.WriteFile(filepath.Join(pluginDir, "README.md"), []byte(readmeContent), 0644); err != nil {
			return fmt.Errorf("failed to create README.md file: %w", err)
		}
		
		// Create output directory
		outDir := filepath.Join(pluginDir, "out")
		if err := os.Mkdir(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create out directory: %w", err)
		}
		
		fmt.Printf("Successfully generated plugin template in %s\n", pluginDir)
		fmt.Println("\nNext steps:")
		fmt.Println("1. Implement your plugin functionality in 'main.go'")
		fmt.Printf("2. Build your plugin with: cd %s && go build -o out/%s.so -buildmode=plugin\n", pluginDir, pluginName)
		fmt.Printf("3. Install your plugin with: daiv plugin install %s/out/%s.so\n", pluginDir, pluginName)
		
		return nil
	},
}

func init() {
	pluginCmd.AddCommand(createPluginCmd)
	createPluginCmd.Flags().String("dir", "", "Directory where the plugin will be created (default is current directory)")
} 
