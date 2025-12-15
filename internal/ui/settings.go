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
	langPref        string   // "en"/"ja"/"both"
}

// NewSettingsModel creates a new SettingsModel.
func NewSettingsModel(stats game.Stats) SettingsModel {
	cfg, _ := config.LoadConfig()

	ti := textinput.New()
	ti.Placeholder = "ジェミニAPIキーを入力"
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
		menu:            []string{"ジェミニAPIキー設定", fmt.Sprintf("言語設定 (現在: %s)", strings.ToUpper(cfg.LangPref)), "保存して終了"},
		originalAPIKey:  currentAPIKey,
		showConfirmExit: false,
		confirmCursor:   0,
		confirmMenu:     []string{"変更を保存", "変更を破棄", "キャンセル"},
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
				case "変更を保存":
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
					return m, func() tea.Msg { return SettingsToTownMsg{} }
				case "変更を破棄":
					return m, func() tea.Msg { return SettingsToTownMsg{} }
				case "キャンセル":
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
				// Cycle language preference: en -> ja -> both -> en
				switch m.langPref {
				case "en":
					m.langPref = "ja"
				case "ja":
					m.langPref = "both"
				default:
					m.langPref = "en"
				}
				// update menu label to show current
				m.menu[1] = fmt.Sprintf("言語設定 (現在: %s)", strings.ToUpper(m.langPref))
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
	b.WriteString(settingsTitleStyle.Render("設定\n"))
	b.WriteString(lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""))

	if m.showConfirmExit {
		b.WriteString("\nAPIキーが変更されました。保存しますか?\n\n")
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

	b.WriteString("アプリケーション設定:\n\n")

	for i, item := range m.menu {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, item))
		if i == 0 {
			b.WriteString(settingsItemStyle.Render(fmt.Sprintf("APIキー: %s\n", m.apiKeyInput.View())))
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
