package views

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

var (
	confirmBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F59E0B")).
			Padding(1, 2).
			Width(50)

	confirmTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F59E0B"))

	confirmHint = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))
)

// ConfirmModel represents a confirmation dialog.
type ConfirmModel struct {
	Title   string
	Message string
	Visible bool
}

// NewConfirmModel creates a new confirmation dialog.
func NewConfirmModel() ConfirmModel {
	return ConfirmModel{}
}

// Show displays the confirmation dialog.
func (m *ConfirmModel) Show(title, message string) {
	m.Title = title
	m.Message = message
	m.Visible = true
}

// Hide closes the confirmation dialog.
func (m *ConfirmModel) Hide() {
	m.Visible = false
}

// View renders the confirmation dialog.
func (m *ConfirmModel) View() string {
	if !m.Visible {
		return ""
	}

	content := fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		confirmTitle.Render(m.Title),
		m.Message,
		confirmHint.Render("[y] Yes  [n/esc] No"),
	)

	return confirmBox.Render(content)
}
