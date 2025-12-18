package services

import (
	"context"
	"fmt"
	"sort"

	"tui-english-quest/internal/db"
	"tui-english-quest/internal/game"
)

const (
	questionsPerSession    = 5
	recentWindowSessions   = 5
	previousWindowSessions = 5
)

// ModeInsight holds per-mode performance metrics.
type ModeInsight struct {
	Mode        string
	Accuracy    float64
	Sessions    int
	Trend       float64
	Description string
}

// ActionSuggestion describes a readable next step for the player.
type ActionSuggestion struct {
	Mode        string
	Title       string
	Description string
	Priority    string
}

// WeaknessReport represents analyzed weak points and recommendations.
type WeaknessReport struct {
	WeakPoints     []ModeInsight
	StrengthPoints []ModeInsight
	Recommendation string
	Summary        string
	ActionPlan     []ActionSuggestion
}

type modeAccum struct {
	Mode          string
	Sessions      int
	Total         int
	Correct       int
	RecentTotal   int
	RecentCorrect int
	PrevTotal     int
	PrevCorrect   int
}

// AnalyzeWeakness analyzes player history and generates a weakness report.
func AnalyzeWeakness(ctx context.Context, gc *GeminiClient, playerID string, stats game.Stats, historyLimit int) (WeaknessReport, error) {
	sessions, err := db.ListSessions(ctx, playerID, historyLimit)
	if err != nil {
		return WeaknessReport{
			Recommendation: fmt.Sprintf("Failed to analyze history: %v", err),
			Summary:        "",
		}, err
	}

	if len(sessions) == 0 {
		return WeaknessReport{
			WeakPoints:     nil,
			StrengthPoints: nil,
			Recommendation: "Play some sessions to get an AI analysis!",
			Summary:        "No sessions available yet.",
		}, nil
	}

	accum := map[string]*modeAccum{}
	totalCorrect, totalQuestions := 0, 0
	recentEnd := min(len(sessions), recentWindowSessions)
	prevEnd := min(len(sessions), recentWindowSessions+previousWindowSessions)

	for i, session := range sessions {
		totalCorrect += session.CorrectCount
		totalQuestions += questionsPerSession
		ma := accum[session.Mode]
		if ma == nil {
			ma = &modeAccum{Mode: session.Mode}
			accum[session.Mode] = ma
		}
		ma.Sessions++
		ma.Total += questionsPerSession
		ma.Correct += session.CorrectCount
		if i < recentEnd {
			ma.RecentTotal += questionsPerSession
			ma.RecentCorrect += session.CorrectCount
		} else if i < prevEnd {
			ma.PrevTotal += questionsPerSession
			ma.PrevCorrect += session.CorrectCount
		}
	}

	insights := make([]ModeInsight, 0, len(accum))
	for _, values := range accum {
		insights = append(insights, buildModeInsight(values))
	}

	sort.Slice(insights, func(i, j int) bool {
		return insights[i].Accuracy < insights[j].Accuracy
	})

	weakPoints, strengthPoints, recommendation := buildWeakAndStrong(insights)
	summary := buildSummary(len(sessions), totalCorrect, totalQuestions)
	actionPlan := buildActionPlan(stats, weakPoints, strengthPoints)

	return WeaknessReport{
		WeakPoints:     weakPoints,
		StrengthPoints: strengthPoints,
		Recommendation: recommendation,
		Summary:        summary,
		ActionPlan:     actionPlan,
	}, nil
}

