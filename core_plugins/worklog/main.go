package main

import (
	extPlugin "github.com/iures/daiv-plugin"
)

// WorklogPlugin implements both plugin.Plugin and plugin.StandupPlugin interfaces
type WorklogPlugin struct {
	// Add any fields needed by the plugin here
}

// Plugin is exported as a symbol for the daiv plugin system to find
// It must be of type plugin.Plugin for the plugin system to recognize it
var Plugin extPlugin.Plugin = &WorklogPlugin{}

// Name returns the unique identifier for this plugin
func (p *WorklogPlugin) Name() string {
	return "worklog"
}

// Manifest returns the plugin manifest
func (p *WorklogPlugin) Manifest() *extPlugin.PluginManifest {
	return &extPlugin.PluginManifest{
		ConfigKeys: []extPlugin.ConfigKey{
			{
				Type:        0, // ConfigTypeString
				Key:         "worklog.path",
				Name:        "Worklog Path",
				Description: "The path to the worklog file",
				Required:    true,
			},
		},
	}
}

// Initialize sets up the plugin with its configuration
func (p *WorklogPlugin) Initialize(settings map[string]interface{}) error {
	return nil
}

// Shutdown performs cleanup when the plugin is being disabled/removed
func (p *WorklogPlugin) Shutdown() error {
	return nil
}

// GetStandupContext implements the StandupPlugin interface
func (p *WorklogPlugin) GetStandupContext(timeRange extPlugin.TimeRange) (extPlugin.StandupContext, error) {
	return extPlugin.StandupContext{
		PluginName: p.Name(),
		Content:    "Example content from worklog plugin",
	}, nil
}
