package github

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
)

type GithubConfig struct {
	Username     string
	Organization string
	Repositories []string
}

func (c *GithubConfig) IsConfigured() bool {
	return c.Username != "" && c.Organization != ""
}

func GetGithubConfig() (*GithubConfig, error) {
	config := GithubConfig{
		Username:     viper.GetString("github.username"),
		Organization: viper.GetString("github.organization"),
		Repositories: viper.GetStringSlice("github.repositories"),
	}
	return &config, nil
}

func newInput(key, title string, value *string) huh.Field {
	return huh.NewInput().
		Key(key).
		Title(title).
		Value(value).
		Validate(func(s string) error {
			if s == "" {
				return fmt.Errorf("%s is required", key)
			}
			return nil
		})
}

func ConfigPrompt(config *GithubConfig) ([]huh.Field, error) {
	var inputs []huh.Field

	if config.Username == "" {
		inputs = append(inputs, newInput("github.username", "GitHub Username", &config.Username))
	}

	if config.Organization == "" {
		inputs = append(inputs, newInput("github.organization", "GitHub Organization", &config.Organization))
	}

	// For repositories, we won't prompt since they're optional and might be better
	// configured directly in the config file due to their list nature

	if len(inputs) > 0 {
		form := huh.NewForm(huh.NewGroup(inputs...))
		if err := form.Run(); err != nil {
			return nil, err
		}
	}

	return inputs, nil
} 
