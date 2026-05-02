package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// Config represents the mktodo configuration
type Config struct {
	TodoFile string          `yaml:"todofile" mapstructure:"todofile"`
	Projects []ProjectConfig `yaml:"projects" mapstructure:"projects"`
}

// ProjectConfig represents a project configuration
type ProjectConfig struct {
	Name   string  `yaml:"name" mapstructure:"name"`
	Title  string  `yaml:"title" mapstructure:"title"`
	Parent *string `yaml:"parent" mapstructure:"parent"`
}

// Load loads configuration from the specified file
func Load(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", path)
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Set defaults
	v.SetDefault("todofile", "README.md")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
