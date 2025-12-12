package ui

import (
	"context"
	"encoding/json" // Added for JSON unmarshalling
	"fmt"
	// "strings" // Removed strings import

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/game"
	"tui-english-quest/internal/services"
	"tui-english-quest/internal/ui/components"
)

var (
	battleStyle         = lipgloss.NewStyle().Padding(1, 2)
	battleTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	questionStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).PaddingBottom(1)
	optionStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedOptionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).PaddingLeft(1)
	answerInputStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	feedbackStyle       = lipgloss.NewStyle().PaddingTop(1)
	correctStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	incorrectStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
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
	answers      []game.VocabAnswer // To store answers for RunVocabSession
}

// NewBattleModel creates a new BattleModel.
func NewBattleModel(stats game.Stats, gc *services.GeminiClient) BattleModel {
	ti := textinput.New()
	ti.Placeholder = "Your answer..."
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
		return BattleQuestionMsg{Questions: vocabEnvelope.Questions}
	}
}

func (m BattleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case BattleQuestionMsg:
		if msg.Err != nil {
			m.feedback = fmt.Sprintf("Error fetching questions: %v", msg.Err)
			m.showFeedback = true
			return m, nil
		}
		m.questions = msg.Questions
		return m, nil

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
				m.currentQuestion++
				if m.currentQuestion >= len(m.questions) {
					// End session, show results
					updatedStats, _, _ := game.RunVocabSession(context.Background(), m.playerStats, m.answers)
					m.playerStats = updatedStats
					return m, func() tea.Msg { return TownToRootMsg{} } // For now, just return to town
				}
				return m, nil
			}

			// Process answer
			currentQ := m.questions[m.currentQuestion]
			isCorrect := (m.answerInput.Value() == currentQ.Options[currentQ.AnswerIndex])
			m.answers = append(m.answers, game.VocabAnswer{Correct: isCorrect})

			if isCorrect {
				m.feedback = "Correct!"
				m.isCorrect = true
				// TODO: Update player stats (EXP, Combo, etc.) - will be handled by RunVocabSession
			} else {
				m.feedback = fmt.Sprintf("Incorrect. The answer was: %s", currentQ.Options[currentQ.AnswerIndex])
				m.isCorrect = false
				// TODO: Update player stats (HP, Combo reset, etc.) - will be handled by RunVocabSession
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
		return "Exiting TUI English Quest...\n"
	}

	s := m.playerStats
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(lipgloss.Width(components.View(s))).Render("TUI English Quest"),
		components.View(s),
	)

	var content string
	if len(m.questions) == 0 {
		content = "Fetching questions...\n"
		if m.showFeedback { // Display error if fetching failed
			content += feedbackStyle.Render(m.feedback)
		}
	} else {
		currentQ := m.questions[m.currentQuestion]
		questionText := questionStyle.Render(fmt.Sprintf("Question %d/%d: What is the meaning of '%s'?", m.currentQuestion+1, len(m.questions), currentQ.Word))

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
			feedbackText += "\nPress Enter to continue..."
		}

		content = lipgloss.JoinVertical(lipgloss.Left,
			questionText,
			optionsText,
			inputField,
			feedbackText,
		)
	}

	footer := "\n[j/k] Move  [Enter] Select/Answer  [Esc] Back to Town  [q/ctrl+c] Quit"

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
		battleStyle.Render(lipgloss.NewStyle().Width(lipgloss.Width(header)-battleStyle.GetHorizontalPadding()).Render(content)),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
		footer,
	)
}
