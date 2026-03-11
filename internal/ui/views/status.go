package views

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/network"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/vpn"
)

var (
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#06B6D4")).
			Padding(0, 1).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#06B6D4"))

	valStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB"))

	connectedDot   = lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E")).Render("●")
	disconnectedDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("○")
)

// StatusView renders the status panel showing active VPNs and interfaces.
func StatusView(vpns []vpn.VPN, ifaces []network.InterfaceInfo, publicIP string, width int) string {
	var b strings.Builder

	// Active VPNs
	var activeNames []string
	for _, v := range vpns {
		if v.Status == vpn.StatusConnected {
			info := v.Name
			if v.IP != "" {
				info += fmt.Sprintf(" (%s)", v.IP)
			}
			activeNames = append(activeNames, info)
		}
	}

	b.WriteString(labelStyle.Render("Active: "))
	if len(activeNames) > 0 {
		b.WriteString(valStyle.Render(strings.Join(activeNames, " + ")))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("None"))
	}
	b.WriteString("\n")

	// Interfaces
	if len(ifaces) > 0 {
		b.WriteString(labelStyle.Render("Interfaces: "))
		var ifaceStrs []string
		for _, iface := range ifaces {
			dot := disconnectedDot
			if iface.Up {
				dot = connectedDot
			}
			addr := ""
			if len(iface.Addrs) > 0 {
				addr = fmt.Sprintf(" [%s]", iface.Addrs[0])
			}
			ifaceStrs = append(ifaceStrs, fmt.Sprintf("%s %s%s", dot, iface.Name, addr))
		}
		b.WriteString(valStyle.Render(strings.Join(ifaceStrs, "  ")))
		b.WriteString("\n")
	}

	// Public IP
	b.WriteString(labelStyle.Render("Public IP: "))
	if publicIP != "" {
		b.WriteString(valStyle.Render(publicIP))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("checking..."))
	}

	return panelStyle.Width(width - 4).Render(b.String())
}
