package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Notebook-style tabs, following the technique from the Bubble Tea `tabs`
// example: active tabs get an "open bottom" that connects into the baseline
// below; inactive tabs stay closed. A rule extends the baseline to full width.

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	b := lipgloss.RoundedBorder()
	b.BottomLeft, b.Bottom, b.BottomRight = left, middle, right
	return b
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")

	inactiveTabStyle = lipgloss.NewStyle().
				Border(inactiveTabBorder, true).
				BorderForeground(borderColor).
				Foreground(lipgloss.Color("245")).
				Padding(0, 2)

	activeTabStyle = inactiveTabStyle.
			Border(activeTabBorder, true).
			BorderForeground(accentColor).
			Foreground(lipgloss.Color("15")).
			Bold(true)

	ruleStyle = lipgloss.NewStyle().Foreground(borderColor)
)

// tabAtX returns the tab kind whose label occupies screen column x on the
// tab-label row. Each tab is title width + 4 padding + 2 border cells, joined
// flush from column 0.
func tabAtX(x int) (kind, bool) {
	col := 0
	for _, k := range order {
		w := lipgloss.Width(specs[k].title) + 6
		if x >= col && x < col+w {
			return k, true
		}
		col += w
	}
	return 0, false
}

func (m Model) tabBar(width int) string {
	if width <= 0 {
		width = 80
	}

	rendered := make([]string, 0, len(order))
	for i, k := range order {
		active := k == m.active
		st := inactiveTabStyle
		if active {
			st = activeTabStyle
		}
		border, _, _, _, _ := st.GetBorder()
		first, last := i == 0, i == len(order)-1
		switch {
		case first && active:
			border.BottomLeft = "│"
		case first && !active:
			border.BottomLeft = "├"
		case last && active:
			border.BottomRight = "│"
		case last && !active:
			border.BottomRight = "┤"
		}
		rendered = append(rendered, st.Border(border).Render(specs[k].title))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Bottom, rendered...)

	// Extend the baseline (the tab bottoms) to the full width.
	if gap := width - lipgloss.Width(row); gap > 0 {
		rule := ruleStyle.Render(strings.Repeat("─", gap))
		row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, rule)
	}
	return row
}
