package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/config"
	"tui-english-quest/internal/game"
	"tui-english-quest/internal/services"
	"tui-english-quest/internal/ui/components"
)

var (
	menuStyle = lipgloss.NewStyle()
	noteStyle = lipgloss.NewStyle().Foreground(components.ColorMuted)
)

// AppState represents the current screen/state of the application.
type AppState int

const (
	StateTop AppState = iota
	StateTown
	StateBattle    // Added StateBattle
	StateDungeon   // Added StateDungeon
	StateTavern    // Added StateTavern
	StateSpelling  // Spelling mock screen
	StateListening // Listening mock screen
	StateAnalysis  // AI Analysis screen
	StateHistory   // History screen
	StateEquipment // Equipment screen
	StateStatus    // Status screen
	StateSettings  // Settings screen
)

// Messages for screen transitions
type AnalysisToTownMsg struct{}
type TownToAnalysisMsg struct{}
type HistoryToTownMsg struct{}
type TownToHistoryMsg struct{}
type EquipmentToTownMsg struct{} // Fixed syntax error
type TownToEquipmentMsg struct{}
type StatusToTownMsg struct{}
type TownToStatusMsg struct{}
type SettingsToTownMsg struct{}
type TownToSettingsMsg struct{}

type TownToTavernMsg struct{}
type TavernToTownMsg struct{}

type TownToSpellingMsg struct{}
type SpellingToTownMsg struct{}

type TownToListeningMsg struct{}
type ListeningToTownMsg struct{}

// RootModel is the top-level model that manages different application states.
type RootModel struct {
	Status       game.Stats
	menu         []string
	cursor       int
	note         string
	state        AppState
	town         TownModel
	battle       BattleModel    // Added BattleModel
	dungeon      DungeonModel   // Added DungeonModel
	tavern       TavernModel    // Added TavernModel
	spelling     SpellingModel  // SpellingModel (mock)
	listening    ListeningModel // ListeningModel (mock)
	analysis     AnalysisModel  // Embed AnalysisModel
	history      HistoryModel
	equipment    EquipmentModel
	status       StatusModel
	settings     SettingsModel
	geminiClient *services.GeminiClient // Add GeminiClient
	LangPref     string
}

// NewRootModel creates the top-level model.
func NewRootModel() RootModel {
	stats := game.DefaultStats()
	cfg, _ := config.LoadConfig()

	// Initialize Gemini Client
	gc, err := services.NewGeminiClient(context.Background())
	if err != nil {
		// Handle error, e.g., log it and proceed without Gemini features or exit
		fmt.Printf("Failed to initialize Gemini Client: %v\n", err)
		// For now, we'll proceed with a nil client, but a real app might exit or disable features.
	}

	return RootModel{
		Status: stats,
		menu: []string{
			"冒険を始める",
			"新しいゲーム",
			"終了",
		},
		cursor: 0,
		note:   "Nで新しいゲームを開始",

		state:        StateTop,
		town:         NewTownModel(stats, gc),    // Pass GeminiClient
		battle:       NewBattleModel(stats, gc),  // Initialize BattleModel
		dungeon:      NewDungeonModel(stats, gc), // Initialize DungeonModel
		tavern:       NewTavernModel(stats, gc, cfg.LangPref),
		spelling:     NewSpellingModel(stats, gc),
		listening:    NewListeningModel(stats, gc),
		analysis:     NewAnalysisModel(stats, gc), // Pass GeminiClient
		history:      NewHistoryModel(stats),
		equipment:    NewEquipmentModel(stats),
		status:       NewStatusModel(stats),
		settings:     NewSettingsModel(stats),
		geminiClient: gc,
		LangPref:     cfg.LangPref,
	}
}

