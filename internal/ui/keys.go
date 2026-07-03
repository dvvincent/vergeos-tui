package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up, Down             key.Binding
	PgUp, PgDn           key.Binding
	Top, Bottom          key.Binding
	NextTab, PrevTab     key.Binding
	Enter, Refresh, Auto key.Binding
	Search               key.Binding
	Start, Stop, Restart key.Binding
	Help, Suspend, Quit  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextTab, k.Enter, k.Search, k.Start, k.Stop, k.Refresh, k.Auto, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PgUp, k.PgDn, k.Top, k.Bottom},
		{k.NextTab, k.PrevTab, k.Enter, k.Search},
		{k.Start, k.Stop, k.Restart, k.Refresh, k.Auto},
		{k.Help, k.Suspend, k.Quit},
	}
}

var keys = keyMap{
	Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	PgUp:    key.NewBinding(key.WithKeys("pgup", "ctrl+u"), key.WithHelp("pgup", "page up")),
	PgDn:    key.NewBinding(key.WithKeys("pgdown", "ctrl+d"), key.WithHelp("pgdn", "page down")),
	Top:     key.NewBinding(key.WithKeys("home", "g"), key.WithHelp("g", "top")),
	Bottom:  key.NewBinding(key.WithKeys("end", "G"), key.WithHelp("G", "bottom")),
	NextTab: key.NewBinding(key.WithKeys("tab", "right", "l"), key.WithHelp("tab", "next tab")),
	PrevTab: key.NewBinding(key.WithKeys("shift+tab", "left"), key.WithHelp("⇧tab", "prev tab")),
	Enter:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "details")),
	Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	Auto:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "auto-refresh")),
	Search:  key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	Start:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start")),
	Stop:    key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "stop")),
	Restart: key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "restart")),
	Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Suspend: key.NewBinding(key.WithKeys("ctrl+z"), key.WithHelp("ctrl+z", "suspend")),
	Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}
