package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/git"
	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/project"
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
		return nil, nil, gitError(messages.Errorf("ERR_NOT_IN_GIT_REPO"))
	}

	cfgPath := repo.ConfigPath()
	if flag := cmd.Flag("config"); flag != nil && flag.Value.String() != "" {
		if cfgPath, err = filepath.Abs(flag.Value.String()); err != nil {
			return nil, nil, configError(messages.Errorf("ERR_CONFIG_LOAD_FAILED", err))
		}
	}

	if _, statErr := os.Stat(cfgPath); errors.Is(statErr, fs.ErrNotExist) {
		return nil, nil, configError(messages.Errorf("ERR_CONFIG_NOT_FOUND", cfgPath))
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, nil, configError(messages.Errorf("ERR_CONFIG_LOAD_FAILED", err))
	}
	if err := cfg.Validate(); err != nil {
		return nil, nil, configError(messages.Errorf("ERR_CONFIG_INVALID", err))
	}

	return repo, cfg, nil
}

// loadDocument loads the TODO file named in the configuration.
func loadDocument(repo *git.Repository, cfg *config.Config) (*markdown.Document, error) {
	todoPath := repo.ResolvePath(cfg.TodoFile)
	doc, err := markdown.LoadDocument(todoPath, cfg)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, fileError(messages.Errorf("ERR_FILE_NOT_FOUND", cfg.TodoFile))
	case err != nil:
		return nil, fileError(messages.Errorf("ERR_FILE_READ_FAILED", cfg.TodoFile, err))
	}
	return doc, nil
}

// saveDocument writes the document back to its file.
func saveDocument(doc *markdown.Document) error {
	if err := markdown.SaveDocument(doc.Path, doc); err != nil {
		return fileError(messages.Errorf("ERR_FILE_WRITE_FAILED", doc.Path, err))
	}
	return nil
}

// findProject looks up a project by dotted path, listing the valid
// projects when it doesn't exist.
func findProject(doc *markdown.Document, path string) (*project.Project, error) {
	proj, err := project.FindByPath(doc.Projects, path)
	if err != nil {
		return nil, usageError(messages.Errorf("ERR_PROJECT_NOT_FOUND", path, getAvailableProjects(doc.Projects)))
	}
	return proj, nil
}

// getAvailableProjects returns the project paths in configuration order.
func getAvailableProjects(projects map[string]*project.Project) string {
	var names []string
	for _, proj := range project.Sorted(projects) {
		names = append(names, proj.FullPath())
	}
	return strings.Join(names, ", ")
}

// projectItems returns the items in the project at path, or every item when
// path is empty.
func projectItems(doc *markdown.Document, path string) ([]*markdown.Item, error) {
	if path == "" {
		return doc.Items, nil
	}
	proj, err := findProject(doc, path)
	if err != nil {
		return nil, err
	}
	return doc.GetItemsByProject(proj), nil
}

// findOneItem returns the single item matching query, in the project at
// projectPath if one is given. No match or several matches is a usage error.
func findOneItem(doc *markdown.Document, projectPath, query string) (*markdown.Item, error) {
	items, err := projectItems(doc, projectPath)
	if err != nil {
		return nil, err
	}

	candidates := markdown.MatchItems(items, query)
	switch len(candidates) {
	case 0:
		return nil, usageError(messages.Errorf("ERR_ITEM_NOT_FOUND", query))
	case 1:
		return candidates[0], nil
	}

	var list strings.Builder
	for i, item := range candidates {
		fmt.Fprintf(&list, "  %d. %s\n", i+1, item.String())
	}
	return nil, usageError(messages.Errorf("ERR_ITEM_AMBIGUOUS", query, strings.TrimRight(list.String(), "\n")))
}

// validateFormat checks an output format flag against the allowed values.
func validateFormat(format string, allowed ...string) error {
	for _, a := range allowed {
		if format == a {
			return nil
		}
	}
	return usageError(messages.Errorf("ERR_INVALID_FORMAT", format, strings.Join(allowed, ", ")))
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
