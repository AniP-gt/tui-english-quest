package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/game"
	"tui-english-quest/internal/ui/components"
)

var (
	statusStyle      = lipgloss.NewStyle().Padding(1, 2)
	statusTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(components.ColorPrimary)
	statusItemStyle  = lipgloss.NewStyle().PaddingLeft(2)
)

// StatusModel displays the player's current status and growth.
type StatusModel struct {
	playerStats game.Stats
}

// NewStatusModel creates a new StatusModel.
func NewStatusModel(stats game.Stats) StatusModel {
	return StatusModel{
		playerStats: stats,
	}
}

func (m StatusModel) Init() tea.Cmd {
	return nil
}

func (m StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "enter":
			return m, func() tea.Msg { return StatusToTownMsg{} }
		}
	}
	return m, nil
}

func (m StatusModel) View() string {
	s := m.playerStats
	header := components.Header(s, true, 0)

	var b strings.Builder
	b.WriteString(statusTitleStyle.Render("Player Status\n"))
	b.WriteString(lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""))

	// Build aligned key-value lines
	labelWidth := 14
	lines := ""
	lines += fmt.Sprintf("%-*s %s\n", labelWidth, "Name:", s.Name)
	lines += fmt.Sprintf("%-*s %s\n", labelWidth, "Class:", s.Class)
	lines += fmt.Sprintf("%-*s %d\n", labelWidth, "Level:", s.Level)
	lines += fmt.Sprintf("%-*s %d / %d\n", labelWidth, "Experience:", s.Exp, s.Next)
	lines += fmt.Sprintf("%-*s %d / %d\n", labelWidth, "HP:", s.HP, s.MaxHP)
	lines += fmt.Sprintf("%-*s %d\n", labelWidth, "Attack:", s.Attack)
	lines += fmt.Sprintf("%-*s %.1f\n", labelWidth, "Defense:", s.Defense)
	lines += fmt.Sprintf("%-*s %d\n", labelWidth, "Combo:", s.Combo)
	lines += fmt.Sprintf("%-*s %d days\n", labelWidth, "Streak:", s.Streak)
	lines += fmt.Sprintf("%-*s %d\n", labelWidth, "Gold:", s.Gold)

	// Badges or achievements (placeholder)
	lines += "\nAchievements:\n"
	lines += "  - First Victory\n"
	lines += "  - Combo Master\n"

	b.WriteString(lines)

	footer := components.Footer("[Enter/Esc] Back to Town", 0)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
		statusStyle.Render(b.String()),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
		footer,
	)
}
