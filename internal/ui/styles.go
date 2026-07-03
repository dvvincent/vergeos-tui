package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// House palette — navy/accent to match the wider VergeOS tooling.
var (
	accentColor = lipgloss.Color("42")  // green
	headerColor = lipgloss.Color("81")  // cyan
	borderColor = lipgloss.Color("240") // grey
	navyColor   = lipgloss.Color("24")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(navyColor).
			Padding(0, 1)

	tabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(accentColor).
			Padding(0, 2)

	tabInactive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Background(lipgloss.Color("236")).
			Padding(0, 2)

	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	autoOnStyle = lipgloss.NewStyle().Foreground(accentColor).Bold(true)

	errStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	hintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)

	spinnerStyle = lipgloss.NewStyle().Foreground(accentColor)

	// detail view
	detailBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 2)
	sectionStyle = lipgloss.NewStyle().Bold(true).Foreground(headerColor)
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(12)
	valueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	// confirm modal
	confirmBox = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("214")).
			Padding(1, 3)
	confirmTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))

	// custom table
	tblHeader   = lipgloss.NewStyle().Bold(true).Foreground(headerColor)
	tblRule     = lipgloss.NewStyle().Foreground(borderColor)
	selectedRow = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(accentColor).Bold(true)
	matchStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)

	// status colors
	stGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	stRed    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	stYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	stGrey   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	// context header
	ctxKey = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	ctxVal = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Bold(true)
	ctxSep = lipgloss.NewStyle().Foreground(borderColor)

	// search bar
	searchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
)

// colorers for the status/health column, per tab.

func vmStatusColor(s string) lipgloss.Style {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "running", "online":
		return stGreen
	case "stopped", "offline":
		return stRed
	case "", "-":
		return stGrey
	default:
		return stYellow
	}
}

func usedPctColor(s string) lipgloss.Style {
	v, ok := parseNum(s)
	if !ok {
		return stGrey
	}
	switch {
	case v >= 90:
		return stRed
	case v >= 70:
		return stYellow
	default:
		return stGreen
	}
}

// barColorer colors a "████░░░ 16%" cell by the trailing percentage.
func barColorer(s string) lipgloss.Style {
	f := strings.Fields(s)
	if len(f) == 0 {
		return stGrey
	}
	return usedPctColor(f[len(f)-1])
}

// bar renders a Lip Gloss progress bar of the given cell width (blocks only;
// color is applied by the caller / colorer).
func bar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int(pct/100*float64(width) + 0.5)
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
