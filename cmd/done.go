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
	Long: `Mark a TODO item as done, or as not done if it already is (toggle).

With --all, every item whose description contains the text is marked
done; items that are already done stay done.

Examples:
  mktodo done "Complete documentation"
  mktodo complete -p lego.technic "Build crane"
  mktodo done --all "release 1.0"`,
	Args: usageArgs(cobra.MinimumNArgs(1)),
	RunE: runDone,
}

func runDone(cmd *cobra.Command, args []string) error {
	projectPath, err := cmd.Flags().GetString("project")
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

	for _, item := range items {
		if all {
			item.Done = true
		} else {
			item.MarkDone()
		}
	}
	if err := saveDocument(doc); err != nil {
		return err
	}

	for _, item := range items {
		if item.Done {
			fmt.Println(messages.Prompt("MSG_MARKED_DONE", item.Description))
		} else {
			fmt.Println(messages.Prompt("MSG_MARKED_UNDONE", item.Description))
		}
	}
	return nil
}

func init() {
	doneCmd.Flags().StringP("project", "p", "", "project path (optional)")
	doneCmd.Flags().BoolP("all", "a", false, "mark every matching item done")
}
