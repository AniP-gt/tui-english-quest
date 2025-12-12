package ui

import (
	"context" // Add context import
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/game"
	"tui-english-quest/internal/services"
	"tui-english-quest/internal/ui/components"
)

var (
	analysisStyle        = lipgloss.NewStyle().Padding(1, 2)
	analysisTitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	analysisSectionStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	analysisItemStyle    = lipgloss.NewStyle().PaddingLeft(2)
)

// AnalysisModel displays the AI weakness analysis.
type AnalysisModel struct {
	playerStats  game.Stats
	report       services.WeaknessReport
	geminiClient *services.GeminiClient // Add GeminiClient
}

// NewAnalysisModel creates a new AnalysisModel.
func NewAnalysisModel(stats game.Stats, gc *services.GeminiClient) AnalysisModel {
	// In a real implementation, playerID would be passed and history fetched.
	report, err := services.AnalyzeWeakness(context.Background(), gc, stats.Name, 200)
	if err != nil {
		// Handle error, e.g., log it and return an empty report or a report with an error message
		report = services.WeaknessReport{
			Recommendation: fmt.Sprintf("Error analyzing weakness: %v", err),
		}
	}
	return AnalysisModel{
		playerStats:  stats,
		report:       report,
		geminiClient: gc,
	}
}

func (m AnalysisModel) Init() tea.Cmd {
	return nil
}

func (m AnalysisModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "enter":
			return m, func() tea.Msg { return TownToRootMsg{} } // Return to Town
		}
	}
	return m, nil
}

func (m AnalysisModel) View() string {
	s := m.playerStats
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(lipgloss.Width(components.View(s))).Render("TUI English Quest"),
		components.View(s),
	)

	var b strings.Builder
	b.WriteString(analysisTitleStyle.Render("AI Analysis\n"))
	b.WriteString(lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""))

	b.WriteString(analysisSectionStyle.Render("\nYour recent performance (last 200 questions)\n"))

	b.WriteString(analysisSectionStyle.Render("\nWeak points:\n"))
	if len(m.report.WeakPoints) == 0 {
		b.WriteString(analysisItemStyle.Render("- None identified\n"))
	} else {
		for _, wp := range m.report.WeakPoints {
			b.WriteString(analysisItemStyle.Render(fmt.Sprintf("- %s\n", wp)))
		}
	}

	b.WriteString(analysisSectionStyle.Render("\nStrengths:\n"))
	if len(m.report.StrengthPoints) == 0 {
		b.WriteString(analysisItemStyle.Render("- None identified\n"))
	} else {
		for _, sp := range m.report.StrengthPoints {
			b.WriteString(analysisItemStyle.Render(fmt.Sprintf("- %s\n", sp)))
		}
	}

	b.WriteString(analysisSectionStyle.Render("\nRecommendations:\n"))
	b.WriteString(analysisItemStyle.Render(fmt.Sprintf("- %s\n", m.report.Recommendation)))

	footer := "\n[Enter] OK  [Esc] Back to Town"

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""), // Separator
		analysisStyle.Render(b.String()),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""), // Separator
		footer,
	)
}
