package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

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
	spellingStyle            = lipgloss.NewStyle().Padding(1, 2)
	spellingTitleStyle       = lipgloss.NewStyle().Bold(true).Foreground(components.ColorPrimary)
	spellingQuestionStyle    = lipgloss.NewStyle().Foreground(components.ColorInfo).PaddingBottom(1)
	spellingOptionStyle      = lipgloss.NewStyle().PaddingLeft(2)
	spellingAnswerInputStyle = lipgloss.NewStyle().Foreground(components.ColorMuted)
	spellingFeedbackStyle    = lipgloss.NewStyle().PaddingTop(1)
	spellingCorrectStyle     = lipgloss.NewStyle().Foreground(components.ColorPrimary)
	spellingIncorrectStyle   = lipgloss.NewStyle().Foreground(components.ColorDanger)
)

// SpellingModel is the TUI for the Spelling Challenge.
type SpellingModel struct {
	playerStats      game.Stats
	geminiClient     *services.GeminiClient
	prompts          []services.SpellingPrompt
	currentQuestion  int
	answerInput      textinput.Model
	isMultipleChoice bool
	mcOptions        []string
	showFeedback     bool
	isCorrect        bool
	feedback         string
	quitting         bool
	hpAnimator       HPAnimator
	answers          []game.SpellingOutcome
}

// SpellingQuestionMsg is sent when questions are fetched.
type SpellingQuestionMsg struct {
	Prompts []services.SpellingPrompt
	Err     error
}

// NewSpellingModel creates a new SpellingModel.
func NewSpellingModel(stats game.Stats, gc *services.GeminiClient) SpellingModel {
	ti := textinput.New()
	ti.Placeholder = i18n.T("spelling_placeholder")
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	return SpellingModel{
		playerStats:      stats,
		geminiClient:     gc,
		prompts:          []services.SpellingPrompt{},
		currentQuestion:  0,
		answerInput:      ti,
		isMultipleChoice: false,
		mcOptions:        nil,
		showFeedback:     false,
		isCorrect:        false,
		feedback:         "",
		quitting:         false,
		hpAnimator:       NewHPAnimator(stats.HP),
		answers:          make([]game.SpellingOutcome, 0, 5),
	}
}

func (m SpellingModel) Init() tea.Cmd { return tea.Batch(textinput.Blink, m.fetchQuestionsCmd()) }

func (m SpellingModel) fetchQuestionsCmd() tea.Cmd {
	return func() tea.Msg {
		payload, err := services.FetchAndValidate(context.Background(), services.ModeSpelling)
		if err != nil {
			return SpellingQuestionMsg{Err: err}
		}

		var env struct {
			Prompts []services.SpellingPrompt `json:"prompts"`
		}
		if err := json.Unmarshal(payload.Content, &env); err != nil {
			return SpellingQuestionMsg{Err: fmt.Errorf("failed to parse spelling prompts: %w", err)}
		}
		cfg, _ := config.LoadConfig()
		N := cfg.QuestionsPerSession
		if N <= 0 {
			N = 5
		}
		ps := env.Prompts
		if len(ps) > N {
			ps = ps[:N]
		}
		return SpellingQuestionMsg{Prompts: ps}
	}
}