func (m RootModel) Init() tea.Cmd { return nil }

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case TownToRootMsg: // Handle message from TownModel to return to Top
		m.state = StateTop
		m.Status = m.town.playerStats // Update RootModel's stats from TownModel
		return m, nil
	case AnalysisToTownMsg: // Handle message from AnalysisModel to return to Town
		m.state = StateTown
		m.Status = m.analysis.playerStats               // Update RootModel's stats from AnalysisModel
		m.town = NewTownModel(m.Status, m.geminiClient) // Refresh TownModel with updated stats
		return m, nil
	case TownToHistoryMsg:
		m.state = StateHistory
		m.history = NewHistoryModel(m.Status)
		return m, nil
	case HistoryToTownMsg:
		m.state = StateTown
		return m, nil
	case TownToEquipmentMsg:
		m.state = StateEquipment
		m.equipment = NewEquipmentModel(m.Status)
		return m, nil
	case EquipmentToTownMsg:
		m.state = StateTown
		return m, nil
	case TownToStatusMsg:
		m.state = StateStatus
		m.status = NewStatusModel(m.Status)
		return m, nil
	case StatusToTownMsg:
		m.state = StateTown
		return m, nil
	case TownToSettingsMsg:
		m.state = StateSettings
		m.settings = NewSettingsModel(m.Status)
		return m, nil
	case SettingsToTownMsg:
		m.state = StateTown
		// reload config in case language or api key changed
		if cfg, err := config.LoadConfig(); err == nil {
			m.LangPref = cfg.LangPref
		}
		// reinitialize models that depend on language pref
		m.tavern = NewTavernModel(m.Status, m.geminiClient, m.LangPref)
		return m, nil
	case TownToBattleMsg: // Added TownToBattleMsg handling
		m.state = StateBattle
		m.battle = NewBattleModel(m.Status, m.geminiClient) // Initialize BattleModel
		return m, m.battle.Init()
	case TownToDungeonMsg: // Added TownToDungeonMsg handling
		m.state = StateDungeon
		m.dungeon = NewDungeonModel(m.Status, m.geminiClient) // Initialize DungeonModel
		return m, m.dungeon.Init()
	case TownToTavernMsg:
		m.state = StateTavern
		m.tavern = NewTavernModel(m.Status, m.geminiClient, m.LangPref)
		return m, m.tavern.Init()
	case TownToSpellingMsg:
		m.state = StateSpelling
		m.spelling = NewSpellingModel(m.Status, m.geminiClient)
		return m, m.spelling.Init()
	case TownToListeningMsg:
		m.state = StateListening
		m.listening = NewListeningModel(m.Status, m.geminiClient)
		return m, m.listening.Init()
	}

	switch m.state {
	case StateTop:
		return m.updateTop(msg)
	case StateTown:
		newTownModel, cmd := m.town.Update(msg)
		m.town = newTownModel.(TownModel)
		m.Status = m.town.playerStats
		return m, cmd
	case StateBattle: // Added StateBattle update
		newBattleModel, cmd := m.battle.Update(msg)
		m.battle = newBattleModel.(BattleModel)
		m.Status = m.battle.playerStats
		return m, cmd
	case StateDungeon: // Added StateDungeon update
		newDungeonModel, cmd := m.dungeon.Update(msg)
		m.dungeon = newDungeonModel.(DungeonModel)
		m.Status = m.dungeon.playerStats
		return m, cmd
	case StateTavern:
		newTavernModel, cmd := m.tavern.Update(msg)
		m.tavern = newTavernModel.(TavernModel)
		m.Status = m.tavern.playerStats
		return m, cmd
	case StateSpelling:
		newSpellingModel, cmd := m.spelling.Update(msg)
		m.spelling = newSpellingModel.(SpellingModel)
		m.Status = m.spelling.playerStats
		return m, cmd
	case StateListening:
		newListeningModel, cmd := m.listening.Update(msg)
		m.listening = newListeningModel.(ListeningModel)
		m.Status = m.listening.playerStats
		return m, cmd
	case StateAnalysis:
		newAnalysisModel, cmd := m.analysis.Update(msg)
		m.analysis = newAnalysisModel.(AnalysisModel)
		m.Status = m.analysis.playerStats
		return m, cmd
	case StateHistory:
		newHistoryModel, cmd := m.history.Update(msg)
		m.history = newHistoryModel.(HistoryModel)
		m.Status = m.history.playerStats
		return m, cmd
	case StateEquipment:
		newEquipmentModel, cmd := m.equipment.Update(msg)
		m.equipment = newEquipmentModel.(EquipmentModel)
		m.Status = m.equipment.playerStats
		return m, cmd
	case StateStatus:
		newStatusModel, cmd := m.status.Update(msg)
		m.status = newStatusModel.(StatusModel)
		m.Status = m.status.playerStats
		return m, cmd
	case StateSettings:
		newSettingsModel, cmd := m.settings.Update(msg)
		m.settings = newSettingsModel.(SettingsModel)
		m.Status = m.settings.playerStats
		return m, cmd
	default:
		return m, nil
	}
}

