package markdown

import "testing"

func TestInsertAndSwap_AcrossSplitLists(t *testing.T) {
	doc := loadString(t, "# TODO\n\n- [ ] a\n\n- [ ] b\n", defaultConfig)
	a, b := doc.Items[0], doc.Items[1]
	if err := doc.SwapItems(a, b); err != nil {
		t.Fatal(err)
	}
	c := doc.Items[0] // b is now first
	if err := doc.InsertItem(doc.Projects["default"], c, &Item{Description: "new"}); err != nil {
		t.Fatal(err)
	}
	want := "# TODO\n\n- [ ] b\n- [ ] new\n\n- [ ] a\n"
	if got := saveAndRead(t, doc); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if doc.Items[0].Description != "b" || doc.Items[1].Description != "new" || doc.Items[2].Description != "a" {
		t.Errorf("Items not in document order: %v", doc.Items)
	}
}

func TestSwapItems_DifferentProjectsRefused(t *testing.T) {
	doc := loadString(t, "# Lego\n\n## Technic\n\n- [ ] t\n\n## City\n\n- [ ] c\n", nestedConfig)
	if err := doc.SwapItems(doc.Items[0], doc.Items[1]); err == nil {
		t.Error("swapping items of different projects should fail")
	}
}

func TestInsertItem_AfterItemOfOtherProjectRefused(t *testing.T) {
	doc := loadString(t, "# Lego\n\n## Technic\n\n- [ ] t\n", nestedConfig)
	before := doc.String()
	if err := doc.InsertItem(doc.Projects["city"], doc.Items[0], &Item{Description: "x"}); err == nil {
		t.Error("inserting after an item of another project should fail")
	}
	if doc.String() != before {
		t.Error("a rejected insert changed the document")
	}
}
