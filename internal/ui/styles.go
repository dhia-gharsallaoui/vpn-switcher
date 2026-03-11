package ui

import "charm.land/lipgloss/v2"

var (
	// Colors
	colorPrimary   = lipgloss.Color("#7C3AED") // purple
	colorSecondary = lipgloss.Color("#06B6D4") // cyan
	colorSuccess   = lipgloss.Color("#22C55E") // green
	colorDanger    = lipgloss.Color("#EF4444") // red
	colorWarning   = lipgloss.Color("#F59E0B") // amber
	colorMuted     = lipgloss.Color("#6B7280") // gray
	colorText      = lipgloss.Color("#F9FAFB") // near-white
	colorDimText   = lipgloss.Color("#9CA3AF") // light gray

	// Header
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText).
			Background(colorPrimary).
			Padding(0, 1)

	tabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText).
			Background(colorPrimary).
			Padding(0, 1)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(colorDimText).
				Padding(0, 1)

	// Status panel
	statusPanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorSecondary).
				Padding(0, 1).
				MarginBottom(1)

	statusLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSecondary)

	statusValueStyle = lipgloss.NewStyle().
				Foreground(colorText)

	// VPN list items
	vpnConnectedStyle = lipgloss.NewStyle().
				Foreground(colorSuccess)

	vpnDisconnectedStyle = lipgloss.NewStyle().
				Foreground(colorMuted)

	vpnSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary)

	// Status bar
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorDimText).
			Padding(0, 1)

	statusBarErrorStyle = lipgloss.NewStyle().
				Foreground(colorDanger).
				Padding(0, 1)

	statusBarSuccessStyle = lipgloss.NewStyle().
				Foreground(colorSuccess).
				Padding(0, 1)

	// Confirm dialog
	confirmBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorWarning).
			Padding(1, 2).
			Width(50)

	confirmTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorWarning)

	// Help
	helpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSecondary)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(colorDimText)

	// Routing table
	routingHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary)

	routingEnabledStyle = lipgloss.NewStyle().
				Foreground(colorSuccess)

	routingDisabledStyle = lipgloss.NewStyle().
				Foreground(colorMuted)

	// General
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorDanger)
)
