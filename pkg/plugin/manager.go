package plugin

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	pluginlib "plugin" // Go standard library plugin
	"strings"
)

// PluginManager handles downloading, installing, and loading plugins
type PluginManager struct {
	pluginsDir string
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(pluginsDir string) (*PluginManager, error) {
	// Create plugins directory if it doesn't exist
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create plugins directory: %w", err)
	}
	
	return &PluginManager{
		pluginsDir: pluginsDir,
	}, nil
}

// InstallFromGitHub downloads and installs a plugin from a GitHub repository
func (pm *PluginManager) InstallFromGitHub(repo string, version string) error {
	// Handle GitHub repo format: username/repo
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid GitHub repository format, expected 'username/repo'")
	}
	
	username, repoName := parts[0], parts[1]
	
	// Create temp directory for cloning
	tempDir, err := os.MkdirTemp("", "daiv-plugin-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Clone the repository
	gitCmd := exec.Command("git", "clone", fmt.Sprintf("https://github.com/%s/%s.git", username, repoName), tempDir)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	
	// If version is specified, checkout that version
	if version != "" {
		checkoutCmd := exec.Command("git", "checkout", version)
		checkoutCmd.Dir = tempDir
		checkoutCmd.Stdout = os.Stdout
		checkoutCmd.Stderr = os.Stderr
		if err := checkoutCmd.Run(); err != nil {
			return fmt.Errorf("failed to checkout version %s: %w", version, err)
		}
	}
	
	// Build the plugin
	buildCmd := exec.Command("go", "build", "-buildmode=plugin", "-o", fmt.Sprintf("%s.so", repoName))
	buildCmd.Dir = tempDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build plugin: %w", err)
	}
	
	// Copy the built plugin to plugins directory
	pluginPath := filepath.Join(tempDir, fmt.Sprintf("%s.so", repoName))
	destPath := filepath.Join(pm.pluginsDir, fmt.Sprintf("%s.so", repoName))
	
	return copyFile(pluginPath, destPath)
}

// InstallFromURL downloads and installs a plugin from a direct URL
func (pm *PluginManager) InstallFromURL(url string) error {
	// Extract the filename from the URL
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]
	
	if !strings.HasSuffix(filename, ".so") && !strings.HasSuffix(filename, ".dll") {
		return fmt.Errorf("plugin file must have .so or .dll extension")
	}
	
	// Download the file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download plugin: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download plugin: HTTP status %d", resp.StatusCode)
	}
	
	// Create the destination file
	destPath := filepath.Join(pm.pluginsDir, filename)
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()
	
	// Copy the downloaded content to the destination file
	if _, err := io.Copy(destFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save plugin file: %w", err)
	}
	
	fmt.Printf("Plugin installed to: %s\n", destPath)
	return nil
}

// InstallFromLocalFile copies a local plugin file to the plugins directory
func (pm *PluginManager) InstallFromLocalFile(filePath string) error {
	// Verify the file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to access plugin file: %w", err)
	}
	
	if fileInfo.IsDir() {
		return fmt.Errorf("expected a file, got a directory")
	}
	
	// Check file extension
	if !strings.HasSuffix(filePath, ".so") && !strings.HasSuffix(filePath, ".dll") {
		return fmt.Errorf("plugin file must have .so or .dll extension")
	}
	
	// Get the filename
	filename := filepath.Base(filePath)
	
	// Create the destination path
	destPath := filepath.Join(pm.pluginsDir, filename)
	
	// Copy the file
	if err := copyFile(filePath, destPath); err != nil {
		return fmt.Errorf("failed to copy plugin file: %w", err)
	}
	
	fmt.Printf("Plugin installed to: %s\n", destPath)
	return nil
}

// LoadPlugins loads all plugins from the plugins directory
func (pm *PluginManager) LoadPlugins() ([]Plugin, error) {
	var plugins []Plugin
	
	// Ensure directory exists
	if _, err := os.Stat(pm.pluginsDir); os.IsNotExist(err) {
		return plugins, nil // No plugins directory, return empty list
	}
	
	// Read all .so files in the plugins directory
	entries, err := os.ReadDir(pm.pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugins directory: %w", err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories
		}
		
		name := entry.Name()
		ext := filepath.Ext(name)
		if ext != ".so" && ext != ".dll" {
			continue // Skip non-plugin files
		}
		
		// Load the plugin
		p, err := pluginlib.Open(filepath.Join(pm.pluginsDir, name))
		if err != nil {
			fmt.Printf("Warning: Failed to load plugin %s: %v\n", name, err)
			continue
		}
		
		// Look up the Plugin symbol
		symPlugin, err := p.Lookup("Plugin")
		if err != nil {
			fmt.Printf("Warning: Plugin %s does not export 'Plugin' symbol: %v\n", name, err)
			continue
		}
		
		// Cast to the Plugin interface
		plug, ok := symPlugin.(Plugin)
		if !ok {
			fmt.Printf("Warning: Plugin %s's exported 'Plugin' is not of type Plugin\n", name)
			continue
		}
		
		// Add to the list of plugins
		plugins = append(plugins, plug)
	}
	
	return plugins, nil
}

// Helper function to copy a file
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()
	
	if _, err = io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	
	return nil
} 