func (m RootModel) updateTop(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q": // T034: 途中離脱
			if m.Status.HP > 0 {
				m.note = "Session interrupted. Progress not saved."
				return m, tea.Quit
			}
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.menu)-1 {
				m.cursor++
			}
		case "enter":
			return m.handleTopEnter()
		case "n":
			m.Status = game.DefaultStats()
			m.note = "New Game started with defaults"
			m.town = NewTownModel(m.Status, m.geminiClient)
			m.state = StateTown
		}
	}
	return m, nil
}

func (m RootModel) handleTopEnter() (tea.Model, tea.Cmd) {
	choice := m.menu[m.cursor]
	switch choice {
	case "Start Adventure":
		m.state = StateTown
		m.town = NewTownModel(m.Status, m.geminiClient)
		return m, nil
	case "New Game":
		m.Status = game.DefaultStats()
		m.note = "New Game started with defaults"
		m.town = NewTownModel(m.Status, m.geminiClient)
		m.state = StateTown
		return m, nil
	case "Quit":
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m RootModel) View() string {
	switch m.state {
	case StateTop:
		return m.viewTop()
	case StateTown:
		return m.viewTown()
	case StateBattle: // Added StateBattle view
		return m.battle.View()
	case StateDungeon: // Added StateDungeon view
		return m.dungeon.View()
	case StateTavern:
		return m.tavern.View()
	case StateSpelling:
		return m.spelling.View()
	case StateListening:
		return m.listening.View()
	case StateAnalysis:
		return m.viewAnalysis()
	case StateHistory:
		return m.viewHistory()
	case StateEquipment:
		return m.viewEquipment()
	case StateStatus:
		return m.viewStatus()
	case StateSettings:
		return m.viewSettings()
	default:
		return "Unknown state"
	}
}

func (m RootModel) viewTop() string {
	// Use shared header and footer components
	header := components.Header(m.Status, true, 0)
	body := ""
	for i, item := range m.menu {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		body += fmt.Sprintf("%s%s\n", cursor, item)
	}
	// Render menu inside a centered box for the Top screen (no blue tone)
	boxWidth := lipgloss.Width(header)
	if boxWidth < 50 {
		boxWidth = 50
	}
	innerWidth := boxWidth - 4 // account for box padding/border

	// Large centered title (use Accent color, not blue)
	title := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Foreground(components.ColorAccent).Bold(true).Render("TUI English Quest")

	// Build centered menu items, highlight selected
	var menuLines []string
	for i, item := range m.menu {
		if i == m.cursor {
			sel := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Background(components.ColorAccent).Foreground(components.ColorBoxDark).Bold(true).Render(item)
			menuLines = append(menuLines, sel)
		} else {
			line := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Foreground(components.ColorMuted).Render(item)
			menuLines = append(menuLines, line)
		}
	}

	content := title + "\n\n" + lipgloss.JoinVertical(lipgloss.Center, menuLines...)
	// Use a dark box (no info/cyan background)
	menuBox := components.Box("", content, "", boxWidth)

	footer := components.Footer("[j/k] 移動  [Enter] 選択  [n] 新しいゲーム  [q] 終了", 0)
	if m.note != "" {
		footer = footer + "\n" + noteStyle.Render(m.note)
	}
	return fmt.Sprintf("%s\n%s\n\n%s\n", header, menuBox, footer)
}

func (m RootModel) viewTown() string {
	return m.town.View()
}

func (m RootModel) viewAnalysis() string {
	return m.analysis.View()
}

func (m RootModel) viewHistory() string {
	return m.history.View()
}

func (m RootModel) viewEquipment() string {
	return m.equipment.View()
}

func (m RootModel) viewStatus() string {
	return m.status.View()
}

func (m RootModel) viewSettings() string {
	return m.settings.View()
}

// runAllModes simulates running all modes sequentially using sample payloads and default answers.
// This function is now only for testing purposes and will be removed later.
func runAllModes() (game.Stats, []game.SessionSummary) { // Changed return type to game.SessionSummary
	ctx := context.Background()
	stats := game.DefaultStats()
	summaries := []game.SessionSummary{} // Changed to game.SessionSummary

	if payload, err := services.FetchAndValidate(ctx, services.ModeVocab); err == nil {
		_ = payload
		ans := make([]game.VocabAnswer, 5) // Changed to game.VocabAnswer
		for i := range ans {
			ans[i] = game.VocabAnswer{Correct: true} // Changed to game.VocabAnswer
		}
		var sum game.SessionSummary                           // Changed to game.SessionSummary
		stats, sum, _ = game.RunVocabSession(ctx, stats, ans) // Changed to game.RunVocabSession
		summaries = append(summaries, sum)
	} else {
		summaries = append(summaries, game.SessionSummary{Mode: services.ModeVocab, Note: err.Error()}) // Changed to game.SessionSummary
	}

	if payload, err := services.FetchAndValidate(ctx, services.ModeGrammar); err == nil {
		_ = payload
		ans := make([]game.GrammarAnswer, 5) // Changed to game.GrammarAnswer
		for i := range ans {
			ans[i] = game.GrammarAnswer{Correct: true} // Changed to game.GrammarAnswer
		}
		var sum game.SessionSummary                             // Changed to game.SessionSummary
		stats, sum, _ = game.RunGrammarSession(ctx, stats, ans) // Changed to game.RunGrammarSession
		summaries = append(summaries, sum)
	} else {
		summaries = append(summaries, game.SessionSummary{Mode: services.ModeGrammar, Note: err.Error()}) // Changed to game.SessionSummary
	}

	if payload, err := services.FetchAndValidate(ctx, services.ModeTavern); err == nil {
		_ = payload
		outs := []game.TavernOutcome{game.OutcomeSuccess, game.OutcomeNormal, game.OutcomeSuccess, game.OutcomeFail, game.OutcomeNormal} // Changed to game.TavernOutcome
		var sum game.SessionSummary                                                                                                      // Changed to game.SessionSummary
		stats, sum, _ = game.RunTavernSession(ctx, stats, outs)                                                                          // Changed to game.RunTavernSession
		summaries = append(summaries, sum)
	} else {
		summaries = append(summaries, game.SessionSummary{Mode: services.ModeTavern, Note: err.Error()}) // Changed to game.SessionSummary
	}

	if payload, err := services.FetchAndValidate(ctx, services.ModeSpelling); err == nil {
		_ = payload
		outs := []game.SpellingOutcome{game.SpellingPerfect, game.SpellingNear, game.SpellingPerfect, game.SpellingFail, game.SpellingNear} // Changed to game.SpellingOutcome
		var sum game.SessionSummary                                                                                                         // Changed to game.SessionSummary
		stats, sum, _ = game.RunSpellingSession(ctx, stats, outs)                                                                           // Changed to game.RunSpellingSession
		summaries = append(summaries, sum)
	} else {
		summaries = append(summaries, game.SessionSummary{Mode: services.ModeSpelling, Note: err.Error()}) // Changed to game.SessionSummary
	}

	if payload, err := services.FetchAndValidate(ctx, services.ModeListening); err == nil {
		_ = payload
		ans := []game.ListeningAnswer{{true}, {true}, {false}, {true}, {true}} // Changed to game.ListeningAnswer
		var sum game.SessionSummary                                            // Changed to game.SessionSummary
		stats, sum, _ = game.RunListeningSession(ctx, stats, ans)              // Changed to game.RunListeningSession
		summaries = append(summaries, sum)
	} else {
		summaries = append(summaries, game.SessionSummary{Mode: services.ModeListening, Note: err.Error()}) // Changed to game.SessionSummary
	}

	return stats, summaries
}

// formatSummaries returns a compact note string.
func formatSummaries(summaries []game.SessionSummary) string { // Changed to game.SessionSummary
	parts := make([]string, 0, len(summaries))
	for _, s := range summaries {
		if s.Note != "" {
			parts = append(parts, fmt.Sprintf("%s: error %s", s.Mode, s.Note))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s ok (EXP %+d, HP %+d, Gold %+d)", s.Mode, s.ExpDelta, s.HPDelta, s.GoldDelta))
	}
	return noteStyle.Render(strings.Join(parts, " | "))
}
