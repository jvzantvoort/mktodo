package report

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/jvzantvoort/mktodo/messages"
)

const (
	dateFormat = "2006-01-02"
	barWidth   = 10
)

// errWriter writes formatted text and remembers the first error, so the
// renderers can write freely and check once at the end.
type errWriter struct {
	w   io.Writer
	err error
}

func (e *errWriter) printf(format string, args ...interface{}) {
	if e.err == nil {
		_, e.err = fmt.Fprintf(e.w, format, args...)
	}
}

// label returns a report label from the messages package.
func label(key string, args ...interface{}) string {
	return messages.Prompt("MSG_REPORT_"+key, args...)
}

// projectLabel names the project an item belongs to.
func projectLabel(item FIXMEItem) string {
	if item.Project == nil {
		return label("NO_PROJECT")
	}
	return item.Project.Title
}

// WriteText renders the report as plain text with bar charts.
func WriteText(w io.Writer, r *Report) error {
	out := &errWriter{w: w}

	title := label("TITLE", r.Generated.Format(dateFormat))
	out.printf("%s\n%s\n\n", title, strings.Repeat("═", utf8.RuneCountInString(title)))

	writeTextHeading(out, label("SUMMARY"))
	s := r.Summary
	out.printf("%-12s %5d\n", label("TOTAL")+":", s.Total)
	out.printf("%-12s %5d (%d%%)\n", label("OPEN")+":", s.Open, s.Percent(s.Open))
	out.printf("%-12s %5d (%d%%)\n", label("DONE")+":", s.Done, s.Percent(s.Done))
	out.printf("%-12s %5d (%d%%)\n\n", label("FIXME")+":", s.FIXME, s.Percent(s.FIXME))

	writeTextHeading(out, label("BY_PROJECT"))
	writeTextProjects(out, r)

	writeTextHeading(out, label("FIXME_ITEMS"))
	if len(r.FIXMEItems) == 0 {
		out.printf("%s\n", label("NO_FIXME"))
	}
	for i, item := range r.FIXMEItems {
		out.printf("%d. [%s] %s%s\n", i+1, projectLabel(item), item.Description, doneSuffix(item.Done, " "+label("ITEM_DONE")))
	}

	return out.err
}

func writeTextHeading(out *errWriter, heading string) {
	out.printf("%s\n%s\n", heading, strings.Repeat("─", utf8.RuneCountInString(heading)))
}

// writeTextProjects writes one bar per project, children indented under
// their parent, with the names padded to a common width.
func writeTextProjects(out *errWriter, r *Report) {
	type row struct {
		name   string
		counts Counts
	}
	rows := make([]row, 0, len(r.Projects)+1)
	for _, p := range r.Projects {
		rows = append(rows, row{strings.Repeat("  ", p.Depth) + p.Title, p.Counts})
	}
	if r.Unassigned.Total > 0 {
		rows = append(rows, row{label("NO_PROJECT"), r.Unassigned})
	}

	width := 0
	for _, row := range rows {
		width = max(width, utf8.RuneCountInString(row.name))
	}

	for _, row := range rows {
		pad := strings.Repeat(" ", width-utf8.RuneCountInString(row.name))
		c := row.counts
		if c.Total == 0 {
			out.printf("%s%s  %s  %s\n", row.name, pad, bar(0), label("NO_ITEMS"))
			continue
		}
		out.printf("%s%s  %s  %d/%d  (%d%%)%s\n", row.name, pad, bar(c.Completion()), c.Done, c.Total, c.Completion(), fixmeSuffix(c.FIXME))
	}
	out.printf("\n")
}

// bar draws a completion bar, e.g. "████████░░" for 80%.
func bar(percent int) string {
	filled := (percent*barWidth + 50) / 100
	return strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
}

func fixmeSuffix(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("  [%d %s]", n, label("FIXME"))
}

func doneSuffix(done bool, suffix string) string {
	if done {
		return suffix
	}
	return ""
}
