package ui

import (
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// Column is a fixed-width table column.
type Column struct {
	Title string
	Width int
}

// Row is one resource row: a key (for actions/detail) and its display cells.
type Row struct {
	Key   int
	Cells []string
}

// resTable is a lightweight, ANSI-aware table with cursor+scroll, live
// filtering, sorting, and per-cell colorization — things bubbles/table can't
// do because it truncates with runewidth (not ANSI-aware).
type resTable struct {
	cols     []Column
	rows     []Row  // all rows
	view     []int  // indices into rows, after filter+sort
	cursor   int    // index into view
	offset   int    // first visible view index
	height   int    // visible row count
	filter   string // active filter (case-insensitive substring)
	sortCol  int    // -1 = none
	sortDesc bool   // sort direction
	colorers map[int]func(string) lipgloss.Style
}

func newResTable(cols []Column, colorers map[int]func(string) lipgloss.Style) resTable {
	return resTable{cols: cols, height: 15, sortCol: -1, colorers: colorers}
}

func (t *resTable) setRows(rows []Row) {
	t.rows = rows
	t.recompute()
}

func (t *resTable) setSize(h int) {
	if h < 1 {
		h = 1
	}
	t.height = h
	t.clampCursor()
}

func (t *resTable) setFilter(f string) {
	t.filter = f
	t.cursor, t.offset = 0, 0
	t.recompute()
}

// cycleSort: click a column → sort asc; same column again → desc; third → off.
func (t *resTable) cycleSort(col int) {
	switch {
	case t.sortCol != col:
		t.sortCol, t.sortDesc = col, false
	case !t.sortDesc:
		t.sortDesc = true
	default:
		t.sortCol = -1
	}
	t.recompute()
}

func (t *resTable) recompute() {
	t.view = t.view[:0]
	f := strings.ToLower(t.filter)
	for i, r := range t.rows {
		if f == "" || rowMatches(r, f) {
			t.view = append(t.view, i)
		}
	}
	if t.sortCol >= 0 {
		sort.SliceStable(t.view, func(a, b int) bool {
			ca := cell(t.rows[t.view[a]], t.sortCol)
			cb := cell(t.rows[t.view[b]], t.sortCol)
			less := naturalLess(ca, cb)
			if t.sortDesc {
				return !less
			}
			return less
		})
	}
	t.clampCursor()
}

func (t *resTable) clampCursor() {
	if t.cursor >= len(t.view) {
		t.cursor = len(t.view) - 1
	}
	if t.cursor < 0 {
		t.cursor = 0
	}
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
	if t.cursor >= t.offset+t.height {
		t.offset = t.cursor - t.height + 1
	}
	if t.offset < 0 {
		t.offset = 0
	}
}

func (t *resTable) moveUp()   { t.cursor--; t.clampCursor() }
func (t *resTable) moveDown() { t.cursor++; t.clampCursor() }
func (t *resTable) pageUp()   { t.cursor -= t.height; t.clampCursor() }
func (t *resTable) pageDown() { t.cursor += t.height; t.clampCursor() }
func (t *resTable) top()      { t.cursor = 0; t.clampCursor() }
func (t *resTable) bottom()   { t.cursor = len(t.view) - 1; t.clampCursor() }

// selectVisible moves the cursor to the vis-th currently-visible row
// (0-based, relative to the scroll offset). Returns false if out of range.
func (t *resTable) selectVisible(vis int) bool {
	if vis < 0 || vis >= t.height {
		return false
	}
	idx := t.offset + vis
	if idx < 0 || idx >= len(t.view) {
		return false
	}
	t.cursor = idx
	t.clampCursor()
	return true
}

func (t *resTable) selected() (Row, bool) {
	if t.cursor >= 0 && t.cursor < len(t.view) {
		return t.rows[t.view[t.cursor]], true
	}
	return Row{}, false
}

func (t *resTable) count() int { return len(t.view) }

// pos is the 1-based cursor position within the visible rows (0 when empty).
func (t *resTable) pos() int {
	if len(t.view) == 0 {
		return 0
	}
	return t.cursor + 1
}

func (t resTable) view_() string {
	var lines []string

	// header (with sort indicator)
	var head strings.Builder
	for i, c := range t.cols {
		title := c.Title
		if i == t.sortCol {
			if t.sortDesc {
				title += " ▼"
			} else {
				title += " ▲"
			}
		}
		head.WriteString(pad(runewidth.Truncate(title, c.Width, "…"), c.Width))
		head.WriteString(" ")
	}
	lines = append(lines, tblHeader.Render(head.String()))
	lines = append(lines, tblRule.Render(strings.Repeat("─", t.totalWidth())))

	end := t.offset + t.height
	if end > len(t.view) {
		end = len(t.view)
	}
	for vi := t.offset; vi < end; vi++ {
		lines = append(lines, t.renderRow(t.rows[t.view[vi]], vi == t.cursor))
	}
	for i := end - t.offset; i < t.height; i++ {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

func (t resTable) renderRow(r Row, selected bool) string {
	cells := make([]string, len(t.cols))
	for i := range t.cols {
		w := t.cols[i].Width
		plain := runewidth.Truncate(cell(r, i), w, "…")
		if selected {
			cells[i] = pad(plain, w)
		} else if fn, ok := t.colorers[i]; ok {
			cells[i] = lipgloss.NewStyle().Width(w).Render(fn(plain).Render(plain))
		} else if t.filter != "" && strings.Contains(strings.ToLower(plain), strings.ToLower(t.filter)) {
			cells[i] = lipgloss.NewStyle().Width(w).Render(matchStyle.Render(plain))
		} else {
			cells[i] = pad(plain, w)
		}
	}
	line := strings.Join(cells, " ")
	if selected {
		return selectedRow.Width(t.totalWidth()).Render(line)
	}
	return line
}

func (t resTable) totalWidth() int {
	w := 0
	for _, c := range t.cols {
		w += c.Width + 1
	}
	if w > 0 {
		w--
	}
	return w
}

// --- helpers ---------------------------------------------------------------

func cell(r Row, i int) string {
	if i >= 0 && i < len(r.Cells) {
		return r.Cells[i]
	}
	return ""
}

func rowMatches(r Row, lowerFilter string) bool {
	for _, c := range r.Cells {
		if strings.Contains(strings.ToLower(c), lowerFilter) {
			return true
		}
	}
	return false
}

// pad renders plain text to a fixed width (space-padded), ANSI-safe.
func pad(s string, w int) string {
	return lipgloss.NewStyle().Width(w).Render(s)
}

// naturalLess compares numerically when both look numeric, else lexically.
func naturalLess(a, b string) bool {
	an, aok := parseNum(a)
	bn, bok := parseNum(b)
	if aok && bok {
		return an < bn
	}
	return strings.ToLower(a) < strings.ToLower(b)
}

func parseNum(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	for _, suf := range []string{"%", "x", " GB", "GB", " MB", "MB"} {
		s = strings.TrimSuffix(s, suf)
	}
	s = strings.TrimSpace(s)
	f, err := strconv.ParseFloat(s, 64)
	return f, err == nil
}
