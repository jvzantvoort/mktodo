package cmd

import (
	"fmt"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/git"
	"github.com/jvzantvoort/mktodo/internal/markdown"
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
	Args: cobra.MinimumNArgs(1),
	RunE: runDone,
}

func runDone(cmd *cobra.Command, args []string) error {
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

	// Build query from args
	query := strings.Join(args, " ")

	// Load document
	todoPath := repo.ResolvePath(cfg.TodoFile)
	doc, err := markdown.LoadDocument(todoPath, cfg)
	if err != nil {
		return fmt.Errorf("loading document: %w", err)
	}

	// Find matching items
	var candidates []*markdown.Item
	if projectPath != "" {
		proj := doc.Projects[projectPath]
		if proj == nil {
			return fmt.Errorf("project %q not found", projectPath)
		}
		// Search within project
		for _, item := range doc.GetItemsByProject(proj) {
			if item.Matches(query) {
				candidates = append(candidates, item)
			}
		}
	} else {
		// Search all items
		candidates = doc.FindItem(query)
	}

	if len(candidates) == 0 {
		return fmt.Errorf("no matching TODO items found for %q", query)
	}

	if len(candidates) > 1 {
		fmt.Printf("Multiple items match %q:\n", query)
		for i, item := range candidates {
			status := " "
			if item.Done {
				status = "X"
			}
			fmt.Printf("  %d. [%s] %s\n", i+1, status, item.Description)
		}
		return fmt.Errorf("please be more specific")
	}

	// Toggle the item
	item := candidates[0]
	item.MarkDone()

	// Update document
	if err := doc.UpdateItem(item); err != nil {
		return fmt.Errorf("updating item: %w", err)
	}

	// Save
	if err := markdown.SaveDocument(todoPath, doc); err != nil {
		return fmt.Errorf("saving document: %w", err)
	}

	// Success message
	if item.Done {
		fmt.Printf("✓ Marked as done: %s\n", item.Description)
	} else {
		fmt.Printf("○ Marked as not done: %s\n", item.Description)
	}

	return nil
}

func init() {
	doneCmd.Flags().StringP("project", "p", "", "project path (optional)")
}
