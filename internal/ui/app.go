package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/config"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/network"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/system"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/ui/components"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/ui/views"
	"github.com/dhia-gharsallaoui/vpn-switcher/internal/vpn"
)

// App is the root Bubble Tea model.
type App struct {
	// Dependencies
	vpnManager *vpn.Manager
	routingMgr *network.RoutingManager
	ipFetcher  *network.IPFetcher
	ifaceMon   *network.InterfaceMonitor
	executor   system.CommandExecutor
	cfg        *config.Config

	// UI state
	activeTab components.ActiveTab
	width     int
	height    int
	keys      KeyMap

	// Sub-views
	vpnList        views.VPNListModel
	routingTab     views.RoutingModel
	diagnosticsTab views.DiagnosticsModel
	confirm        views.ConfirmModel
	helpOn         bool

	// Shared state
	spinner    spinner.Model
	loading    bool
	publicIP   string
	vpns       []vpn.VPN
	interfaces []network.InterfaceInfo
	statusMsg  string
	statusErr  bool

	// Pending action for confirm dialog
	pendingAction tea.Cmd
}

// NewApp creates a new App with all dependencies injected.
func NewApp(
	mgr *vpn.Manager,
	rmgr *network.RoutingManager,
	ipf *network.IPFetcher,
	imon *network.InterfaceMonitor,
	exec system.CommandExecutor,
	cfg *config.Config,
) App {
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(spinnerStyle),
	)

	routingModel := views.NewRoutingModel()
	// Load routing rules from config
	routingRules := make([]network.RoutingRule, len(cfg.RoutingRules))
	for i, r := range cfg.RoutingRules {
		routingRules[i] = network.RoutingRule{
			ID:           r.ID,
			Description:  r.Description,
			Enabled:      r.Enabled,
			Type:         network.RuleType(r.Type),
			DestCIDR:     r.DestCIDR,
			Domains:      r.Domains,
			VPNInterface: r.VPNInterface,
			Table:        r.Table,
		}
	}
	routingModel.SetRules(routingRules)

	// Build diagnostics model with custom tables from routing rules
	diagModel := views.NewDiagnosticsModel()
	for _, r := range cfg.RoutingRules {
		if r.Table != "" {
			diagModel.AddCustomTable(r.Table)
		}
	}

	return App{
		vpnManager:     mgr,
		routingMgr:     rmgr,
		ipFetcher:      ipf,
		ifaceMon:       imon,
		executor:       exec,
		cfg:            cfg,
		activeTab:      components.TabVPNs,
		keys:           DefaultKeyMap,
		vpnList:        views.NewVPNListModel(),
		routingTab:     routingModel,
		diagnosticsTab: diagModel,
		confirm:        views.NewConfirmModel(),
		spinner:        s,
		loading:        true,
	}
}

func (a App) timeouts() config.TimeoutConfig {
	return a.cfg.General.Timeouts
}

func (a App) vpnProfiles() []vpn.VPNProfile {
	profiles := make([]vpn.VPNProfile, len(a.cfg.VPNProfiles))
	for i, p := range a.cfg.VPNProfiles {
		profiles[i] = vpn.VPNProfile{
			Name:       p.Name,
			ConfigPath: p.ConfigPath,
			Provider:   p.Provider,
			ExitNode:   p.ExitNode,
			Interface:  p.Interface,
		}
	}
	return profiles
}

