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
	playerStats     game.Stats
	apiKeyInput     textinput.Model
	cursor          int
	menu            []string
	originalAPIKey  string   // Store the API key when entering settings
	showConfirmExit bool     // Whether to show the exit confirmation
	confirmCursor   int      // Cursor for the exit confirmation menu
	confirmMenu     []string // Menu for exit confirmation
}

// NewSettingsModel creates a new SettingsModel.
func NewSettingsModel(stats game.Stats) SettingsModel {
	ti := textinput.New()
	ti.Placeholder = "Enter your Gemini API key"
	ti.CharLimit = 100
	ti.Width = 50

	// Load current API key from environment for comparison
	currentAPIKey := os.Getenv("GEMINI_API_KEY")
	ti.SetValue(currentAPIKey) // Set initial value of text input

	return SettingsModel{
		playerStats:     stats,
		apiKeyInput:     ti,
		cursor:          0,
		menu:            []string{"Set Gemini API Key", "Save and Exit"},
		originalAPIKey:  currentAPIKey,
		showConfirmExit: false,
		confirmCursor:   0,
		confirmMenu:     []string{"Save Changes", "Discard Changes", "Cancel"},
	}
}

func (m SettingsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle confirmation dialog if active
	if m.showConfirmExit {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.confirmCursor > 0 {
					m.confirmCursor--
				}
			case "down", "j":
				if m.confirmCursor < len(m.confirmMenu)-1 {
					m.confirmCursor++
				}
			case "enter":
				switch m.confirmMenu[m.confirmCursor] {
				case "Save Changes":
					// Save logic
					apiKey := m.apiKeyInput.Value()
					if apiKey != "" {
						envContent := fmt.Sprintf("DB_PATH=./db.sqlite\nGEMINI_API_KEY=%s\n", apiKey)
						if err := os.WriteFile(".env", []byte(envContent), 0644); err != nil {
							// TODO: Handle error, show message to user
						}
					}
					return m, func() tea.Msg { return SettingsToTownMsg{} }
				case "Discard Changes":
					return m, func() tea.Msg { return SettingsToTownMsg{} }
				case "Cancel":
					m.showConfirmExit = false
					m.confirmCursor = 0 // Reset cursor
					return m, nil
				}
			}
		}
		return m, nil // Consume input while confirmation is active
	}

	// Handle main settings menu
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			// If API key changed, show confirmation
			if m.apiKeyInput.Value() != m.originalAPIKey {
				m.showConfirmExit = true
				return m, nil
			}
			return m, func() tea.Msg { return SettingsToTownMsg{} }
		case "esc":
			// If API key changed, show confirmation
			if m.apiKeyInput.Value() != m.originalAPIKey {
				m.showConfirmExit = true
				return m, nil
			}
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
	header := components.Header(s, true, 0)

	var b strings.Builder
	b.WriteString(settingsTitleStyle.Render("Settings\n"))
	b.WriteString(lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""))

	if m.showConfirmExit {
		b.WriteString("\nAPI key has changed. Do you want to save?\n\n")
		for i, item := range m.confirmMenu {
			cursor := "  "
			if i == m.confirmCursor {
				cursor = "> "
			}
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, item))
		}
		footer := components.Footer("[j/k] Navigate  [Enter] Select", 0)
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
			settingsStyle.Render(b.String()),
			lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
			footer,
		)
	}

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

	footer := components.Footer("[j/k] Navigate  [Enter] Select  [Esc] Back to Town", 0)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
		settingsStyle.Render(b.String()),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
		footer,
	)
}
