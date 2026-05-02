package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/git"
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
	Args: cobra.MinimumNArgs(1),
	RunE: runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Check we're in a git repository
	repo, err := git.FindRepository()
	if err != nil {
		return messages.Errorf("ERR_NOT_IN_GIT_REPO")
	}

	// Load configuration
	cfg, err := config.Load(repo.ConfigPath())
	if err != nil {
		return messages.Errorf("ERR_CONFIG_LOAD_FAILED", err)
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
			fmt.Print(messages.Prompt("PROMPT_CREATE_FILE", cfg.TodoFile))
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				return messages.Errorf("ERR_OPERATION_CANCELLED")
			}
		}
	}

	// Load or create document
	var doc *markdown.Document
	if fileExists {
		doc, err = markdown.LoadDocument(todoPath, cfg)
		if err != nil {
			return messages.Errorf("ERR_FILE_READ_FAILED", todoPath, err)
		}
	} else {
		// Create new document with project structure
		projects, err := project.BuildHierarchy(cfg)
		if err != nil {
			return messages.Errorf("ERR_CONFIG_INVALID", err)
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
	proj, err := project.FindByPath(doc.Projects, projectPath)
	if err != nil {
		return messages.Errorf("ERR_PROJECT_NOT_FOUND", projectPath, getAvailableProjects(doc.Projects))
	}

	// Add the item
	item, err := doc.AddItem(proj, description)
	if err != nil {
		return messages.Errorf("ERR_OPERATION_FAILED", err)
	}

	// Save the document
	if err := markdown.SaveDocument(todoPath, doc); err != nil {
		return messages.Errorf("ERR_FILE_WRITE_FAILED", todoPath, err)
	}

	// Success message
	fmt.Println(messages.Prompt("MSG_ADDED", item.Description))
	if item.IsFIXME {
		fmt.Println(messages.Prompt("MSG_MARKED_FIXME"))
	}
	fmt.Println(messages.Prompt("MSG_TO_PROJECT", proj.Title))

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

// getAvailableProjects returns a comma-separated list of project names
func getAvailableProjects(projects map[string]*project.Project) string {
	var names []string
	for name := range projects {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

func init() {
	addCmd.Flags().StringP("project", "p", "default", "project path (e.g., lego.technic)")
	addCmd.Flags().BoolP("yes", "y", false, "skip confirmations")
}
