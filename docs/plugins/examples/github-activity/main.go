package main

import (
	"github.com/iures/daiv-github-activity/plugin" // Import the plugin package
	plug "github.com/iures/daivplug"
)

// Plugin is exported as a symbol for the daiv plugin system to find
var Plugin plug.Plugin = plugin.New() 
