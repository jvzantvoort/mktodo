package cmd

import (
	"fmt"

	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/git"
	"github.com/jvzantvoort/mktodo/internal/markdown"
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
	RunE: runOpen,
}

func runOpen(cmd *cobra.Command, args []string) error {
	// Check we're in a git repository
	repo, err := git.FindRepository()
	if err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	// Load configuration
	cfg, err := config.Load(repo.ConfigPath())
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Get flags
	projectPath, _ := cmd.Flags().GetString("project")

	// Load document
	todoPath := repo.ResolvePath(cfg.TodoFile)
	doc, err := markdown.LoadDocument(todoPath, cfg)
	if err != nil {
		return fmt.Errorf("loading document: %w", err)
	}

	// Get items
	var items []*markdown.Item
	if projectPath != "" {
		proj := doc.Projects[projectPath]
		if proj == nil {
			return fmt.Errorf("project %q not found", projectPath)
		}
		items = doc.GetItemsByProject(proj)
	} else {
		items = doc.Items
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
		fmt.Println("🔴 FIXME Items:")
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
		fmt.Println("Open Items:")
		for _, item := range regularItems {
			projectInfo := ""
			if item.Project != nil {
				projectInfo = fmt.Sprintf(" (%s)", item.Project.FullPath())
			}
			fmt.Printf("  - %s%s\n", item.Description, projectInfo)
		}
	}

	if len(items) == 0 {
		fmt.Println("✓ No open TODO items!")
	} else {
		fmt.Printf("\nTotal: %d open items", len(items))
		if len(fixmeItems) > 0 {
			fmt.Printf(" (%d FIXME)", len(fixmeItems))
		}
		fmt.Println()
	}

	return nil
}

func init() {
	openCmd.Flags().StringP("project", "p", "", "filter by project")
}