func (m SpellingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case SpellingQuestionMsg:
		if msg.Err != nil {
			m.feedback = fmt.Sprintf(i18n.T("error_fetching_questions"), msg.Err)
			m.showFeedback = true
			return m, nil
		}
		m.prompts = msg.Prompts
		return m, nil

	case hpTickMsg:
		return m, m.hpAnimator.Tick(m.playerStats.HP)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg { return SpellingToTownMsg{} }
		case "tab":
			// Toggle multiple-choice for current prompt without inserting tab chars
			if !m.isMultipleChoice {
				m.isMultipleChoice = true
				m.mcOptions = generateMCOptions(m.prompts[m.currentQuestion].CorrectSpelling)
			} else {
				m.isMultipleChoice = false
				m.mcOptions = nil
			}
			return m, nil
		case "enter":
			if m.showFeedback {
				// Next question or finish
				m.showFeedback = false
				m.answerInput.SetValue("")
				m.currentQuestion++
				if m.currentQuestion >= len(m.prompts) {
					// End session
					updatedStats, _, _ := game.RunSpellingSession(context.Background(), m.playerStats, m.answers)
					m.playerStats = updatedStats
					return m, func() tea.Msg { return SpellingToTownMsg{} }
				}
				m = m.refreshMCOptionsForCurrentPrompt()
				return m, nil
			}
			// Process answer for fill-in
			if m.isMultipleChoice {
				// ignore enter for MC
				return m, nil
			}
			current := m.prompts[m.currentQuestion]
			user := strings.TrimSpace(m.answerInput.Value())
			if strings.EqualFold(user, current.CorrectSpelling) {
				m.feedback = i18n.T("correct_feedback")
				m.isCorrect = true
				m.answers = append(m.answers, game.SpellingPerfect)
			} else if isNear(user, current.CorrectSpelling) {
				m.feedback = fmt.Sprintf(i18n.T("spelling_almost_correct"), current.CorrectSpelling)
				m.isCorrect = false
				m.answers = append(m.answers, game.SpellingNear)
			} else {
				m.feedback = fmt.Sprintf(i18n.T("spelling_incorrect"), current.CorrectSpelling)
				m.isCorrect = false
				m.answers = append(m.answers, game.SpellingFail)
				prevHP := m.playerStats.HP
				// Immediate HP update for UX
				m.playerStats.MaxHP = game.MaxHPForLevel(m.playerStats.Level)
				M := game.AllowedMisses(len(m.prompts))
				dmg := game.DamagePerMiss(m.playerStats.MaxHP, M)
				m.playerStats = game.ApplyDamage(m.playerStats, dmg)
				m.playerStats = game.ResetCombo(m.playerStats)
				m.showFeedback = true
				return m, m.hpAnimator.StartAnimation(prevHP, m.playerStats.HP)
			}
			// Auto-finalize when we've answered all prompts
			if len(m.answers) == len(m.prompts) {
				updatedStats, _, _ := game.RunSpellingSession(context.Background(), m.playerStats, m.answers)
				m.playerStats = updatedStats
				m.hpAnimator.Sync(m.playerStats.HP)
				m.showFeedback = true
				return m, func() tea.Msg { return SpellingToTownMsg{} }
			}
			m.showFeedback = true
			return m, nil

		case "1", "2", "3", "4":
			// Handle MC selection
			if !m.isMultipleChoice || m.showFeedback {
				return m, nil
			}
			idx := int(msg.String()[0] - '1')
			if idx < 0 || idx >= len(m.mcOptions) {
				return m, nil
			}
			selected := m.mcOptions[idx]
			current := m.prompts[m.currentQuestion]
			if strings.EqualFold(selected, current.CorrectSpelling) {
				m.feedback = i18n.T("correct_feedback")
				m.isCorrect = true
				m.answers = append(m.answers, game.SpellingPerfect)
			} else if isNear(selected, current.CorrectSpelling) {
				m.feedback = fmt.Sprintf(i18n.T("spelling_almost_correct"), current.CorrectSpelling)
				m.isCorrect = false
				m.answers = append(m.answers, game.SpellingNear)
			} else {
				m.feedback = fmt.Sprintf(i18n.T("spelling_incorrect"), current.CorrectSpelling)
				m.isCorrect = false
				m.answers = append(m.answers, game.SpellingFail)
				prevHP := m.playerStats.HP
				// Immediate HP update for UX
				m.playerStats.MaxHP = game.MaxHPForLevel(m.playerStats.Level)
				M := game.AllowedMisses(len(m.prompts))
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

	// Handle text input when not showing feedback and not in MC mode
	if !m.showFeedback && !m.isMultipleChoice {
		m.answerInput, cmd = m.answerInput.Update(msg)
	}

	return m, cmd
}

func (m SpellingModel) View() string {
	if m.quitting {
		return i18n.T("exiting_message") + "\n"
	}

	displayStats := m.playerStats
	displayStats.HP = m.hpAnimator.Display()
	header := components.Header(displayStats, true, 0)

	var content string
	if len(m.prompts) == 0 {
		content = i18n.FetchingFor("spelling") + "\n"
		if m.showFeedback {
			content += spellingFeedbackStyle.Render(m.feedback)
		}
	} else {
		// Guard: currentQuestion may equal len(prompts) transiently after Update triggers session finalize
		if m.currentQuestion >= len(m.prompts) {
			content := spellingTitleStyle.Render(i18n.T("session_complete")) + "\n\n"
			if m.showFeedback {
				content += spellingFeedbackStyle.Render(m.feedback) + "\n"
			}
			footer := components.Footer(i18n.T("footer_spelling"), 0)
			return lipgloss.JoinVertical(lipgloss.Left,
				header,
				lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
				spellingStyle.Render(lipgloss.NewStyle().Width(lipgloss.Width(header)-spellingStyle.GetHorizontalPadding()).Render(content)),
				lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
				footer,
			)
		}

		current := m.prompts[m.currentQuestion]
		questionText := spellingQuestionStyle.Render(fmt.Sprintf(i18n.T("spelling_question_progress"), m.currentQuestion+1, len(m.prompts), current.JAHint))

		// Calculate content width based on header width
		contentWidth := lipgloss.Width(header) - spellingStyle.GetHorizontalPadding()

		if m.isMultipleChoice && m.mcOptions != nil {
			// Render options
			var rendered []string
			for i, opt := range m.mcOptions {
				rendered = append(rendered, spellingOptionStyle.Width(contentWidth-spellingOptionStyle.GetHorizontalPadding()).Render(fmt.Sprintf("%d. %s", i+1, opt)))
			}
			optionsText := lipgloss.JoinVertical(lipgloss.Left, rendered...)
			inputField := "" // no free input
			feedbackText := ""
			if m.showFeedback {
				if m.isCorrect {
					feedbackText = spellingFeedbackStyle.Render(spellingCorrectStyle.Render(m.feedback))
				} else {
					feedbackText = spellingFeedbackStyle.Render(spellingIncorrectStyle.Render(m.feedback))
				}
				feedbackText += "\n" + i18n.T("press_enter_continue")
			}
			content = lipgloss.JoinVertical(lipgloss.Left, questionText, optionsText, inputField, feedbackText)
		} else {
			// Fill-in mode
			inputField := spellingAnswerInputStyle.Render(fmt.Sprintf("\n%s\n", m.answerInput.View()))
			feedbackText := ""
			if m.showFeedback {
				if m.isCorrect {
					feedbackText = spellingFeedbackStyle.Render(spellingCorrectStyle.Render(m.feedback))
				} else {
					feedbackText = spellingFeedbackStyle.Render(spellingIncorrectStyle.Render(m.feedback))
				}
				feedbackText += "\n" + i18n.T("press_enter_continue")
			}
			content = lipgloss.JoinVertical(lipgloss.Left, questionText, inputField, feedbackText)
		}
	}

	footer := components.Footer(i18n.T("footer_spelling"), 0)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).Width(lipgloss.Width(header)).Render(""),
		spellingStyle.Render(lipgloss.NewStyle().Width(lipgloss.Width(header)-spellingStyle.GetHorizontalPadding()).Render(content)),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Width(lipgloss.Width(header)).Render(""),
		footer,
	)
}

