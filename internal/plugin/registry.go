package plugin

import (
	"fmt"
	"sync"
)

// Registry manages the registration and access of plugins
type Registry struct {
	mu       sync.RWMutex
	Plugins  map[string]Plugin
}

var (
	globalRegistry = &Registry{
		Plugins: make(map[string]Plugin),
	}
)

// GetRegistry returns the global plugin registry
func GetRegistry() *Registry {
	return globalRegistry
}

func (r *Registry) GetStandupPlugins() []StandupPlugin {
	standupPlugins := []StandupPlugin{}

	for _, plugin := range r.Plugins {
		standupPlugin, ok := plugin.(StandupPlugin)
		if ok {
			standupPlugins = append(standupPlugins, standupPlugin)
		}
	}

	return standupPlugins
}

func (r *Registry) Register(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plugin.Name()
	if _, exists := r.Plugins[name]; exists {
		return fmt.Errorf("reporter plugin %s is already registered", name)
	}

	// Then initialize it
	if err := Initialize(plugin); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %w", name, err)
	}

  r.Plugins[name] = plugin

	return nil
}

// Retrieve plugin by name
func (r *Registry) Get(name string) (Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
  plugin, ok := r.Plugins[name]
	return plugin, ok
}

// ShutdownAll gracefully shuts down all plugins
func (r *Registry) ShutdownAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for name, plugin := range r.Plugins {
		if err := plugin.Shutdown(); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown plugin %s: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}
	return nil
}