func (a App) Init() tea.Cmd {
	interval := time.Duration(a.cfg.General.IPCheckInterval) * time.Second
	return tea.Batch(
		discoverVPNsCmd(a.vpnManager, a.vpnProfiles(), a.timeouts()),
		fetchPublicIPCmd(a.ipFetcher, a.timeouts()),
		listInterfacesCmd(a.ifaceMon),
		tickCmd(interval),
		a.spinner.Tick,
	)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tea.KeyPressMsg:
		return a.handleKey(msg)

	case tea.PasteMsg:
		return a.handlePaste(msg)

	case spinner.TickMsg:
		if a.loading {
			var cmd tea.Cmd
			a.spinner, cmd = a.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case VPNListMsg:
		a.loading = false
		if msg.Err != nil {
			a.statusMsg = fmt.Sprintf("Discovery failed: %v", msg.Err)
			a.statusErr = true
		} else {
			a.vpns = msg.VPNs
			a.vpnList.SetVPNs(msg.VPNs)
			a.statusMsg = fmt.Sprintf("Found %d VPN(s)", len(msg.VPNs))
			a.statusErr = false
		}

	case ConnectResultMsg:
		a.loading = false
		if msg.Err != nil {
			a.statusMsg = fmt.Sprintf("Connect failed: %v", msg.Err)
			a.statusErr = true
		} else {
			a.statusMsg = fmt.Sprintf("Connected to %s", msg.VPN.Name)
			a.statusErr = false
			cmds = append(cmds,
				discoverVPNsCmd(a.vpnManager, a.vpnProfiles(), a.timeouts()),
				fetchPublicIPCmd(a.ipFetcher, a.timeouts()),
				listInterfacesCmd(a.ifaceMon),
			)
		}

	case DisconnectResultMsg:
		a.loading = false
		if msg.Err != nil {
			a.statusMsg = fmt.Sprintf("Disconnect failed: %v", msg.Err)
			a.statusErr = true
		} else {
			a.statusMsg = fmt.Sprintf("Disconnected from %s", msg.VPN.Name)
			a.statusErr = false
			cmds = append(cmds,
				discoverVPNsCmd(a.vpnManager, a.vpnProfiles(), a.timeouts()),
				fetchPublicIPCmd(a.ipFetcher, a.timeouts()),
				listInterfacesCmd(a.ifaceMon),
			)
		}

	case PublicIPMsg:
		if msg.Err == nil {
			a.publicIP = msg.IP
		}

	case InterfaceListMsg:
		if msg.Err == nil {
			a.interfaces = msg.Interfaces
		}

	case RoutingRuleAppliedMsg:
		if msg.Err != nil {
			a.statusMsg = fmt.Sprintf("Apply rule failed: %v", msg.Err)
			a.statusErr = true
		} else {
			a.statusMsg = fmt.Sprintf("Applied rule: %s", msg.Rule.DestCIDR)
			a.statusErr = false
		}

	case RoutingRuleRemovedMsg:
		if msg.Err != nil {
			a.statusMsg = fmt.Sprintf("Remove rule failed: %v", msg.Err)
			a.statusErr = true
		} else {
			a.statusMsg = fmt.Sprintf("Removed rule: %s", msg.Rule.DestCIDR)
			a.statusErr = false
		}

	case RouteListMsg:
		if msg.Err != nil {
			a.statusMsg = fmt.Sprintf("Route list failed: %v", msg.Err)
			a.statusErr = true
		} else {
			a.diagnosticsTab.SetRoutes(msg.Routes, msg.Table)
		}

	case PingResultMsg:
		if msg.Err != nil {
			// Still show the failed ping in the table
			msg.Result.Success = false
			a.diagnosticsTab.AddPingResult(msg.Result)
		} else {
			a.diagnosticsTab.AddPingResult(msg.Result)
		}

	case DNSResultMsg:
		if msg.Err != nil {
			a.diagnosticsTab.SetDNSError(msg.Err.Error())
			a.statusMsg = fmt.Sprintf("DNS lookup failed: %v", msg.Err)
			a.statusErr = true
		} else if len(msg.Results) == 0 {
			a.diagnosticsTab.SetDNSError(fmt.Sprintf("No %s records found for %s", a.diagnosticsTab.DNSType, a.diagnosticsTab.DNSQuery))
			a.statusMsg = "DNS: no records found"
			a.statusErr = false
		} else {
			a.diagnosticsTab.SetDNSResults(msg.Results)
			a.statusMsg = fmt.Sprintf("DNS: %d result(s)", len(msg.Results))
			a.statusErr = false
		}

	case TracerouteResultMsg:
		a.loading = false
		if msg.Err != nil {
			a.diagnosticsTab.SetTraceError(msg.Err.Error())
			a.statusMsg = fmt.Sprintf("Traceroute failed: %v", msg.Err)
			a.statusErr = true
		} else {
			a.diagnosticsTab.SetTraceHops(msg.Hops)
			a.statusMsg = fmt.Sprintf("Traceroute: %d hop(s)", len(msg.Hops))
			a.statusErr = false
		}

	case StatusMsg:
		a.statusMsg = msg.Text
		a.statusErr = msg.IsError

	case TickMsg:
		interval := time.Duration(a.cfg.General.IPCheckInterval) * time.Second
		cmds = append(cmds,
			discoverVPNsCmd(a.vpnManager, a.vpnProfiles(), a.timeouts()),
			fetchPublicIPCmd(a.ipFetcher, a.timeouts()),
			listInterfacesCmd(a.ifaceMon),
			tickCmd(interval),
		)
	}

	return a, tea.Batch(cmds...)
}

