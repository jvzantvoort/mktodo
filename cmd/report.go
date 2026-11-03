package cmd

import (
	"bytes"
	"io"
	"time"

	"github.com/jvzantvoort/mktodo/internal/markdown"
	"github.com/jvzantvoort/mktodo/internal/report"
	"github.com/jvzantvoort/mktodo/messages"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a TODO report",
	Long: `Generate a report of TODO items per project: summary statistics,
a completion bar per project, and the list of FIXME items.

Formats:
  text      plain text with bar charts (default)
  markdown  for embedding in documentation
  html      a standalone page with inline CSS
  json      for scripts

Examples:
  mktodo report
  mktodo report --format markdown -o report.md
  mktodo report --format html -o report.html`,
	Args: usageArgs(cobra.NoArgs),
	RunE: runReport,
}

// reportFormats maps each --format value to its renderer.
var reportFormats = map[string]func(io.Writer, *report.Report) error{
	"text":     report.WriteText,
	"markdown": report.WriteMarkdown,
	"html":     report.WriteHTML,
	"json":     report.WriteJSON,
}

// reportNow is the report's timestamp; tests replace it.
var reportNow = time.Now

func runReport(cmd *cobra.Command, _ []string) error {
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}
	outputFile, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}
	if err = validateFormat(format, "text", "markdown", "html", "json"); err != nil {
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

	var buf bytes.Buffer
	if err = reportFormats[format](&buf, report.Build(doc, reportNow())); err != nil {
		return messages.Errorf("ERR_OPERATION_FAILED", err)
	}

	if outputFile == "" {
		_, err = cmd.OutOrStdout().Write(buf.Bytes())
		return err
	}
	if err = markdown.NewWriter(outputFile).WriteString(buf.String()); err != nil {
		return fileError(messages.Errorf("ERR_FILE_WRITE_FAILED", outputFile, err))
	}
	_, err = cmd.OutOrStdout().Write([]byte(messages.Prompt("MSG_REPORT_WRITTEN", outputFile) + "\n"))
	return err
}

func init() {
	reportCmd.Flags().String("format", "text", "output format: text|markdown|html|json")
	reportCmd.Flags().StringP("output", "o", "", "write to file instead of stdout")
}
