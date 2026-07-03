// Package vrg is the backend: it shells out to the `vrg` CLI with structured
// (-q -o json) output and parses the results into Go types. Keeping this
// separate from the UI means it can be exercised headlessly (see --selftest).
package vrg

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ---- types ----------------------------------------------------------------

// VM mirrors the fields `vrg -o json vm list` returns. Note the `$key` tag —
// VergeOS uses $key as the primary key.
type VM struct {
	Key         int    `json:"$key"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	CPUCores    int    `json:"cpu_cores"`
	RAM         int    `json:"ram"` // megabytes
	NodeName    string `json:"node_name"`
	ClusterName string `json:"cluster_name"`
	OSFamily    string `json:"os_family"`
	Description string `json:"description"`
}

type Network struct {
	Key         int    `json:"$key"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	IPAddress   string `json:"ipaddress"`
	Network     string `json:"network"` // CIDR
	Gateway     string `json:"gateway"`
	Running     bool   `json:"running"`
	DHCPEnabled bool   `json:"dhcp_enabled"`
	DHCPStart   string `json:"dhcp_start"`
	DHCPStop    string `json:"dhcp_stop"`
	Domain      string `json:"domain"`
	DNS         string `json:"dns"`
	MTU         int    `json:"mtu"`
}

type Tenant struct {
	Key         int    `json:"$key"`
	Name        string `json:"name"`
	State       string `json:"state"`
	NetworkName string `json:"network_name"`
	UIAddressIP string `json:"ui_address_ip"`
	IsIsolated  bool   `json:"is_isolated"`
	URL         string `json:"url"`
	UUID        string `json:"uuid"`
	Note        string `json:"note"`
}

type StorageTier struct {
	Tier                 int     `json:"tier"`
	CapacityGB           float64 `json:"capacity_gb"`
	UsedGB               float64 `json:"used_gb"`
	FreeGB               float64 `json:"free_gb"`
	UsedPercent          float64 `json:"used_percent"`
	DedupeRatio          float64 `json:"dedupe_ratio"`
	DedupeSavingsPercent float64 `json:"dedupe_savings_percent"`
	ReadOps              float64 `json:"read_ops"`
	WriteOps             float64 `json:"write_ops"`
	Description          string  `json:"description"`
}

type Drive struct {
	Key       int     `json:"$key"`
	Name      string  `json:"name"`
	Media     string  `json:"media"`
	Interface string  `json:"interface"`
	SizeGB    float64 `json:"size_gb"`
	Tier      string  `json:"tier"` // vrg returns tier as a string, e.g. "1"
	Enabled   bool    `json:"enabled"`
}

type NIC struct {
	Key         int    `json:"$key"`
	Name        string `json:"name"`
	Interface   string `json:"interface"`
	IPAddress   string `json:"ip_address"`
	MACAddress  string `json:"mac_address"`
	NetworkName string `json:"network_name"`
	Enabled     bool   `json:"enabled"`
}

// VMDetail bundles a VM with its drives and NICs for the detail view.
type VMDetail struct {
	VM     VM
	Drives []Drive
	NICs   []NIC
}

// SysInfo is the cluster-context summary from `vrg system info`.
type SysInfo struct {
	Host        string `json:"host"`
	Version     string `json:"version"`
	CloudName   string `json:"cloud_name"`
	VMsTotal    int    `json:"vms_total"`
	VMsOnline   int    `json:"vms_online"`
	NodesTotal  int    `json:"nodes_total"`
	NodesOnline int    `json:"nodes_online"`
	AlarmsTotal int    `json:"alarms_total"`
}

// ---- command runners ------------------------------------------------------

// runJSON invokes `vrg -q -o json <args...>`. Global flags (-q, -o) MUST come
// before the subcommand, which is why we prepend them here.
func runJSON(args ...string) ([]byte, error) {
	full := append([]string{"-q", "-o", "json"}, args...)
	cmd := exec.Command("vrg", full...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			msg := strings.TrimSpace(string(ee.Stderr))
			if msg == "" {
				msg = "no stderr"
			}
			return nil, fmt.Errorf("vrg %s (exit %d): %s",
				strings.Join(args, " "), ee.ExitCode(), msg)
		}
		return nil, fmt.Errorf("vrg %s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}

// run invokes a mutating `vrg -q <args...>` where we don't need JSON back.
func run(args ...string) error {
	cmd := exec.Command("vrg", append([]string{"-q"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return nil
}

func list[T any](args ...string) ([]T, error) {
	out, err := runJSON(args...)
	if err != nil {
		return nil, err
	}
	var v []T
	if err := json.Unmarshal(out, &v); err != nil {
		return nil, fmt.Errorf("parse %s: %w", strings.Join(args, " "), err)
	}
	return v, nil
}

// ---- list endpoints -------------------------------------------------------

func ListVMs() ([]VM, error)              { return list[VM]("vm", "list") }
func ListNetworks() ([]Network, error)    { return list[Network]("network", "list") }
func ListTenants() ([]Tenant, error)      { return list[Tenant]("tenant", "list") }
func ListStorage() ([]StorageTier, error) { return list[StorageTier]("storage", "list") }

func ListDrives(vmKey int) ([]Drive, error) {
	return list[Drive]("vm", "drive", "list", fmt.Sprint(vmKey))
}
func ListNICs(vmKey int) ([]NIC, error) {
	return list[NIC]("vm", "nic", "list", fmt.Sprint(vmKey))
}

// SystemInfo returns the cluster-context summary.
func SystemInfo() (SysInfo, error) {
	out, err := runJSON("system", "info")
	if err != nil {
		return SysInfo{}, err
	}
	var s SysInfo
	if err := json.Unmarshal(out, &s); err != nil {
		return SysInfo{}, fmt.Errorf("parse system info: %w", err)
	}
	return s, nil
}

// GetVMDetail loads a VM's drives and NICs.
func GetVMDetail(v VM) (VMDetail, error) {
	d := VMDetail{VM: v}
	drives, err := ListDrives(v.Key)
	if err != nil {
		return d, err
	}
	nics, err := ListNICs(v.Key)
	if err != nil {
		return d, err
	}
	d.Drives, d.NICs = drives, nics
	return d, nil
}

// ---- VM lifecycle actions -------------------------------------------------

func StartVM(key int) error   { return run("vm", "start", fmt.Sprint(key)) }
func StopVM(key int) error    { return run("vm", "stop", fmt.Sprint(key)) }
func RestartVM(key int) error { return run("vm", "restart", fmt.Sprint(key)) }
