package todo

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		wantItem    bool
		wantDone    bool
		wantFIXME   bool
		wantDesc    string
	}{
		{
			name:     "basic open TODO",
			line:     "- [ ] Complete documentation",
			wantItem: true,
			wantDone: false,
			wantFIXME: false,
			wantDesc: "Complete documentation",
		},
		{
			name:     "basic done TODO",
			line:     "- [X] Finished task",
			wantItem: true,
			wantDone: true,
			wantDesc: "Finished task",
		},
		{
			name:     "done with lowercase x",
			line:     "- [x] Another finished task",
			wantItem: true,
			wantDone: true,
			wantDesc: "Another finished task",
		},
		{
			name:     "with asterisk",
			line:     "* [ ] Task with asterisk",
			wantItem: true,
			wantDone: false,
			wantDesc: "Task with asterisk",
		},
		{
			name:     "with plus",
			line:     "+ [ ] Task with plus",
			wantItem: true,
			wantDone: false,
			wantDesc: "Task with plus",
		},
		{
			name:     "FIXME item uppercase",
			line:     "- [ ] FIXME: Security vulnerability",
			wantItem: true,
			wantFIXME: true,
			wantDesc: "FIXME: Security vulnerability",
		},
		{
			name:     "FIXME item lowercase",
			line:     "- [ ] fixme: needs refactoring",
			wantItem: true,
			wantFIXME: true,
			wantDesc: "fixme: needs refactoring",
		},
		{
			name:     "FIXME item mixed case",
			line:     "- [ ] FiXmE: broken feature",
			wantItem: true,
			wantFIXME: true,
			wantDesc: "FiXmE: broken feature",
		},
		{
			name:     "with leading whitespace",
			line:     "  - [ ] Indented task",
			wantItem: true,
			wantDesc: "Indented task",
		},
		{
			name:     "not a TODO - no checkbox",
			line:     "- Regular list item",
			wantItem: false,
		},
		{
			name:     "not a TODO - plain text",
			line:     "This is just a sentence",
			wantItem: false,
		},
		{
			name:     "not a TODO - header",
			line:     "# Header",
			wantItem: false,
		},
		{
			name:     "empty line",
			line:     "",
			wantItem: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := Parse(tt.line)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if tt.wantItem {
				if item == nil {
					t.Fatal("Parse() returned nil, expected item")
				}
				if item.Done != tt.wantDone {
					t.Errorf("Done = %v, want %v", item.Done, tt.wantDone)
				}
				if item.IsFIXME != tt.wantFIXME {
					t.Errorf("IsFIXME = %v, want %v", item.IsFIXME, tt.wantFIXME)
				}
				if item.Description != tt.wantDesc {
					t.Errorf("Description = %q, want %q", item.Description, tt.wantDesc)
				}
			} else {
				if item != nil {
					t.Errorf("Parse() returned item, expected nil")
				}
			}
		})
	}
}

func TestItem_String(t *testing.T) {
	tests := []struct {
		name string
		item *Item
		want string
	}{
		{
			name: "open item",
			item: &Item{
				Description: "Test task",
				Done:        false,
			},
			want: "- [ ] Test task",
		},
		{
			name: "done item",
			item: &Item{
				Description: "Completed task",
				Done:        true,
			},
			want: "- [X] Completed task",
		},
		{
			name: "FIXME item",
			item: &Item{
				Description: "FIXME: broken",
				Done:        false,
				IsFIXME:     true,
			},
			want: "- [ ] FIXME: broken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.item.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestItem_RoundTrip(t *testing.T) {
	lines := []string{
		"- [ ] Open task",
		"- [X] Done task",
		"- [ ] FIXME: broken feature",
		"* [ ] With asterisk",
		"+ [x] Plus and done",
	}

	for _, line := range lines {
		t.Run(line, func(t *testing.T) {
			item, err := Parse(line)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if item == nil {
				t.Fatal("Parse() returned nil")
			}

			// Convert back to string and parse again
			str := item.String()
			item2, err := Parse(str)
			if err != nil {
				t.Fatalf("Parse() round-trip error = %v", err)
			}
			if item2 == nil {
				t.Fatal("Parse() round-trip returned nil")
			}

			// Check that semantics are preserved
			if item2.Done != item.Done {
				t.Errorf("round-trip Done mismatch: %v != %v", item2.Done, item.Done)
			}
			if item2.Description != item.Description {
				t.Errorf("round-trip Description mismatch: %q != %q", item2.Description, item.Description)
			}
		})
	}
}

func TestItem_MarkDone(t *testing.T) {
	item := &Item{Description: "Test", Done: false}
	
	if item.Done {
		t.Error("item should start as not done")
	}
	
	item.MarkDone()
	if !item.Done {
		t.Error("item should be done after MarkDone()")
	}
	
	item.MarkDone()
	if item.Done {
		t.Error("item should be undone after second MarkDone()")
	}
}

func TestItem_Matches(t *testing.T) {
	item := &Item{Description: "Complete the documentation for the API"}

	tests := []struct {
		query string
		want  bool
	}{
		{"documentation", true},
		{"Documentation", true},
		{"DOCUMENTATION", true},
		{"API", true},
		{"api", true},
		{"complete", true},
		{"nonexistent", false},
		{"xyz", false},
		{"", true}, // Empty query matches everything
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got := item.Matches(tt.query)
			if got != tt.want {
				t.Errorf("Matches(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}
