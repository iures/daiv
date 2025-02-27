package plugin

import (
	"daiv/internal/utils"
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
)

func Initialize(plugin Plugin) error {
	configParams := getConfigParams(plugin.Name())

	missingConfigKeys := missingConfigKeys(plugin.Manifest().ConfigKeys, configParams)

	if len(missingConfigKeys) > 0 {
		changedConfigKeys, err := promptConfigKeys(missingConfigKeys)
		if err != nil {
			return err
		}

		saveChanges(changedConfigKeys)

		err = plugin.Initialize()
		if err != nil {
			return err
		}
	}

	return nil
}

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
	return huh.NewInput().
		Key(key.Key).
		Title(key.Name).
		Value(nil).
		Validate(func(s string) error {
			return nil
		})
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

func saveConfigChanges(pluginName string, changedConfigKeys map[string]interface{}) error {
	for key, value := range changedConfigKeys {
		viper.Set(fmt.Sprintf("%s.%s", pluginName, key), value)
	}
	return viper.WriteConfig()
}

func getConfigParams(pluginName string) map[string]interface{} {
	configPath := fmt.Sprintf("plugins.%s", pluginName)
	return viper.Sub(configPath).AllSettings()
}
