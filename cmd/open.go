package cmd

import (
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "List open TODO items",
	Long: `List only open (not completed) TODO items, FIXME items first.

Equivalent to: mktodo list --open

Examples:
  mktodo open
  mktodo open --fixme
  mktodo open -p lego.technic`,
	Args: usageArgs(cobra.NoArgs),
	RunE: runOpen,
}

func runOpen(cmd *cobra.Command, _ []string) error {
	opts, err := readListFlags(cmd)
	if err != nil {
		return err
	}
	opts.open = true
	return listItems(cmd, opts)
}

func init() {
	addListFlags(openCmd)
}