func (a App) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Confirm dialog takes priority
	if a.confirm.Visible {
		return a.handleConfirmKey(msg)
	}

	// Help overlay
	if a.helpOn {
		if key.Matches(msg, a.keys.Help) || key.Matches(msg, a.keys.Quit) || key.Matches(msg, a.keys.Cancel) {
			a.helpOn = false
		}
		return a, nil
	}

	// Routing form mode
	if a.activeTab == components.TabRouting && a.routingTab.Adding {
		return a.handleRoutingFormKey(msg)
	}

	// DNS input mode
	if a.activeTab == components.TabDiagnostics && a.diagnosticsTab.DNSInput {
		return a.handleDNSInputKey(msg)
	}

	// Global keys
	switch {
	case key.Matches(msg, a.keys.Quit):
		return a, tea.Quit
	case key.Matches(msg, a.keys.Help):
		a.helpOn = !a.helpOn
		return a, nil
	case key.Matches(msg, a.keys.TabNext):
		switch a.activeTab {
		case components.TabVPNs:
			a.activeTab = components.TabRouting
		case components.TabRouting:
			a.activeTab = components.TabDiagnostics
			return a, listRoutesCmd(a.executor, a.diagnosticsTab.RouteTable)
		default:
			a.activeTab = components.TabVPNs
		}
		return a, nil
	case key.Matches(msg, a.keys.TabPrev):
		switch a.activeTab {
		case components.TabVPNs:
			a.activeTab = components.TabDiagnostics
			return a, listRoutesCmd(a.executor, a.diagnosticsTab.RouteTable)
		case components.TabRouting:
			a.activeTab = components.TabVPNs
		default:
			a.activeTab = components.TabRouting
		}
		return a, nil
	case key.Matches(msg, a.keys.Refresh):
		a.loading = true
		a.statusMsg = "Refreshing..."
		a.statusErr = false
		return a, tea.Batch(
			discoverVPNsCmd(a.vpnManager, a.vpnProfiles(), a.timeouts()),
			fetchPublicIPCmd(a.ipFetcher, a.timeouts()),
			listInterfacesCmd(a.ifaceMon),
			a.spinner.Tick,
		)
	}

	// Tab-specific keys
	switch a.activeTab {
	case components.TabVPNs:
		return a.handleVPNKey(msg)
	case components.TabRouting:
		return a.handleRoutingKey(msg)
	case components.TabDiagnostics:
		return a.handleDiagnosticsKey(msg)
	}

	return a, nil
}

func (a App) handleVPNKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Up):
		a.vpnList.MoveUp()
	case key.Matches(msg, a.keys.Down):
		a.vpnList.MoveDown()
	case key.Matches(msg, a.keys.Connect):
		if selected, ok := a.vpnList.Selected(); ok {
			if selected.Status == vpn.StatusConnected {
				a.statusMsg = "Already connected"
				a.statusErr = false
				return a, nil
			}
			// Check if any VPN is active
			hasActive := false
			for _, v := range a.vpns {
				if v.Status == vpn.StatusConnected {
					hasActive = true
					break
				}
			}
			if hasActive && !a.cfg.General.AllowMultiVPN {
				a.confirm.Show("Switch VPN?", fmt.Sprintf("Disconnect current VPN and connect to %s?", selected.Name))
				a.pendingAction = switchVPNCmd(a.vpnManager, selected, a.timeouts())
				return a, nil
			}
			a.loading = true
			a.statusMsg = fmt.Sprintf("Connecting to %s...", selected.Name)
			a.statusErr = false
			var cmd tea.Cmd
			if a.cfg.General.AllowMultiVPN {
				cmd = connectMultiVPNCmd(a.vpnManager, selected, a.timeouts())
			} else {
				cmd = connectVPNCmd(a.vpnManager, selected, a.timeouts())
			}
			return a, tea.Batch(cmd, a.spinner.Tick)
		}
	case key.Matches(msg, a.keys.Disconnect):
		if selected, ok := a.vpnList.Selected(); ok {
			if selected.Status != vpn.StatusConnected {
				a.statusMsg = "Not connected"
				a.statusErr = false
				return a, nil
			}
			a.loading = true
			a.statusMsg = fmt.Sprintf("Disconnecting %s...", selected.Name)
			a.statusErr = false
			return a, tea.Batch(disconnectVPNCmd(a.vpnManager, selected, a.timeouts()), a.spinner.Tick)
		}
	}
	return a, nil
}

