package views

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/network"
)

const (
	maxRoutePanelRows = 10
	maxDNSPanelRows   = 6
	maxPingPanelRows  = 6
	maxTraceRows      = 8
	panelCount        = 4 // routes, dns, ping, traceroute
)

var (
	diagPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6B7280")).
			Padding(0, 1)

	diagActiveBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#06B6D4")).
				Padding(0, 1)

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
)

// DiagnosticsModel manages the diagnostics tab state.
type DiagnosticsModel struct {
	// Route table
	Routes      []network.Route
	RouteTable  string
	RouteTables []string
	RouteCursor int
	RouteScroll int

	// DNS
	DNSResults []network.DNSResult
	DNSQuery   string
	DNSType    string
	DNSInput   bool
	DNSField   string
	DNSCursor  int // cursor position within DNSField
	DNSError   string

	// Ping
	PingResults []network.PingResult

	// Traceroute
	TraceHops  []network.Hop
	TraceError string

	// Panel focus: 0=routes, 1=dns, 2=ping, 3=traceroute
	ActivePanel int
}

// NewDiagnosticsModel creates a new diagnostics model.
func NewDiagnosticsModel() DiagnosticsModel {
	return DiagnosticsModel{
		RouteTable:  "main",
		RouteTables: []string{"main", "local"},
		DNSType:     "A",
	}
}

// --- Data setters ---

func (m *DiagnosticsModel) SetRoutes(routes []network.Route, table string) {
	m.Routes = routes
	m.RouteTable = table
	m.RouteScroll = 0
	if m.RouteCursor >= len(routes) && len(routes) > 0 {
		m.RouteCursor = len(routes) - 1
	}
}

func (m *DiagnosticsModel) SetDNSResults(results []network.DNSResult) {
	m.DNSResults = results
	m.DNSError = ""
}

func (m *DiagnosticsModel) SetDNSError(err string) {
	m.DNSError = err
	m.DNSResults = nil
}

func (m *DiagnosticsModel) AddPingResult(result network.PingResult) {
	for i, r := range m.PingResults {
		if r.Interface == result.Interface {
			m.PingResults[i] = result
			return
		}
	}
	m.PingResults = append(m.PingResults, result)
}

func (m *DiagnosticsModel) SetTraceHops(hops []network.Hop) {
	m.TraceHops = hops
	m.TraceError = ""
}

func (m *DiagnosticsModel) SetTraceError(err string) {
	m.TraceError = err
	m.TraceHops = nil
}

// --- Table cycling ---

func (m *DiagnosticsModel) CycleTable() string {
	if len(m.RouteTables) == 0 {
		return "main"
	}
	for i, t := range m.RouteTables {
		if t == m.RouteTable {
			next := (i + 1) % len(m.RouteTables)
			m.RouteTable = m.RouteTables[next]
			m.RouteCursor = 0
			m.RouteScroll = 0
			return m.RouteTable
		}
	}
	m.RouteTable = m.RouteTables[0]
	m.RouteCursor = 0
	m.RouteScroll = 0
	return m.RouteTable
}

func (m *DiagnosticsModel) AddCustomTable(name string) {
	for _, t := range m.RouteTables {
		if t == name {
			return
		}
	}
	m.RouteTables = append(m.RouteTables, name)
}

// --- DNS input ---

func (m *DiagnosticsModel) StartDNSInput() {
	m.DNSInput = true
	m.DNSField = ""
	m.DNSCursor = 0
}

func (m *DiagnosticsModel) CancelDNSInput() {
	m.DNSInput = false
}

func (m *DiagnosticsModel) FinishDNSInput() string {
	m.DNSInput = false
	m.DNSQuery = m.DNSField
	return m.DNSQuery
}

func (m *DiagnosticsModel) TypeDNSChar(ch string) {
	m.DNSField = m.DNSField[:m.DNSCursor] + ch + m.DNSField[m.DNSCursor:]
	m.DNSCursor += len(ch)
}

