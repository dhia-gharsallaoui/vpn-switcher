package views

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/network"
)

var (
	diagPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#06B6D4")).
			Padding(0, 1).
			MarginBottom(1)

	diagHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	diagHighlightStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#06B6D4"))

	diagSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#22C55E"))

	diagErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444"))

	diagDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	diagValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9FAFB"))

	diagActivePanelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#06B6D4"))

	diagInactivePanelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF"))

	diagInputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#06B6D4")).
			Padding(0, 1)
)

// DiagnosticsModel manages the diagnostics tab state.
type DiagnosticsModel struct {
	// Route table
	Routes      []network.Route
	RouteTable  string // "main", "local", etc.
	RouteTables []string
	RouteCursor int

	// DNS
	DNSResults []network.DNSResult
	DNSQuery   string
	DNSType    string // "A", "AAAA", etc.
	DNSInput   bool   // true when user is typing a domain
	DNSField   string // current input text
	DNSError   string

	// Ping
	PingResults []network.PingResult

	// Traceroute
	TraceHops  []network.Hop
	TraceError string

	// Panel focus
	ActivePanel int // 0=routes, 1=dns, 2=ping
}

// NewDiagnosticsModel creates a new diagnostics model.
func NewDiagnosticsModel() DiagnosticsModel {
	return DiagnosticsModel{
		RouteTable:  "main",
		RouteTables: []string{"main", "local"},
		DNSType:     "A",
	}
}

// SetRoutes updates the displayed routes.
func (m *DiagnosticsModel) SetRoutes(routes []network.Route, table string) {
	m.Routes = routes
	m.RouteTable = table
	if m.RouteCursor >= len(routes) && len(routes) > 0 {
		m.RouteCursor = len(routes) - 1
	}
}

// SetDNSResults updates the DNS results.
func (m *DiagnosticsModel) SetDNSResults(results []network.DNSResult) {
	m.DNSResults = results
	m.DNSError = ""
}

// SetDNSError records a DNS lookup error.
func (m *DiagnosticsModel) SetDNSError(err string) {
	m.DNSError = err
	m.DNSResults = nil
}

// AddPingResult adds or updates a ping result for an interface.
func (m *DiagnosticsModel) AddPingResult(result network.PingResult) {
	for i, r := range m.PingResults {
		if r.Interface == result.Interface {
			m.PingResults[i] = result
			return
		}
	}
	m.PingResults = append(m.PingResults, result)
}

// SetTraceHops updates the traceroute results.
func (m *DiagnosticsModel) SetTraceHops(hops []network.Hop) {
	m.TraceHops = hops
	m.TraceError = ""
}

// SetTraceError records a traceroute error.
func (m *DiagnosticsModel) SetTraceError(err string) {
	m.TraceError = err
	m.TraceHops = nil
}

// CycleTable advances to the next routing table name.
func (m *DiagnosticsModel) CycleTable() string {
	if len(m.RouteTables) == 0 {
		return "main"
	}
	for i, t := range m.RouteTables {
		if t == m.RouteTable {
			next := (i + 1) % len(m.RouteTables)
			m.RouteTable = m.RouteTables[next]
			m.RouteCursor = 0
			return m.RouteTable
		}
	}
	m.RouteTable = m.RouteTables[0]
	m.RouteCursor = 0
	return m.RouteTable
}

// AddCustomTable adds a table name if not already present.
func (m *DiagnosticsModel) AddCustomTable(name string) {
	for _, t := range m.RouteTables {
		if t == name {
			return
		}
	}
	m.RouteTables = append(m.RouteTables, name)
}

// StartDNSInput enters DNS input mode.
func (m *DiagnosticsModel) StartDNSInput() {
	m.DNSInput = true
	m.DNSField = ""
}

// CancelDNSInput exits DNS input mode.
func (m *DiagnosticsModel) CancelDNSInput() {
	m.DNSInput = false
}

// FinishDNSInput exits DNS input mode and sets the query.
func (m *DiagnosticsModel) FinishDNSInput() string {
	m.DNSInput = false
	m.DNSQuery = m.DNSField
	return m.DNSQuery
}