func (a App) handleRoutingKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Up):
		a.routingTab.MoveUp()
	case key.Matches(msg, a.keys.Down):
		a.routingTab.MoveDown()
	case key.Matches(msg, a.keys.AddRule):
		a.routingTab.StartAdd()
	case key.Matches(msg, a.keys.ToggleRule):
		if a.routingTab.ToggleSelected() {
			if rule, ok := a.routingTab.Selected(); ok {
				if rule.Enabled {
					return a, applyRoutingRuleCmd(a.routingMgr, rule)
				}
				return a, removeRoutingRuleCmd(a.routingMgr, rule)
			}
		}
	case key.Matches(msg, a.keys.DeleteRule):
		if rule, ok := a.routingTab.Selected(); ok {
			a.confirm.Show("Delete Rule?", fmt.Sprintf("Delete routing rule for %s?", rule.DestCIDR))
			a.pendingAction = tea.Batch(
				removeRoutingRuleCmd(a.routingMgr, rule),
				func() tea.Msg {
					return StatusMsg{Text: "Rule deleted"}
				},
			)
			return a, nil
		}
	}
	return a, nil
}

func (a App) handleRoutingFormKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.routingTab.CancelAdd()
	case "enter":
		if !a.routingTab.Validate() {
			return a, nil
		}
		rule := a.routingTab.BuildRule()
		a.routingTab.Rules = append(a.routingTab.Rules, rule)
		a.routingTab.CancelAdd()
		return a, applyRoutingRuleCmd(a.routingMgr, rule)
	case "tab":
		a.routingTab.MoveDown()
	case "shift+tab":
		a.routingTab.MoveUp()
	case "backspace":
		a.routingTab.Backspace()
	default:
		s := msg.String()
		if s != "" && !strings.HasPrefix(s, "ctrl+") && !strings.HasPrefix(s, "alt+") {
			a.routingTab.TypeChar(s)
		}
	}
	return a, nil
}

func (a App) handleDiagnosticsKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.PanelNext):
		a.diagnosticsTab.NextPanel()
		return a, nil
	case key.Matches(msg, a.keys.PanelPrev):
		a.diagnosticsTab.PrevPanel()
		return a, nil
	case key.Matches(msg, a.keys.Up):
		a.diagnosticsTab.MoveUp()
	case key.Matches(msg, a.keys.Down):
		a.diagnosticsTab.MoveDown()
	case key.Matches(msg, a.keys.SwitchTable):
		table := a.diagnosticsTab.CycleTable()
		return a, listRoutesCmd(a.executor, table)
	case key.Matches(msg, a.keys.DNSLookup):
		a.diagnosticsTab.StartDNSInput()
	case key.Matches(msg, a.keys.PingAll):
		a.diagnosticsTab.PingResults = nil // clear old results

		// Use resolved IP from DNS results when available
		target := a.diagnosticsTab.ResolvedIP()

		var cmds []tea.Cmd
		if target != "" {
			// Ping the resolved IP from each VPN interface
			for _, iface := range a.interfaces {
				if iface.Up && len(iface.Addrs) > 0 {
					cmds = append(cmds, pingInterfaceCmd(a.executor, iface.Name, target))
				}
			}
			// Also ping via default route
			cmds = append(cmds, pingTargetCmd(a.executor, target))
		} else {
			// No domain — ping each interface's gateway
			for _, iface := range a.interfaces {
				if iface.Up && len(iface.Addrs) > 0 {
					cmds = append(cmds, pingGatewayCmd(a.executor, iface.Name))
				}
			}
		}

		if len(cmds) == 0 {
			a.statusMsg = "No interfaces with addresses to ping"
			a.statusErr = false
			return a, nil
		}
		label := target
		if label == "" {
			label = "gateways"
		}
		a.statusMsg = fmt.Sprintf("Pinging %s via %d path(s)...", label, len(cmds))
		a.statusErr = false
		return a, tea.Batch(cmds...)
	case key.Matches(msg, a.keys.TracerouteKey):
		// Use resolved IP from DNS, or fall back to 8.8.8.8
		target := a.diagnosticsTab.ResolvedIP()
		if target == "" {
			target = "8.8.8.8"
		}
		iface := ""
		for _, i := range a.interfaces {
			if i.Up && len(i.Addrs) > 0 {
				iface = i.Name
				break
			}
		}
		a.loading = true
		a.statusMsg = fmt.Sprintf("Traceroute to %s...", target)
		a.statusErr = false
		return a, tea.Batch(tracerouteCmd(a.executor, target, iface), a.spinner.Tick)
	}
	return a, nil
}

