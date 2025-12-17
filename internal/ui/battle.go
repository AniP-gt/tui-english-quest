package ui

import (
	"context"
	"encoding/json" // Added for JSON unmarshalling
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
	battleStyle         = lipgloss.NewStyle().Padding(1, 2)
	battleTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(components.ColorDanger)
	questionStyle       = lipgloss.NewStyle().Foreground(components.ColorInfo).PaddingBottom(1)
	optionStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedOptionStyle = lipgloss.NewStyle().Foreground(components.ColorPrimary).Bold(true).PaddingLeft(1)
	answerInputStyle    = lipgloss.NewStyle().Foreground(components.ColorMuted)
	feedbackStyle       = lipgloss.NewStyle().PaddingTop(1)
	correctStyle        = lipgloss.NewStyle().Foreground(components.ColorPrimary)
	incorrectStyle      = lipgloss.NewStyle().Foreground(components.ColorDanger)
)

// BattleModel represents the vocabulary battle screen.
type BattleModel struct {
	playerStats     game.Stats
	geminiClient    *services.GeminiClient
	questions       []services.VocabQuestion
	currentQuestion int
	answerInput     textinput.Model
	// selectedOption int // For multiple choice - removed
	mode         string
	feedback     string
	isCorrect    bool
	showFeedback bool
	quitting     bool
	hpAnimator   HPAnimator
	answers      []game.VocabAnswer // To store answers for RunVocabSession
}

// NewBattleModel creates a new BattleModel.
func NewBattleModel(stats game.Stats, gc *services.GeminiClient) BattleModel {
	ti := textinput.New()
	ti.Placeholder = i18n.T("battle_placeholder")
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	return BattleModel{
		playerStats:     stats,
		geminiClient:    gc,
		questions:       []services.VocabQuestion{},
		currentQuestion: 0,
		answerInput:     ti,
		// selectedOption: 0, // removed
		mode:         services.ModeVocab, // This model is specifically for Vocab
		feedback:     "",
		isCorrect:    false,
		showFeedback: false,
		quitting:     false,
		hpAnimator:   NewHPAnimator(stats.HP),
		answers:      make([]game.VocabAnswer, 0, 5), // Initialize answers slice
	}
}

// BattleQuestionMsg is a message to indicate questions have been fetched.
type BattleQuestionMsg struct {
	Questions []services.VocabQuestion
	Err       error
}

func (m BattleModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.fetchQuestionsCmd())
}

func (m BattleModel) fetchQuestionsCmd() tea.Cmd {
	return func() tea.Msg {
		payload, err := services.FetchAndValidate(context.Background(), m.mode)
		if err != nil {
			return BattleQuestionMsg{Err: err}
		}

		var vocabEnvelope struct {
			Questions []services.VocabQuestion `json:"questions"`
		}
		if err := json.Unmarshal(payload.Content, &vocabEnvelope); err != nil {
			return BattleQuestionMsg{Err: fmt.Errorf("failed to parse vocab questions: %w", err)}
		}
		// read configured questions per session from config
		cfg, _ := config.LoadConfig()
		N := cfg.QuestionsPerSession
		if N <= 0 {
			N = 5
		}
		qs := vocabEnvelope.Questions
		if len(qs) > N {
			qs = qs[:N]
		}
		return BattleQuestionMsg{Questions: qs}
	}
}

