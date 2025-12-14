package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/game"
)

var (
	statusBarStyle = lipgloss.NewStyle().
		Foreground(ColorMuted).
		Background(ColorBoxDark).
		Padding(0, 1).
		Height(1)
)

// View renders the status bar using HPBar from this package.
func View(s game.Stats) string {
	// Use a default HP width; the caller can wrap if needed.
	hpW := 10
	hp := HPBar(s.HP, s.MaxHP, hpW)
	status := fmt.Sprintf("LV:%d EXP:%d/%d HP:%s Gold:%d Streak:%d", s.Level, s.Exp, s.Next, hp, s.Gold, s.Streak)
	if s.Combo > 0 {
		status += fmt.Sprintf(" Combo:%d", s.Combo)
	}
	return statusBarStyle.Render(status)
}
