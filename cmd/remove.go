package cmd

import (
	"fmt"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/git"
	"github.com/jvzantvoort/mktodo/internal/markdown"
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
	Args: cobra.MinimumNArgs(1),
	RunE: runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
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
	skipConfirm, _ := cmd.Flags().GetBool("yes")

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
		for _, item := range doc.GetItemsByProject(proj) {
			if item.Matches(query) {
				candidates = append(candidates, item)
			}
		}
	} else {
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

	item := candidates[0]

	// Confirm deletion
	if !skipConfirm {
		fmt.Printf("Remove: %s\n", item.Description)
		fmt.Print("Are you sure? [y/N] ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			return fmt.Errorf("cancelled")
		}
	}

	// Remove the item
	if err := doc.RemoveItem(item); err != nil {
		return fmt.Errorf("removing item: %w", err)
	}

	// Save
	if err := markdown.SaveDocument(todoPath, doc); err != nil {
		return fmt.Errorf("saving document: %w", err)
	}

	fmt.Printf("✗ Removed: %s\n", item.Description)

	return nil
}

func init() {
	removeCmd.Flags().StringP("project", "p", "", "project path (optional)")
	removeCmd.Flags().BoolP("yes", "y", false, "skip confirmation")
}
