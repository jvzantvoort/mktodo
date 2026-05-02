package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/git"
	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/project"
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
	Args: cobra.MinimumNArgs(1),
	RunE: runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
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

	// Build description from args
	description := strings.Join(args, " ")

	// Check if file exists
	todoPath := repo.ResolvePath(cfg.TodoFile)
	fileExists := true
	if _, err := os.Stat(todoPath); os.IsNotExist(err) {
		fileExists = false
		if !skipConfirm {
			fmt.Printf("File %s does not exist. Create it? [y/N] ", cfg.TodoFile)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				return fmt.Errorf("cancelled")
			}
		}
	}

	// Load or create document
	var doc *markdown.Document
	if fileExists {
		doc, err = markdown.LoadDocument(todoPath, cfg)
		if err != nil {
			return fmt.Errorf("loading document: %w", err)
		}
	} else {
		// Create new document with project structure
		projects, err := project.BuildHierarchy(cfg)
		if err != nil {
			return fmt.Errorf("building hierarchy: %w", err)
		}
		doc = &markdown.Document{
			Path:     todoPath,
			Sections: []*markdown.Section{},
			Projects: projects,
			Items:    []*markdown.Item{},
		}
		// Add project headers
		for _, proj := range project.GetRootProjects(projects) {
			addProjectHeaders(doc, proj)
		}
	}

	// Find the target project
	proj := doc.Projects[projectPath]
	if proj == nil {
		return fmt.Errorf("project %q not found", projectPath)
	}

	// Add the item
	item, err := doc.AddItem(proj, description)
	if err != nil {
		return fmt.Errorf("adding item: %w", err)
	}

	// Save the document
	if err := markdown.SaveDocument(todoPath, doc); err != nil {
		return fmt.Errorf("saving document: %w", err)
	}

	// Success message
	fmt.Printf("Added: %s\n", item.Description)
	if item.IsFIXME {
		fmt.Println("  (marked as FIXME)")
	}
	fmt.Printf("  to project: %s\n", proj.Title)

	return nil
}

// addProjectHeaders recursively adds project headers to a new document
func addProjectHeaders(doc *markdown.Document, proj *project.Project) {
	header := strings.Repeat("#", proj.Level) + " " + proj.Title
	doc.Sections = append(doc.Sections, &markdown.Section{
		Type:    markdown.SectionHeader,
		Level:   proj.Level,
		Title:   proj.Title,
		Project: proj,
		Lines:   []string{header, ""},
	})

	for _, child := range proj.Children {
		addProjectHeaders(doc, child)
	}
}

func init() {
	addCmd.Flags().StringP("project", "p", "default", "project path (e.g., lego.technic)")
	addCmd.Flags().BoolP("yes", "y", false, "skip confirmations")
}
