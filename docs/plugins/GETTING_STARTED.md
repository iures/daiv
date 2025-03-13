# Getting Started with Daiv Plugins

This guide will walk you through creating your first Daiv plugin from scratch. By the end, you'll have a working plugin that provides context to the Daiv LLM.

## Prerequisites

Before you begin, make sure you have:

- Go installed (version 1.18 or later)
- Daiv installed and configured
- Basic knowledge of Go programming

## Step 1: Create a New Plugin

The easiest way to create a new plugin is to use the built-in `daiv plugin create` command:

```bash
daiv plugin create hello-world
```

This will generate a new plugin with the name `daiv-hello-world` in the current directory. The command will:

1. Create the plugin directory structure
2. Initialize a Git repository
3. Create a Makefile for building and installing
4. Build the plugin automatically

## Step 2: Explore the Generated Files

Let's take a look at the files that were generated:

```
daiv-hello-world/
├── main.go                   # Plugin entry point
├── plugin/
│   ├── plugin.go             # Core plugin implementation
│   └── contexts/             # Context providers directory
│       └── standup.go        # Standup context provider
├── out/                      # Compiled plugin output
├── .gitignore                # Git ignore file
├── Makefile                  # Build automation
└── README.md                 # Documentation
```

The most important files are:

- **main.go**: The entry point that exports the Plugin interface
- **plugin/plugin.go**: The core plugin implementation
- **plugin/contexts/standup.go**: The standup context provider

## Step 3: Customize the Plugin

Let's modify the standup context provider to return a custom message. Open `plugin/contexts/standup.go` and update the `GetStandupContext` function:

```go
// GetStandupContext generates the standup context for the plugin
func GetStandupContext(pluginName string, timeRange types.TimeRange) (types.StandupContext, error) {
    // Create a custom message
    message := fmt.Sprintf(
        "Hello from %s! Today is %s.\n\n"+
        "This is my first Daiv plugin. It provides context for the time range:\n"+
        "- Start: %s\n"+
        "- End: %s",
        pluginName,
        time.Now().Format("Monday, January 2, 2006"),
        timeRange.Start.Format("15:04:05"),
        timeRange.End.Format("15:04:05"),
    )
    
    return types.StandupContext{
        PluginName: pluginName,
        Content:    message,
    }, nil
}
```

Don't forget to add the necessary imports at the top of the file:

```go
import (
    "fmt"
    "time"
    "github.com/iures/daivplug/types"
)
```

## Step 4: Build and Install the Plugin

Now that you've customized your plugin, let's build and install it:

```bash
cd daiv-hello-world
make install
```

This will compile your plugin and install it to `~/.daiv/plugins/`, where Daiv will automatically detect and load it.

## Step 5: Test Your Plugin

To test your plugin, run the Daiv standup command:

```bash
daiv standup
```

You should see your custom message included in the standup report!

## Next Steps

Now that you've created your first plugin, here are some ideas to enhance it:

### Add Configuration

Update the `Manifest()` method in `plugin/plugin.go` to add configuration options:

```go
// Manifest returns the plugin manifest
func (p *HelloWorldPlugin) Manifest() *plug.PluginManifest {
    return &plug.PluginManifest{
        ConfigKeys: []plug.ConfigKey{
            {
                Type:        0, // ConfigTypeString
                Key:         "daiv-hello-world.username",
                Name:        "Your Name",
                Description: "Your name to personalize the greeting",
                Required:    true,
            },
        },
    }
}
```

Then update your `Initialize()` method to store the configuration:

```go
// Initialize sets up the plugin with its configuration
func (p *HelloWorldPlugin) Initialize(settings map[string]interface{}) error {
    // Get the username from settings
    if username, ok := settings["daiv-hello-world.username"].(string); ok {
        p.username = username
    }
    return nil
}
```

Don't forget to add the `username` field to your plugin struct:

```go
// HelloWorldPlugin implements the Plugin interface
type HelloWorldPlugin struct{
    username string
}
```

### Fetch External Data

Enhance your plugin to fetch data from an external API. For example, you could add weather information to your standup report:

```go
// GetStandupContext generates the standup context for the plugin
func GetStandupContext(pluginName string, timeRange types.TimeRange) (types.StandupContext, error) {
    // Fetch weather data
    weather, err := fetchWeatherData()
    if err != nil {
        return types.StandupContext{}, err
    }
    
    // Create a custom message with weather information
    message := fmt.Sprintf(
        "Hello from %s! Today is %s.\n\n"+
        "Current weather: %s, %.1f°C\n\n"+
        "This is my first Daiv plugin. It provides context for the time range:\n"+
        "- Start: %s\n"+
        "- End: %s",
        pluginName,
        time.Now().Format("Monday, January 2, 2006"),
        weather.Condition,
        weather.Temperature,
        timeRange.Start.Format("15:04:05"),
        timeRange.End.Format("15:04:05"),
    )
    
    return types.StandupContext{
        PluginName: pluginName,
        Content:    message,
    }, nil
}

// fetchWeatherData fetches weather data from an API
func fetchWeatherData() (*WeatherData, error) {
    // Make API request
    // Parse response
    // Return weather data
}

// WeatherData represents weather information
type WeatherData struct {
    Temperature float64
    Condition   string
}
```

## Learn More

For more information about developing Daiv plugins, check out:

- [Plugin System Documentation](README.md)
- [Plugin API Reference](README.md#plugin-api-reference)
- [Example Plugins](README.md#example-plugins)
- [Advanced Plugin Development](README.md#advanced-plugin-development) 