// generateMCOptions creates 4 options including correct by simple mutations.
func generateMCOptions(correct string) []string {
	rand.Seed(time.Now().UnixNano())
	opts := map[string]struct{}{}
	opts[correct] = struct{}{}
	mutators := []func(string) string{
		func(s string) string { return swapOne(s) },
		func(s string) string { return removeOne(s) },
		func(s string) string { return duplicateOne(s) },
		func(s string) string { return replaceOne(s) },
	}
	// generate until we have 4 options or attempts exhausted
	attempts := 0
	for len(opts) < 4 && attempts < 20 {
		m := mutators[rand.Intn(len(mutators))]
		cand := m(correct)
		if cand == "" {
			attempts++
			continue
		}
		opts[cand] = struct{}{}
		attempts++
	}
	// Collect and shuffle
	list := make([]string, 0, len(opts))
	for k := range opts {
		list = append(list, k)
	}
	rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
	if len(list) > 4 {
		list = list[:4]
	}
	return list
}

func swapOne(s string) string {
	r := []rune(s)
	if len(r) < 2 {
		return ""
	}
	i := rand.Intn(len(r) - 1)
	r[i], r[i+1] = r[i+1], r[i]
	return string(r)
}

func removeOne(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return ""
	}
	i := rand.Intn(len(r))
	return string(append(r[:i], r[i+1:]...))
}

func duplicateOne(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return ""
	}
	i := rand.Intn(len(r))
	r = append(r[:i+1], r[i:]...)
	return string(r)
}

func replaceOne(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return ""
	}
	i := rand.Intn(len(r))
	// replace with nearby character: a simple shift
	r[i] = r[i] + 1
	return string(r)
}

func (m SpellingModel) refreshMCOptionsForCurrentPrompt() SpellingModel {
	if !m.isMultipleChoice || m.currentQuestion >= len(m.prompts) {
		m.mcOptions = nil
		return m
	}
	m.mcOptions = generateMCOptions(m.prompts[m.currentQuestion].CorrectSpelling)
	return m
}

// isNear returns true for small edit differences (Levenshtein <=1)
func isNear(a, b string) bool {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	// quick length check
	if abs(len(a)-len(b)) > 1 {
		return false
	}
	// compute simple Levenshtein up to 1
	if levenshtein(a, b) <= 1 {
		return true
	}
	return false
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func levenshtein(a, b string) int {
	// classic dynamic programming
	la := len(a)
	lb := len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	dp := make([][]int, la+1)
	for i := range dp {
		dp[i] = make([]int, lb+1)
	}
	for i := 0; i <= la; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		dp[0][j] = j
	}
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			dp[i][j] = min(dp[i-1][j]+1, dp[i][j-1]+1, dp[i-1][j-1]+cost)
		}
	}
	return dp[la][lb]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
