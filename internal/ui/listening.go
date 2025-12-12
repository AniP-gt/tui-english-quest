package ui

import (
	"context"

	"tui-english-quest/internal/db"
	"tui-english-quest/internal/game"
)

// ListeningAnswer represents correctness.
type ListeningAnswer struct {
	Correct bool
}

// RunListeningSession applies listening rules for 5 questions.
func RunListeningSession(ctx context.Context, stats game.Stats, answers []ListeningAnswer) (game.Stats, SessionSummary, error) {
	// Simulate device check
	if !isAudioDeviceAvailable() {
		return stats, SessionSummary{Mode: "listening", Note: "Audio device not available. Skipping."}, nil
	}

	summary := SessionSummary{Mode: "listening"}
	before := stats
	expDelta := 0
	hpDelta := 0
	for _, a := range answers {
		if a.Correct {
			expDelta += 4
			summary.Correct++
		} else {
			expDelta += 2
			hpDelta -= 6
		}
	}
	stats = game.GainExp(stats, expDelta)
	stats.HP += hpDelta
	if stats.HP < 0 {
		stats.HP = 0
	}
	stats, fainted := applyFaint(stats)
	summary.ExpDelta = expDelta
	summary.HPDelta = hpDelta
	summary.Fainted = fainted
	summary.LeveledUp = leveledUp(before, stats)

	rec := db.SessionRecord{
		Mode:         "listening",
		CorrectCount: summary.Correct,
		ExpGained:    expDelta,
		HPDelta:      hpDelta,
		Fainted:      fainted,
		LeveledUp:    summary.LeveledUp,
	}
	_ = db.SaveSession(ctx, rec)
	return stats, summary, nil
}

// isAudioDeviceAvailable simulates checking for an audio device.
// In a real application, this would involve platform-specific checks.
func isAudioDeviceAvailable() bool {
	// For now, always return true.
	return true
}
