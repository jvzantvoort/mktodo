package cmd

import (
	"fmt"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/messages"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <description>",
	Aliases: []string{"rm", "destroy"},
	Short:   "Remove a TODO item",
	Long: `Remove a TODO item from the specified project.

With --all, every item whose description contains the text is removed,
after one confirmation listing them.

Examples:
  mktodo remove "Complete documentation"
  mktodo rm -p lego "Old task"
  mktodo rm -p lego.technic "Build crane" --yes
  mktodo rm --all "obsolete"`,
	Args: usageArgs(cobra.MinimumNArgs(1)),
	RunE: runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	projectPath, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}
	skipConfirm, err := cmd.Flags().GetBool("yes")
	if err != nil {
		return err
	}
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}
	query := strings.Join(args, " ")

	repo, cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}
	doc, err := loadDocument(repo, cfg)
	if err != nil {
		return err
	}
	items, err := findItems(doc, projectPath, query, all)
	if err != nil {
		return err
	}

	if !skipConfirm && !confirmRemove(items) {
		return messages.Errorf("ERR_OPERATION_CANCELLED")
	}

	for _, item := range items {
		if err := doc.RemoveItem(item); err != nil {
			return messages.Errorf("ERR_OPERATION_FAILED", err)
		}
	}
	if err := saveDocument(doc); err != nil {
		return err
	}

	for _, item := range items {
		fmt.Println(messages.Prompt("MSG_REMOVED", item.Description))
	}
	return nil
}

// confirmRemove asks before removing; several items are listed first.
func confirmRemove(items []*markdown.Item) bool {
	if len(items) == 1 {
		return confirm(messages.Prompt("PROMPT_CONFIRM_REMOVE", items[0].Description))
	}
	for _, item := range items {
		fmt.Printf("  %s\n", item.String())
	}
	return confirm(messages.Prompt("PROMPT_CONFIRM_REMOVE_ALL", len(items)))
}

func init() {
	removeCmd.Flags().StringP("project", "p", "", "project path (optional)")
	removeCmd.Flags().BoolP("yes", "y", false, "skip confirmation")
	removeCmd.Flags().BoolP("all", "a", false, "remove every matching item")
}
