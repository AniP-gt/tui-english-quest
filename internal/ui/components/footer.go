package components

import (
	"github.com/charmbracelet/lipgloss"
)

// Footer renders control hints.
func Footer(controls string, width int) string {
	style := lipgloss.NewStyle().Foreground(ColorMuted).Padding(0, 1)
	return style.Render(controls)
}
