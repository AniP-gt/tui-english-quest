package ui

import (
	"context"
	"time"

	"tui-english-quest/internal/db"
	"tui-english-quest/internal/game"
)

// SpellingOutcome indicates result quality.
type SpellingOutcome string

const (
	SpellingPerfect SpellingOutcome = "perfect"
	SpellingNear    SpellingOutcome = "near"
	SpellingFail    SpellingOutcome = "fail"
)

// RunSpellingSession applies spelling challenge rules for 5 prompts.
func RunSpellingSession(ctx context.Context, stats game.Stats, outcomes []SpellingOutcome) (game.Stats, SessionSummary, error) {
	startedAt := time.Now()
	summary := SessionSummary{Mode: "spelling"}
	before := stats
	expDelta := 0
	hpDelta := 0
	for _, o := range outcomes {
		switch o {
		case SpellingPerfect:
			expDelta += 5
			summary.Correct++
		case SpellingNear:
			expDelta += 2
			hpDelta -= 5
		case SpellingFail:
			expDelta += 1
			hpDelta -= 12
		default:
			expDelta += 1
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

	endedAt := time.Now()
	rec := db.NewSessionRecord("spelling", startedAt, endedAt)
	rec.CorrectCount = summary.Correct
	rec.ExpGained = summary.ExpDelta
	rec.HPDelta = summary.HPDelta
	rec.Fainted = summary.Fainted
	rec.LeveledUp = summary.LeveledUp
	_ = db.SaveSession(ctx, rec)
	return stats, summary, nil
}