func (m *DiagnosticsModel) BackspaceDNS() {
	if m.DNSCursor > 0 {
		m.DNSField = m.DNSField[:m.DNSCursor-1] + m.DNSField[m.DNSCursor:]
		m.DNSCursor--
	}
}

func (m *DiagnosticsModel) MoveDNSCursorLeft() {
	if m.DNSCursor > 0 {
		m.DNSCursor--
	}
}

func (m *DiagnosticsModel) MoveDNSCursorRight() {
	if m.DNSCursor < len(m.DNSField) {
		m.DNSCursor++
	}
}

// --- Resolved IP ---

func (m *DiagnosticsModel) ResolvedIP() string {
	for _, r := range m.DNSResults {
		if r.Type == "A" || r.Type == "AAAA" {
			return r.Value
		}
	}
	return m.DNSQuery
}

// --- Navigation ---

func (m *DiagnosticsModel) MoveUp() {
	if m.ActivePanel == 0 {
		if m.RouteCursor > 0 {
			m.RouteCursor--
			if m.RouteCursor < m.RouteScroll {
				m.RouteScroll = m.RouteCursor
			}
		}
	}
}

func (m *DiagnosticsModel) MoveDown() {
	if m.ActivePanel == 0 {
		if m.RouteCursor < len(m.Routes)-1 {
			m.RouteCursor++
			if m.RouteCursor >= m.RouteScroll+maxRoutePanelRows {
				m.RouteScroll = m.RouteCursor - maxRoutePanelRows + 1
			}
		}
	}
}

func (m *DiagnosticsModel) NextPanel() {
	m.ActivePanel = (m.ActivePanel + 1) % panelCount
}

func (m *DiagnosticsModel) PrevPanel() {
	m.ActivePanel = (m.ActivePanel + panelCount - 1) % panelCount
}

// --- Rendering ---

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

func (m *DiagnosticsModel) panelStyle(panelIndex int) lipgloss.Style {
	if m.ActivePanel == panelIndex {
		return diagActiveBorderStyle
	}
	return diagPanelStyle
}

func (m *DiagnosticsModel) panelTitle(name string, panelIndex int) string {
	if m.ActivePanel == panelIndex {
		return diagActivePanelStyle.Render(fmt.Sprintf("▸ %s", name))
	}
	return diagInactivePanelStyle.Render(fmt.Sprintf("  %s", name))
}

func scrollIndicator(offset, total, visible int) string {
	if total <= visible {
		return ""
	}
	return diagDimStyle.Render(fmt.Sprintf("  [%d-%d of %d] ↑↓ scroll", offset+1, min(offset+visible, total), total))
}

