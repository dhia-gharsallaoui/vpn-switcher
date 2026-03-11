package views

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
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

// HelpView renders the help overlay from actual key bindings.
func HelpView(bindings []key.Binding) string {
	var b strings.Builder
	b.WriteString(helpTitle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	for _, kb := range bindings {
		h := kb.Help()
		if h.Key == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("%s %s\n", helpKey.Render(h.Key), helpDesc.Render(h.Desc)))
	}

	return helpBox.Render(b.String())
}
