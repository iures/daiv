package plugin

import (
	"fmt"
	"sync"
)

// Registry manages the registration and access of plugins
type Registry struct {
	mu       sync.RWMutex
	plugins  map[string]Plugin
	reporters map[string]Reporter
}

var (
	globalRegistry = &Registry{
		plugins:   make(map[string]Plugin),
		reporters: make(map[string]Reporter),
	}
)

// GetRegistry returns the global plugin registry
func GetRegistry() *Registry {
	return globalRegistry
}

// RegisterReporter adds a new reporter plugin to the registry
func (r *Registry) RegisterReporter(reporter Reporter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := reporter.Name()
	if _, exists := r.reporters[name]; exists {
		return fmt.Errorf("reporter plugin %s is already registered", name)
	}

	r.reporters[name] = reporter
	r.plugins[name] = reporter

	Initialize(reporter)

	return nil
}

// GetReporter retrieves a reporter plugin by name
func (r *Registry) GetReporter(name string) (Reporter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	reporter, exists := r.reporters[name]
	return reporter, exists
}

// GetEnabledReporters returns all enabled reporter plugins
func (r *Registry) GetEnabledReporters() []Reporter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var enabledReporters []Reporter
	for _, reporter := range r.reporters {
		enabledReporters = append(enabledReporters, reporter)
	}
	return enabledReporters
}

// ShutdownAll gracefully shuts down all plugins
func (r *Registry) ShutdownAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for name, plugin := range r.plugins {
		if err := plugin.Shutdown(); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown plugin %s: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}
	return nil
} 