// TypeDNSChar adds a character to the DNS input.
func (m *DiagnosticsModel) TypeDNSChar(ch string) {
	m.DNSField += ch
}

// BackspaceDNS removes the last character from DNS input.
func (m *DiagnosticsModel) BackspaceDNS() {
	if len(m.DNSField) > 0 {
		m.DNSField = m.DNSField[:len(m.DNSField)-1]
	}
}

// MoveUp moves cursor up in the active panel.
func (m *DiagnosticsModel) MoveUp() {
	switch m.ActivePanel {
	case 0:
		if m.RouteCursor > 0 {
			m.RouteCursor--
		}
	}
}

// MoveDown moves cursor down in the active panel.
func (m *DiagnosticsModel) MoveDown() {
	switch m.ActivePanel {
	case 0:
		if m.RouteCursor < len(m.Routes)-1 {
			m.RouteCursor++
		}
	}
}

// NextPanel cycles to the next sub-panel.
func (m *DiagnosticsModel) NextPanel() {
	m.ActivePanel = (m.ActivePanel + 1) % 3
}

// PrevPanel cycles to the previous sub-panel.
func (m *DiagnosticsModel) PrevPanel() {
	m.ActivePanel = (m.ActivePanel + 2) % 3
}

// View renders the diagnostics tab with all three panels stacked vertically.
func (m *DiagnosticsModel) View(width int) string {
	var b strings.Builder

	b.WriteString(m.renderRoutePanel(width))
	b.WriteString("\n")
	b.WriteString(m.renderDNSPanel(width))
	b.WriteString("\n")
	b.WriteString(m.renderPingPanel(width))

	if len(m.TraceHops) > 0 || m.TraceError != "" {
		b.WriteString("\n")
		b.WriteString(m.renderTraceroutePanel(width))
	}

	return b.String()
}

func (m *DiagnosticsModel) panelTitle(name string, panelIndex int) string {
	if m.ActivePanel == panelIndex {
		return diagActivePanelStyle.Render(fmt.Sprintf("▸ %s", name))
	}
	return diagInactivePanelStyle.Render(fmt.Sprintf("  %s", name))
}

func (m *DiagnosticsModel) renderRoutePanel(width int) string {
	var b strings.Builder

	title := m.panelTitle(fmt.Sprintf("Route Table [%s]", m.RouteTable), 0)
	b.WriteString(title)
	b.WriteString("\n")

	if len(m.Routes) == 0 {
		b.WriteString(diagDimStyle.Render("  No routes. Press t to switch tables."))
		return diagPanelStyle.Width(width - 4).Render(b.String())
	}

	// Header
	b.WriteString(diagHeaderStyle.Render(fmt.Sprintf("  %-22s %-16s %-12s %-8s %-8s %s", "Destination", "Via", "Dev", "Scope", "Metric", "Proto")))
	b.WriteString("\n")

	for i, r := range m.Routes {
		via := r.Via
		if via == "" {
			via = "-"
		}
		scope := r.Scope
		if scope == "" {
			scope = "-"
		}
		metric := r.Metric
		if metric == "" {
			metric = "-"
		}
		proto := r.Proto
		if proto == "" {
			proto = "-"
		}

		line := fmt.Sprintf("%-22s %-16s %-12s %-8s %-8s %s", r.Dest, via, r.Dev, scope, metric, proto)

		if i == m.RouteCursor && m.ActivePanel == 0 {
			b.WriteString(diagHighlightStyle.Render(fmt.Sprintf("▸ %s", line)))
		} else {
			b.WriteString(diagValueStyle.Render(fmt.Sprintf("  %s", line)))
		}
		b.WriteString("\n")
	}

	return diagPanelStyle.Width(width - 4).Render(b.String())
}

