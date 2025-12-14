package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/game"
)

// Header renders the application header with optional status line.
func Header(s game.Stats, showStatus bool, width int) string {
	left := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("TUI English Quest")
	var right string
	if showStatus {
		// Use existing status view in this package
		right = View(s)
	} else {
		right = ""
	}
	// Join with spacing
	head := lipgloss.JoinHorizontal(lipgloss.Top, left, lipgloss.NewStyle().PaddingLeft(1).Render(right))
	// Add separator line under header
	sep := lipgloss.NewStyle().Width(lipgloss.Width(head)).Border(lipgloss.NormalBorder()).Render("")
	return fmt.Sprintf("%s\n%s", head, sep)
}
