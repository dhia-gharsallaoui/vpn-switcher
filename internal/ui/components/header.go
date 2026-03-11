package components

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

var (
	colorPrimary = lipgloss.Color("#7C3AED")
	colorText    = lipgloss.Color("#F9FAFB")
	colorDim     = lipgloss.Color("#9CA3AF")

	headerBg = lipgloss.NewStyle().
			Background(colorPrimary).
			Foreground(colorText).
			Bold(true).
			Padding(0, 1)

	tabActive = lipgloss.NewStyle().
			Background(colorPrimary).
			Foreground(colorText).
			Bold(true).
			Padding(0, 1)

	tabInactive = lipgloss.NewStyle().
			Background(colorPrimary).
			Foreground(colorDim).
			Padding(0, 1)
)

// ActiveTab identifies the current tab.
type ActiveTab int

const (
	TabVPNs       ActiveTab = iota
	TabRouting
	TabDiagnostics
)

// Header renders the top bar with title and tab indicators.
func Header(title string, activeTab ActiveTab, width int) string {
	tabs := []struct {
		label string
		tab   ActiveTab
	}{
		{"[1] VPNs", TabVPNs},
		{"[2] Routes", TabRouting},
		{"[3] Diagnostics", TabDiagnostics},
	}

	var tabStr string
	for _, t := range tabs {
		if t.tab == activeTab {
			tabStr += tabActive.Render(t.label)
		} else {
			tabStr += tabInactive.Render(t.label)
		}
	}

	titleRendered := headerBg.Render(fmt.Sprintf(" %s", title))

	gap := width - lipgloss.Width(titleRendered) - lipgloss.Width(tabStr)
	if gap < 0 {
		gap = 1
	}

	line := lipgloss.NewStyle().
		Background(colorPrimary).
		Width(width).
		Render(titleRendered + lipgloss.NewStyle().Background(colorPrimary).Width(gap).Render("") + tabStr)

	return line
}