func (m BattleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case BattleQuestionMsg:
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
		case "ctrl+c", "q":
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
					// End session, show results
					updatedStats, _, _ := game.RunVocabSession(context.Background(), m.playerStats, m.answers)
					m.playerStats = updatedStats
					m.hpAnimator.Sync(m.playerStats.HP)
					return m, func() tea.Msg { return TownToRootMsg{} } // For now, just return to town
				}

				m.currentQuestion++
				return m, nil
			}

			// Process answer
			// Defensive: ensure currentQuestion is within bounds
			if m.currentQuestion < 0 || m.currentQuestion >= len(m.questions) {
				// Out-of-range state: ignore input and reset feedback
				m.feedback = i18n.T("battle_error_state")
				m.showFeedback = true
				m.isCorrect = false
				return m, nil
			}
			currentQ := m.questions[m.currentQuestion]
			isCorrect := (m.answerInput.Value() == currentQ.Options[currentQ.AnswerIndex])
			m.answers = append(m.answers, game.VocabAnswer{Correct: isCorrect})

			// If this was the last answer, finalize session immediately
			if len(m.answers) == len(m.questions) {
				updatedStats, _, _ := game.RunVocabSession(context.Background(), m.playerStats, m.answers)
				m.playerStats = updatedStats
				m.hpAnimator.Sync(m.playerStats.HP)
				m.showFeedback = true
				m.answerInput.SetValue("")
				return m, func() tea.Msg { return TownToRootMsg{} }
			}

			if isCorrect {
				m.feedback = i18n.T("correct_feedback")
				m.isCorrect = true
				// TODO: Update player stats (EXP, Combo, etc.) - will be handled by RunVocabSession at session end
			} else {
				m.feedback = fmt.Sprintf(i18n.T("battle_incorrect_answer"), currentQ.Options[currentQ.AnswerIndex])
				m.isCorrect = false
				prevHP := m.playerStats.HP
				// Immediate HP update for UX: compute damage and apply to playerStats
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

			// case "up", "k": // removed
			// 	if !m.answerInput.Focused() && m.selectedOption > 0 {
			// 		m.selectedOption--
			// 	}
			// case "down", "j": // removed
			// 	if !m.answerInput.Focused() && m.selectedOption < len(m.questions[m.currentQuestion].Options)-1 {
			// 		m.selectedOption++
			// 	}
		}
	}

	// Handle text input
	if !m.showFeedback {
		m.answerInput, cmd = m.answerInput.Update(msg)
	}

	return m, cmd
}

func (m BattleModel) View() string {
	if m.quitting {
		return i18n.T("exiting_message") + "\n"
	}

	displayStats := m.playerStats
	displayStats.HP = m.hpAnimator.Display()
	header := components.Header(displayStats, true, 0)

	var content string
	if len(m.questions) == 0 {
		content = i18n.FetchingFor("battle") + "\n"
		if m.showFeedback { // Display error if fetching failed
			content += feedbackStyle.Render(m.feedback)
		}
	} else {
		// Guard: currentQuestion may equal len(questions) transiently after Update finalizes the session
		if m.currentQuestion >= len(m.questions) {
			content = i18n.T("session_complete") + "\n\n"
			if m.showFeedback {
				content += feedbackStyle.Render(m.feedback)
			}
			footer := components.Footer(i18n.T("footer_battle"), 0)
			return lipgloss.JoinVertical(lipgloss.Left,
				header,
				lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
				battleStyle.Render(lipgloss.NewStyle().Width(lipgloss.Width(header)-battleStyle.GetHorizontalPadding()).Render(content)),
				lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
				footer,
			)
		}

		currentQ := m.questions[m.currentQuestion]
		questionText := questionStyle.Render(fmt.Sprintf(i18n.T("battle_question_format"), m.currentQuestion+1, len(m.questions), currentQ.Word))

		// Calculate content width based on header width
		contentWidth := lipgloss.Width(header) - battleStyle.GetHorizontalPadding()

		var renderedOptions []string
		for _, opt := range currentQ.Options {
			renderedOptions = append(renderedOptions, optionStyle.Width(contentWidth-optionStyle.GetHorizontalPadding()).Render(fmt.Sprintf("  %s", opt)))
		}
		optionsText := lipgloss.JoinVertical(lipgloss.Left, renderedOptions...)

		inputField := answerInputStyle.Render(fmt.Sprintf("\n%s\n", m.answerInput.View()))

		feedbackText := ""
		if m.showFeedback {
			if m.isCorrect {
				feedbackText = feedbackStyle.Render(correctStyle.Render(m.feedback))
			} else {
				feedbackText = feedbackStyle.Render(incorrectStyle.Render(m.feedback))
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

	footer := components.Footer(i18n.T("footer_battle"), 0)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
		battleStyle.Render(lipgloss.NewStyle().Width(lipgloss.Width(header)-battleStyle.GetHorizontalPadding()).Render(content)),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
		footer,
	)
}
