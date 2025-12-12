package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/game"
	"tui-english-quest/internal/ui/components"
)

var (
	settingsStyle      = lipgloss.NewStyle().Padding(1, 2)
	settingsTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	settingsItemStyle  = lipgloss.NewStyle().PaddingLeft(2)
)

// SettingsModel displays and manages application settings.
type SettingsModel struct {
	playerStats game.Stats
	apiKeyInput textinput.Model
	cursor      int
	menu        []string
}

// NewSettingsModel creates a new SettingsModel.
func NewSettingsModel(stats game.Stats) SettingsModel {
	ti := textinput.New()
	ti.Placeholder = "Enter your Gemini API key"
	ti.CharLimit = 100
	ti.Width = 50

	return SettingsModel{
		playerStats: stats,
		apiKeyInput: ti,
		cursor:      0,
		menu:        []string{"Set Gemini API Key", "Save and Exit"},
	}
}

func (m SettingsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, func() tea.Msg { return SettingsToTownMsg{} }
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.menu)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor == 0 {
				// Focus API key input
				m.apiKeyInput.Focus()
			} else if m.cursor == 1 {
				// Save and exit
				apiKey := m.apiKeyInput.Value()
				if apiKey != "" {
					// Save to .env file
					envContent := fmt.Sprintf("DB_PATH=./db.sqlite\nGEMINI_API_KEY=%s\n", apiKey)
					if err := os.WriteFile(".env", []byte(envContent), 0644); err != nil {
						// Handle error, perhaps show message
					}
				}
				return m, func() tea.Msg { return SettingsToTownMsg{} }
			}
		}
	}

	if m.apiKeyInput.Focused() {
		m.apiKeyInput, cmd = m.apiKeyInput.Update(msg)
	}

	return m, cmd
}

func (m SettingsModel) View() string {
	s := m.playerStats
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(lipgloss.Width(components.View(s))).Render("TUI English Quest"),
		components.View(s),
	)

	var b strings.Builder
	b.WriteString(settingsTitleStyle.Render("Settings\n"))
	b.WriteString(lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""))

	b.WriteString("Configure application settings:\n\n")

	for i, item := range m.menu {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, item))
		if i == 0 {
			b.WriteString(settingsItemStyle.Render(fmt.Sprintf("API Key: %s\n", m.apiKeyInput.View())))
		}
	}

	footer := "\n[j/k] Navigate  [Enter] Select  [Esc] Back to Town"

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
		settingsStyle.Render(b.String()),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
		footer,
	)
}
