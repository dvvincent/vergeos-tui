// Package ui is the Bubble Tea front-end. It follows the Elm architecture:
// a Model holds state, Update reacts to Msgs (key presses, data-loaded events),
// and View renders. All vrg calls run as tea.Cmds so the UI never blocks.
package ui

import (
	"fmt"
	"strconv"
	"time"

	"vergeos-tui/internal/vrg"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Screen geometry for mouse hit-testing. The View lays out: row 0 = title +
// header, rows 1-3 = the 3-line notebook tab bar (labels on row 2), row 4 =
// table header, row 5 = rule, rows 6+ = data. Assumes the header fits on one
// line (true for terminals wide enough not to wrap it).
const (
	rowTabLabels = 2
	rowFirstData = 6
)

// --- resource kinds (tabs) -------------------------------------------------

type kind int

const (
	kVMs kind = iota
	kNetworks
	kTenants
	kStorage
	nKinds
)

type spec struct {
	title    string
	columns  []Column
	colorers map[int]func(string) lipgloss.Style
}

var specs = map[kind]spec{
	kVMs: {"VMs", []Column{
		{"KEY", 5}, {"NAME", 22}, {"STATUS", 9}, {"CPU", 4}, {"RAM", 7}, {"NODE", 7}, {"CLUSTER", 10},
	}, map[int]func(string) lipgloss.Style{2: vmStatusColor}},
	kNetworks: {"Networks", []Column{
		{"KEY", 5}, {"NAME", 22}, {"TYPE", 10}, {"IP", 16}, {"STATE", 8}, {"DHCP", 5}, {"DNS", 8},
	}, map[int]func(string) lipgloss.Style{4: vmStatusColor}},
	kTenants: {"Tenants", []Column{
		{"KEY", 5}, {"NAME", 22}, {"STATE", 10}, {"NETWORK", 18}, {"UI IP", 15}, {"ISOLATED", 8},
	}, map[int]func(string) lipgloss.Style{2: vmStatusColor}},
	kStorage: {"Storage", []Column{
		{"TIER", 5}, {"CAPACITY", 12}, {"USED", 12}, {"FREE", 12}, {"USAGE", 19}, {"DEDUPE", 8},
	}, map[int]func(string) lipgloss.Style{4: barColorer}},
}

var order = []kind{kVMs, kNetworks, kTenants, kStorage}

// --- messages & commands ---------------------------------------------------

type rowsMsg struct {
	k     kind
	rows  []Row
	count int
	vms   []vrg.VM
	nets  []vrg.Network
	tens  []vrg.Tenant
	stor  []vrg.StorageTier
}
type loadErrMsg struct {
	k   kind
	err error
}
type sysMsg struct{ s vrg.SysInfo }
type detailMsg struct{ d vrg.VMDetail }
type detailErrMsg struct{ err error }
type actionMsg struct {
	verb, name string
	err        error
}
type tickMsg time.Time

func loadCmd(k kind) tea.Cmd {
	return func() tea.Msg {
		switch k {
		case kVMs:
			vms, err := vrg.ListVMs()
			if err != nil {
				return loadErrMsg{k, err}
			}
			rows := make([]Row, 0, len(vms))
			for _, v := range vms {
				rows = append(rows, Row{v.Key, []string{
					fmt.Sprint(v.Key), v.Name, dash(v.Status),
					fmt.Sprint(v.CPUCores), ramStr(v.RAM), dash(v.NodeName), dash(v.ClusterName),
				}})
			}
			return rowsMsg{k: k, rows: rows, count: len(vms), vms: vms}
		case kNetworks:
			ns, err := vrg.ListNetworks()
			if err != nil {
				return loadErrMsg{k, err}
			}
			rows := make([]Row, 0, len(ns))
			for _, n := range ns {
				rows = append(rows, Row{n.Key, []string{
					fmt.Sprint(n.Key), n.Name, dash(n.Type), dash(n.IPAddress),
					runState(n.Running), yesno(n.DHCPEnabled), dash(n.DNS),
				}})
			}
			return rowsMsg{k: k, rows: rows, count: len(ns), nets: ns}
		case kTenants:
			ts, err := vrg.ListTenants()
			if err != nil {
				return loadErrMsg{k, err}
			}
			rows := make([]Row, 0, len(ts))
			for _, t := range ts {
				rows = append(rows, Row{t.Key, []string{
					fmt.Sprint(t.Key), t.Name, dash(t.State), dash(t.NetworkName),
					dash(t.UIAddressIP), yesno(t.IsIsolated),
				}})
			}
			return rowsMsg{k: k, rows: rows, count: len(ts), tens: ts}
		case kStorage:
			ss, err := vrg.ListStorage()
			if err != nil {
				return loadErrMsg{k, err}
			}
			rows := make([]Row, 0, len(ss))
			for _, s := range ss {
				usage := bar(s.UsedPercent, 12) + fmt.Sprintf(" %3.0f%%", s.UsedPercent)
				rows = append(rows, Row{s.Tier, []string{
					fmt.Sprint(s.Tier), gb(s.CapacityGB), gb(s.UsedGB), gb(s.FreeGB),
					usage, fmt.Sprintf("%.1fx", s.DedupeRatio),
				}})
			}
			return rowsMsg{k: k, rows: rows, count: len(ss), stor: ss}
		}
		return nil
	}
}

func sysCmd() tea.Msg {
	s, err := vrg.SystemInfo()
	if err != nil {
		return nil // header is best-effort
	}
	return sysMsg{s}
}

func detailCmd(v vrg.VM) tea.Cmd {
	return func() tea.Msg {
		d, err := vrg.GetVMDetail(v)
		if err != nil {
			return detailErrMsg{err}
		}
		return detailMsg{d}
	}
}

func actionCmd(verb string, key int, name string) tea.Cmd {
	return func() tea.Msg {
		var err error
		switch verb {
		case "start":
			err = vrg.StartVM(key)
		case "stop":
			err = vrg.StopVM(key)
		case "restart":
			err = vrg.RestartVM(key)
		}
		return actionMsg{verb, name, err}
	}
}

// refreshInterval is how often auto-refresh reloads. The UI ticks once a
// second so the status bar can show a live countdown to the next reload.
const refreshInterval = 5

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// --- model -----------------------------------------------------------------

type mode int

const (
	modeList mode = iota
	modeDetail
	modeConfirm
)

type tabState struct {
	table   resTable
	loading bool
	loaded  bool
	err     error
	vms     []vrg.VM
	nets    []vrg.Network
	tens    []vrg.Tenant
	stor    []vrg.StorageTier
}

type pending struct {
	verb, name string
	key        int
}

type Model struct {
	tabs        map[kind]*tabState
	active      kind
	mode        mode
	spinner     spinner.Model
	help        help.Model
	width       int
	height      int
	status      string
	autoRefresh bool
	refreshIn   int // seconds until the next auto-refresh

	sys       vrg.SysInfo
	sysLoaded bool

	searching bool
	search    textinput.Model

	detailKind kind
	detail     *vrg.VMDetail
	detailNet  *vrg.Network
	detailTen  *vrg.Tenant
	detailStor *vrg.StorageTier
	detailErr  error
	vp         viewport.Model // scrolls the detail panel
	pending    pending
}

func New() Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = spinnerStyle

	tabs := make(map[kind]*tabState, nKinds)
	for _, k := range order {
		s := specs[k]
		rt := newResTable(s.columns, s.colorers)
		tabs[k] = &tabState{table: rt}
	}

	ti := textinput.New()
	ti.Prompt = "/"
	ti.Placeholder = "filter…"
	ti.CharLimit = 64

	m := Model{tabs: tabs, active: kVMs, spinner: sp, help: help.New(), search: ti, vp: viewport.New(0, 0)}
	m.tabs[kVMs].loading = true
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, loadCmd(kVMs), sysCmd, tea.SetWindowTitle("VergeOS TUI"))
}

