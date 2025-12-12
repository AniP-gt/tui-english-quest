package services

import (
	"context"
	"fmt"

	"tui-english-quest/internal/db"
)

// WeaknessReport represents analyzed weak points.
type WeaknessReport struct {
	WeakPoints     []string
	StrengthPoints []string
	Recommendation string
}

// AnalyzeWeakness analyzes player history and generates a weakness report.
func AnalyzeWeakness(ctx context.Context, playerID string, historyLimit int) WeaknessReport {
	// For now, use a simple analysis based on mode and correct count.
	// In a real implementation, this would involve more sophisticated logic
	// to identify specific words/grammar points.

	sessions, err := db.ListSessions(ctx, playerID, historyLimit)
	if err != nil {
		return WeaknessReport{
			WeakPoints:     []string{"analysis error"},
			Recommendation: fmt.Sprintf("Failed to analyze history: %v", err),
		}
	}

	if len(sessions) == 0 {
		return WeaknessReport{
			WeakPoints:     []string{"no data"},
			Recommendation: "Play some sessions to get an AI analysis!",
		}
	}

	modeScores := make(map[string]struct {
		Correct int
		Total   int
	})

	for _, s := range sessions {
		score := modeScores[s.Mode]
		score.Correct += s.CorrectCount
		score.Total += 5 // Assuming 5 questions per session
		modeScores[s.Mode] = score
	}

	weakPoints := []string{}
	strengthPoints := []string{}
	recommendation := "Keep up the good work!"

	// Simple logic: mode with lowest correct percentage is a weak point
	minScore := 1.0
	maxScore := 0.0
	weakMode := ""
	strongMode := ""

	for mode, score := range modeScores {
		if score.Total == 0 {
			continue
		}
		percentage := float64(score.Correct) / float64(score.Total)
		if percentage < minScore {
			minScore = percentage
			weakMode = mode
		}
		if percentage > maxScore {
			maxScore = percentage
			strongMode = mode
		}
	}

	if weakMode != "" && minScore < 0.7 { // Threshold for weakness
		weakPoints = append(weakPoints, weakMode)
		recommendation = fmt.Sprintf("Focus on %s. Try playing %s sessions.", weakMode, weakMode)
	}
	if strongMode != "" && maxScore > 0.8 { // Threshold for strength
		strengthPoints = append(strengthPoints, strongMode)
	}

	if len(weakPoints) == 0 && len(strengthPoints) == 0 {
		recommendation = "No clear patterns yet. Keep playing!"
	} else if len(weakPoints) == 0 {
		recommendation = "Great job! You're strong in all areas."
	}

	return WeaknessReport{
		WeakPoints:     weakPoints,
		StrengthPoints: strengthPoints,
		Recommendation: recommendation,
	}
}
