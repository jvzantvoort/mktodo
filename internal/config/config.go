// Package config loads and validates the .mktodo.yml configuration.
package config

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

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
	if err := checkKeys(v.AllSettings()); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

var (
	topLevelKeys = []string{"todofile", "projects"}
	projectKeys  = []string{"name", "title", "parent"}
)

// checkKeys refuses keys mktodo doesn't know, so a typo such as "projets"
// is reported instead of silently ignored (spec: validate YAML strictly).
func checkKeys(settings map[string]interface{}) error {
	problems := &ValidationError{}
	for _, key := range unknownKeys(settings, topLevelKeys) {
		problems.add("Unknown key '%s' (expected: %s)", key, strings.Join(topLevelKeys, ", "))
	}

	projects, _ := settings["projects"].([]interface{})
	for i, entry := range projects {
		fields := toStringMap(entry)
		if fields == nil {
			problems.add("Project at position %d is not a list of keys", i+1)
			continue
		}
		label := fmt.Sprintf("Project at position %d", i+1)
		if name, ok := fields["name"].(string); ok && name != "" {
			label = fmt.Sprintf("Project '%s'", name)
		}
		for _, key := range unknownKeys(fields, projectKeys) {
			problems.add("%s: unknown key '%s' (expected: %s)", label, key, strings.Join(projectKeys, ", "))
		}
	}
	return problems.orNil()
}

func unknownKeys(fields map[string]interface{}, known []string) []string {
	var unknown []string
	for key := range fields {
		if !slices.Contains(known, key) {
			unknown = append(unknown, key)
		}
	}
	sort.Strings(unknown)
	return unknown
}

func toStringMap(v interface{}) map[string]interface{} {
	switch m := v.(type) {
	case map[string]interface{}:
		return m
	case map[interface{}]interface{}:
		out := make(map[string]interface{}, len(m))
		for k, val := range m {
			out[strings.ToLower(fmt.Sprint(k))] = val
		}
		return out
	}
	return nil
}
