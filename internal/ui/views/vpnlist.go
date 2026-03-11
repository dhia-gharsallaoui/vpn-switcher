package views

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/vpn"
)

var (
	cursorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
	itemStyle       = lipgloss.NewStyle().PaddingLeft(2)
	selectedStyle   = lipgloss.NewStyle().PaddingLeft(0).Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	connStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E"))
	disconnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	connectingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
	vpnListTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F9FAFB")).MarginBottom(1)
)

// VPNListModel holds the state for the VPN list view.
type VPNListModel struct {
	VPNs   []vpn.VPN
	Cursor int
}

// NewVPNListModel creates a new VPN list model.
func NewVPNListModel() VPNListModel {
	return VPNListModel{}
}

// SetVPNs updates the VPN list.
func (m *VPNListModel) SetVPNs(vpns []vpn.VPN) {
	m.VPNs = vpns
	if m.Cursor >= len(vpns) && len(vpns) > 0 {
		m.Cursor = len(vpns) - 1
	}
}

// Selected returns the currently selected VPN, if any.
func (m *VPNListModel) Selected() (vpn.VPN, bool) {
	if len(m.VPNs) == 0 || m.Cursor < 0 || m.Cursor >= len(m.VPNs) {
		return vpn.VPN{}, false
	}
	return m.VPNs[m.Cursor], true
}

// MoveUp moves the cursor up.
func (m *VPNListModel) MoveUp() {
	if m.Cursor > 0 {
		m.Cursor--
	}
}

// MoveDown moves the cursor down.
func (m *VPNListModel) MoveDown() {
	if m.Cursor < len(m.VPNs)-1 {
		m.Cursor++
	}
}

// View renders the VPN list.
func (m *VPNListModel) View(width int) string {
	if len(m.VPNs) == 0 {
		return itemStyle.Render("No VPNs discovered. Press r to refresh.")
	}

	var b strings.Builder
	b.WriteString(vpnListTitle.Render("VPN Connections"))
	b.WriteString("\n")

	for i, v := range m.VPNs {
		// Status indicator
		var statusStr string
		switch v.Status {
		case vpn.StatusConnected:
			statusStr = connStyle.Render("[connected]")
		case vpn.StatusConnecting:
			statusStr = connectingStyle.Render("[connecting...]")
		case vpn.StatusDisconnecting:
			statusStr = connectingStyle.Render("[disconnecting...]")
		default:
			statusStr = disconnStyle.Render("[off]")
		}

		// Provider prefix
		name := v.Name
		if v.Provider == vpn.ProviderOpenVPN {
			name = fmt.Sprintf("OpenVPN: %s", v.Name)
		}

		line := fmt.Sprintf("%s  %s", name, statusStr)

		if i == m.Cursor {
			b.WriteString(selectedStyle.Render(fmt.Sprintf("▸ %s", line)))
		} else {
			b.WriteString(itemStyle.Render(fmt.Sprintf("  %s", line)))
		}
		b.WriteString("\n")
	}

	return b.String()
}
