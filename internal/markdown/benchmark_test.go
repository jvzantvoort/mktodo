package markdown

import (
	"fmt"
	"strings"
	"testing"
)

// BenchmarkParseSections benchmarks section parsing performance
func BenchmarkParseSections(b *testing.B) {
	content := `# Project Header

Some introductory text here.

## TODO Items

- [ ] Task 1
- [X] Task 2
- [ ] FIXME: broken feature
- [ ] Task 3

## Another Section

More content here.

### Nested TODO

- [ ] Nested task 1
- [ ] Nested task 2
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseSections(content)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWriteDocument benchmarks document writing performance
func BenchmarkWriteDocument(b *testing.B) {
	sections := []*Section{
		{
			Type:  SectionHeader,
			Level: 1,
			Title: "TODO",
			Lines: []string{"# TODO"},
		},
		{
			Type:  SectionOther,
			Lines: []string{"", "- [ ] Task 1", "- [X] Task 2", ""},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var output strings.Builder
		for _, section := range sections {
			for _, line := range section.Lines {
				output.WriteString(line)
				output.WriteString("\n")
			}
		}
		_ = output.String()
	}
}

// BenchmarkLargeDocument benchmarks parsing of large documents
func BenchmarkLargeDocument(b *testing.B) {
	// Generate a document with many sections and tasks
	var builder strings.Builder
	for i := 0; i < 10; i++ {
		builder.WriteString(fmt.Sprintf("# Project %c\n\n", rune('A'+i)))
		
		for j := 0; j < 20; j++ {
			builder.WriteString(fmt.Sprintf("- [ ] Task %d\n", j))
		}
		builder.WriteString("\n")
	}
	
	content := builder.String()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseSections(content)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFindSection benchmarks section finding
func BenchmarkFindSection(b *testing.B) {
	content := `# TODO

- [ ] Task 1

## Subproject

- [ ] Task 2

### Nested

- [ ] Task 3
`

	sections, err := ParseSections(content)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Find section with title "TODO" and level 1
		for _, section := range sections {
			if section.Title == "TODO" && section.Level == 1 {
				break
			}
		}
	}
}
