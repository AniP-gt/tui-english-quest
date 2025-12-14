package components

import (
	"github.com/charmbracelet/lipgloss"
)

// Box renders a titled box with a background tone.
func Box(title, content, tone string, width int) string {
	bg := ColorBoxDark
	switch tone {
	case "info":
		bg = ColorInfo
	case "warn":
		bg = ColorAccent
	case "danger":
		bg = ColorDanger
	}
	style := lipgloss.NewStyle().Background(bg).Border(lipgloss.RoundedBorder()).Padding(1, 1).Width(width)
	if title != "" {
		return style.Render(title + "\n" + content)
	}
	return style.Render(content)
}
