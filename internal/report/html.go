package report

import (
	"html/template"
	"io"
)

// htmlTemplate is a standalone page with inline CSS and no external
// resources. html/template escapes all user text.
var htmlTemplate = template.Must(template.New("report").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>
  :root {
    --bg: #ffffff; --fg: #1f2328; --muted: #59636e; --line: #d1d9e0;
    --card: #f6f8fa; --bar: #d1d9e0; --done: #1a7f37; --fixme: #cf222e;
  }
  @media (prefers-color-scheme: dark) {
    :root {
      --bg: #0d1117; --fg: #e6edf3; --muted: #9198a1; --line: #3d444d;
      --card: #151b23; --bar: #3d444d; --done: #3fb950; --fixme: #f85149;
    }
  }
  * { box-sizing: border-box; }
  body { margin: 0; padding: 24px 16px; background: var(--bg); color: var(--fg);
         font: 15px/1.5 system-ui, -apple-system, "Segoe UI", sans-serif; }
  main { max-width: 860px; margin: 0 auto; }
  h1 { font-size: 1.6rem; margin: 0 0 24px; }
  h2 { font-size: 1.15rem; margin: 32px 0 12px; }
  .cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 12px; }
  .card { background: var(--card); border: 1px solid var(--line); border-radius: 8px; padding: 12px 16px; }
  .card .value { font-size: 1.6rem; font-weight: 600; }
  .card .label, .muted { color: var(--muted); }
  .fixme { color: var(--fixme); }
  .table-wrap { overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; }
  th, td { text-align: left; padding: 8px; border-bottom: 1px solid var(--line); white-space: nowrap; }
  th.num, td.num { text-align: right; }
  td.name { white-space: normal; }
  .bar { width: 120px; height: 10px; background: var(--bar); border-radius: 5px; overflow: hidden; }
  .bar span { display: block; height: 100%; background: var(--done); }
  ol { padding-left: 24px; }
</style>
</head>
<body>
<main>
<h1>{{.Title}}</h1>

<h2>{{.L.Summary}}</h2>
<div class="cards">
  <div class="card"><div class="value">{{.R.Summary.Total}}</div><div class="label">{{.L.Total}}</div></div>
  <div class="card"><div class="value">{{.R.Summary.Open}}</div><div class="label">{{.L.Open}} ({{.R.Summary.Percent .R.Summary.Open}}%)</div></div>
  <div class="card"><div class="value">{{.R.Summary.Done}}</div><div class="label">{{.L.Done}} ({{.R.Summary.Percent .R.Summary.Done}}%)</div></div>
  <div class="card"><div class="value fixme">{{.R.Summary.FIXME}}</div><div class="label">{{.L.FIXME}} ({{.R.Summary.Percent .R.Summary.FIXME}}%)</div></div>
</div>

<h2>{{.L.ByProject}}</h2>
<div class="table-wrap">
<table>
<thead><tr><th>{{.L.Project}}</th><th>{{.L.Completion}}</th><th class="num">{{.L.Done}}</th><th class="num">{{.L.Total}}</th><th class="num">{{.L.FIXME}}</th></tr></thead>
<tbody>
{{- range .Rows}}
<tr>
  <td class="name" style="padding-left: {{.Indent}}px">{{.Name}}</td>
  <td>{{if .Total}}<div class="bar" role="img" aria-label="{{.Completion}}%"><span style="width: {{.Completion}}%"></span></div>{{else}}<span class="muted">{{$.L.NoItems}}</span>{{end}}</td>
  <td class="num">{{.Done}}</td>
  <td class="num">{{.Total}}</td>
  <td class="num{{if .FIXME}} fixme{{end}}">{{.FIXME}}</td>
</tr>
{{- end}}
</tbody>
</table>
</div>

<h2>{{.L.FIXMEItems}}</h2>
{{- if .FIXMEItems}}
<ol>
{{- range .FIXMEItems}}
  <li><strong>{{.Project}}:</strong> {{.Description}}{{if .Done}} <span class="muted">{{$.L.ItemDone}}</span>{{end}}</li>
{{- end}}
</ol>
{{- else}}
<p class="muted">{{.L.NoFIXME}}</p>
{{- end}}
</main>
</body>
</html>
`))

type htmlRow struct {
	Name   string
	Indent int
	Counts
}

type htmlFIXME struct {
	Project     string
	Description string
	Done        bool
}

type htmlLabels struct {
	Summary, ByProject, FIXMEItems, Total, Open, Done, FIXME,
	Project, Completion, NoItems, NoFIXME, ItemDone string
}

// WriteHTML renders the report as a standalone HTML page.
func WriteHTML(w io.Writer, r *Report) error {
	data := struct {
		Title      string
		L          htmlLabels
		R          *Report
		Rows       []htmlRow
		FIXMEItems []htmlFIXME
	}{
		Title: label("TITLE", r.Generated.Format(dateFormat)),
		L: htmlLabels{
			Summary: label("SUMMARY"), ByProject: label("BY_PROJECT"), FIXMEItems: label("FIXME_ITEMS"),
			Total: label("TOTAL"), Open: label("OPEN"), Done: label("DONE"), FIXME: label("FIXME"),
			Project: label("PROJECT"), Completion: label("COMPLETION"), NoItems: label("NO_ITEMS"),
			NoFIXME: label("NO_FIXME"), ItemDone: label("ITEM_DONE"),
		},
		R: r,
	}

	for _, p := range r.Projects {
		data.Rows = append(data.Rows, htmlRow{Name: p.Title, Indent: 8 + 20*p.Depth, Counts: p.Counts})
	}
	if r.Unassigned.Total > 0 {
		data.Rows = append(data.Rows, htmlRow{Name: label("NO_PROJECT"), Indent: 8, Counts: r.Unassigned})
	}
	for _, item := range r.FIXMEItems {
		data.FIXMEItems = append(data.FIXMEItems, htmlFIXME{projectLabel(item), item.Description, item.Done})
	}

	return htmlTemplate.Execute(w, data)
}
