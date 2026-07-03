package ui

import "testing"

func sampleTable() resTable {
	t := newResTable([]Column{{"KEY", 5}, {"NAME", 20}, {"RAM", 6}}, nil)
	t.setRows([]Row{
		{32, []string{"32", "adminjump", "8GB"}},
		{36, []string{"36", "k3s-server-01", "8GB"}},
		{48, []string{"48", "k3s-server-02", "4GB"}},
		{46, []string{"46", "talos-01", "8GB"}},
	})
	return t
}

func TestFilter(t *testing.T) {
	tbl := sampleTable()
	if tbl.count() != 4 {
		t.Fatalf("want 4, got %d", tbl.count())
	}
	tbl.setFilter("k3s")
	if tbl.count() != 2 {
		t.Fatalf("filter k3s: want 2, got %d", tbl.count())
	}
	tbl.setFilter("TALOS") // case-insensitive
	if tbl.count() != 1 {
		t.Fatalf("filter TALOS: want 1, got %d", tbl.count())
	}
	tbl.setFilter("")
	if tbl.count() != 4 {
		t.Fatalf("clear filter: want 4, got %d", tbl.count())
	}
}

func TestSort(t *testing.T) {
	tbl := sampleTable()
	tbl.cycleSort(1) // by NAME asc
	if r, _ := tbl.selected(); r.Cells[1] != "adminjump" {
		t.Fatalf("asc: first should be adminjump, got %q", r.Cells[1])
	}
	tbl.cycleSort(1) // NAME desc
	if r, _ := tbl.selected(); r.Cells[1] != "talos-01" {
		t.Fatalf("desc: first should be talos-01, got %q", r.Cells[1])
	}
	tbl.cycleSort(1) // off — back to insertion order
	if tbl.sortCol != -1 {
		t.Fatalf("third cycle should disable sort, got sortCol=%d", tbl.sortCol)
	}
	// numeric sort on KEY
	tbl.cycleSort(0)
	if r, _ := tbl.selected(); r.Key != 32 {
		t.Fatalf("numeric asc by key: want 32, got %d", r.Key)
	}
}

func TestNavigation(t *testing.T) {
	tbl := sampleTable()
	tbl.setSize(2) // only 2 visible → exercise scroll
	tbl.moveDown()
	tbl.moveDown()
	tbl.moveDown() // cursor at last (index 3)
	if r, _ := tbl.selected(); r.Key != 46 {
		t.Fatalf("want last row key 46, got %d", r.Key)
	}
	if tbl.offset == 0 {
		t.Fatalf("expected scroll offset > 0 with height 2")
	}
	tbl.top()
	if r, _ := tbl.selected(); r.Key != 32 || tbl.offset != 0 {
		t.Fatalf("top() should reset to first row & offset 0")
	}
}

func TestParseNum(t *testing.T) {
	cases := map[string]float64{"8GB": 8, "512MB": 512, "16%": 16, "0.1x": 0.1, "  42 ": 42}
	for in, want := range cases {
		got, ok := parseNum(in)
		if !ok || got != want {
			t.Fatalf("parseNum(%q) = %v,%v want %v", in, got, ok, want)
		}
	}
	if _, ok := parseNum("adminjump"); ok {
		t.Fatalf("parseNum should reject non-numeric")
	}
}
