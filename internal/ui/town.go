package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/game"
	"tui-english-quest/internal/services"
	"tui-english-quest/internal/ui/components"
)

var (
	townMenu = []string{
		"⚔  Vocabulary Battle",
		"🏰 Grammar Dungeon",
		"🍺 Conversation Tavern",
		"🪄 Spelling Challenge",
		"🔊 Listening Cave",
		"🎒 Equipment",
		"🧠 AI Analysis",
		"📖 History",
		"🎒 Status",
		"⚙  Settings",
	}

	townMenuStyle     = lipgloss.NewStyle().Padding(1, 0)
	townItemStyle     = lipgloss.NewStyle().PaddingLeft(2)
	townCursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")) // Pink
	townAdviceStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	menuItemStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Width(25).Align(lipgloss.Center)
	menuSelectedStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Width(25).Align(lipgloss.Center).Background(lipgloss.Color("62")).Foreground(lipgloss.Color("0"))
)

// TownModel handles the town/home screen.
type TownModel struct {
	playerStats game.Stats
	menu        []string
	cursor      int
	aiAdvice    services.WeaknessReport // Placeholder for AI advice
}

// NewTownModel creates a new TownModel.
func NewTownModel(stats game.Stats, gc *services.GeminiClient) TownModel {
	// TODO: Fetch actual AI advice based on player history
	// For now, use a placeholder report.
	// In a real implementation, playerID would be passed and history fetched.
	aiReport, err := services.AnalyzeWeakness(context.Background(), gc, stats.Name, 200) // Use player name as ID, limit 200
	if err != nil {
		aiReport = services.WeaknessReport{
			Recommendation: fmt.Sprintf("Error getting AI advice: %v", err),
		}
	}
	return TownModel{
		playerStats: stats,
		menu:        townMenu,
		cursor:      0,
		aiAdvice:    aiReport,
	}
}

// TownToRootMsg signals to the RootModel to return to the top screen.
type TownToRootMsg struct{}

// TownToBattleMsg signals to the RootModel to transition to the battle screen.

// TownToDungeonMsg signals to the RootModel to transition to the dungeon screen.

func (m TownModel) Init() tea.Cmd {
	return nil
}

func (m TownModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return TownToRootMsg{} } // Signal to return to RootModel
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.menu)-1 {
				m.cursor++
			}
		case "enter":
			switch m.menu[m.cursor] {
			case "⚔  Vocabulary Battle": // Handle Vocabulary Battle selection
				return m, func() tea.Msg { return TownToBattleMsg{} }
			case "🏰 Grammar Dungeon": // Handle Grammar Dungeon selection
				return m, func() tea.Msg { return TownToDungeonMsg{} }
			case "🎒 Equipment":
				return m, func() tea.Msg { return TownToEquipmentMsg{} }
			case "🧠 AI Analysis":
				return m, func() tea.Msg { return TownToAnalysisMsg{} }
			case "📖 History":
				return m, func() tea.Msg { return TownToHistoryMsg{} }
			case "🎒 Status":
				return m, func() tea.Msg { return TownToStatusMsg{} }
			case "⚙  Settings":
				return m, func() tea.Msg { return TownToSettingsMsg{} }
			default:
				return m, func() tea.Msg { return TownToRootMsg{} }
			}
		}
	}
	return m, nil
}

func (m TownModel) View() string {
	s := m.playerStats
	// Use shared header
	header := components.Header(s, true, 0)

	// Render menu using shared Menu component
	menuBody := "Where do you want to go?\n\n"
	menuBody += components.Menu(m.menu, m.cursor, 2, 0)

	advice := fmt.Sprintf("\nTip / AI Advice\n  Weak points: %s\n  Recommendation: %s",
		strings.Join(m.aiAdvice.WeakPoints, ", "), m.aiAdvice.Recommendation)

	footer := components.Footer("[j/k] Move  [Enter] Select  [q] Quit", 0)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""), // Separator
		townMenuStyle.Render(menuBody),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true).Render(advice),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""), // Separator
		footer,
	)
}
