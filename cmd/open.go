package cmd

import (
	"fmt"

	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/messages"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "List open TODO items",
	Long: `List only open (not completed) TODO items.

Equivalent to: mktodo list --open

Examples:
  mktodo open
  mktodo open -p lego.technic`,
	Args: usageArgs(cobra.NoArgs),
	RunE: runOpen,
}

func runOpen(cmd *cobra.Command, args []string) error {
	repo, cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	// Get flags
	projectPath, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}

	doc, err := loadDocument(repo, cfg)
	if err != nil {
		return err
	}
	items, err := projectItems(doc, projectPath)
	if err != nil {
		return err
	}

	// Filter to open items only
	items = filterOpen(items)

	// Show FIXME items first
	fixmeItems := []*markdown.Item{}
	regularItems := []*markdown.Item{}
	for _, item := range items {
		if item.IsFIXME {
			fixmeItems = append(fixmeItems, item)
		} else {
			regularItems = append(regularItems, item)
		}
	}

	if len(fixmeItems) > 0 {
		fmt.Println(messages.Prompt("MSG_FIXME_ITEMS"))
		for _, item := range fixmeItems {
			projectInfo := ""
			if item.Project != nil {
				projectInfo = fmt.Sprintf(" (%s)", item.Project.FullPath())
			}
			fmt.Printf("  - %s%s\n", item.Description, projectInfo)
		}
		fmt.Println()
	}

	if len(regularItems) > 0 {
		fmt.Println(messages.Prompt("MSG_OPEN_ITEMS"))
		for _, item := range regularItems {
			projectInfo := ""
			if item.Project != nil {
				projectInfo = fmt.Sprintf(" (%s)", item.Project.FullPath())
			}
			fmt.Printf("  - %s%s\n", item.Description, projectInfo)
		}
	}

	switch {
	case len(items) == 0:
		fmt.Println(messages.Prompt("MSG_NO_OPEN"))
	case len(fixmeItems) > 0:
		fmt.Println()
		fmt.Println(messages.Prompt("MSG_TOTAL_OPEN", len(items), len(fixmeItems)))
	default:
		fmt.Println()
		fmt.Println(messages.Prompt("MSG_TOTAL_OPEN_NO_FIXME", len(items)))
	}

	return nil
}

func init() {
	openCmd.Flags().StringP("project", "p", "", "filter by project")
}