func buildModeInsight(acc *modeAccum) ModeInsight {
	accuracy := 0.0
	if acc.Total > 0 {
		accuracy = float64(acc.Correct) / float64(acc.Total)
	}
	desc := fmt.Sprintf("%d sessions, %.0f%% accuracy", acc.Sessions, accuracy*100)
	if acc.RecentTotal > 0 || acc.PrevTotal > 0 {
		recentAvg := 0.0
		prevAvg := 0.0
		if acc.RecentTotal > 0 {
			recentAvg = float64(acc.RecentCorrect) / float64(acc.RecentTotal)
		}
		if acc.PrevTotal > 0 {
			prevAvg = float64(acc.PrevCorrect) / float64(acc.PrevTotal)
		}
		desc = fmt.Sprintf("Recent %.0f%% vs prior %.0f%%", recentAvg*100, prevAvg*100)
	}
	return ModeInsight{
		Mode:        acc.Mode,
		Accuracy:    accuracy,
		Sessions:    acc.Sessions,
		Trend:       calculateTrend(acc),
		Description: desc,
	}
}

func buildWeakAndStrong(insights []ModeInsight) (weak []ModeInsight, strong []ModeInsight, recommendation string) {
	for _, insight := range insights {
		if insight.Accuracy < 0.75 && len(weak) < 2 {
			weak = append(weak, insight)
		}
	}
	for i := len(insights) - 1; i >= 0 && len(strong) < 2; i-- {
		if insights[i].Accuracy > 0.85 {
			strong = append(strong, insights[i])
		}
	}

	switch {
	case len(weak) > 0:
		recommendation = fmt.Sprintf("Focus on %s. Try playing %s sessions.", weak[0].Mode, weak[0].Mode)
	case len(strong) > 0:
		recommendation = "Great job! You're strong in all areas."
	default:
		recommendation = "No clear patterns yet. Keep playing!"
	}
	return
}

func buildSummary(sessionCount, correct, questions int) string {
	if sessionCount == 0 || questions == 0 {
		return "No data to summarize yet."
	}
	overall := float64(correct) / float64(questions) * 100
	return fmt.Sprintf("Analyzed %d sessions (%d questions) with %.0f%% accuracy overall.", sessionCount, questions, overall)
}

func buildActionPlan(stats game.Stats, weak []ModeInsight, strong []ModeInsight) []ActionSuggestion {
	var plan []ActionSuggestion
	if stats.MaxHP > 0 && stats.HP < stats.MaxHP/2 {
		plan = append(plan, ActionSuggestion{
			Title:       "Recover HP",
			Description: fmt.Sprintf("HP is %d/%d. Run a lighter mode to rebuild HP before tackling harder fights.", stats.HP, stats.MaxHP),
			Priority:    "high",
		})
	}
	if len(weak) > 0 {
		entry := weak[0]
		plan = append(plan, ActionSuggestion{
			Mode:        entry.Mode,
			Title:       fmt.Sprintf("Focus on %s", entry.Mode),
			Description: fmt.Sprintf("Accuracy %.0f%%. Spend two sessions reviewing %s mode mistakes.", entry.Accuracy*100, entry.Mode),
			Priority:    "high",
		})
	} else {
		plan = append(plan, ActionSuggestion{
			Title:       "Keep the pace",
			Description: "No pronounced weak points—rotate through high-accuracy modes to maintain streaks.",
			Priority:    "medium",
		})
	}
	if stats.Streak >= 3 {
		plan = append(plan, ActionSuggestion{
			Title:       "Protect streak",
			Description: fmt.Sprintf("Streak %d days. Pick quick, high-accuracy runs to lock it in.", stats.Streak),
			Priority:    "medium",
		})
	}
	if len(strong) > 0 {
		top := strong[0]
		plan = append(plan, ActionSuggestion{
			Mode:        top.Mode,
			Title:       fmt.Sprintf("Use %s for bonus EXP", top.Mode),
			Description: fmt.Sprintf("You're strong in %s. Lean on it for a confident run.", top.Mode),
			Priority:    "low",
		})
	}
	return plan
}

func calculateTrend(acc *modeAccum) float64 {
	if acc.RecentTotal == 0 || acc.PrevTotal == 0 {
		return 0
	}
	recentAvg := float64(acc.RecentCorrect) / float64(acc.RecentTotal)
	prevAvg := float64(acc.PrevCorrect) / float64(acc.PrevTotal)
	return recentAvg - prevAvg
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
