package ui

import (
	"context"
	"encoding/json"
	"fmt"
	// "strings" // Removed strings import

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/config"
	"tui-english-quest/internal/game"
	"tui-english-quest/internal/i18n"
	"tui-english-quest/internal/services"
	"tui-english-quest/internal/ui/components"
)

var (
	dungeonStyle            = lipgloss.NewStyle().Padding(1, 2)
	dungeonTitleStyle       = lipgloss.NewStyle().Bold(true).Foreground(components.ColorPurple) // Purple
	questionStyleDungeon    = lipgloss.NewStyle().Foreground(components.ColorInfo).PaddingBottom(1)
	optionStyleDungeon      = lipgloss.NewStyle().PaddingLeft(2)
	answerInputStyleDungeon = lipgloss.NewStyle().Foreground(components.ColorMuted)
	feedbackStyleDungeon    = lipgloss.NewStyle().PaddingTop(1)
	correctStyleDungeon     = lipgloss.NewStyle().Foreground(components.ColorPrimary)
	incorrectStyleDungeon   = lipgloss.NewStyle().Foreground(components.ColorDanger)
)

// DungeonModel represents the grammar dungeon screen.
type DungeonModel struct {
	playerStats     game.Stats
	geminiClient    *services.GeminiClient
	questions       []services.GrammarTrap
	currentQuestion int
	answerInput     textinput.Model
	mode            string
	feedback        string
	isCorrect       bool
	showFeedback    bool
	quitting        bool
	hpAnimator      HPAnimator
	answers         []game.GrammarAnswer // To store answers for RunGrammarSession
}

// NewDungeonModel creates a new DungeonModel.
func NewDungeonModel(stats game.Stats, gc *services.GeminiClient) DungeonModel {
	ti := textinput.New()
	ti.Placeholder = i18n.T("dungeon_placeholder")
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	return DungeonModel{
		playerStats:     stats,
		geminiClient:    gc,
		questions:       []services.GrammarTrap{},
		currentQuestion: 0,
		answerInput:     ti,
		mode:            services.ModeGrammar, // This model is specifically for Grammar
		feedback:        "",
		isCorrect:       false,
		showFeedback:    false,
		quitting:        false,
		hpAnimator:      NewHPAnimator(stats.HP),
		answers:         make([]game.GrammarAnswer, 0, 5), // Initialize answers slice
	}
}

// DungeonQuestionMsg is a message to indicate questions have been fetched.
type DungeonQuestionMsg struct {
	Questions []services.GrammarTrap
	Err       error
}

func (m DungeonModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.fetchQuestionsCmd())
}

func (m DungeonModel) fetchQuestionsCmd() tea.Cmd {
	return func() tea.Msg {
		payload, err := services.FetchAndValidate(context.Background(), m.mode)
		if err != nil {
			return DungeonQuestionMsg{Err: err}
		}

		var grammarEnvelope struct {
			Traps []services.GrammarTrap `json:"traps"`
		}
		if err := json.Unmarshal(payload.Content, &grammarEnvelope); err != nil {
			return DungeonQuestionMsg{Err: fmt.Errorf("failed to parse grammar questions: %w", err)}
		}
		cfg, _ := config.LoadConfig()
		N := cfg.QuestionsPerSession
		if N <= 0 {
			N = 5
		}
		qs := grammarEnvelope.Traps
		if len(qs) > N {
			qs = qs[:N]
		}
		return DungeonQuestionMsg{Questions: qs}
	}
}

