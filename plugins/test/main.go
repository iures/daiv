package main

import (
	plug "github.com/iures/daivplug"
)

// TestPlugin implements the Plugin interface
type TestPlugin struct{
	// Add any fields needed by the plugin here
}

// Plugin is exported as a symbol for the daiv plugin system to find
var Plugin plug.Plugin = &TestPlugin{}

// Name returns the unique identifier for this plugin
func (p *TestPlugin) Name() string {
	return "test"
}

// Manifest returns the plugin manifest
func (p *TestPlugin) Manifest() *plug.PluginManifest {
	return &plug.PluginManifest{
		ConfigKeys: []plug.ConfigKey{
			{
				Type:        0, // ConfigTypeString
				Key:         "test.apikey",
				Name:        "API Key",
				Description: "API key for the service",
				Required:    true,
			},
			// Add more config keys as needed
		},
	}
}

// Initialize sets up the plugin with its configuration
func (p *TestPlugin) Initialize(settings map[string]interface{}) error {
	// Process configuration settings
	// apiKey := settings["test.apikey"].(string)
	// TODO: Initialize your plugin with the settings
	return nil
}

// Shutdown performs cleanup when the plugin is being disabled/removed
func (p *TestPlugin) Shutdown() error {
	// TODO: Clean up any resources
	return nil
}

// GetStandupContext implements the StandupPlugin interface
func (p *TestPlugin) GetStandupContext(timeRange plug.TimeRange) (plug.StandupContext, error) {
	// TODO: Implement your plugin-specific standup context generation
	return plug.StandupContext{
		PluginName: p.Name(),
		Content:    "Example content from test plugin",
	}, nil
}
