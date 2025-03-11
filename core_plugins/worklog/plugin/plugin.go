package plugin

import (
	daivPlugin "github.com/iures/daiv/pkg/plugin"
)

// Name returns the unique identifier for this plugin
func (p *daivPlugin.Plugin) Name() string {
	return "worklog"
}

// Manifest returns the plugin manifest
func (p *daivPlugin.Plugin) Manifest() *daivPlugin.PluginManifest {
	return &daivPlugin.PluginManifest{
		ConfigKeys: []daivPlugin.ConfigKey{
			{
				Type:        daivPlugin.ConfigTypeString,
				Key:         "worklog.apikey",
				Name:        "API Key",
				Description: "API key for the service",
				Required:    true,
			},
			// Add more config keys as needed
		},
	}
}

// Initialize sets up the plugin with its configuration
func (p *daivPlugin.Plugin) Initialize(settings map[string]interface{}) error {
	// Process configuration settings
	// apiKey := settings["worklog.apikey"].(string)
	// TODO: Initialize your plugin with the settings
	return nil
}

// Shutdown performs cleanup when the plugin is being disabled/removed
func (p *daivPlugin.Plugin) Shutdown() error {
	// TODO: Clean up any resources
	return nil
}

// GetStandupContext implements the StandupPlugin interface
func (p *daivPlugin.Plugin) GetStandupContext(timeRange daivPlugin.TimeRange) (daivPlugin.StandupContext, error) {
	// TODO: Implement your plugin-specific standup context generation
	return daivPlugin.StandupContext{
		PluginName: p.Name(),
		Content:    "Example content from worklog plugin",
	}, nil
}
