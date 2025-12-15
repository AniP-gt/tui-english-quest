package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tui-english-quest/internal/game"
	"tui-english-quest/internal/services"
	"tui-english-quest/internal/ui/components"
)

var (
	listeningStyle      = lipgloss.NewStyle().Padding(1, 2)
	listeningTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(components.ColorPrimary)
)

// ListeningModel is an interactive TUI for the Listening Cave.
type ListeningModel struct {
	playerStats  game.Stats
	geminiClient *services.GeminiClient
	items        []services.ListeningItem
	currentIndex int
	selected     int
	answers      []game.ListeningAnswer
	feedback     string
	showFeedback bool
	quitting     bool
}

// NewListeningModel creates a new ListeningModel.
func NewListeningModel(stats game.Stats, gc *services.GeminiClient) ListeningModel {
	return ListeningModel{
		playerStats:  stats,
		geminiClient: gc,
		items:        []services.ListeningItem{},
		currentIndex: 0,
		selected:     0,
		answers:      make([]game.ListeningAnswer, 0, 5),
	}
}

func (m ListeningModel) Init() tea.Cmd {
	return tea.Batch(m.fetchQuestionsCmd())
}

// Question fetch message
type ListeningQuestionMsg struct {
	Items []services.ListeningItem
	Err   error
}

func (m ListeningModel) fetchQuestionsCmd() tea.Cmd {
	return func() tea.Msg {
		payload, err := services.FetchAndValidate(context.Background(), services.ModeListening)
		if err != nil {
			return ListeningQuestionMsg{Err: err}
		}
		var env struct {
			Audio []services.ListeningItem `json:"audio"`
		}
		if err := json.Unmarshal(payload.Content, &env); err != nil {
			return ListeningQuestionMsg{Err: err}
		}
		return ListeningQuestionMsg{Items: env.Audio}
	}
}

func (m ListeningModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case ListeningQuestionMsg:
		if msg.Err != nil {
			m.feedback = fmt.Sprintf("Error fetching listening items: %v", msg.Err)
			m.showFeedback = true
			return m, nil
		}
		m.items = msg.Items
		// speak first prompt
		if len(m.items) > 0 {
			_ = services.Speak(m.items[0].Prompt)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg { return ListeningToTownMsg{} }
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < 3 {
				m.selected++
			}
		case "r":
			// replay audio
			if m.currentIndex < len(m.items) {
				_ = services.Speak(m.items[m.currentIndex].Prompt)
			}
		case "1", "2", "3", "4":
			// choose numeric option
			n := int(msg.String()[0] - '1')
			m.selected = n
			fallthrough
		case "enter":
			if m.showFeedback {
				// advance
				m.showFeedback = false
				m.selected = 0
				m.currentIndex++
				if m.currentIndex >= len(m.items) {
					// end session
					updatedStats, _, _ := game.RunListeningSession(context.Background(), m.playerStats, m.answers)
					m.playerStats = updatedStats
					return m, func() tea.Msg { return ListeningToTownMsg{} }
				}
				// speak next prompt
				_ = services.Speak(m.items[m.currentIndex].Prompt)
				return m, nil
			}
			// submit answer for current question
			if m.currentIndex < len(m.items) {
				item := m.items[m.currentIndex]
				isCorrect := m.selected == item.AnswerIndex
				m.answers = append(m.answers, game.ListeningAnswer{Correct: isCorrect})
				if isCorrect {
					m.feedback = "Correct!"
				} else {
					m.feedback = fmt.Sprintf("Incorrect. Answer: %s", item.Options[item.AnswerIndex])
				}
				m.showFeedback = true
				return m, nil
			}
		}
	}

	return m, cmd
}

func (m ListeningModel) View() string {
	if m.quitting {
		return "Exiting TUI English Quest...\n"
	}
	header := components.Header(m.playerStats, true, 0)

	if len(m.items) == 0 {
		content := "Fetching listening items...\n"
		if m.showFeedback {
			content += m.feedback + "\n"
		}
		footer := components.Footer("[r] Replay  [Enter] Answer/Continue  [Esc/q] Back to Town", 0)
		return lipgloss.JoinVertical(lipgloss.Left,
			header,
			lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
			listeningStyle.Render(content),
			lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
			footer,
		)
	}

	item := m.items[m.currentIndex]
	qText := listeningTitleStyle.Render(fmt.Sprintf("Listening %d/%d", m.currentIndex+1, len(m.items))) + "\n\n"
	qText += fmt.Sprintf("(Press [r] to replay)\n\n")

	var opts []string
	for i, o := range item.Options {
		prefix := "  "
		if i == m.selected {
			prefix = "> "
		}
		opts = append(opts, fmt.Sprintf("%s%d) %s", prefix, i+1, o))
	}

	optText := strings.Join(opts, "\n")

	feedbackText := ""
	if m.showFeedback {
		feedbackText = "\n" + m.feedback + "\nPress Enter to continue..."
	}

	footer := components.Footer("[j/k] Move  [1-4] Quick select  [r] Replay  [Enter] Answer/Continue  [Esc/q] Back to Town", 0)

	content := lipgloss.JoinVertical(lipgloss.Left,
		qText,
		optText,
		feedbackText,
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
		listeningStyle.Render(content),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
		footer,
	)
}
