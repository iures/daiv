package github

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
)

func InitializeGithub() error {
	config, err := GetGithubConfig()
	if err != nil {
		return err
	}

	if !config.IsConfigured() {
		inputs, err := ConfigPrompt(config)
		if err != nil {
			return err
		}

		if len(inputs) > 0 {
			if err := saveChanges(inputs); err != nil {
				return err
			}
		}
	}

	return nil
}

func saveChanges(inputs []huh.Field) error {
	if len(inputs) == 0 {
		return nil
	}

	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}

	cacheConfig := viper.New()
	configPath := filepath.Join(cacheDir, "config.yaml")
	cacheConfig.SetConfigFile(configPath)
	cacheConfig.ReadInConfig()

	for _, input := range inputs {
		viper.Set(input.GetKey(), input.GetValue())
		cacheConfig.Set(input.GetKey(), input.GetValue())
	}

	return cacheConfig.WriteConfigAs(configPath)
}

func getCacheDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	daivDir := filepath.Join(cacheDir, "daiv")
	if err := os.MkdirAll(daivDir, 0755); err != nil {
		return "", err
	}
	return daivDir, nil
} 
