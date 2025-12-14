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
	outer := lipgloss.NewStyle().Background(bg).Border(lipgloss.RoundedBorder()).Padding(1, 1).Width(width)
	titleStyled := title
	contentStyled := content
	// Render content with muted foreground on the box background.
	if title != "" {
		titleStyled = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true).Render(title)
	}
	if content != "" {
		// content may already include centered/colored lines, so avoid double-styling; only apply muted if plain
		contentStyled = content
	}
	if title != "" {
		return outer.Render(titleStyled + "\n" + contentStyled)
	}
	return outer.Render(contentStyled)
}
