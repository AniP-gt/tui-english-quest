package ui

import (
	"context" // Add context import
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
	analysisStyle        = lipgloss.NewStyle().Padding(1, 2)
	analysisTitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(components.ColorPrimary)
	analysisSectionStyle = lipgloss.NewStyle().Bold(true).Foreground(components.ColorMuted)
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
	report, err := services.AnalyzeWeakness(context.Background(), gc, stats.Name, stats, 200)
	if err != nil {
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
			return m, func() tea.Msg { return AnalysisToTownMsg{} }
		}
	}
	return m, nil
}

func (m AnalysisModel) View() string {
	s := m.playerStats
	header := components.Header(s, true, 0)

	var b strings.Builder
	b.WriteString(analysisTitleStyle.Render(i18n.T("analysis_title") + "\n"))
	b.WriteString(lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""))

	summary := m.report.Summary
	if summary == "" {
		summary = i18n.T("analysis_list_none")
	}
	b.WriteString(analysisSectionStyle.Render("\n" + i18n.T("analysis_summary") + "\n"))
	b.WriteString(analysisItemStyle.Render(fmt.Sprintf("- %s\n", summary)))

	b.WriteString(analysisSectionStyle.Render("\n" + i18n.T("analysis_weak_points") + "\n"))
	if len(m.report.WeakPoints) == 0 {
		b.WriteString(analysisItemStyle.Render("- " + i18n.T("analysis_list_none") + "\n"))
	} else {
		for _, wp := range m.report.WeakPoints {
			line := fmt.Sprintf("- %s: %.0f%% (%s)", modeLabel(wp.Mode), wp.Accuracy*100, formatTrend(wp.Trend))
			if wp.Description != "" {
				line = fmt.Sprintf("%s\n  %s", line, wp.Description)
			}
			b.WriteString(analysisItemStyle.Render(line + "\n"))
		}
	}

	b.WriteString(analysisSectionStyle.Render("\n" + i18n.T("analysis_strengths") + "\n"))
	if len(m.report.StrengthPoints) == 0 {
		b.WriteString(analysisItemStyle.Render("- " + i18n.T("analysis_list_none") + "\n"))
	} else {
		for _, sp := range m.report.StrengthPoints {
			line := fmt.Sprintf("- %s: %.0f%% (%s)", modeLabel(sp.Mode), sp.Accuracy*100, formatTrend(sp.Trend))
			if sp.Description != "" {
				line = fmt.Sprintf("%s\n  %s", line, sp.Description)
			}
			b.WriteString(analysisItemStyle.Render(line + "\n"))
		}
	}

	b.WriteString(analysisSectionStyle.Render("\n" + i18n.T("analysis_action_plan") + "\n"))
	if len(m.report.ActionPlan) == 0 {
		b.WriteString(analysisItemStyle.Render("- " + i18n.T("analysis_action_plan_empty") + "\n"))
	} else {
		for _, plan := range m.report.ActionPlan {
			label := plan.Title
			if plan.Mode != "" {
				label = fmt.Sprintf("%s (%s)", label, modeLabel(plan.Mode))
			}
			priority := plan.Priority
			if priority == "" {
				priority = "medium"
			}
			b.WriteString(analysisItemStyle.Render(fmt.Sprintf("- %s: %s [%s]\n", label, plan.Description, strings.Title(priority))))
		}
	}

	recommendation := m.report.Recommendation
	if recommendation == "" {
		recommendation = i18n.T("analysis_list_none")
	}
	b.WriteString(analysisSectionStyle.Render("\n" + i18n.T("analysis_recommendations") + "\n"))
	b.WriteString(analysisItemStyle.Render(fmt.Sprintf("- %s\n", recommendation)))

	footer := components.Footer(i18n.T("footer_analysis"), 0)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""), // Separator
		analysisStyle.Render(b.String()),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""), // Separator
		footer,
	)
}

func formatTrend(trend float64) string {
	switch {
	case trend > 0.015:
		return fmt.Sprintf("+%.0f%%", trend*100)
	case trend < -0.015:
		return fmt.Sprintf("-%.0f%%", -trend*100)
	default:
		return "stable"
	}
}
