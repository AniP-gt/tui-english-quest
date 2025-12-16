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
	b.WriteString(lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render("") + "\n")

	// Build aligned key-value lines
	labelWidth := 14
	lines := ""
	lines += components.RenderKeyValue("Name:", s.Name, labelWidth) + "\n"
	lines += components.RenderKeyValue("Class:", s.Class, labelWidth) + "\n"
	lines += components.RenderKeyValue("Level:", fmt.Sprintf("%d", s.Level), labelWidth) + "\n"
	lines += components.RenderKeyValue("Experience:", fmt.Sprintf("%d / %d", s.Exp, s.Next), labelWidth) + "\n"
	lines += components.RenderKeyValue("HP:", fmt.Sprintf("%d / %d", s.HP, s.MaxHP), labelWidth) + "\n"
	lines += components.RenderKeyValue("Attack:", fmt.Sprintf("%d", s.Attack), labelWidth) + "\n"
	lines += components.RenderKeyValue("Defense:", fmt.Sprintf("%.1f", s.Defense), labelWidth) + "\n"
	lines += components.RenderKeyValue("Combo:", fmt.Sprintf("%d", s.Combo), labelWidth) + "\n"
	lines += components.RenderKeyValue("Streak:", fmt.Sprintf("%d days", s.Streak), labelWidth) + "\n"
	lines += components.RenderKeyValue("Gold:", fmt.Sprintf("%d", s.Gold), labelWidth) + "\n"

	// Badges or achievements (placeholder)
	lines += "\nAchievements:\n\n"
	achievements := []string{"First Victory", "Combo Master"}
	lines += components.RenderBulletList(achievements, 2)

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