func (m *DiagnosticsModel) renderDNSPanel(width int) string {
	var b strings.Builder

	queryInfo := ""
	if m.DNSQuery != "" {
		queryInfo = fmt.Sprintf(" — %s %s", m.DNSQuery, m.DNSType)
	}
	title := m.panelTitle(fmt.Sprintf("DNS Lookup%s", queryInfo), 1)
	b.WriteString(title)
	b.WriteString("\n")

	if m.DNSInput {
		b.WriteString(diagValueStyle.Render("  Domain: "))
		b.WriteString(diagHighlightStyle.Render(m.DNSField))
		b.WriteString(diagHighlightStyle.Render("█"))
		b.WriteString("\n")
		b.WriteString(diagDimStyle.Render("  [enter] lookup  [esc] cancel"))
		return diagPanelStyle.Width(width - 4).Render(b.String())
	}

	if m.DNSError != "" {
		b.WriteString(diagErrorStyle.Render(fmt.Sprintf("  Error: %s", m.DNSError)))
		return diagPanelStyle.Width(width - 4).Render(b.String())
	}

	if len(m.DNSResults) == 0 {
		b.WriteString(diagDimStyle.Render("  Press / to perform a DNS lookup."))
		return diagPanelStyle.Width(width - 4).Render(b.String())
	}

	// Header
	b.WriteString(diagHeaderStyle.Render(fmt.Sprintf("  %-30s %-6s %-6s %-40s %s", "Domain", "Type", "TTL", "Value", "Server")))
	b.WriteString("\n")

	for _, r := range m.DNSResults {
		server := r.Server
		if server == "" {
			server = "-"
		}
		line := fmt.Sprintf("  %-30s %-6s %-6d %-40s %s", r.Domain, r.Type, r.TTL, r.Value, server)
		b.WriteString(diagValueStyle.Render(line))
		b.WriteString("\n")
	}

	if len(m.DNSResults) > 0 {
		b.WriteString(diagDimStyle.Render(fmt.Sprintf("  Query time: %s", m.DNSResults[0].ResponseTime)))
	}

	return diagPanelStyle.Width(width - 4).Render(b.String())
}

func (m *DiagnosticsModel) renderPingPanel(width int) string {
	var b strings.Builder

	title := m.panelTitle("Ping Results", 2)
	b.WriteString(title)
	b.WriteString("\n")

	if len(m.PingResults) == 0 {
		b.WriteString(diagDimStyle.Render("  Press p to ping all interfaces."))
		return diagPanelStyle.Width(width - 4).Render(b.String())
	}

	// Header
	b.WriteString(diagHeaderStyle.Render(fmt.Sprintf("  %-16s %-20s %-12s %s", "Interface", "Target", "Latency", "Status")))
	b.WriteString("\n")

	for _, r := range m.PingResults {
		var statusStr, latStr string
		if r.Success {
			statusStr = diagSuccessStyle.Render("OK")
			latStr = fmt.Sprintf("%.2f ms", float64(r.Latency.Microseconds())/1000.0)
		} else {
			statusStr = diagErrorStyle.Render("timeout")
			latStr = "-"
		}

		line := fmt.Sprintf("  %-16s %-20s %-12s %s", r.Interface, r.Target, latStr, statusStr)
		b.WriteString(diagValueStyle.Render(line))
		b.WriteString("\n")
	}

	return diagPanelStyle.Width(width - 4).Render(b.String())
}

func (m *DiagnosticsModel) renderTraceroutePanel(width int) string {
	var b strings.Builder

	b.WriteString(diagHeaderStyle.Render("  Traceroute"))
	b.WriteString("\n")

	if m.TraceError != "" {
		b.WriteString(diagErrorStyle.Render(fmt.Sprintf("  Error: %s", m.TraceError)))
		return diagPanelStyle.Width(width - 4).Render(b.String())
	}

	b.WriteString(diagHeaderStyle.Render(fmt.Sprintf("  %-4s %-20s %s", "Hop", "IP", "Latency")))
	b.WriteString("\n")

	for _, h := range m.TraceHops {
		var latStr string
		if h.IP == "*" {
			latStr = "*"
		} else {
			latStr = fmt.Sprintf("%.2f ms", float64(h.Latency.Microseconds())/1000.0)
		}

		line := fmt.Sprintf("  %-4d %-20s %s", h.Number, h.IP, latStr)
		b.WriteString(diagValueStyle.Render(line))
		b.WriteString("\n")
	}

	return diagPanelStyle.Width(width - 4).Render(b.String())
}
