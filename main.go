// vergeos-tui — a terminal UI for managing VergeOS, driving the `vrg` CLI in
// the background. Built on Bubble Tea (charmbracelet/bubbletea).
package main

import (
	"fmt"
	"os"

	"vergeos-tui/internal/ui"
	"vergeos-tui/internal/vrg"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// --selftest exercises the vrg backend without a TTY, so the data path
	// can be verified in CI / non-interactive shells.
	if len(os.Args) > 1 && os.Args[1] == "--selftest" {
		vms, err := vrg.ListVMs()
		if err != nil {
			fmt.Fprintln(os.Stderr, "selftest FAILED:", err)
			os.Exit(1)
		}
		fmt.Printf("VMs:       %d\n", len(vms))
		if nets, err := vrg.ListNetworks(); err != nil {
			fmt.Fprintln(os.Stderr, "networks:", err)
			os.Exit(1)
		} else {
			fmt.Printf("Networks:  %d\n", len(nets))
		}
		if ts, err := vrg.ListTenants(); err != nil {
			fmt.Fprintln(os.Stderr, "tenants:", err)
			os.Exit(1)
		} else {
			fmt.Printf("Tenants:   %d\n", len(ts))
		}
		if ss, err := vrg.ListStorage(); err != nil {
			fmt.Fprintln(os.Stderr, "storage:", err)
			os.Exit(1)
		} else {
			for _, s := range ss {
				fmt.Printf("Storage:   tier %d  %.0f/%.0f GB (%.0f%%)  dedupe %.1fx\n",
					s.Tier, s.UsedGB, s.CapacityGB, s.UsedPercent, s.DedupeRatio)
			}
		}
		// detail path (drives + nics) for the first VM
		if len(vms) > 0 {
			d, err := vrg.GetVMDetail(vms[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, "detail:", err)
				os.Exit(1)
			}
			fmt.Printf("Detail:    %s → %d drives, %d NICs\n", d.VM.Name, len(d.Drives), len(d.NICs))
		}
		fmt.Println("selftest OK — all backends parsed")
		return
	}

	// --preview [tab] renders one static frame to stdout (no TTY required).
	// tab ∈ vms|networks|tenants|storage|detail (default vms).
	if len(os.Args) > 1 && os.Args[1] == "--preview" {
		tab := "vms"
		if len(os.Args) > 2 {
			tab = os.Args[2]
		}
		fmt.Println(ui.RenderPreview(100, 24, tab))
		return
	}

	p := tea.NewProgram(ui.New(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func ram(mb int) string {
	if mb >= 1024 && mb%1024 == 0 {
		return fmt.Sprintf("%dGB", mb/1024)
	}
	return fmt.Sprintf("%dMB", mb)
}
