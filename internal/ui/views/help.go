package views

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	helpBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(1, 2).
		Width(50)

	helpTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginBottom(1)

	helpKey = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#06B6D4")).
		Width(12)

	helpDesc = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))
)

type helpEntry struct {
	key  string
	desc string
}

// HelpView renders the help overlay.
func HelpView() string {
	entries := []helpEntry{
		{"j/k, ↑↓", "Navigate list"},
		{"enter", "Connect VPN"},
		{"d", "Disconnect VPN"},
		{"r", "Refresh status"},
		{"tab", "Switch tab"},
		{"a", "Add routing rule"},
		{"x", "Delete routing rule"},
		{"space", "Toggle rule on/off"},
		{"?", "Toggle help"},
		{"q, Ctrl+C", "Quit"},
	}

	var b strings.Builder
	b.WriteString(helpTitle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	for _, e := range entries {
		b.WriteString(fmt.Sprintf("%s %s\n", helpKey.Render(e.key), helpDesc.Render(e.desc)))
	}

	return helpBox.Render(b.String())
}