func (a App) handleDNSInputKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.diagnosticsTab.CancelDNSInput()
	case "enter":
		domain := a.diagnosticsTab.FinishDNSInput()
		if domain == "" {
			return a, nil
		}
		a.statusMsg = fmt.Sprintf("Looking up %s...", domain)
		a.statusErr = false
		return a, dnsLookupCmd(a.executor, domain, a.diagnosticsTab.DNSType)
	case "backspace":
		a.diagnosticsTab.BackspaceDNS()
	case "left":
		a.diagnosticsTab.MoveDNSCursorLeft()
	case "right":
		a.diagnosticsTab.MoveDNSCursorRight()
	case "up", "down", "tab", "shift+tab":
		// ignore navigation keys in input mode
	default:
		s := msg.String()
		if s != "" && !strings.HasPrefix(s, "ctrl+") && !strings.HasPrefix(s, "alt+") && !strings.HasPrefix(s, "shift+") {
			a.diagnosticsTab.TypeDNSChar(s)
		}
	}
	return a, nil
}

func (a App) handlePaste(msg tea.PasteMsg) (tea.Model, tea.Cmd) {
	text := msg.String()
	if text == "" {
		return a, nil
	}

	if a.activeTab == components.TabDiagnostics && a.diagnosticsTab.DNSInput {
		a.diagnosticsTab.TypeDNSChar(text)
		return a, nil
	}

	if a.activeTab == components.TabRouting && a.routingTab.Adding {
		a.routingTab.TypeChar(text)
		return a, nil
	}

	return a, nil
}

func (a App) handleConfirmKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Confirm):
		a.confirm.Hide()
		a.loading = true
		pending := a.pendingAction
		a.pendingAction = nil
		return a, tea.Batch(pending, a.spinner.Tick)
	case key.Matches(msg, a.keys.Cancel):
		a.confirm.Hide()
		a.pendingAction = nil
		a.statusMsg = "Cancelled"
		a.statusErr = false
	}
	return a, nil
}

func (a App) View() tea.View {
	if a.width == 0 {
		return tea.NewView("Initializing...")
	}

	var b strings.Builder

	// Header
	b.WriteString(components.Header("vpn-switcher", a.activeTab, a.width))
	b.WriteString("\n")

	// Status panel
	b.WriteString(views.StatusView(a.vpns, a.interfaces, a.publicIP, a.width))
	b.WriteString("\n")

	// Loading spinner
	if a.loading {
		b.WriteString(fmt.Sprintf("  %s %s\n\n", a.spinner.View(), a.statusMsg))
	}

	// Active tab content
	switch a.activeTab {
	case components.TabVPNs:
		b.WriteString(a.vpnList.View(a.width))
	case components.TabRouting:
		b.WriteString(a.routingTab.View(a.width))
	case components.TabDiagnostics:
		b.WriteString(a.diagnosticsTab.View(a.width))
	}

	// Confirm dialog overlay
	if a.confirm.Visible {
		b.WriteString("\n")
		b.WriteString(a.confirm.View())
	}

	// Help overlay
	if a.helpOn {
		b.WriteString("\n")
		b.WriteString(views.HelpView(a.keys.AllBindings()))
	}

	// Calculate remaining height for padding
	contentHeight := strings.Count(b.String(), "\n")
	if a.height > contentHeight+3 {
		padding := a.height - contentHeight - 3
		b.WriteString(strings.Repeat("\n", padding))
	}

	// Status bar
	b.WriteString(components.StatusBar(a.statusMsg, a.statusErr, a.publicIP, a.width))
	b.WriteString("\n")

	// Help bar
	var bindings []string
	switch a.activeTab {
	case components.TabVPNs:
		for _, kb := range a.keys.HelpBindings() {
			h := kb.Help()
			bindings = append(bindings, fmt.Sprintf("%s %s", helpKeyStyle.Render(h.Key), helpDescStyle.Render(h.Desc)))
		}
	case components.TabRouting:
		for _, kb := range a.keys.RoutingHelpBindings() {
			h := kb.Help()
			bindings = append(bindings, fmt.Sprintf("%s %s", helpKeyStyle.Render(h.Key), helpDescStyle.Render(h.Desc)))
		}
	case components.TabDiagnostics:
		for _, kb := range a.keys.DiagnosticsHelpBindings() {
			h := kb.Help()
			bindings = append(bindings, fmt.Sprintf("%s %s", helpKeyStyle.Render(h.Key), helpDescStyle.Render(h.Desc)))
		}
	}
	b.WriteString(components.HelpBar(bindings, a.width))

	content := lipgloss.NewStyle().MaxWidth(a.width).MaxHeight(a.height).Render(b.String())

	var v tea.View
	v.SetContent(content)
	v.AltScreen = true
	return v
}