func (m *Model) cur() *tabState { return m.tabs[m.active] }

func (m *Model) selectedVM() (vrg.VM, bool) {
	ts := m.tabs[kVMs]
	if m.active != kVMs {
		return vrg.VM{}, false
	}
	r, ok := ts.table.selected()
	if !ok {
		return vrg.VM{}, false
	}
	for _, v := range ts.vms {
		if v.Key == r.Key {
			return v, true
		}
	}
	return vrg.VM{}, false
}

func (m *Model) selectedNet() (vrg.Network, bool) {
	if r, ok := m.tabs[kNetworks].table.selected(); ok {
		for _, n := range m.tabs[kNetworks].nets {
			if n.Key == r.Key {
				return n, true
			}
		}
	}
	return vrg.Network{}, false
}

func (m *Model) selectedTen() (vrg.Tenant, bool) {
	if r, ok := m.tabs[kTenants].table.selected(); ok {
		for _, t := range m.tabs[kTenants].tens {
			if t.Key == r.Key {
				return t, true
			}
		}
	}
	return vrg.Tenant{}, false
}

func (m *Model) selectedStor() (vrg.StorageTier, bool) {
	if r, ok := m.tabs[kStorage].table.selected(); ok {
		for _, s := range m.tabs[kStorage].stor {
			if s.Tier == r.Key {
				return s, true
			}
		}
	}
	return vrg.StorageTier{}, false
}

