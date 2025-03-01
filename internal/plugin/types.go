package plugin

import (
	"time"
)

// ConfigType defines the type of configuration input
type ConfigType int

const (
	// ConfigTypeString represents a simple string input
	ConfigTypeString ConfigType = iota
	// ConfigTypePassword represents a password input that should be masked
	ConfigTypePassword
	// ConfigTypeMultiline represents a multiline text input
	ConfigTypeMultiline
	// ConfigTypeSelect represents a dropdown selection
	ConfigTypeSelect
	// ConfigTypeMultiSelect represents a multi-selection input
	ConfigTypeMultiSelect
	// ConfigTypeBoolean represents a boolean toggle
	ConfigTypeBoolean
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
	Type        ConfigType
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
