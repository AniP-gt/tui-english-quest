package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/config"
	"tui-english-quest/internal/game"
	"tui-english-quest/internal/i18n"
	"tui-english-quest/internal/ui/components"
)

var (
	settingsStyle      = lipgloss.NewStyle().Padding(1, 2)
	settingsTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(components.ColorPrimary)
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
	langPref        string   // "en"/"ja"
}

// NewSettingsModel creates a new SettingsModel.
func NewSettingsModel(stats game.Stats) SettingsModel {
	cfg, _ := config.LoadConfig()

	ti := textinput.New()
	ti.Placeholder = i18n.T("settings_api_placeholder")
	ti.CharLimit = 100
	ti.Width = 50

	// Load current API key from environment for comparison; also allow config stored key
	currentAPIKey := os.Getenv("GEMINI_API_KEY")
	if currentAPIKey == "" {
		currentAPIKey = cfg.ApiKey
	}
	ti.SetValue(currentAPIKey) // Set initial value of text input

	return SettingsModel{
		playerStats:     stats,
		apiKeyInput:     ti,
		cursor:          0,
		menu:            []string{i18n.T("settings_menu_api"), fmt.Sprintf(i18n.T("settings_menu_lang_current"), strings.ToUpper(cfg.LangPref)), i18n.T("settings_save")},
		originalAPIKey:  currentAPIKey,
		showConfirmExit: false,
		confirmCursor:   0,
		confirmMenu:     []string{i18n.T("confirm_save_opt1"), i18n.T("confirm_save_opt2"), i18n.T("confirm_save_opt3")},
		langPref:        cfg.LangPref,
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
				case i18n.T("confirm_save_opt1"):
					// Save logic: write API key and config
					apiKey := m.apiKeyInput.Value()
					cfg := config.Config{LangPref: m.langPref, ApiKey: apiKey}
					if err := config.SaveConfig(cfg); err != nil {
						// TODO: show error
					}
					// also set env for current process
					if apiKey != "" {
						_ = os.Setenv("GEMINI_API_KEY", apiKey)
					}
					// apply language immediately for current process
					i18n.SetLang(m.langPref)
					return m, func() tea.Msg { return SettingsToTownMsg{} }

				case i18n.T("confirm_save_opt2"):
					return m, func() tea.Msg { return SettingsToTownMsg{} }
				case i18n.T("confirm_save_opt3"):
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
			// otherwise return
			return m, func() tea.Msg { return SettingsToTownMsg{} }
		case "esc":
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
			switch m.cursor {
			case 0:
				// Focus API key input
				m.apiKeyInput.Focus()
			case 1:
				// Toggle language preference: en <-> ja
				if m.langPref == "en" {
					m.langPref = "ja"
				} else {
					m.langPref = "en"
				}
				// update menu label to show current
				m.menu[1] = fmt.Sprintf(i18n.T("settings_menu_lang_current"), strings.ToUpper(m.langPref))

			case 2:
				// Save and exit
				apiKey := m.apiKeyInput.Value()
				cfg := config.Config{LangPref: m.langPref, ApiKey: apiKey}
				if err := config.SaveConfig(cfg); err != nil {
					// handle error
				}
				if apiKey != "" {
					_ = os.Setenv("GEMINI_API_KEY", apiKey)
				}
				// apply language immediately
				i18n.SetLang(m.langPref)
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
	b.WriteString(settingsTitleStyle.Render(i18n.T("settings_title") + "\n"))
	b.WriteString(lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render("") + "\n")

	if m.showConfirmExit {
		b.WriteString("\n" + i18n.T("confirm_save") + "\n\n")
		for i, item := range m.confirmMenu {
			cursor := "  "
			if i == m.confirmCursor {
				cursor = "> "
			}
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, item))
		}
		footer := components.Footer(i18n.T("footer_settings_confirm"), 0)
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
			settingsStyle.Render(b.String()),
			lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
			footer,
		)
	}

	b.WriteString(i18n.T("settings_prompt") + "\n\n")

	for i, item := range m.menu {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		// Render each menu item using settingsItemStyle for consistent padding
		b.WriteString(settingsItemStyle.Render(fmt.Sprintf("%s%s", cursor, item)) + "\n")
		if i == 0 { // This is for API key input
			b.WriteString(settingsItemStyle.Render(fmt.Sprintf("%s: %s", i18n.T("api_label"), m.apiKeyInput.View())) + "\n")
		}
	}

	footer := components.Footer(i18n.T("footer_settings_main"), 0)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
		settingsStyle.Render(b.String()),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
		footer,
	)
}
