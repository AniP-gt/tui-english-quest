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
	listeningStyle      = lipgloss.NewStyle().Padding(1, 2)
	listeningTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(components.ColorPrimary)
)

// ListeningModel is a simple mock TUI for the Listening Cave.
type ListeningModel struct {
	playerStats  game.Stats
	geminiClient *services.GeminiClient
}

// NewListeningModel creates a new ListeningModel.
func NewListeningModel(stats game.Stats, gc *services.GeminiClient) ListeningModel {
	return ListeningModel{
		playerStats:  stats,
		geminiClient: gc,
	}
}

func (m ListeningModel) Init() tea.Cmd { return nil }

func (m ListeningModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return ListeningToTownMsg{} }
		case "enter":
			// Run a mock session quickly and return to town with a note.
			return m, func() tea.Msg { return ListeningToTownMsg{} }
		}
	}
	return m, nil
}

func (m ListeningModel) View() string {
	header := components.Header(m.playerStats, true, 0)

	var b strings.Builder
	b.WriteString(listeningTitleStyle.Render("Listening Cave (mock)\n"))
	b.WriteString("\nThis is a mock screen for the Listening Cave.\n")
	b.WriteString("Press [Enter] to run a mock session or [Esc] to return to Town.\n")

	footer := components.Footer("[Enter] Run mock  [Esc/q] Back to Town", 0)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
		listeningStyle.Render(b.String()),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
		footer,
	)
}
