package ui

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"tui-english-quest/internal/game"
	"tui-english-quest/internal/i18n"
	"tui-english-quest/internal/services"
	"tui-english-quest/internal/ui/components"
)

// TavernModel represents the conversation tavern UI.
type TavernModel struct {
	playerStats      game.Stats
	geminiClient     *services.GeminiClient
	npcName          string
	npcOpening       string
	turns            []services.TavernTurn
	evaluationRubric []string

	playerUtterances []string
	evaluations      []services.TavernEvaluation

	currentTurn int
	input       textinput.Model

	loading      bool
	feedback     string
	showFeedback bool
	quitting     bool

	langPref string // "en"/"ja"
}

type TavernQuestionMsg struct {
	NPCName          string
	NPCOpening       string
	EvaluationRubric []string
	Turns            []services.TavernTurn
	Err              error
}

type TavernEvalMsg struct {
	Evaluations []services.TavernEvaluation
	Err         error
}

func NewTavernModel(stats game.Stats, gc *services.GeminiClient, langPref string) TavernModel {
	ti := textinput.New()
	ti.Placeholder = i18n.T("tavern_placeholder")

	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 60

	return TavernModel{
		playerStats:      stats,
		geminiClient:     gc,
		turns:            []services.TavernTurn{},
		evaluationRubric: []string{},
		playerUtterances: make([]string, 0, 5),
		evaluations:      []services.TavernEvaluation{},
		currentTurn:      0,
		input:            ti,
		langPref:         langPref,
	}
}

func (m TavernModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.fetchTavernCmd())
}

func (m TavernModel) fetchTavernCmd() tea.Cmd {
	return func() tea.Msg {
		payload, err := services.FetchAndValidate(context.Background(), services.ModeTavern)
		if err != nil {
			return TavernQuestionMsg{Err: err}
		}

		var env struct {
			NPCName          string                `json:"npc_name"`
			NPCOpening       string                `json:"npc_opening"`
			EvaluationRubric []string              `json:"evaluation_rubric"`
			Turns            []services.TavernTurn `json:"turns"`
		}
		if err := json.Unmarshal(payload.Content, &env); err != nil {
			return TavernQuestionMsg{Err: fmt.Errorf("failed to parse tavern payload: %w", err)}
		}
		return TavernQuestionMsg{
			NPCName:          env.NPCName,
			NPCOpening:       env.NPCOpening,
			EvaluationRubric: env.EvaluationRubric,
			Turns:            env.Turns,
			Err:              nil,
		}
	}
}

func (m TavernModel) batchEvaluateCmd() tea.Cmd {
	return func() tea.Msg {
		evals, err := m.geminiClient.BatchEvaluateTavern(context.Background(), m.evaluationRubric, m.npcOpening, m.turns, m.playerUtterances, m.langPref)
		if err != nil {
			return TavernEvalMsg{Err: err}
		}
		return TavernEvalMsg{Evaluations: evals}
	}
}

func (m TavernModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case TavernQuestionMsg:
		if msg.Err != nil {
			m.feedback = fmt.Sprintf(i18n.T("error_fetching_questions"), msg.Err)
			m.showFeedback = true
			return m, nil
		}
		m.npcName = msg.NPCName
		m.npcOpening = msg.NPCOpening
		m.evaluationRubric = msg.EvaluationRubric
		m.turns = msg.Turns
		return m, nil

	case TavernEvalMsg:
		if msg.Err != nil {
			m.feedback = i18n.T("tavern_eval_default_fail")
			m.evaluations = make([]services.TavernEvaluation, 5)
			for i := range m.evaluations {
				m.evaluations[i] = services.TavernEvaluation{
					Outcome: "normal",
					Reason:  i18n.T("tavern_eval_default_fail"),
				}
			}
		} else {
			m.evaluations = msg.Evaluations
		}

		outcomes := make([]TavernOutcome, len(m.evaluations))
		for i, e := range m.evaluations {
			switch e.Outcome {
			case "success":
				outcomes[i] = OutcomeSuccess
			case "normal":
				outcomes[i] = OutcomeNormal
			case "fail":
				outcomes[i] = OutcomeFail
			default:
				outcomes[i] = OutcomeNormal
			}
		}

		updatedStats, summary, _ := RunTavernSession(context.Background(), m.playerStats, outcomes)
		m.playerStats = updatedStats
		m.feedback = fmt.Sprintf(i18n.T("tavern_finished_format"), summary.ExpDelta, summary.GoldDelta, summary.Correct)
		m.showFeedback = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg { return TownToRootMsg{} }
		case "enter":
			if m.showFeedback {
				if m.currentTurn >= len(m.turns) {
					return m, func() tea.Msg { return TownToRootMsg{} }
				}
				m.showFeedback = false
				m.input.SetValue("")
				return m, nil
			}

			ut := m.input.Value()
			m.playerUtterances = append(m.playerUtterances, ut)
			m.input.SetValue("")
			m.currentTurn++

			if m.currentTurn >= len(m.turns) {
				m.loading = true
				return m, m.batchEvaluateCmd()
			}
			return m, nil
		}
	}

	if !m.showFeedback && !m.loading {
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m TavernModel) View() string {
	if m.quitting {
		return i18n.T("tavern_exiting") + "\n"
	}
	header := components.Header(m.playerStats, true, 0)

	var content string
	if len(m.turns) == 0 {
		content = i18n.FetchingFor("tavern") + "\n"
		if m.showFeedback {
			content += m.feedback
		}
	} else {
		if m.currentTurn == 0 {
			content += fmt.Sprintf("%s\n\n", m.npcOpening)
		}
		if m.currentTurn < len(m.turns) {
			content += fmt.Sprintf(i18n.T("tavern_npc_line"), m.npcName, m.turns[m.currentTurn].NPCReply) + "\n\n"
			content += fmt.Sprintf(i18n.T("tavern_player_turn"), m.currentTurn+1, len(m.turns), m.input.View())
		} else {
			content += i18n.T("tavern_evaluations") + "\n"
			for i, ev := range m.evaluations {
				content += fmt.Sprintf(i18n.T("tavern_eval_line"), i+1, ev.Outcome, ev.Reason)
			}
			content += "\n" + m.feedback + "\n"
			content += i18n.T("press_enter_return")
		}

		if m.showFeedback {
			content += "\n" + m.feedback + "\n"
		}
	}

	footer := components.Footer("[Enter] Send  [Esc] Back to Town  [q/ctrl+c] Quit", 0)
	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}
