package components

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

var (
	colorSuccess = lipgloss.Color("#22C55E")
	colorDanger  = lipgloss.Color("#EF4444")
	colorMuted   = lipgloss.Color("#6B7280")

	barStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Padding(0, 1)

	barErrorStyle = lipgloss.NewStyle().
			Foreground(colorDanger).
			Padding(0, 1)

	barSuccessStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Padding(0, 1)
)

// StatusBar renders the bottom status bar.
func StatusBar(message string, isError bool, publicIP string, width int) string {
	var msgRendered string
	if isError {
		msgRendered = barErrorStyle.Render(message)
	} else if message != "" {
		msgRendered = barSuccessStyle.Render(message)
	} else {
		msgRendered = barStyle.Render("Ready")
	}

	ipStr := ""
	if publicIP != "" {
		ipStr = barStyle.Render(fmt.Sprintf("IP: %s", publicIP))
	}

	gap := width - lipgloss.Width(msgRendered) - lipgloss.Width(ipStr)
	if gap < 0 {
		gap = 1
	}

	return msgRendered + lipgloss.NewStyle().Width(gap).Render("") + ipStr
}

// HelpBar renders the quick help hints at the bottom.
func HelpBar(bindings []string, width int) string {
	hint := ""
	for i, b := range bindings {
		if i > 0 {
			hint += "  "
		}
		hint += b
	}
	return lipgloss.NewStyle().
		Foreground(colorMuted).
		Padding(0, 1).
		Width(width).
		Render(hint)
}
