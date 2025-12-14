package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/game"
	"tui-english-quest/internal/services"
	"tui-english-quest/internal/ui/components"
)

var (
	spellingStyle      = lipgloss.NewStyle().Padding(1, 2)
	spellingTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(components.ColorPrimary)
)

// SpellingModel is a simple mock TUI for the Spelling Challenge.
type SpellingModel struct {
	playerStats  game.Stats
	geminiClient *services.GeminiClient
}

// NewSpellingModel creates a new SpellingModel.
func NewSpellingModel(stats game.Stats, gc *services.GeminiClient) SpellingModel {
	return SpellingModel{
		playerStats:  stats,
		geminiClient: gc,
	}
}

func (m SpellingModel) Init() tea.Cmd { return nil }

func (m SpellingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return SpellingToTownMsg{} }
		case "enter":
			// Run a mock session quickly and return to town with a note.
			// In a real implementation, you'd call RunSpellingSession and update stats.
			return m, func() tea.Msg { return SpellingToTownMsg{} }
		}
	}
	return m, nil
}

func (m SpellingModel) View() string {
	header := components.Header(m.playerStats, true, 0)

	var b strings.Builder
	b.WriteString(spellingTitleStyle.Render("Spelling Challenge (mock)\n"))
	b.WriteString("\nThis is a mock screen for the Spelling Challenge.\n")
	b.WriteString("Press [Enter] to run a mock session or [Esc] to return to Town.\n")

	footer := components.Footer("[Enter] Run mock  [Esc/q] Back to Town", 0)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
		spellingStyle.Render(b.String()),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
		footer,
	)
}
