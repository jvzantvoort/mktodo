package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/messages"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List TODO items",
	Long: `List TODO items with optional filtering.

FIXME items are listed first; otherwise items appear in file order.

Examples:
  mktodo list
  mktodo ls -o
  mktodo ls --fixme
  mktodo ls -p lego --format json`,
	Args: usageArgs(cobra.NoArgs),
	RunE: runList,
}

// listOptions selects and formats the items shown by list and open.
type listOptions struct {
	project string
	open    bool
	done    bool
	fixme   bool
	format  string
}

func runList(cmd *cobra.Command, _ []string) error {
	opts, err := readListFlags(cmd)
	if err != nil {
		return err
	}
	if opts.open, err = cmd.Flags().GetBool("open"); err != nil {
		return err
	}
	if opts.done, err = cmd.Flags().GetBool("done"); err != nil {
		return err
	}
	return listItems(cmd, opts)
}

// readListFlags reads the flags that list and open share.
func readListFlags(cmd *cobra.Command) (listOptions, error) {
	var opts listOptions
	var err error
	if opts.project, err = cmd.Flags().GetString("project"); err != nil {
		return opts, err
	}
	if opts.fixme, err = cmd.Flags().GetBool("fixme"); err != nil {
		return opts, err
	}
	if opts.format, err = cmd.Flags().GetString("format"); err != nil {
		return opts, err
	}
	return opts, nil
}

// listItems prints the selected items. Flags are checked before any file
// is read, so usage errors are reported first.
func listItems(cmd *cobra.Command, opts listOptions) error {
	if opts.open && opts.done {
		return usageError(messages.Errorf("ERR_CONFLICTING_FLAGS", "open", "done"))
	}
	if err := validateFormat(opts.format, "text", "json", "yaml"); err != nil {
		return err
	}

	repo, cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}
	doc, err := loadDocument(repo, cfg)
	if err != nil {
		return err
	}
	items, err := projectItems(doc, opts.project)
	if err != nil {
		return err
	}

	items = sortFIXMEFirst(filterItems(items, opts))

	out := cmd.OutOrStdout()
	switch opts.format {
	case "json":
		return outputJSON(out, items)
	case "yaml":
		return outputYAML(out, items)
	default:
		return outputText(out, items, opts)
	}
}

// filterItems keeps the items that match the status and FIXME filters.
func filterItems(items []*markdown.Item, opts listOptions) []*markdown.Item {
	var result []*markdown.Item
	for _, item := range items {
		if (opts.open && item.Done) || (opts.done && !item.Done) || (opts.fixme && !item.IsFIXME) {
			continue
		}
		result = append(result, item)
	}
	return result
}

// sortFIXMEFirst moves FIXME items to the front, keeping file order within
// each group (spec FR-5: "FIXME first, then order in file").
func sortFIXMEFirst(items []*markdown.Item) []*markdown.Item {
	sorted := append([]*markdown.Item(nil), items...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].IsFIXME && !sorted[j].IsFIXME
	})
	return sorted
}

func outputText(out io.Writer, items []*markdown.Item, opts listOptions) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(out, messages.Prompt(emptyListMessage(opts)))
		return err
	}

	for _, item := range items {
		checkbox := "[ ]"
		if item.Done {
			checkbox = "[X]"
		}

		fixmeTag := ""
		if item.IsFIXME {
			fixmeTag = " [FIXME]"
		}

		projectInfo := ""
		if item.Project != nil {
			projectInfo = fmt.Sprintf(" (%s)", item.Project.FullPath())
		}

		if _, err := fmt.Fprintf(out, "%s %s%s%s\n", checkbox, item.Description, fixmeTag, projectInfo); err != nil {
			return err
		}
	}
	return nil
}

func emptyListMessage(opts listOptions) string {
	switch {
	case opts.open:
		return "MSG_NO_OPEN_ITEMS"
	case opts.done:
		return "MSG_NO_DONE_ITEMS"
	default:
		return "MSG_NO_ITEMS"
	}
}

type itemOutput struct {
	Description string `json:"description" yaml:"description"`
	Done        bool   `json:"done" yaml:"done"`
	IsFIXME     bool   `json:"fixme" yaml:"fixme"`
	Project     string `json:"project,omitempty" yaml:"project,omitempty"`
	LineNumber  int    `json:"line_number,omitempty" yaml:"line_number,omitempty"`
}

func toItemOutput(items []*markdown.Item) []itemOutput {
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
	return output
}

func outputJSON(out io.Writer, items []*markdown.Item) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(toItemOutput(items))
}

func outputYAML(out io.Writer, items []*markdown.Item) error {
	enc := yaml.NewEncoder(out)
	defer func() { _ = enc.Close() }()
	return enc.Encode(toItemOutput(items))
}

// addListFlags registers the flags that list and open share.
func addListFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("project", "p", "", "filter by project")
	cmd.Flags().BoolP("fixme", "f", false, "show only FIXME items")
	cmd.Flags().String("format", "text", "output format: text|json|yaml")
}

func init() {
	addListFlags(listCmd)
	listCmd.Flags().BoolP("open", "o", false, "show only open items")
	listCmd.Flags().BoolP("done", "d", false, "show only completed items")
}
