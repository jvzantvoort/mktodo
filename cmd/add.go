package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/project"
	"github.com/jvzantvoort/mktodo/messages"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add <description>",
	Aliases: []string{"create"},
	Short:   "Add a new TODO item",
	Long: `Add a new TODO item to the specified project.

Examples:
  mktodo add "Complete documentation"
  mktodo add -p lego.technic "Build crane model"
  mktodo add "FIXME: Security vulnerability"`,
	Args: usageArgs(cobra.MinimumNArgs(1)),
	RunE: runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
	repo, cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	// Get flags
	projectPath, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}
	skipConfirm, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}

	description, err := checkDescription(strings.Join(args, " "))
	if err != nil {
		return err
	}
	path, err := todoPath(repo, cfg)
	if err != nil {
		return err
	}

	doc, err := loadOrCreateDocument(path, cfg, skipConfirm)
	if err != nil {
		return err
	}

	proj, err := findProject(doc, projectPath)
	if err != nil {
		return err
	}

	// The file exists but has no header for this project: ask before adding one.
	if !doc.HasSection(proj) && !skipConfirm &&
		!confirm(messages.Prompt("PROMPT_CREATE_SECTION", proj.Title, cfg.TodoFile)) {
		return messages.Errorf("ERR_OPERATION_CANCELLED")
	}

	item, err := doc.AddItem(proj, description)
	if err != nil {
		return messages.Errorf("ERR_OPERATION_FAILED", err)
	}

	if err := saveDocument(doc); err != nil {
		return err
	}

	// Success message
	fmt.Println(messages.Prompt("MSG_ADDED", item.Description))
	if item.IsFIXME {
		fmt.Println(messages.Prompt("MSG_MARKED_FIXME"))
	}
	fmt.Println(messages.Prompt("MSG_TO_PROJECT", proj.Title))

	return nil
}

// loadOrCreateDocument loads the TODO file. If it doesn't exist, it asks
// before creating it (unless skipConfirm) and starts a document with a
// header for every configured project.
func loadOrCreateDocument(todoPath string, cfg *config.Config, skipConfirm bool) (*markdown.Document, error) {
	doc, err := markdown.LoadDocument(todoPath, cfg)
	if err == nil {
		return doc, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return nil, fileError(messages.Errorf("ERR_FILE_READ_FAILED", todoPath, err))
	}

	if !skipConfirm && !confirm(messages.Prompt("PROMPT_CREATE_FILE", cfg.TodoFile)) {
		return nil, messages.Errorf("ERR_OPERATION_CANCELLED")
	}

	projects, err := project.BuildHierarchy(cfg)
	if err != nil {
		return nil, configError(messages.Errorf("ERR_CONFIG_INVALID", err))
	}
	doc = markdown.NewDocument(todoPath, projects)
	for _, proj := range project.Sorted(projects) {
		if err := doc.EnsureSection(proj); err != nil {
			return nil, messages.Errorf("ERR_OPERATION_FAILED", err)
		}
	}
	return doc, nil
}

func init() {
	addCmd.Flags().StringP("project", "p", "default", "project path (e.g., lego.technic)")
	addCmd.Flags().BoolP("yes", "y", false, "skip confirmations")
}
