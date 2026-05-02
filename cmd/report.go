package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jvzantvoort/mktodo/internal/config"
	"github.com/jvzantvoort/mktodo/internal/git"
	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/project"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a TODO report",
	Long: `Generate a graphical report of TODO items per project.

Examples:
  mktodo report
  mktodo report --format markdown -o report.md`,
	RunE: runReport,
}

type projectStats struct {
	Name  string
	Title string
	Level int
	Total int
	Open  int
	Done  int
	FIXME int
}

func runReport(cmd *cobra.Command, args []string) error {
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
	format, _ := cmd.Flags().GetString("format")
	outputFile, _ := cmd.Flags().GetString("output")

	// Load document
	todoPath := repo.ResolvePath(cfg.TodoFile)
	doc, err := markdown.LoadDocument(todoPath, cfg)
	if err != nil {
		return fmt.Errorf("loading document: %w", err)
	}

	// Calculate stats for each project
	stats := make(map[string]*projectStats)
	for name, proj := range doc.Projects {
		items := doc.GetItemsByProject(proj)
		s := &projectStats{
			Name:  name,
			Title: proj.Title,
			Level: proj.Level,
			Total: len(items),
		}
		for _, item := range items {
			if item.Done {
				s.Done++
			} else {
				s.Open++
			}
			if item.IsFIXME {
				s.FIXME++
			}
		}
		stats[name] = s
	}

	// Generate output
	var output string
	switch format {
	case "json":
		output, err = generateJSONReport(stats)
	case "markdown":
		output = generateMarkdownReport(stats, doc)
	default:
		output = generateTextReport(stats, doc)
	}

	if err != nil {
		return fmt.Errorf("generating report: %w", err)
	}

	// Write output
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		fmt.Printf("Report written to %s\n", outputFile)
	} else {
		fmt.Print(output)
	}

	return nil
}

func generateTextReport(stats map[string]*projectStats, doc *markdown.Document) string {
	var sb strings.Builder

	sb.WriteString("TODO Report\n")
	sb.WriteString("===========\n\n")

	// Get root projects
	roots := project.GetRootProjects(doc.Projects)
	for _, root := range roots {
		generateTextProjectReport(&sb, root, stats, 0)
	}

	// Overall summary
	totalItems := len(doc.Items)
	openItems := len(doc.GetOpenItems())
	doneItems := len(doc.GetDoneItems())
	fixmeItems := len(doc.GetFIXMEItems())

	sb.WriteString("\nOverall Summary\n")
	sb.WriteString("---------------\n")
	sb.WriteString(fmt.Sprintf("Total Items:  %d\n", totalItems))
	sb.WriteString(fmt.Sprintf("Open:         %d\n", openItems))
	sb.WriteString(fmt.Sprintf("Done:         %d\n", doneItems))
	sb.WriteString(fmt.Sprintf("FIXME:        %d\n", fixmeItems))
	if totalItems > 0 {
		completion := float64(doneItems) / float64(totalItems) * 100
		sb.WriteString(fmt.Sprintf("Completion:   %.1f%%\n", completion))
	}

	return sb.String()
}

func generateTextProjectReport(sb *strings.Builder, proj *project.Project, stats map[string]*projectStats, depth int) {
	indent := strings.Repeat("  ", depth)
	s := stats[proj.Name]

	if s.Total > 0 {
		completion := float64(s.Done) / float64(s.Total) * 100
		fixmeTag := ""
		if s.FIXME > 0 {
			fixmeTag = fmt.Sprintf(" [%d FIXME]", s.FIXME)
		}
		sb.WriteString(fmt.Sprintf("%s%s: %d items (%.0f%% done)%s\n",
			indent, proj.Title, s.Total, completion, fixmeTag))
	} else {
		sb.WriteString(fmt.Sprintf("%s%s: no items\n", indent, proj.Title))
	}

	for _, child := range proj.Children {
		generateTextProjectReport(sb, child, stats, depth+1)
	}
}

func generateMarkdownReport(stats map[string]*projectStats, doc *markdown.Document) string {
	var sb strings.Builder

	sb.WriteString("# TODO Report\n\n")

	// Get root projects
	roots := project.GetRootProjects(doc.Projects)
	for _, root := range roots {
		generateMarkdownProjectReport(&sb, root, stats, 2)
	}

	// Overall summary
	totalItems := len(doc.Items)
	openItems := len(doc.GetOpenItems())
	doneItems := len(doc.GetDoneItems())
	fixmeItems := len(doc.GetFIXMEItems())

	sb.WriteString("\n## Overall Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Items:** %d\n", totalItems))
	sb.WriteString(fmt.Sprintf("- **Open:** %d\n", openItems))
	sb.WriteString(fmt.Sprintf("- **Done:** %d\n", doneItems))
	sb.WriteString(fmt.Sprintf("- **FIXME:** %d\n", fixmeItems))
	if totalItems > 0 {
		completion := float64(doneItems) / float64(totalItems) * 100
		sb.WriteString(fmt.Sprintf("- **Completion:** %.1f%%\n", completion))
	}

	return sb.String()
}

func generateMarkdownProjectReport(sb *strings.Builder, proj *project.Project, stats map[string]*projectStats, level int) {
	header := strings.Repeat("#", level)
	s := stats[proj.Name]

	sb.WriteString(fmt.Sprintf("%s %s\n\n", header, proj.Title))

	if s.Total > 0 {
		completion := float64(s.Done) / float64(s.Total) * 100
		sb.WriteString(fmt.Sprintf("- Total: %d items\n", s.Total))
		sb.WriteString(fmt.Sprintf("- Open: %d\n", s.Open))
		sb.WriteString(fmt.Sprintf("- Done: %d\n", s.Done))
		if s.FIXME > 0 {
			sb.WriteString(fmt.Sprintf("- FIXME: %d\n", s.FIXME))
		}
		sb.WriteString(fmt.Sprintf("- Completion: %.1f%%\n\n", completion))
	} else {
		sb.WriteString("No items.\n\n")
	}

	for _, child := range proj.Children {
		generateMarkdownProjectReport(sb, child, stats, level+1)
	}
}

func generateJSONReport(stats map[string]*projectStats) (string, error) {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func init() {
	reportCmd.Flags().String("format", "text", "output format: text|markdown|json")
	reportCmd.Flags().StringP("output", "o", "", "output file (default: stdout)")
}
