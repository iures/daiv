# Daiv Plugin System

The Daiv plugin system allows you to extend the functionality of Daiv by creating custom plugins that provide additional context to the LLM. This document explains how to create, build, install, and use plugins with Daiv.

## Table of Contents

- [Overview](#overview)
- [Creating a Plugin](#creating-a-plugin)
- [Plugin Structure](#plugin-structure)
- [Building a Plugin](#building-a-plugin)
- [Installing a Plugin](#installing-a-plugin)
- [Plugin Configuration](#plugin-configuration)
- [Plugin API Reference](#plugin-api-reference)
- [Example Plugins](#example-plugins)
- [Advanced Plugin Development](#advanced-plugin-development)
- [Troubleshooting](#troubleshooting)

## Overview

Daiv plugins are Go shared libraries (`.so` files) that implement the Plugin interface defined in the `github.com/iures/daivplug` package. Plugins can provide additional context to the LLM, such as information from external systems like GitHub, Jira, or other services.

Currently, plugins can provide the following contexts:
- **Standup Context**: Information to include in your daily standup reports

In the future, plugins may support additional contexts as the Daiv ecosystem grows.

## Creating a Plugin

The easiest way to create a new plugin is to use the built-in `daiv plugin create` command:

```bash
daiv plugin create myplugin
```

This will generate a new plugin with the name `daiv-myplugin` in the current directory, with the following structure:

```
daiv-myplugin/
├── main.go                   # Plugin entry point
├── plugin/
│   ├── plugin.go             # Core plugin implementation
│   └── contexts/             # Context providers directory
│       └── standup.go        # Standup context provider
├── out/                      # Compiled plugin output
├── .gitignore                # Git ignore file
├── Makefile                  # Build automation
└── README.md                 # Documentation
```

The command will also:
1. Initialize a Git repository in the plugin directory
2. Create a Makefile for building and installing the plugin
3. Build the plugin for you automatically

## Plugin Structure

A Daiv plugin follows a modular structure designed for maintainability and extensibility:

### Main Entry Point (main.go)

The `main.go` file is the entry point for your plugin. It's very simple and just exports the Plugin interface:

```go
package main

import (
	plug "github.com/iures/daivplug"
	"github.com/yourname/daiv-myplugin/plugin" // Import the plugin package
)

// Plugin is exported as a symbol for the daiv plugin system to find
var Plugin plug.Plugin = plugin.New()
```

### Core Plugin Implementation (plugin/plugin.go)

The `plugin.go` file contains the core plugin implementation, including configuration, lifecycle management, and routing to context providers:

```go
package plugin

import (
	plug "github.com/iures/daivplug"
	"github.com/iures/daivplug/types"
	"github.com/yourname/daiv-myplugin/plugin/contexts" // Import contexts package
)

// MyPlugin implements the Plugin interface
type MyPlugin struct{
	// Add any fields needed by the plugin here
}

// New creates a new instance of the plugin
func New() *MyPlugin {
	return &MyPlugin{}
}

// Name returns the unique identifier for this plugin
func (p *MyPlugin) Name() string {
	return "daiv-myplugin"
}

// Manifest returns the plugin manifest
func (p *MyPlugin) Manifest() *plug.PluginManifest {
	return &plug.PluginManifest{
		ConfigKeys: []plug.ConfigKey{
			{
				Type:        0, // ConfigTypeString
				Key:         "daiv-myplugin.apikey",
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
	// apiKey := settings["daiv-myplugin.apikey"].(string)
	return nil
}

// Shutdown performs cleanup when the plugin is being disabled/removed
func (p *MyPlugin) Shutdown() error {
	// Clean up any resources
	return nil
}

// GetStandupContext implements the StandupPlugin interface
func (p *MyPlugin) GetStandupContext(timeRange types.TimeRange) (types.StandupContext, error) {
	// Delegate to the standup context implementation
	return contexts.GetStandupContext(p.Name(), timeRange)
}
```

### Context Providers (plugin/contexts/standup.go)

Context providers generate specific content for different LLM interactions. Each context is separated into its own file for better organization:

```go
package contexts

import (
	"github.com/iures/daivplug/types"
)

// GetStandupContext generates the standup context for the plugin
func GetStandupContext(pluginName string, timeRange types.TimeRange) (types.StandupContext, error) {
	// Implement your plugin-specific standup context generation
	
	// Example implementation
	return types.StandupContext{
		PluginName: pluginName,
		Content:    "Example content from " + pluginName + " plugin",
	}, nil
}
```

This modular approach allows you to:
1. Keep the main plugin code clean and focused on core functionality
2. Add new context providers without modifying existing code
3. Organize related functionality in separate files
4. Test each context provider independently

## Building a Plugin

To build your plugin, you can use the Makefile generated by the `daiv plugin create` command:

```bash
cd daiv-myplugin
make build
```

This will compile your plugin into a shared library (`daiv-myplugin.so`) in the `out` directory.

## Installing a Plugin

You can install your plugin using the Makefile:

```bash
make install
```

This will copy the compiled plugin to `~/.daiv/plugins/`, where Daiv will automatically detect and load it.

Alternatively, you can use the `daiv plugin install` command:

```bash
daiv plugin install ./out/daiv-myplugin.so
```

## Plugin Configuration

Plugins can define configuration keys using the `Manifest()` method. When a user installs a plugin, Daiv will prompt them to provide values for required configuration keys.

Configuration keys can have the following types:
- **ConfigTypeString**: Simple string input
- **ConfigTypePassword**: Password input that should be masked
- **ConfigTypeMultiline**: Multiline text input
- **ConfigTypeMultiSelect**: Dropdown selection
- **ConfigTypeBoolean**: Boolean toggle

Example configuration key definition:

```go
ConfigKey{
    Type:        ConfigTypeString,
    Key:         "myplugin.apikey",
    Name:        "API Key",
    Description: "API key for the service",
    Required:    true,
    Secret:      true,
    EnvVar:      "MYPLUGIN_API_KEY",
}
```

## Plugin API Reference

### Plugin Interface

All plugins must implement the following interface:

```go
type Plugin interface {
    // Returns the manifest for this plugin
    Manifest() *PluginManifest
    // Name returns the unique identifier for this plugin
    Name() string
    // Initialize sets up the plugin with its configuration
    Initialize(settings map[string]interface{}) error
    // Shutdown performs cleanup when the plugin is being disabled/removed
    Shutdown() error
}
```

### StandupPlugin Interface

If your plugin provides standup context, it should implement the StandupPlugin interface:

```go
type StandupPlugin interface {
    Plugin

    GetStandupContext(timeRange TimeRange) (StandupContext, error)
}
```

### Types

- **TimeRange**: Represents a period for report generation
  ```go
  type TimeRange struct {
      Start time.Time
      End   time.Time
  }
  ```

- **StandupContext**: The context provided for standup reports
  ```go
  type StandupContext struct {
      PluginName string
      Content    string
  }
  ```

## Example Plugins

Here are some example plugins that demonstrate different use cases:

### GitHub Activity Plugin

A plugin that fetches your recent GitHub activity and includes it in your standup report:

```go
// plugin/contexts/standup.go
package contexts

import (
    "fmt"
    "time"
    "github.com/google/go-github/v45/github"
    "github.com/iures/daivplug/types"
    "golang.org/x/oauth2"
)

// GetStandupContext fetches GitHub activity for the standup report
func GetStandupContext(pluginName string, timeRange types.TimeRange) (types.StandupContext, error) {
    // Get GitHub token from environment or configuration
    token := os.Getenv("GITHUB_TOKEN")
    
    // Create GitHub client
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
    tc := oauth2.NewClient(context.Background(), ts)
    client := github.NewClient(tc)
    
    // Get user events
    events, _, err := client.Activity.ListEventsPerformedByUser(
        context.Background(), 
        "your-username", 
        false, 
        &github.ListOptions{},
    )
    if err != nil {
        return types.StandupContext{}, err
    }
    
    // Filter events by time range and format them
    var content string
    for _, event := range events {
        eventTime := event.GetCreatedAt()
        if eventTime.After(timeRange.Start) && eventTime.Before(timeRange.End) {
            content += fmt.Sprintf("- %s on %s at %s\n", 
                event.GetType(), 
                event.GetRepo().GetName(),
                eventTime.Format("15:04"),
            )
        }
    }
    
    return types.StandupContext{
        PluginName: pluginName,
        Content:    content,
    }, nil
}
```

### Weather Plugin

A plugin that includes local weather information in your standup:

```go
// plugin/contexts/standup.go
package contexts

import (
    "fmt"
    "net/http"
    "encoding/json"
    "github.com/iures/daivplug/types"
)

type WeatherResponse struct {
    Current struct {
        TempC float64 `json:"temp_c"`
        Condition struct {
            Text string `json:"text"`
        } `json:"condition"`
    } `json:"current"`
    Location struct {
        Name string `json:"name"`
    } `json:"location"`
}

// GetStandupContext fetches weather information for the standup report
func GetStandupContext(pluginName string, timeRange types.TimeRange) (types.StandupContext, error) {
    // Get API key from environment or configuration
    apiKey := os.Getenv("WEATHER_API_KEY")
    
    // Make API request
    resp, err := http.Get(fmt.Sprintf(
        "https://api.weatherapi.com/v1/current.json?key=%s&q=London&aqi=no", 
        apiKey,
    ))
    if err != nil {
        return types.StandupContext{}, err
    }
    defer resp.Body.Close()
    
    // Parse response
    var weather WeatherResponse
    if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
        return types.StandupContext{}, err
    }
    
    // Format content
    content := fmt.Sprintf(
        "Current weather in %s: %.1f°C, %s", 
        weather.Location.Name,
        weather.Current.TempC,
        weather.Current.Condition.Text,
    )
    
    return types.StandupContext{
        PluginName: pluginName,
        Content:    content,
    }, nil
}
```

## Advanced Plugin Development

### Adding External Dependencies

If your plugin requires external dependencies, you can add them to your `go.mod` file:

```bash
cd daiv-myplugin
go get github.com/some/dependency
```

### Testing Your Plugin

Create tests in a `plugin/plugin_test.go` file to ensure your plugin works correctly:

```go
package plugin

import (
    "testing"
    "time"
    
    "github.com/iures/daivplug/types"
)

func TestGetStandupContext(t *testing.T) {
    p := New()
    
    timeRange := types.TimeRange{
        Start: time.Now().Add(-24 * time.Hour),
        End:   time.Now(),
    }
    
    ctx, err := p.GetStandupContext(timeRange)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    
    if ctx.PluginName != p.Name() {
        t.Errorf("Expected plugin name %s, got %s", p.Name(), ctx.PluginName)
    }
    
    if ctx.Content == "" {
        t.Error("Expected non-empty content")
    }
}
```

### Creating Context-Specific Tests

You can also test individual context providers:

```go
// plugin/contexts/standup_test.go
package contexts

import (
    "testing"
    "time"
    
    "github.com/iures/daivplug/types"
)

func TestGetStandupContext(t *testing.T) {
    timeRange := types.TimeRange{
        Start: time.Now().Add(-24 * time.Hour),
        End:   time.Now(),
    }
    
    ctx, err := GetStandupContext("test-plugin", timeRange)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    
    if ctx.PluginName != "test-plugin" {
        t.Errorf("Expected plugin name %s, got %s", "test-plugin", ctx.PluginName)
    }
    
    if ctx.Content == "" {
        t.Error("Expected non-empty content")
    }
}
```

### Publishing Your Plugin

To make your plugin available to others:
1. Push your plugin to a public GitHub repository
2. Users can install it with `daiv plugin install username/repo-name`

## Troubleshooting

### Plugin Not Loading

If your plugin is not being loaded by Daiv:

1. Make sure it's in the correct location (`~/.daiv/plugins/`)
2. Check that the plugin is compiled for the correct architecture
3. Verify that the plugin exports a `Plugin` variable that implements the Plugin interface
4. Look for error messages in the Daiv logs

### Configuration Issues

If your plugin has configuration issues:

1. Check that your configuration keys are properly defined in the `Manifest()` method
2. Verify that required configuration values are being set
3. Make sure your plugin handles the configuration properly in the `Initialize()` method

### Build Errors

If you encounter errors when building your plugin:

1. Make sure you have Go installed (version 1.18 or later)
2. Verify that your plugin imports the correct version of `github.com/iures/daivplug`
3. Check that your plugin follows the structure expected by the build system

### Common Error Messages

- **"plugin was built with a different version of package X"**: This usually means your plugin was built with a different version of a dependency than what Daiv is using. Try rebuilding your plugin with the correct versions.
- **"plugin exports no symbol named Plugin"**: Make sure your main.go file exports a variable named `Plugin` that implements the Plugin interface.
- **"could not open plugin file"**: Check file permissions and make sure the .so file exists at the specified path.
