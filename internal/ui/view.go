package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" VergeOS TUI ") + "  " + m.header() + "\n")
	b.WriteString(m.tabBar(m.width) + "\n")

	switch m.mode {
	case modeDetail:
		b.WriteString(m.vp.View() + "\n")
	case modeConfirm:
		b.WriteString(m.confirmView() + "\n")
	default:
		ts := m.cur()
		switch {
		case ts.err != nil:
			b.WriteString(errStyle.Render("Error: "+ts.err.Error()) + "\n")
			b.WriteString(hintStyle.Render("Is vrg configured?  Try:  vrg system info") + "\n")
		case ts.loading && !ts.loaded:
			b.WriteString("  " + m.spinner.View() + "loading " + specs[m.active].title + "…\n")
		default:
			b.WriteString(ts.table.view_() + "\n")
		}
	}

	b.WriteString("\n" + m.statusBar() + "\n")
	b.WriteString(m.help.View(keys))
	return b.String()
}

// header renders the cluster-context bar.
func (m Model) header() string {
	if !m.sysLoaded {
		return ctxKey.Render("connecting…")
	}
	s := m.sys
	alarms := ctxVal.Render(fmt.Sprint(s.AlarmsTotal))
	if s.AlarmsTotal > 0 {
		alarms = stRed.Bold(true).Render(fmt.Sprintf("%d ⚠", s.AlarmsTotal))
	}
	sep := ctxSep.Render("  ·  ")
	return strings.Join([]string{
		ctxKey.Render("host ") + ctxVal.Render(strings.TrimPrefix(s.Host, "https://")),
		ctxKey.Render("cloud ") + ctxVal.Render(s.CloudName),
		ctxKey.Render("VMs ") + ctxVal.Render(fmt.Sprintf("%d/%d", s.VMsOnline, s.VMsTotal)),
		ctxKey.Render("nodes ") + ctxVal.Render(fmt.Sprintf("%d/%d", s.NodesOnline, s.NodesTotal)),
		ctxKey.Render("alarms ") + alarms,
		ctxKey.Render("v") + ctxKey.Render(s.Version),
	}, sep)
}

func (m Model) statusBar() string {
	if m.searching {
		return searchStyle.Render(m.search.View()) +
			statusStyle.Render(fmt.Sprintf("   %d matches   ", m.cur().table.count())) +
			hintStyle.Render("enter — keep · esc — clear")
	}
	s := m.status
	if m.cur().loading && m.cur().loaded {
		s = m.spinner.View() + s
	}
	parts := []string{statusStyle.Render(s)}
	if m.mode == modeList {
		if n := m.cur().table.count(); n > 0 {
			parts = append(parts, hintStyle.Render(fmt.Sprintf("[%d/%d]", m.cur().table.pos(), n)))
		}
	}
	if f := m.search.Value(); f != "" {
		parts = append(parts, hintStyle.Render(fmt.Sprintf("filter %q (%d)", f, m.cur().table.count())))
	}
	if m.autoRefresh {
		parts = append(parts, autoOnStyle.Render(fmt.Sprintf("● auto — next in %ds", m.refreshIn)))
	}
	return strings.Join(parts, "   ")
}

// --- value helpers ---------------------------------------------------------

func dash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func yesno(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func runState(running bool) string {
	if running {
		return "running"
	}
	return "stopped"
}

func ramStr(mb int) string {
	if mb >= 1024 && mb%1024 == 0 {
		return fmt.Sprintf("%dGB", mb/1024)
	}
	return fmt.Sprintf("%dMB", mb)
}

func gb(v float64) string { return fmt.Sprintf("%.1f GB", v) }

func verbPast(v string) string {
	switch v {
	case "start":
		return "started"
	case "stop":
		return "stopped"
	case "restart":
		return "restarted"
	}
	return v
}

// --- preview (headless) ----------------------------------------------------

var tabByName = map[string]kind{
	"vms": kVMs, "networks": kNetworks, "tenants": kTenants, "storage": kStorage,
}

// RenderPreview builds a single static frame (used by `vtui --preview [tab]`).
func RenderPreview(width, height int, tab string) string {
	m := New()
	tm, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	m = tm.(Model)
	tm, _ = m.Update(sysCmd())
	m = tm.(Model)

	// detail previews: "detail" (VM), "netdetail", "tendetail", "stordetail"
	detailFor := map[string]kind{
		"detail": kVMs, "netdetail": kNetworks, "tendetail": kTenants, "stordetail": kStorage,
	}
	if dk, isDetail := detailFor[tab]; isDetail {
		m.active = dk
		tm, _ = m.Update(loadCmd(dk)())
		m = tm.(Model)
		if dk == kVMs {
			if v, ok := m.selectedVM(); ok {
				m.mode = modeDetail
				tm, _ = m.Update(detailCmd(v)())
				m = tm.(Model)
			}
		} else {
			m.openDetail() // instant detail (network/tenant/storage)
		}
		return m.View()
	}

	k, ok := tabByName[tab]
	if !ok {
		k = kVMs
	}
	m.active = k
	tm, _ = m.Update(loadCmd(k)())
	m = tm.(Model)
	return m.View()
}
