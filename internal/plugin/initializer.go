package plugin

import (
	"daiv/internal/utils"
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
)

// Initialize handles plugin initialization by ensuring all required config is present
func Initialize(plugin Plugin) error {
	configParams := getConfigParams(plugin.Name())
	
	// Set plugin name for all config keys
	configKeys := plugin.Manifest().ConfigKeys
	for i := range configKeys {
		configKeys[i].PluginName = plugin.Name()
	}
	
	missingConfigKeys := missingConfigKeys(configKeys, configParams)

	if len(missingConfigKeys) > 0 {
		changedConfigKeys, err := promptConfigKeys(missingConfigKeys)
		if err != nil {
			return err
		}

		err = saveChanges(changedConfigKeys)
		if err != nil {
			return err
		}

		// After config is saved, call plugin.Initialize() to let the plugin finish setup
		settings := getPluginSettings(plugin.Name())
		err = plugin.Initialize(settings)
		if err != nil {
			return err
		}
	}

	return nil
}

func getPluginSettings(pluginName string) map[string]interface{} {
	return getConfigParams(pluginName)
}

// saveChanges saves the changed configuration to the cache directory
func saveChanges(inputs []huh.Field) error {
	if len(inputs) == 0 {
		return nil
	}

	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return err
	}

	cacheConfig := viper.New()
	configPath := filepath.Join(cacheDir, "config.yaml")
	cacheConfig.SetConfigFile(configPath)
	cacheConfig.ReadInConfig()

	for _, input := range inputs {
		// Set in both viper instances to ensure consistency
		viper.Set(input.GetKey(), input.GetValue())
		cacheConfig.Set(input.GetKey(), input.GetValue())
	}

	return cacheConfig.WriteConfigAs(configPath)
}

func promptConfigKeys(missingConfigKeys []ConfigKey) ([]huh.Field, error) {
	var inputs []huh.Field

	for _, key := range missingConfigKeys {
		inputs = append(inputs, promptConfigKey(key))
	}

	if len(inputs) > 0 {
		form := huh.NewForm(huh.NewGroup(inputs...))
		if err := form.Run(); err != nil {
			return nil, err
		}
	}

	return inputs, nil
}

func promptConfigKey(key ConfigKey) huh.Field {
	value := ""
	input := huh.NewInput().
		Key(fmt.Sprintf("plugins.%s.%s", key.PluginName, key.Key)).
		Title(key.Name).
		Value(&value)
	
	if key.Required {
		input = input.Validate(func(s string) error {
			if s == "" {
				return fmt.Errorf("this field is required")
			}
			return nil
		})
	}
	
	return input
}

func missingConfigKeys(configKeys []ConfigKey, configParams map[string]interface{}) []ConfigKey {
	missingKeys := []ConfigKey{}

	for _, key := range configKeys {
		if _, ok := configParams[key.Key]; !ok && key.Required {
			missingKeys = append(missingKeys, key)
		}
	}

	return missingKeys
}

func getConfigParams(pluginName string) map[string]interface{} {
	configPath := fmt.Sprintf("plugins.%s", pluginName)
	sub := viper.Sub(configPath)

	if sub == nil {
		return make(map[string]interface{})
	}

	return sub.AllSettings()
}
