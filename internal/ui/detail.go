package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// detailView dispatches to the per-kind detail panel.
func (m Model) detailView() string {
	switch m.detailKind {
	case kNetworks:
		return m.netDetailView()
	case kTenants:
		return m.tenDetailView()
	case kStorage:
		return m.storDetailView()
	default:
		return m.vmDetailView()
	}
}

// vmDetailView renders the selected VM's summary, drives, and NICs.
func (m Model) vmDetailView() string {
	if m.detailErr != nil {
		return errStyle.Render("Error loading detail: " + m.detailErr.Error())
	}
	if m.detail == nil {
		return "  " + m.spinner.View() + "loading detail…"
	}
	d := m.detail
	var b strings.Builder

	b.WriteString(sectionStyle.Render("VM  ·  "+d.VM.Name) + "\n\n")
	b.WriteString(kv("Key", fmt.Sprint(d.VM.Key)))
	b.WriteString(kv("Status", dash(d.VM.Status)))
	b.WriteString(kv("CPU", fmt.Sprintf("%d cores", d.VM.CPUCores)))
	b.WriteString(kv("RAM", ramStr(d.VM.RAM)))
	b.WriteString(kv("Node", dash(d.VM.NodeName)))
	b.WriteString(kv("Cluster", dash(d.VM.ClusterName)))
	b.WriteString(kv("OS", dash(d.VM.OSFamily)))

	b.WriteString("\n" + sectionStyle.Render(fmt.Sprintf("Drives (%d)", len(d.Drives))) + "\n")
	if len(d.Drives) == 0 {
		b.WriteString(hintStyle.Render("  none") + "\n")
	}
	for _, dr := range d.Drives {
		b.WriteString(fmt.Sprintf("  %-16s  %-11s  %-11s  %6.1f GB  tier %s\n",
			trunc(dr.Name, 16), dr.Media, dr.Interface, dr.SizeGB, dash(dr.Tier)))
	}

	b.WriteString("\n" + sectionStyle.Render(fmt.Sprintf("NICs (%d)", len(d.NICs))) + "\n")
	if len(d.NICs) == 0 {
		b.WriteString(hintStyle.Render("  none") + "\n")
	}
	for _, n := range d.NICs {
		b.WriteString(fmt.Sprintf("  %-16s  %-12s  %-16s  %s\n",
			trunc(n.Name, 16), dash(n.NetworkName), dash(n.IPAddress), dash(n.MACAddress)))
	}

	b.WriteString("\n" + hintStyle.Render("esc/enter — back  ·  ↑/↓ scroll"))
	return detailBox.Render(b.String())
}

// netDetailView renders a network's configuration.
func (m Model) netDetailView() string {
	if m.detailNet == nil {
		return hintStyle.Render("no network selected")
	}
	n := *m.detailNet
	var b strings.Builder
	b.WriteString(sectionStyle.Render("Network  ·  "+n.Name) + "\n\n")
	b.WriteString(kv("Key", fmt.Sprint(n.Key)))
	b.WriteString(kv("Type", dash(n.Type)))
	b.WriteString(kv("State", vmStatusColor(runState(n.Running)).Render(runState(n.Running))))
	b.WriteString(kv("Address", dash(n.IPAddress)))
	b.WriteString(kv("Network", dash(n.Network)))
	b.WriteString(kv("Gateway", dash(n.Gateway)))
	b.WriteString(kv("MTU", fmt.Sprint(n.MTU)))
	b.WriteString("\n" + sectionStyle.Render("DHCP / DNS") + "\n")
	b.WriteString(kv("DHCP", yesno(n.DHCPEnabled)))
	if n.DHCPEnabled && (n.DHCPStart != "" || n.DHCPStop != "") {
		b.WriteString(kv("Range", dash(n.DHCPStart)+" – "+dash(n.DHCPStop)))
	}
	b.WriteString(kv("Domain", dash(n.Domain)))
	b.WriteString(kv("DNS", dash(n.DNS)))
	b.WriteString("\n" + hintStyle.Render("esc/enter — back  ·  ↑/↓ scroll"))
	return detailBox.Render(b.String())
}

// tenDetailView renders a tenant's summary.
func (m Model) tenDetailView() string {
	if m.detailTen == nil {
		return hintStyle.Render("no tenant selected")
	}
	t := *m.detailTen
	var b strings.Builder
	b.WriteString(sectionStyle.Render("Tenant  ·  "+t.Name) + "\n\n")
	b.WriteString(kv("Key", fmt.Sprint(t.Key)))
	b.WriteString(kv("State", vmStatusColor(t.State).Render(dash(t.State))))
	b.WriteString(kv("Network", dash(t.NetworkName)))
	b.WriteString(kv("UI IP", dash(t.UIAddressIP)))
	b.WriteString(kv("Isolated", yesno(t.IsIsolated)))
	b.WriteString(kv("URL", dash(t.URL)))
	b.WriteString(kv("UUID", dash(t.UUID)))
	if strings.TrimSpace(t.Note) != "" {
		b.WriteString(kv("Note", t.Note))
	}
	b.WriteString("\n" + hintStyle.Render("esc/enter — back  ·  ↑/↓ scroll"))
	return detailBox.Render(b.String())
}

// storDetailView renders a storage tier with a prominent usage bar.
func (m Model) storDetailView() string {
	if m.detailStor == nil {
		return hintStyle.Render("no tier selected")
	}
	s := *m.detailStor
	var b strings.Builder
	b.WriteString(sectionStyle.Render(fmt.Sprintf("Storage Tier %d", s.Tier)) + "\n\n")

	// big usage bar (40 cells) colored by threshold
	width := 40
	pctStr := fmt.Sprintf("%.1f%%", s.UsedPercent)
	barStr := usedPctColor(pctStr).Render(bar(s.UsedPercent, width))
	b.WriteString("  " + barStr + "  " + valueStyle.Render(pctStr) + "\n\n")

	b.WriteString(kv("Capacity", gb(s.CapacityGB)))
	b.WriteString(kv("Used", gb(s.UsedGB)))
	b.WriteString(kv("Free", gb(s.FreeGB)))
	b.WriteString(kv("Dedupe", fmt.Sprintf("%.1fx  (%.0f%% saved)", s.DedupeRatio, s.DedupeSavingsPercent)))
	b.WriteString(kv("Read ops", fmt.Sprintf("%.0f/s", s.ReadOps)))
	b.WriteString(kv("Write ops", fmt.Sprintf("%.0f/s", s.WriteOps)))
	b.WriteString("\n" + hintStyle.Render("esc/enter — back  ·  ↑/↓ scroll"))
	return detailBox.Render(b.String())
}

// confirmView renders the action confirmation modal.
func (m Model) confirmView() string {
	p := m.pending
	verb := strings.ToUpper(p.verb[:1]) + p.verb[1:]
	body := confirmTitle.Render(fmt.Sprintf("%s VM  “%s”  (key %d) ?", verb, p.name, p.key)) +
		"\n\n" + hintStyle.Render("y / enter — confirm     n / esc — cancel")
	box := confirmBox.Render(body)
	// center-ish
	return lipgloss.NewStyle().Padding(1, 0, 0, 4).Render(box)
}

func kv(label, value string) string {
	return "  " + labelStyle.Render(label) + valueStyle.Render(value) + "\n"
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}
