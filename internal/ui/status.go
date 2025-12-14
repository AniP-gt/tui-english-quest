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
	statusTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
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

	b.WriteString(statusItemStyle.Render(fmt.Sprintf("Name: %s\n", s.Name)))
	b.WriteString(statusItemStyle.Render(fmt.Sprintf("Class: %s\n", s.Class)))
	b.WriteString(statusItemStyle.Render(fmt.Sprintf("Level: %d\n", s.Level)))
	b.WriteString(statusItemStyle.Render(fmt.Sprintf("Experience: %d / %d\n", s.Exp, s.Next)))
	b.WriteString(statusItemStyle.Render(fmt.Sprintf("HP: %d / %d\n", s.HP, s.MaxHP)))
	b.WriteString(statusItemStyle.Render(fmt.Sprintf("Attack: %d\n", s.Attack)))
	b.WriteString(statusItemStyle.Render(fmt.Sprintf("Defense: %.1f\n", s.Defense)))
	b.WriteString(statusItemStyle.Render(fmt.Sprintf("Combo: %d\n", s.Combo)))
	b.WriteString(statusItemStyle.Render(fmt.Sprintf("Streak: %d days\n", s.Streak)))
	b.WriteString(statusItemStyle.Render(fmt.Sprintf("Gold: %d\n", s.Gold)))

	// Badges or achievements (placeholder)
	b.WriteString(statusItemStyle.Render("\nAchievements:\n"))
	b.WriteString(statusItemStyle.Render("- First Victory\n"))
	b.WriteString(statusItemStyle.Render("- Combo Master\n"))

	footer := components.Footer("[Enter/Esc] Back to Town", 0)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
		statusStyle.Render(b.String()),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
		footer,
	)
}
