package config

import (
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		wantErr    bool
		validate   func(*testing.T, *Config)
	}{
		{
			name:       "simple config",
			configFile: "simple.yml",
			wantErr:    false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.TodoFile != "README.md" {
					t.Errorf("TodoFile = %s, want README.md", cfg.TodoFile)
				}
				if len(cfg.Projects) != 1 {
					t.Errorf("len(Projects) = %d, want 1", len(cfg.Projects))
				}
				if len(cfg.Projects) > 0 {
					if cfg.Projects[0].Name != "default" {
						t.Errorf("Projects[0].Name = %s, want default", cfg.Projects[0].Name)
					}
					if cfg.Projects[0].Title != "TODO" {
						t.Errorf("Projects[0].Title = %s, want TODO", cfg.Projects[0].Title)
					}
				}
			},
		},
		{
			name:       "nested config",
			configFile: "nested.yml",
			wantErr:    false,
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.Projects) != 3 {
					t.Errorf("len(Projects) = %d, want 3", len(cfg.Projects))
				}
			},
		},
		{
			name:       "invalid YAML",
			configFile: "invalid.yml",
			wantErr:    true,
			validate:   nil,
		},
		{
			name:       "missing file",
			configFile: "nonexistent.yml",
			wantErr:    true,
			validate:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("../../testdata/configs", tt.configFile)
			cfg, err := Load(path)

			if tt.wantErr {
				if err == nil {
					t.Error("Load() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Load() unexpected error: %v", err)
				return
			}

			if cfg == nil {
				t.Error("Load() returned nil config")
				return
			}

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestConfigDefaults(t *testing.T) {
	// Create a minimal config file
	path := filepath.Join("../../testdata/configs", "simple.yml")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify default is applied
	if cfg.TodoFile == "" {
		t.Error("TodoFile should have default value")
	}
}
