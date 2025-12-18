package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/game"
	"tui-english-quest/internal/i18n"
	"tui-english-quest/internal/services"
	"tui-english-quest/internal/ui/components"
)

var (
	townMenuStyle     = lipgloss.NewStyle().Padding(1, 0)
	townItemStyle     = lipgloss.NewStyle().PaddingLeft(2)
	townCursorStyle   = lipgloss.NewStyle().Foreground(components.ColorAccent) // Pink
	townAdviceStyle   = lipgloss.NewStyle().Foreground(components.ColorMuted).Italic(true)
	menuItemStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Width(25).Align(lipgloss.Center)
	menuSelectedStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).Width(25).Align(lipgloss.Center).Background(components.ColorPrimary).Foreground(components.ColorBoxDark)
)

func townMenuLabels() []string {
	return []string{
		i18n.MenuLabel("town_menu_vocab_battle"),
		i18n.MenuLabel("town_menu_grammar_dungeon"),
		i18n.MenuLabel("town_menu_conversation_tavern"),
		i18n.MenuLabel("town_menu_spelling_challenge"),
		i18n.MenuLabel("town_menu_listening_cave"),
		i18n.MenuLabel("town_menu_equipment"),
		i18n.MenuLabel("town_menu_ai_analysis"),
		i18n.MenuLabel("town_menu_history"),
		i18n.MenuLabel("town_menu_status"),
		i18n.MenuLabel("town_menu_settings"),
	}
}

// TownModel handles the town/home screen.
type TownModel struct {
	playerStats game.Stats
	menuKeys    []string
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
			Recommendation: fmt.Sprintf(i18n.T("error_ai_advice"), err),
		}
	}
	return TownModel{
		playerStats: stats,
		menuKeys: []string{
			"town_menu_vocab_battle",
			"town_menu_grammar_dungeon",
			"town_menu_conversation_tavern",
			"town_menu_spelling_challenge",
			"town_menu_listening_cave",
			"town_menu_ai_analysis",
			"town_menu_history",
			"town_menu_status",
			"town_menu_settings",
		},
		cursor:   0,
		aiAdvice: aiReport,
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
			if m.cursor < len(m.menuKeys)-1 {
				m.cursor++
			}
		case "enter":
			switch m.cursor {
			case 0:
				return m, func() tea.Msg { return TownToBattleMsg{} }
			case 1:
				return m, func() tea.Msg { return TownToDungeonMsg{} }
			case 2:
				return m, func() tea.Msg { return TownToTavernMsg{} }
			case 3:
				return m, func() tea.Msg { return TownToSpellingMsg{} }
			case 4:
				return m, func() tea.Msg { return TownToListeningMsg{} }
			case 5:
				return m, func() tea.Msg { return TownToAnalysisMsg{} }
			case 6:
				return m, func() tea.Msg { return TownToHistoryMsg{} }
			case 7:
				return m, func() tea.Msg { return TownToStatusMsg{} }
			case 8:
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
	menuBody := i18n.T("town_menu_prompt") + "\n\n"
	// Build labels from keys each render
	labels := make([]string, len(m.menuKeys))
	for i, k := range m.menuKeys {
		labels[i] = i18n.T(k)
	}
	menuBody += components.Menu(labels, m.cursor, 2, 0)

	advice := fmt.Sprintf(i18n.T("town_ai_advice_format"), strings.Join(m.aiAdvice.WeakPoints, ", "), m.aiAdvice.Recommendation)

	footer := components.Footer(i18n.T("footer_town"), 0)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""), // Separator
		townMenuStyle.Render(menuBody),
		lipgloss.NewStyle().Foreground(components.ColorMuted).Italic(true).Render(advice),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""), // Separator
		footer,
	)
}
