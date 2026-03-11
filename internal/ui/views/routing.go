package views

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/network"
)

var (
	rtTitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F9FAFB")).MarginBottom(1)
	rtHeaderStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	rtEnabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E"))
	rtDisabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	rtSelectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	rtNormalStyle   = lipgloss.NewStyle().PaddingLeft(2)
	rtFormStyle     = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#06B6D4")).
			Padding(0, 1).
			MarginTop(1)
)

// RoutingModel manages the routing rules view.
type RoutingModel struct {
	Rules  []network.RoutingRule
	Cursor int
	Adding bool

	// Form fields for adding a new rule
	FormCIDR      string
	FormInterface string
	FormTable     string
	FormField     int // 0=CIDR, 1=interface, 2=table
}

// NewRoutingModel creates a new routing model.
func NewRoutingModel() RoutingModel {
	return RoutingModel{}
}

// SetRules updates the routing rules.
func (m *RoutingModel) SetRules(rules []network.RoutingRule) {
	m.Rules = rules
	if m.Cursor >= len(rules) && len(rules) > 0 {
		m.Cursor = len(rules) - 1
	}
}

// Selected returns the currently selected rule.
func (m *RoutingModel) Selected() (network.RoutingRule, bool) {
	if len(m.Rules) == 0 || m.Cursor < 0 || m.Cursor >= len(m.Rules) {
		return network.RoutingRule{}, false
	}
	return m.Rules[m.Cursor], true
}

// MoveUp moves the cursor up.
func (m *RoutingModel) MoveUp() {
	if m.Adding {
		if m.FormField > 0 {
			m.FormField--
		}
		return
	}
	if m.Cursor > 0 {
		m.Cursor--
	}
}

// MoveDown moves the cursor down.
func (m *RoutingModel) MoveDown() {
	if m.Adding {
		if m.FormField < 2 {
			m.FormField++
		}
		return
	}
	if m.Cursor < len(m.Rules)-1 {
		m.Cursor++
	}
}

// ToggleSelected toggles the enabled state of the selected rule.
func (m *RoutingModel) ToggleSelected() bool {
	if m.Cursor >= 0 && m.Cursor < len(m.Rules) {
		m.Rules[m.Cursor].Enabled = !m.Rules[m.Cursor].Enabled
		return true
	}
	return false
}

// StartAdd enters add mode.
func (m *RoutingModel) StartAdd() {
	m.Adding = true
	m.FormCIDR = ""
	m.FormInterface = "tun0"
	m.FormTable = "100"
	m.FormField = 0
}

// CancelAdd exits add mode.
func (m *RoutingModel) CancelAdd() {
	m.Adding = false
}

// TypeChar adds a character to the current form field.
func (m *RoutingModel) TypeChar(ch string) {
	switch m.FormField {
	case 0:
		m.FormCIDR += ch
	case 1:
		m.FormInterface += ch
	case 2:
		m.FormTable += ch
	}
}

// Backspace deletes the last character from the current form field.
func (m *RoutingModel) Backspace() {
	switch m.FormField {
	case 0:
		if len(m.FormCIDR) > 0 {
			m.FormCIDR = m.FormCIDR[:len(m.FormCIDR)-1]
		}
	case 1:
		if len(m.FormInterface) > 0 {
			m.FormInterface = m.FormInterface[:len(m.FormInterface)-1]
		}
	case 2:
		if len(m.FormTable) > 0 {
			m.FormTable = m.FormTable[:len(m.FormTable)-1]
		}
	}
}

// BuildRule creates a RoutingRule from the form fields.
func (m *RoutingModel) BuildRule() network.RoutingRule {
	return network.RoutingRule{
		ID:           fmt.Sprintf("rule-%d", len(m.Rules)+1),
		Description:  fmt.Sprintf("Route %s via %s", m.FormCIDR, m.FormInterface),
		Enabled:      true,
		Type:         network.RuleTypeSubnet,
		DestCIDR:     m.FormCIDR,
		VPNInterface: m.FormInterface,
		Table:        m.FormTable,
	}
}

// View renders the routing rules table.
func (m *RoutingModel) View(width int) string {
	var b strings.Builder

	b.WriteString(rtTitleStyle.Render("Routing Rules"))
	b.WriteString("\n")

	if len(m.Rules) == 0 && !m.Adding {
		b.WriteString(rtNormalStyle.Render("No routing rules configured. Press a to add one."))
		return b.String()
	}

	// Header
	b.WriteString(rtHeaderStyle.Render(fmt.Sprintf("  %-20s %-15s %-8s %s", "Destination", "Interface", "Table", "Status")))
	b.WriteString("\n")

	// Rules
	for i, rule := range m.Rules {
		statusStr := rtEnabledStyle.Render("[on] ")
		if !rule.Enabled {
			statusStr = rtDisabledStyle.Render("[off]")
		}

		line := fmt.Sprintf("%-20s %-15s %-8s %s", rule.DestCIDR, rule.VPNInterface, rule.Table, statusStr)

		if i == m.Cursor && !m.Adding {
			b.WriteString(rtSelectedStyle.Render(fmt.Sprintf("▸ %s", line)))
		} else {
			b.WriteString(rtNormalStyle.Render(fmt.Sprintf("  %s", line)))
		}
		b.WriteString("\n")
	}

	// Add form
	if m.Adding {
		b.WriteString(m.renderForm())
	}

	return b.String()
}

func (m *RoutingModel) renderForm() string {
	fields := []struct {
		label string
		value string
	}{
		{"CIDR", m.FormCIDR},
		{"Interface", m.FormInterface},
		{"Table", m.FormTable},
	}

	var b strings.Builder
	b.WriteString("Add Routing Rule\n\n")

	for i, f := range fields {
		cursor := "  "
		if i == m.FormField {
			cursor = "▸ "
		}
		b.WriteString(fmt.Sprintf("%s%-12s: %s", cursor, f.label, f.value))
		if i == m.FormField {
			b.WriteString("█") // cursor indicator
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render("[enter] save  [esc] cancel  [tab] next field"))

	return rtFormStyle.Render(b.String())
}