func (m *DiagnosticsModel) renderRoutePanel(width int) string {
	var b strings.Builder

	title := m.panelTitle(fmt.Sprintf("Route Table [%s]", m.RouteTable), 0)
	b.WriteString(title)
	b.WriteString("\n")

	if len(m.Routes) == 0 {
		b.WriteString(diagDimStyle.Render("  No routes. Press t to switch tables."))
		return m.panelStyle(0).Width(width - 4).Render(b.String())
	}

	b.WriteString(diagHeaderStyle.Render(fmt.Sprintf("  %-22s %-16s %-12s %-8s %-8s %s", "Destination", "Via", "Dev", "Scope", "Metric", "Proto")))
	b.WriteString("\n")

	end := m.RouteScroll + maxRoutePanelRows
	if end > len(m.Routes) {
		end = len(m.Routes)
	}

	for i := m.RouteScroll; i < end; i++ {
		r := m.Routes[i]
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

	ind := scrollIndicator(m.RouteScroll, len(m.Routes), maxRoutePanelRows)
	if ind != "" {
		b.WriteString(ind)
	}

	return m.panelStyle(0).Width(width - 4).Render(b.String())
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
		// Render input field with cursor
		before := m.DNSField[:m.DNSCursor]
		after := m.DNSField[m.DNSCursor:]
		b.WriteString(diagValueStyle.Render("  Domain: "))
		b.WriteString(diagHighlightStyle.Render(before))
		b.WriteString(diagHighlightStyle.Render("█"))
		b.WriteString(diagHighlightStyle.Render(after))
		b.WriteString("\n")
		b.WriteString(diagDimStyle.Render("  [enter] lookup  [esc] cancel  [←/→] move cursor"))
		return m.panelStyle(1).Width(width - 4).Render(b.String())
	}

	if m.DNSError != "" {
		b.WriteString(diagErrorStyle.Render(fmt.Sprintf("  Error: %s", m.DNSError)))
		return m.panelStyle(1).Width(width - 4).Render(b.String())
	}

	if len(m.DNSResults) == 0 {
		if m.DNSQuery != "" {
			b.WriteString(diagDimStyle.Render("  No results found. Press / to try another domain."))
		} else {
			b.WriteString(diagDimStyle.Render("  Press / to perform a DNS lookup."))
		}
		return m.panelStyle(1).Width(width - 4).Render(b.String())
	}

	b.WriteString(diagHeaderStyle.Render(fmt.Sprintf("  %-30s %-6s %-6s %-40s %s", "Domain", "Type", "TTL", "Value", "Server")))
	b.WriteString("\n")

	visible := maxDNSPanelRows
	if len(m.DNSResults) < visible {
		visible = len(m.DNSResults)
	}
	for i := 0; i < visible; i++ {
		r := m.DNSResults[i]
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
	if len(m.DNSResults) > maxDNSPanelRows {
		b.WriteString(diagDimStyle.Render(fmt.Sprintf("  ... and %d more", len(m.DNSResults)-maxDNSPanelRows)))
	}

	return m.panelStyle(1).Width(width - 4).Render(b.String())
}

func (m *DiagnosticsModel) renderPingPanel(width int) string {
	var b strings.Builder

	title := m.panelTitle("Ping Results", 2)
	b.WriteString(title)
	b.WriteString("\n")

	if len(m.PingResults) == 0 {
		b.WriteString(diagDimStyle.Render("  Press p to ping all interfaces."))
		return m.panelStyle(2).Width(width - 4).Render(b.String())
	}

	b.WriteString(diagHeaderStyle.Render(fmt.Sprintf("  %-16s %-20s %-12s %s", "Interface", "Target", "Latency", "Status")))
	b.WriteString("\n")

	visible := maxPingPanelRows
	if len(m.PingResults) < visible {
		visible = len(m.PingResults)
	}
	for i := 0; i < visible; i++ {
		r := m.PingResults[i]
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

	return m.panelStyle(2).Width(width - 4).Render(b.String())
}

func (m *DiagnosticsModel) renderTraceroutePanel(width int) string {
	var b strings.Builder

	title := m.panelTitle("Traceroute", 3)
	b.WriteString(title)
	b.WriteString("\n")

	if m.TraceError != "" {
		b.WriteString(diagErrorStyle.Render(fmt.Sprintf("  Error: %s", m.TraceError)))
		return m.panelStyle(3).Width(width - 4).Render(b.String())
	}

	b.WriteString(diagHeaderStyle.Render(fmt.Sprintf("  %-4s %-20s %s", "Hop", "IP", "Latency")))
	b.WriteString("\n")

	visible := maxTraceRows
	if len(m.TraceHops) < visible {
		visible = len(m.TraceHops)
	}
	for i := 0; i < visible; i++ {
		h := m.TraceHops[i]
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

	if len(m.TraceHops) > maxTraceRows {
		b.WriteString(diagDimStyle.Render(fmt.Sprintf("  ... and %d more hops", len(m.TraceHops)-maxTraceRows)))
	}

	return m.panelStyle(3).Width(width - 4).Render(b.String())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
