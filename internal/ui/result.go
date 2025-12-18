package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/game"
	"tui-english-quest/internal/i18n"
	"tui-english-quest/internal/ui/components"
)

var (
	resultBoxStyle = lipgloss.NewStyle().Padding(1, 2)
)

// ResultModel shows the session summary before returning to town.
type ResultModel struct {
	stats   game.Stats
	summary game.SessionSummary
}

// NewResultModel builds a ResultModel for the given stats and summary.
func NewResultModel(stats game.Stats, summary game.SessionSummary) ResultModel {
	return ResultModel{stats: stats, summary: summary}
}

// Init satisfies the tea.Model interface.
func (m ResultModel) Init() tea.Cmd {
	return nil
}

func (m ResultModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc", "q":
			return m, func() tea.Msg { return ResultToTownMsg{} }
		}
	}
	return m, nil
}

func (m ResultModel) View() string {
	displayStats := m.stats
	header := components.Header(displayStats, true, 0)

	titleKey := fmt.Sprintf("result_title_%s", m.summary.Mode)
	title := fmt.Sprintf("%s %s", i18n.T("result_title"), i18n.T(titleKey))

	lines := []string{
		title,
		fmt.Sprintf(i18n.T("result_exp_gain"), m.summary.ExpDelta),
		fmt.Sprintf(i18n.T("result_correct"), m.summary.Correct),
	}

	if m.summary.HPDelta != 0 {
		lines = append(lines, fmt.Sprintf(i18n.T("result_hp_delta"), m.summary.HPDelta))
	}
	if m.summary.GoldDelta != 0 {
		lines = append(lines, fmt.Sprintf(i18n.T("result_gold_delta"), m.summary.GoldDelta))
	}
	if m.summary.DefenseDelta != 0 {
		lines = append(lines, fmt.Sprintf(i18n.T("result_defense_delta"), m.summary.DefenseDelta))
	}

	if strings.TrimSpace(m.summary.Note) != "" {
		lines = append(lines, fmt.Sprintf(i18n.T("result_note"), m.summary.Note))
	}

	if m.summary.LeveledUp {
		lines = append(lines, lipgloss.NewStyle().Foreground(components.ColorPrimary).Render(i18n.T("result_leveled_up")))
	}
	if m.summary.Fainted {
		lines = append(lines, lipgloss.NewStyle().Foreground(components.ColorDanger).Render(i18n.T("result_fainted")))
	}

	body := lipgloss.JoinVertical(lipgloss.Left, lines...)

	width := lipgloss.Width(header) - resultBoxStyle.GetHorizontalPadding()
	if width <= 0 {
		width = 40
	}
	resultBox := resultBoxStyle.Width(width).Render(body)

	topSeparator := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render("")
	bottomSeparator := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render("")
	footer := components.Footer(i18n.T("result_footer"), 0)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		topSeparator,
		resultBox,
		bottomSeparator,
		footer,
	)
}
