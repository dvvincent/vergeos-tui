package ui

import "testing"

// Tab widths = title width + 6 (padding 2+2, border 1+1), joined from col 0:
// VMs[0,9) Networks[9,23) Tenants[23,36) Storage[36,49).
func TestTabAtX(t *testing.T) {
	cases := []struct {
		x    int
		want kind
		ok   bool
	}{
		{0, kVMs, true},
		{5, kVMs, true},
		{9, kNetworks, true},
		{22, kNetworks, true},
		{23, kTenants, true},
		{40, kStorage, true},
		{200, 0, false}, // past the last tab
		{-1, 0, false},
	}
	for _, c := range cases {
		got, ok := tabAtX(c.x)
		if ok != c.ok || (ok && got != c.want) {
			t.Fatalf("tabAtX(%d) = %v,%v want %v,%v", c.x, got, ok, c.want, c.ok)
		}
	}
}

func TestSelectVisible(t *testing.T) {
	tbl := sampleTable() // 4 rows, keys 32,36,48,46
	if !tbl.selectVisible(2) {
		t.Fatal("selectVisible(2) should succeed")
	}
	if r, _ := tbl.selected(); r.Key != 48 {
		t.Fatalf("row 2 should be key 48, got %d", r.Key)
	}
	if tbl.selectVisible(-1) {
		t.Fatal("negative index must fail")
	}
	if tbl.selectVisible(4) {
		t.Fatal("index past last row must fail")
	}
	if tbl.selectVisible(tbl.height) {
		t.Fatal("index at/after visible height must fail")
	}
}
