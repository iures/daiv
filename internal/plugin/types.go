package plugin

import (
	"context"
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

// Plugin defines the base interface that all plugins must implement
type Plugin interface {
	// Name returns the unique identifier for this plugin
	Name() string
	// Initialize sets up the plugin with its configuration
	Initialize(config map[string]interface{}) error
	// Shutdown performs cleanup when the plugin is being disabled/removed
	Shutdown() error
}

// Reporter defines the interface for plugins that generate reports
type Reporter interface {
	Plugin
	// GenerateReport produces a report for the given time range
	GenerateReport(ctx context.Context, timeRange TimeRange) (Report, error)
} 
