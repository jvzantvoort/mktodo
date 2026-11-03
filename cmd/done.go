package cmd

import (
	"fmt"
	"strings"

	"github.com/jvzantvoort/mktodo/messages"
	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:     "done <description>",
	Aliases: []string{"complete"},
	Short:   "Mark a TODO item as done",
	Long: `Mark a TODO item as done (toggle checkbox state).

Examples:
  mktodo done "Complete documentation"
  mktodo complete -p lego.technic "Build crane"`,
	Args: usageArgs(cobra.MinimumNArgs(1)),
	RunE: runDone,
}

func runDone(cmd *cobra.Command, args []string) error {
	repo, cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	// Get flags
	projectPath, err := cmd.Flags().GetString("project")
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

	item.MarkDone()
	if err := saveDocument(doc); err != nil {
		return err
	}

	if item.Done {
		fmt.Println(messages.Prompt("MSG_MARKED_DONE", item.Description))
	} else {
		fmt.Println(messages.Prompt("MSG_MARKED_UNDONE", item.Description))
	}
	return nil
}

func init() {
	doneCmd.Flags().StringP("project", "p", "", "project path (optional)")
}
