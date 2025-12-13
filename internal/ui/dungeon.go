package ui

import (
	"context"
	"encoding/json"
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
	dungeonStyle            = lipgloss.NewStyle().Padding(1, 2)
	dungeonTitleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13")) // Purple
	questionStyleDungeon    = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).PaddingBottom(1)
	optionStyleDungeon      = lipgloss.NewStyle().PaddingLeft(2)
	answerInputStyleDungeon = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	feedbackStyleDungeon    = lipgloss.NewStyle().PaddingTop(1)
	correctStyleDungeon     = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	incorrectStyleDungeon   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
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
	answers         []game.GrammarAnswer // To store answers for RunGrammarSession
}

// NewDungeonModel creates a new DungeonModel.
func NewDungeonModel(stats game.Stats, gc *services.GeminiClient) DungeonModel {
	ti := textinput.New()
	ti.Placeholder = "Your answer..."
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
		return DungeonQuestionMsg{Questions: grammarEnvelope.Traps}
	}
}

func (m DungeonModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case DungeonQuestionMsg:
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
					updatedStats, _, _ := game.RunGrammarSession(context.Background(), m.playerStats, m.answers)
					m.playerStats = updatedStats
					return m, func() tea.Msg { return TownToRootMsg{} } // For now, just return to town
				}
				return m, nil
			}

			// Process answer
			currentQ := m.questions[m.currentQuestion]
			isCorrect := (m.answerInput.Value() == currentQ.Options[currentQ.AnswerIndex])
			m.answers = append(m.answers, game.GrammarAnswer{Correct: isCorrect})

			if isCorrect {
				m.feedback = "Correct!"
				m.isCorrect = true
				// TODO: Update player stats (EXP, Combo, etc.) - will be handled by RunGrammarSession
			} else {
				m.feedback = fmt.Sprintf("Incorrect. The answer was: %s", currentQ.Options[currentQ.AnswerIndex])
				m.isCorrect = false
				// TODO: Update player stats (HP, Combo reset, etc.) - will be handled by RunGrammarSession
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

func (m DungeonModel) View() string {
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
			content += feedbackStyleDungeon.Render(m.feedback)
		}
	} else {
		currentQ := m.questions[m.currentQuestion]
		questionText := questionStyleDungeon.Render(fmt.Sprintf("Question %d/%d: %s", m.currentQuestion+1, len(m.questions), currentQ.Question))

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
		dungeonStyle.Render(lipgloss.NewStyle().Width(lipgloss.Width(header)-dungeonStyle.GetHorizontalPadding()).Render(content)),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
		footer,
	)
}