func (m DungeonModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case DungeonQuestionMsg:
		if msg.Err != nil {
			m.feedback = fmt.Sprintf(i18n.T("error_fetching_questions"), msg.Err)
			m.showFeedback = true
			return m, nil
		}
		m.questions = msg.Questions
		return m, nil

	case hpTickMsg:
		return m, m.hpAnimator.Tick(m.playerStats.HP)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg { return TownToRootMsg{} } // Return to Town

		case "enter":
			if m.showFeedback {
				// Move to next question or end session
				m.showFeedback = false
				m.answerInput.SetValue("")
				// If the next index would be past the last question, end session now
				if m.currentQuestion+1 >= len(m.questions) {
					return m.finalizeGrammarSession()
				}

				m.currentQuestion++
				return m, nil
			}

			// Process answer
			// Defensive: ensure currentQuestion is within bounds
			if m.currentQuestion < 0 || m.currentQuestion >= len(m.questions) {
				m.feedback = i18n.T("dungeon_error_state")
				m.showFeedback = true
				m.isCorrect = false
				return m, nil
			}
			currentQ := m.questions[m.currentQuestion]
			isCorrect := (m.answerInput.Value() == currentQ.Options[currentQ.AnswerIndex])
			m.answers = append(m.answers, game.GrammarAnswer{Correct: isCorrect})

			// Auto-finalize when answers reach configured count
			if len(m.answers) == len(m.questions) {
				return m.finalizeGrammarSession()
			}

			if isCorrect {
				m.feedback = i18n.T("correct_feedback")
				m.isCorrect = true
				// TODO: Update player stats (EXP, Combo, etc.) - will be handled by RunGrammarSession
			} else {
				m.feedback = fmt.Sprintf(i18n.T("dungeon_incorrect_answer"), currentQ.Options[currentQ.AnswerIndex])
				m.isCorrect = false
				prevHP := m.playerStats.HP
				// Immediate HP update for UX
				m.playerStats.MaxHP = game.MaxHPForLevel(m.playerStats.Level)
				M := game.AllowedMisses(len(m.questions))
				dmg := game.DamagePerMiss(m.playerStats.MaxHP, M)
				m.playerStats = game.ApplyDamage(m.playerStats, dmg)
				m.playerStats = game.ResetCombo(m.playerStats)
				m.showFeedback = true
				return m, m.hpAnimator.StartAnimation(prevHP, m.playerStats.HP)
			}
			m.showFeedback = true
			return m, nil

		}
	}

	// Handle text input
	if !m.showFeedback {
		m.answerInput, cmd = m.answerInput.Update(msg)
	}

	return m, cmd
}

func (m DungeonModel) finalizeGrammarSession() (DungeonModel, tea.Cmd) {
	updatedStats, summary, err := game.RunGrammarSession(context.Background(), m.playerStats, m.answers)
	if err != nil {
		m.feedback = fmt.Sprintf("Session error: %v", err)
		m.showFeedback = true
		return m, nil
	}
	m.playerStats = updatedStats
	m.hpAnimator.Sync(m.playerStats.HP)
	m.showFeedback = true
	m.answerInput.SetValue("")
	return m, func() tea.Msg { return SessionResultMsg{Stats: m.playerStats, Summary: summary} }
}

func (m DungeonModel) View() string {
	if m.quitting {
		return i18n.T("exiting_message") + "\n"
	}

	displayStats := m.playerStats
	displayStats.HP = m.hpAnimator.Display()
	header := components.Header(displayStats, true, 0)

	var content string
	if len(m.questions) == 0 {
		content = i18n.FetchingFor("dungeon") + "\n"
		if m.showFeedback { // Display error if fetching failed
			content += feedbackStyleDungeon.Render(m.feedback)
		}
	} else {
		// Guard: currentQuestion may equal len(questions) transiently after Update finalizes the session
		if m.currentQuestion >= len(m.questions) {
			content = i18n.T("session_complete") + "\n\n"
			if m.showFeedback {
				content += feedbackStyleDungeon.Render(m.feedback)
			}
			footer := components.Footer(i18n.T("footer_dungeon"), 0)
			return lipgloss.JoinVertical(lipgloss.Left,
				header,
				lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
				dungeonStyle.Render(lipgloss.NewStyle().Width(lipgloss.Width(header)-dungeonStyle.GetHorizontalPadding()).Render(content)),
				lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
				footer,
			)
		}

		currentQ := m.questions[m.currentQuestion]
		questionText := questionStyleDungeon.Render(fmt.Sprintf(i18n.T("dungeon_question_progress"), m.currentQuestion+1, len(m.questions), currentQ.Question))

		// Calculate content width based on header width
		contentWidth := lipgloss.Width(header) - dungeonStyle.GetHorizontalPadding()

		var renderedOptions []string
		for _, opt := range currentQ.Options {
			renderedOptions = append(renderedOptions, optionStyleDungeon.Width(contentWidth-optionStyleDungeon.GetHorizontalPadding()).Render(fmt.Sprintf("  %s", opt)))
		}
		optionsText := lipgloss.JoinVertical(lipgloss.Left, renderedOptions...)

		inputField := answerInputStyleDungeon.Render(fmt.Sprintf("\n%s\n", m.answerInput.View()))

		feedbackText := ""
		if m.showFeedback {
			if m.isCorrect {
				feedbackText = feedbackStyleDungeon.Render(correctStyleDungeon.Render(m.feedback))
			} else {
				feedbackText = feedbackStyleDungeon.Render(incorrectStyleDungeon.Render(m.feedback))
			}
			feedbackText += "\n" + i18n.T("press_enter_continue")
		}

		content = lipgloss.JoinVertical(lipgloss.Left,
			questionText,
			optionsText,
			inputField,
			feedbackText,
		)
	}

	footer := components.Footer(i18n.T("footer_dungeon"), 0)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
		dungeonStyle.Render(lipgloss.NewStyle().Width(lipgloss.Width(header)-dungeonStyle.GetHorizontalPadding()).Render(content)),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
		footer,
	)
}
