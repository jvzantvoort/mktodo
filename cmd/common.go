package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/git"
	"github.com/jvzantvoort/mktodo/messages"
	"github.com/spf13/cobra"
)

// loadConfig finds the git repository and loads and validates the
// configuration. The configuration is read from the --config flag when it is
// set (relative to the current directory), otherwise from .mktodo.yml in the
// repository root.
func loadConfig(cmd *cobra.Command) (*git.Repository, *config.Config, error) {
	repo, err := git.FindRepository()
	if err != nil {
		return nil, nil, messages.Errorf("ERR_NOT_IN_GIT_REPO")
	}

	cfgPath := repo.ConfigPath()
	if flag := cmd.Flag("config"); flag != nil && flag.Value.String() != "" {
		if cfgPath, err = filepath.Abs(flag.Value.String()); err != nil {
			return nil, nil, messages.Errorf("ERR_CONFIG_LOAD_FAILED", err)
		}
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, nil, messages.Errorf("ERR_CONFIG_LOAD_FAILED", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, nil, messages.Errorf("ERR_CONFIG_INVALID", err)
	}

	return repo, cfg, nil
}

// confirm prints prompt and reports whether the user answered "y". An empty
// answer or a read error (such as no terminal) counts as "no".
func confirm(prompt string) bool {
	fmt.Print(prompt)
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(response), "y")
}
