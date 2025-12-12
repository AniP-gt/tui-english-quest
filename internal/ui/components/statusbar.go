package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/game"
)

var (
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Background(lipgloss.Color("235")).
			Padding(0, 1).
			Height(1)

	hpBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")). // Red
			Background(lipgloss.Color("237")).
			Width(10) // Fixed width for HP bar
)

// View renders the status bar.
func View(s game.Stats) string {
	hpRatio := float64(s.HP) / float64(s.MaxHP)
	hpBarFilled := int(hpRatio * float64(hpBarStyle.GetWidth()))
	hpBar := strings.Repeat("█", hpBarFilled) + strings.Repeat("░", hpBarStyle.GetWidth()-hpBarFilled)

	status := fmt.Sprintf("LV:%d EXP:%d/%d HP:%s Gold:%d Streak:%d",
		s.Level, s.Exp, s.Next, hpBarStyle.Render(hpBar), s.Gold, s.Streak)

	// Add Combo if it's not zero
	if s.Combo > 0 {
		status += fmt.Sprintf(" Combo:%d", s.Combo)
	}

	return statusBarStyle.Render(status)
}
