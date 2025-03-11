package main

import (
	"time"
)

// Plugin is exported as a symbol for the daiv plugin system to find
var Plugin plugin = &WorklogPlugin{}

// TimeRange represents a period for report generation
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// StandupContext represents the standup information
type StandupContext struct {
	PluginName string
	Content    string
}

// ConfigKey represents configuration metadata
type ConfigKey struct {
	Type        int // Using int to represent ConfigType
	Key         string
	Value       interface{}
	Name        string
	Description string
	Required    bool
	Secret      bool
	EnvVar      string
}

// PluginManifest represents plugin metadata
type PluginManifest struct {
	ConfigKeys []ConfigKey
}

// Plugin interface matches the one in daiv/pkg/plugin
type plugin interface {
	Name() string
	Manifest() *PluginManifest
	Initialize(map[string]interface{}) error
	Shutdown() error
	GetStandupContext(TimeRange) (StandupContext, error)
}

// WorklogPlugin implements the Plugin interface
type WorklogPlugin struct{}

// Name returns the unique identifier for this plugin
func (p *WorklogPlugin) Name() string {
	return "worklog"
}

// Manifest returns the plugin manifest
func (p *WorklogPlugin) Manifest() *PluginManifest {
	return &PluginManifest{
		ConfigKeys: []ConfigKey{
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
func (p *WorklogPlugin) GetStandupContext(timeRange TimeRange) (StandupContext, error) {
	return StandupContext{
		PluginName: p.Name(),
		Content:    "Example content from worklog plugin",
	}, nil
}

func main() {} // Required for building as a plugin
