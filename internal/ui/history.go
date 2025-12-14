package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/db"
	"tui-english-quest/internal/game"
	"tui-english-quest/internal/ui/components"
)

var (
	historyStyle       = lipgloss.NewStyle().Padding(1, 2)
	historyTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(components.ColorPrimary)
	historyItemStyle   = lipgloss.NewStyle().PaddingLeft(2)
	historyHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(components.ColorMuted)
)

// HistoryModel displays the player's session history.
type HistoryModel struct {
	playerStats game.Stats
	sessions    []db.SessionRecord
	cursor      int
}

// NewHistoryModel creates a new HistoryModel.
func NewHistoryModel(stats game.Stats) HistoryModel {
	// Fetch recent sessions (limit 20 for display)
	ctx := context.Background()
	sessions, err := db.ListSessions(ctx, stats.Name, 20) // Using name as playerID
	if err != nil {
		// Handle error, perhaps show empty list
		sessions = []db.SessionRecord{}
	}
	return HistoryModel{
		playerStats: stats,
		sessions:    sessions,
		cursor:      0,
	}
}

func (m HistoryModel) Init() tea.Cmd {
	return nil
}

func (m HistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "enter":
			return m, func() tea.Msg { return HistoryToTownMsg{} }
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.sessions)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m HistoryModel) View() string {
	s := m.playerStats
	header := components.Header(s, true, 0)

	var b strings.Builder
	b.WriteString(historyTitleStyle.Render("Session History\n"))
	b.WriteString(lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""))

	if len(m.sessions) == 0 {
		b.WriteString(historyItemStyle.Render("No sessions found.\n"))
	} else {
		// Header
		b.WriteString(historyHeaderStyle.Render(fmt.Sprintf("%-12s %-10s %-8s %-6s %-6s %-8s\n", "Date", "Mode", "Score", "EXP", "Gold", "HP Δ")))
		b.WriteString(strings.Repeat("-", 60) + "\n")

		// Sessions
		for i, session := range m.sessions {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}
			date := session.EndedAt.Format("01/02 15:04")
			mode := session.Mode
			score := fmt.Sprintf("%d/5", session.CorrectCount)
			exp := fmt.Sprintf("%+d", session.ExpGained)
			gold := fmt.Sprintf("%+d", session.GoldDelta)
			hp := fmt.Sprintf("%+d", session.HPDelta)
			if session.Fainted {
				hp += " 💀"
			}
			if session.LeveledUp {
				exp += " ↑"
			}
			line := fmt.Sprintf("%s%-12s %-10s %-8s %-6s %-6s %-8s\n", cursor, date, mode, score, exp, gold, hp)
			b.WriteString(line)
		}
	}

	footer := components.Footer("[j/k] Navigate  [Enter/Esc] Back to Town", 0)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
		historyStyle.Render(b.String()),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
		footer,
	)
}
