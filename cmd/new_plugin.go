package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
The plugin name will be automatically prefixed with 'daiv-' if not already present.

Example:
  daiv plugin create myplugin
  daiv plugin create daiv-myplugin --dir ~/projects/daiv-plugins`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rawPluginName := args[0]
		titleCaser := cases.Title(language.English)
		
		// Ensure plugin name has daiv- prefix
		pluginName := rawPluginName
		if !strings.HasPrefix(pluginName, "daiv-") {
			pluginName = "daiv-" + pluginName
		}
		
		// Get the directory from flag or use current directory
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			// Use current directory if not specified
			dir = "."
		}
		
		// Create a valid Go identifier from the plugin name
		// Remove the daiv- prefix for the Go identifier
		nameWithoutPrefix := strings.TrimPrefix(pluginName, "daiv-")
		goIdent := strings.ReplaceAll(nameWithoutPrefix, "-", "")
		goIdent = titleCaser.String(goIdent)
		
		// Create full path for the plugin directory
		pluginDir := filepath.Join(dir, pluginName)
		
		// Create plugin directory
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			return fmt.Errorf("failed to create plugin directory: %w", err)
		}
		
		// Create go.mod file
		goModContent := fmt.Sprintf(`module %s

go 1.21

require github.com/iures/daivplug v0.0.3

// For local development, uncomment and update the path to your local daiv repository:
// replace github.com/iures/daivplug => /absolute/path/to/local/daiv
`, pluginName)

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
daiv plugin install YOUR_GITHUB_USERNAME/%s
` + "```" + `

### From Source

1. Clone the repository:
   ` + "```" + `
   git clone https://github.com/YOUR_GITHUB_USERNAME/%s.git
   cd %s
   ` + "```" + `

2. Build the plugin:
   ` + "```" + `
   make install
   ` + "```" + `
   
   Or manually:
   ` + "```" + `
   go build -o out/%s.so -buildmode=plugin
   daiv plugin install ./out/%s.so
   ` + "```" + `

## Configuration

This plugin requires the following configuration:

- %s.apikey: API key for the service

You can configure these settings when you first run daiv after installing the plugin.

## Usage

After installation, the plugin will be automatically loaded when you start daiv.

## Development

This plugin includes a Makefile with the following commands:

- ` + "`make build`" + `: Build the plugin
- ` + "`make install`" + `: Build and install the plugin
- ` + "`make clean`" + `: Clean build artifacts
- ` + "`make tidy`" + `: Run go mod tidy

`, 
			titleCaser.String(strings.ReplaceAll(pluginName, "-", " ")), 
			pluginName, 
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

		// Create .gitignore file
		gitignoreContent := `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with 'go test -c'
*.test

# Output of the go coverage tool
*.out

# Dependency directories
/vendor/
/go.sum

# Go workspace file
go.work

# IDE specific files
.idea/
.vscode/
*.swp
*.swo

# OS specific files
.DS_Store
Thumbs.db

# Plugin build output
/out/
`

		if err := os.WriteFile(filepath.Join(pluginDir, ".gitignore"), []byte(gitignoreContent), 0644); err != nil {
			return fmt.Errorf("failed to create .gitignore file: %w", err)
		}
		
		// Create Makefile
		makefileContent := fmt.Sprintf(`PLUGIN_NAME=%s

.PHONY: build install clean tidy

install: build
	cp ./out/$(PLUGIN_NAME).so ~/.daiv/plugins/

build: tidy
	go build -o ./out/$(PLUGIN_NAME).so -buildmode=plugin main.go

tidy: clean
	go mod tidy

clean:
	rm -f ./out/$(PLUGIN_NAME).so
	rm -f ~/.daiv/plugins/$(PLUGIN_NAME).so
`, pluginName)

		if err := os.WriteFile(filepath.Join(pluginDir, "Makefile"), []byte(makefileContent), 0644); err != nil {
			return fmt.Errorf("failed to create Makefile: %w", err)
		}

		// Initialize git repository (always)
		// Save current directory
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Change to plugin directory
		if err := os.Chdir(pluginDir); err != nil {
			return fmt.Errorf("failed to change to plugin directory: %w", err)
		}

		// Initialize git repository
		gitInit := exec.Command("git", "init")
		if err := gitInit.Run(); err != nil {
			// Change back to original directory before returning error
			os.Chdir(currentDir)
			return fmt.Errorf("failed to initialize git repository: %w", err)
		}

		// Change back to original directory
		if err := os.Chdir(currentDir); err != nil {
			return fmt.Errorf("failed to change back to original directory: %w", err)
		}
		
		fmt.Printf("Successfully generated plugin template in %s\n", pluginDir)
		fmt.Println("\nPlugin name has been prefixed with 'daiv-' as per convention.")
		fmt.Println("Git repository has been initialized with a .gitignore file.")
		fmt.Println("A Makefile has been created with build, install, clean, and tidy commands.")
		
		fmt.Println("\nNext steps:")
		fmt.Println("1. Implement your plugin functionality in 'main.go'")
		fmt.Printf("2. Build and install your plugin with: cd %s && make install\n", pluginDir)
		
		return nil
	},
}

func init() {
	pluginCmd.AddCommand(createPluginCmd)
	createPluginCmd.Flags().String("dir", "", "Directory where the plugin will be created (default is current directory)")
} 
