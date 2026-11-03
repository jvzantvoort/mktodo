package cmd

import (
	"fmt"
	"strings"

	"github.com/jvzantvoort/mktodo/messages"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <description>",
	Aliases: []string{"rm", "destroy"},
	Short:   "Remove a TODO item",
	Long: `Remove a TODO item from the specified project.

Examples:
  mktodo remove "Complete documentation"
  mktodo rm -p lego "Old task"
  mktodo rm -p lego.technic "Build crane" --yes`,
	Args: usageArgs(cobra.MinimumNArgs(1)),
	RunE: runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
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

	// Build query from args
	query := strings.Join(args, " ")

	doc, err := loadDocument(repo, cfg)
	if err != nil {
		return err
	}

	item, err := findOneItem(doc, projectPath, query)
	if err != nil {
		return err
	}

	if !skipConfirm && !confirm(messages.Prompt("PROMPT_CONFIRM_REMOVE", item.Description)) {
		return messages.Errorf("ERR_OPERATION_CANCELLED")
	}

	if err := doc.RemoveItem(item); err != nil {
		return messages.Errorf("ERR_OPERATION_FAILED", err)
	}
	if err := saveDocument(doc); err != nil {
		return err
	}

	fmt.Println(messages.Prompt("MSG_REMOVED", item.Description))
	return nil
}

func init() {
	removeCmd.Flags().StringP("project", "p", "", "project path (optional)")
	removeCmd.Flags().BoolP("yes", "y", false, "skip confirmation")
}
