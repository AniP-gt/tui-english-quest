package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HPBar returns a string representation of an HP bar with the given width.
func HPBar(current, max, width int) string {
	if width <= 0 {
		width = 10
	}
	pct := 0.0
	if max > 0 {
		pct = float64(current) / float64(max)
	}
	filled := int(pct * float64(width))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	// Choose color by threshold
	var style lipgloss.Style
	if pct <= 0.25 {
		style = lipgloss.NewStyle().Foreground(ColorDanger)
	} else if pct <= 0.5 {
		style = lipgloss.NewStyle().Foreground(ColorOrange)
	} else {
		style = lipgloss.NewStyle().Foreground(ColorPrimary)
	}
	return style.Render(bar)
}

// Small helper to format current/max as text
func HPText(current, max int) string {
	return fmt.Sprintf("%d/%d", current, max)
}
