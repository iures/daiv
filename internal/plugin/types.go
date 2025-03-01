package plugin

import (
	"time"
)

// TimeRange represents a period for report generation
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// Report represents the output from a plugin
type Report struct {
	PluginName string
	Content    string
	Metadata   map[string]interface{}
}

type ConfigKey struct {
	Type
	Key         string
	Name        string
	Description string
	Required    bool
	Secret      bool
	PluginName  string // Name of the plugin this config belongs to
}

type PluginManifest struct {
	ConfigKeys []ConfigKey
}

// Plugin defines the base interface that all plugins must implement
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

// Reporter defines the interface for plugins that generate reports
type StandupPlugin interface {
	Plugin

	GetStandupContext(timeRange TimeRange) (string, error)
}
