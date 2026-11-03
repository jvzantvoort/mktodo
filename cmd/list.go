package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/project"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List TODO items",
	Long: `List TODO items with optional filtering.

Examples:
  mktodo list
  mktodo ls -o
  mktodo ls -p lego --format json`,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	repo, cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	// Get flags
	projectPath, err := cmd.Flags().GetString("project")
	if err != nil {
		return err
	}
	showOpen, err := cmd.Flags().GetBool("open")
	if err != nil {
		return err
	}
	showDone, err := cmd.Flags().GetBool("done")
	if err != nil {
		return err
	}
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}

	// Load document
	todoPath := repo.ResolvePath(cfg.TodoFile)
	doc, err := markdown.LoadDocument(todoPath, cfg)
	if err != nil {
		return fmt.Errorf("loading document: %w", err)
	}

	// Get items
	var items []*markdown.Item
	if projectPath != "" {
		proj, err := project.FindByPath(doc.Projects, projectPath)
		if err != nil {
			return fmt.Errorf("project %q not found", projectPath)
		}
		items = doc.GetItemsByProject(proj)
	} else {
		items = doc.Items
	}

	// Filter by status
	if showOpen {
		items = filterOpen(items)
	} else if showDone {
		items = filterDone(items)
	}

	// Output based on format
	switch format {
	case "json":
		return outputJSON(items)
	case "yaml":
		return outputYAML(items)
	default:
		return outputText(items)
	}
}

func filterOpen(items []*markdown.Item) []*markdown.Item {
	var result []*markdown.Item
	for _, item := range items {
		if !item.Done {
			result = append(result, item)
		}
	}
	return result
}

func filterDone(items []*markdown.Item) []*markdown.Item {
	var result []*markdown.Item
	for _, item := range items {
		if item.Done {
			result = append(result, item)
		}
	}
	return result
}

func outputText(items []*markdown.Item) error {
	if len(items) == 0 {
		fmt.Println("No TODO items found.")
		return nil
	}

	for _, item := range items {
		// Status indicator
		checkbox := "[ ]"
		if item.Done {
			checkbox = "[X]"
		}

		// FIXME indicator
		fixmeTag := ""
		if item.IsFIXME {
			fixmeTag = " [FIXME]"
		}

		// Project info
		projectInfo := ""
		if item.Project != nil {
			projectInfo = fmt.Sprintf(" (%s)", item.Project.FullPath())
		}

		fmt.Printf("%s %s%s%s\n", checkbox, item.Description, fixmeTag, projectInfo)
	}

	return nil
}

type itemOutput struct {
	Description string `json:"description" yaml:"description"`
	Done        bool   `json:"done" yaml:"done"`
	IsFIXME     bool   `json:"fixme" yaml:"fixme"`
	Project     string `json:"project,omitempty" yaml:"project,omitempty"`
	LineNumber  int    `json:"line_number,omitempty" yaml:"line_number,omitempty"`
}

func outputJSON(items []*markdown.Item) error {
	output := make([]itemOutput, len(items))
	for i, item := range items {
		output[i] = itemOutput{
			Description: item.Description,
			Done:        item.Done,
			IsFIXME:     item.IsFIXME,
			LineNumber:  item.LineNumber,
		}
		if item.Project != nil {
			output[i].Project = item.Project.FullPath()
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func outputYAML(items []*markdown.Item) error {
	output := make([]itemOutput, len(items))
	for i, item := range items {
		output[i] = itemOutput{
			Description: item.Description,
			Done:        item.Done,
			IsFIXME:     item.IsFIXME,
			LineNumber:  item.LineNumber,
		}
		if item.Project != nil {
			output[i].Project = item.Project.FullPath()
		}
	}

	enc := yaml.NewEncoder(os.Stdout)
	return enc.Encode(output)
}

func init() {
	listCmd.Flags().StringP("project", "p", "", "filter by project")
	listCmd.Flags().BoolP("open", "o", false, "show only open items")
	listCmd.Flags().BoolP("done", "d", false, "show only completed items")
	listCmd.Flags().String("format", "text", "output format: text|json|yaml")
}
