package plugin

import (
	"daiv/internal/utils"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
)

// Initialize handles plugin initialization by ensuring all required config is present
func Initialize(plugin Plugin) error {
	configParams := getConfigParams(plugin.Name())
	configKeys := plugin.Manifest().ConfigKeys

	missingConfigKeys := missingConfigKeys(configKeys, configParams)

	if len(missingConfigKeys) > 0 {
		err := promptConfigKeys(missingConfigKeys)
		if err != nil {
			return err
		}

		err = saveChanges(plugin.Name(), missingConfigKeys)
		if err != nil {
			return err
		}

		// After config is saved, call plugin.Initialize() to let the plugin finish setup
		settings := getConfigParams(plugin.Name())
		err = plugin.Initialize(settings)
		if err != nil {
			return err
		}
	}

	return nil
}

// saveChanges saves the changed configuration to the cache directory
func saveChanges(pluginName string, changedConfigKeys []ConfigKey) error {
	if len(changedConfigKeys) == 0 {
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

	for _, key := range changedConfigKeys {
		configKey := fmt.Sprintf("plugins.%s.%s", pluginName, key.Key)
		value := key.Value

		fmt.Println("key", key.Key)
		fmt.Println("value", value)

		if key.Type == ConfigTypeMultiline {
			if value == nil {
				value = []string{} // Handle nil value for multiline config
			} else if strValue, ok := value.(string); ok {
				value = strings.Split(strValue, "\n")
			} else if strValues, ok := value.([]string); ok {
				// Already in the right format
				value = strValues
			}
		}

		viper.Set(configKey, value)
		cacheConfig.Set(configKey, value)
	}

	return cacheConfig.WriteConfigAs(configPath)
}

func promptConfigKeys(missingConfigKeys []ConfigKey) error {
	var inputs []huh.Field
	
	// Create maps to store the values
	stringValues := make(map[string]*string)
	boolValues := make(map[string]*bool)
	textValues := make(map[string]*string)

	for _, key := range missingConfigKeys {
		switch key.Type {
		case ConfigTypeString:
			var value string
			if v, ok := key.Value.(string); ok && v != "" {
				value = v
			}
			stringValues[key.Key] = &value
			inputs = append(inputs, createStringInput(key, &value))
			
		case ConfigTypeBoolean:
			var value bool
			if v, ok := key.Value.(bool); ok {
				value = v
			}
			boolValues[key.Key] = &value
			inputs = append(inputs, createBooleanInput(key, &value))
			
		case ConfigTypeMultiline:
			var value string
			if v, ok := key.Value.([]string); ok && len(v) > 0 {
				value = strings.Join(v, "\n")
			}
			textValues[key.Key] = &value
			inputs = append(inputs, createMultilineInput(key, &value))
			
		default:
			// Default to string input
			var value string
			if v, ok := key.Value.(string); ok && v != "" {
				value = v
			}
			stringValues[key.Key] = &value
			inputs = append(inputs, createStringInput(key, &value))
		}
	}

	if len(inputs) > 0 {
		form := huh.NewForm(huh.NewGroup(inputs...))
		if err := form.Run(); err != nil {
			return err
		}

		// Update the original ConfigKey values
		for i, key := range missingConfigKeys {
			switch key.Type {
			case ConfigTypeString:
				if value, ok := stringValues[key.Key]; ok && value != nil {
					missingConfigKeys[i].Value = *value
				}
			case ConfigTypeBoolean:
				if value, ok := boolValues[key.Key]; ok && value != nil {
					missingConfigKeys[i].Value = *value
				}
			case ConfigTypeMultiline:
				if value, ok := textValues[key.Key]; ok && value != nil {
					// Convert the multiline string back to a string array
					if *value != "" {
						missingConfigKeys[i].Value = strings.Split(*value, "\n")
					} else {
						missingConfigKeys[i].Value = []string{}
					}
				}
			default:
				if value, ok := stringValues[key.Key]; ok && value != nil {
					missingConfigKeys[i].Value = *value
				}
			}
		}
	}

	return nil
}

func createStringInput(key ConfigKey, value *string) huh.Field {
	input := huh.NewInput().
		Key(key.Key).
		Title(key.Name).
		Description(key.Description).
		Value(value)

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

func createBooleanInput(key ConfigKey, value *bool) huh.Field {
	return huh.NewConfirm().
		Key(key.Key).
		Title(key.Name).
		Description(key.Description).
		Value(value)
}

func createMultilineInput(key ConfigKey, value *string) huh.Field {
	return huh.NewText().
		Key(key.Key).
		Title(key.Name).
		Description(key.Description).
		Value(value).
		Lines(5).
		ShowLineNumbers(true)
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