// openDetail drills into the selected resource. VMs load drives/NICs
// asynchronously; the others render instantly from data already in hand.
func (m *Model) openDetail() tea.Cmd {
	var cmd tea.Cmd
	switch m.active {
	case kVMs:
		if v, ok := m.selectedVM(); ok {
			m.mode, m.detailKind = modeDetail, kVMs
			m.detail, m.detailErr = nil, nil
			cmd = tea.Batch(m.spinner.Tick, detailCmd(v))
		}
	case kNetworks:
		if n, ok := m.selectedNet(); ok {
			m.mode, m.detailKind, m.detailNet = modeDetail, kNetworks, &n
		}
	case kTenants:
		if t, ok := m.selectedTen(); ok {
			m.mode, m.detailKind, m.detailTen = modeDetail, kTenants, &t
		}
	case kStorage:
		if s, ok := m.selectedStor(); ok {
			m.mode, m.detailKind, m.detailStor = modeDetail, kStorage, &s
		}
	}
	if m.mode == modeDetail {
		m.vp.SetContent(m.detailView())
		m.vp.GotoTop()
	}
	return cmd
}

func (m *Model) resizeTables() {
	h := m.height - 9 // header + tabs + status + help
	if h < 3 {
		h = 3
	}
	for _, k := range order {
		m.tabs[k].table.setSize(h)
	}
	// Detail viewport occupies the same area minus title(1)+tabs(3)+status(2).
	vpH := m.height - 7
	if vpH < 3 {
		vpH = 3
	}
	m.vp.Width, m.vp.Height = m.width, vpH
	if m.mode == modeDetail {
		m.vp.SetContent(m.detailView())
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Keep the search input's cursor blinking while it's focused (non-key msgs).
	if m.searching {
		if _, isKey := msg.(tea.KeyMsg); !isKey {
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resizeTables()

	case tea.KeyMsg:
		// handleKey fully owns key handling (there's no fall-through table).
		return m, m.handleKey(msg)

	case rowsMsg:
		ts := m.tabs[msg.k]
		ts.table.setRows(msg.rows)
		ts.loading, ts.loaded, ts.err = false, true, nil
		ts.vms, ts.nets, ts.tens, ts.stor = msg.vms, msg.nets, msg.tens, msg.stor
		if msg.k == m.active {
			m.status = fmt.Sprintf("%d %s", ts.table.count(), specs[msg.k].title)
		}

	case loadErrMsg:
		ts := m.tabs[msg.k]
		ts.loading, ts.err = false, msg.err

	case sysMsg:
		m.sys, m.sysLoaded = msg.s, true
		if msg.s.CloudName != "" {
			cmds = append(cmds, tea.SetWindowTitle("VergeOS TUI — "+msg.s.CloudName))
		}

	case detailMsg:
		m.detail, m.detailErr = &msg.d, nil
		m.vp.SetContent(m.detailView())

	case detailErrMsg:
		m.detailErr = msg.err
		m.vp.SetContent(m.detailView())

	case tea.MouseMsg:
		return m, m.handleMouse(msg)

	case actionMsg:
		if msg.err != nil {
			m.status = errStyle.Render(fmt.Sprintf("%s %s failed: %v", msg.verb, msg.name, msg.err))
		} else {
			m.status = fmt.Sprintf("%s %s — ok", verbPast(msg.verb), msg.name)
		}
		m.cur().loading = true
		cmds = append(cmds, m.spinner.Tick, loadCmd(m.active), sysCmd)

	case tickMsg:
		if m.autoRefresh {
			if m.refreshIn > 0 {
				m.refreshIn--
			}
			if m.refreshIn <= 0 && m.mode == modeList && !m.cur().loading {
				cmds = append(cmds, loadCmd(m.active), sysCmd)
				m.refreshIn = refreshInterval
			}
			cmds = append(cmds, tickCmd())
		}

	case spinner.TickMsg:
		if m.anyLoading() {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) handleKey(msg tea.KeyMsg) tea.Cmd {
	s := msg.String()
	if s == "ctrl+c" {
		return tea.Quit
	}
	if s == "ctrl+z" {
		return tea.Suspend
	}

	// Search input mode: the textinput bubble owns the keys.
	if m.searching {
		switch s {
		case "enter": // keep the filter, exit input
			m.searching = false
			m.search.Blur()
		case "esc": // clear the filter, exit input
			m.searching = false
			m.search.Blur()
			m.search.SetValue("")
			m.cur().table.setFilter("")
		default:
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			m.cur().table.setFilter(m.search.Value())
			return cmd
		}
		return nil
	}

	switch m.mode {
	case modeConfirm:
		switch s {
		case "y", "Y", "enter":
			p := m.pending
			m.mode = modeList
			m.status = fmt.Sprintf("%sing %s…", p.verb, p.name)
			return tea.Batch(m.spinner.Tick, actionCmd(p.verb, p.key, p.name))
		case "n", "N", "esc", "q":
			m.mode = modeList
			m.status = "cancelled"
		}
		return nil

	case modeDetail:
		switch s {
		case "esc", "enter", "q", "backspace", "h", "left":
			m.mode = modeList
			m.detail, m.detailErr = nil, nil
			m.detailNet, m.detailTen, m.detailStor = nil, nil, nil
		case "up", "k":
			m.vp.LineUp(1)
		case "down", "j":
			m.vp.LineDown(1)
		case "pgup", "ctrl+u":
			m.vp.HalfViewUp()
		case "pgdown", "ctrl+d":
			m.vp.HalfViewDown()
		case "g", "home":
			m.vp.GotoTop()
		case "G", "end":
			m.vp.GotoBottom()
		}
		return nil
	}

	// modeList
	switch {
	case key.Matches(msg, keys.Quit):
		return tea.Quit
	case key.Matches(msg, keys.Search):
		m.searching = true
		m.search.Focus()
		return textinput.Blink
	case key.Matches(msg, keys.NextTab):
		m.switchTab(1)
		return m.loadIfNeeded()
	case key.Matches(msg, keys.PrevTab):
		m.switchTab(-1)
		return m.loadIfNeeded()
	case key.Matches(msg, keys.Up):
		m.cur().table.moveUp()
	case key.Matches(msg, keys.Down):
		m.cur().table.moveDown()
	case key.Matches(msg, keys.PgUp):
		m.cur().table.pageUp()
	case key.Matches(msg, keys.PgDn):
		m.cur().table.pageDown()
	case key.Matches(msg, keys.Top):
		m.cur().table.top()
	case key.Matches(msg, keys.Bottom):
		m.cur().table.bottom()
	case key.Matches(msg, keys.Refresh):
		m.cur().loading = true
		m.status = "refreshing…"
		return tea.Batch(m.spinner.Tick, loadCmd(m.active), sysCmd)
	case key.Matches(msg, keys.Auto):
		m.autoRefresh = !m.autoRefresh
		if m.autoRefresh {
			m.refreshIn = refreshInterval
			return tickCmd()
		}
	case key.Matches(msg, keys.Help):
		m.help.ShowAll = !m.help.ShowAll
	case key.Matches(msg, keys.Enter):
		return m.openDetail()
	case key.Matches(msg, keys.Start):
		return m.confirmAction("start")
	case key.Matches(msg, keys.Stop):
		return m.confirmAction("stop")
	case key.Matches(msg, keys.Restart):
		return m.confirmAction("restart")
	default:
		// number keys sort by column
		if n, err := strconv.Atoi(s); err == nil && n >= 1 && n <= len(specs[m.active].columns) {
			m.cur().table.cycleSort(n - 1)
		} else {
		}
	}
	return nil
}

// handleMouse maps clicks and wheel events onto tab/row selection and detail
// scrolling. Wheel works in any mode; clicks act only in list mode.
func (m *Model) handleMouse(e tea.MouseMsg) tea.Cmd {
	if m.searching {
		return nil
	}
	switch e.Button {
	case tea.MouseButtonWheelUp:
		if m.mode == modeDetail {
			m.vp.LineUp(3)
		} else {
			m.cur().table.moveUp()
		}
		return nil
	case tea.MouseButtonWheelDown:
		if m.mode == modeDetail {
			m.vp.LineDown(3)
		} else {
			m.cur().table.moveDown()
		}
		return nil
	}

	if e.Button != tea.MouseButtonLeft || e.Action != tea.MouseActionPress || m.mode != modeList {
		return nil
	}

	// Click a tab to switch to it.
	if e.Y == rowTabLabels {
		if k, ok := tabAtX(e.X); ok && k != m.active {
			m.active = k
			return m.loadIfNeeded()
		}
		return nil
	}

	// Click a data row: first click selects it, a second click drills in.
	if e.Y >= rowFirstData {
		t := &m.cur().table
		prev := t.cursor
		if t.selectVisible(e.Y-rowFirstData) && t.cursor == prev {
			return m.openDetail()
		}
	}
	return nil
}

func (m *Model) confirmAction(verb string) tea.Cmd {
	if v, ok := m.selectedVM(); ok {
		m.pending = pending{verb: verb, name: v.Name, key: v.Key}
		m.mode = modeConfirm
	}
	return nil
}

func (m *Model) switchTab(dir int) {
	idx := 0
	for i, k := range order {
		if k == m.active {
			idx = i
		}
	}
	m.active = order[(idx+dir+len(order))%len(order)]
}

func (m *Model) loadIfNeeded() tea.Cmd {
	ts := m.cur()
	if !ts.loaded && !ts.loading {
		ts.loading = true
		return tea.Batch(m.spinner.Tick, loadCmd(m.active))
	}
	m.status = fmt.Sprintf("%d %s", ts.table.count(), specs[m.active].title)
	return nil
}

func (m Model) anyLoading() bool {
	for _, k := range order {
		if m.tabs[k].loading {
			return true
		}
	}
	return m.mode == modeDetail && m.detailKind == kVMs && m.detail == nil && m.detailErr == nil
}
