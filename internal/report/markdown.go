package report

import (
	"io"
	"strconv"
	"strings"
)

// WriteMarkdown renders the report as markdown, for embedding in docs.
func WriteMarkdown(w io.Writer, r *Report) error {
	out := &errWriter{w: w}

	out.printf("# %s\n\n", label("TITLE", r.Generated.Format(dateFormat)))

	out.printf("## %s\n\n", label("SUMMARY"))
	s := r.Summary
	out.printf("- **%s:** %d\n", label("TOTAL"), s.Total)
	out.printf("- **%s:** %d (%d%%)\n", label("OPEN"), s.Open, s.Percent(s.Open))
	out.printf("- **%s:** %d (%d%%)\n", label("DONE"), s.Done, s.Percent(s.Done))
	out.printf("- **%s:** %d (%d%%)\n\n", label("FIXME"), s.FIXME, s.Percent(s.FIXME))

	out.printf("## %s\n\n", label("BY_PROJECT"))
	out.printf("| %s | %s | %s | %s | %s |\n", label("PROJECT"), label("DONE"), label("TOTAL"), label("COMPLETION"), label("FIXME"))
	out.printf("|---|---:|---:|---:|---:|\n")
	for _, p := range r.Projects {
		indent := strings.Repeat("&nbsp;&nbsp;", p.Depth)
		writeMarkdownRow(out, indent+escapeMarkdown(p.Title), p.Counts)
	}
	if r.Unassigned.Total > 0 {
		writeMarkdownRow(out, escapeMarkdown(label("NO_PROJECT")), r.Unassigned)
	}
	out.printf("\n")

	out.printf("## %s\n\n", label("FIXME_ITEMS"))
	if len(r.FIXMEItems) == 0 {
		out.printf("%s\n", label("NO_FIXME"))
	}
	for i, item := range r.FIXMEItems {
		out.printf("%d. **%s:** %s%s\n", i+1, escapeMarkdown(projectLabel(item)), escapeMarkdown(item.Description),
			doneSuffix(item.Done, " _"+label("ITEM_DONE")+"_"))
	}

	return out.err
}

func writeMarkdownRow(out *errWriter, name string, c Counts) {
	completion := "–"
	if c.Total > 0 {
		completion = strconv.Itoa(c.Completion()) + "%"
	}
	out.printf("| %s | %d | %d | %s | %d |\n", name, c.Done, c.Total, completion, c.FIXME)
}

// escapeMarkdown keeps user text from breaking tables or adding formatting.
var markdownEscaper = strings.NewReplacer(
	`\`, `\\`, "|", `\|`, "*", `\*`, "_", `\_`, "`", "\\`", "[", `\[`, "]", `\]`, "<", "&lt;", ">", "&gt;",
)

func escapeMarkdown(s string) string {
	return markdownEscaper.Replace(s)
}
