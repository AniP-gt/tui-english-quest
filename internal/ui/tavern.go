package ui

import (
	"context"

	"tui-english-quest/internal/db"
	"tui-english-quest/internal/game"
)

// TavernOutcome represents per-turn evaluation.
type TavernOutcome string

const (
	OutcomeSuccess TavernOutcome = "success"
	OutcomeNormal  TavernOutcome = "normal"
	OutcomeFail    TavernOutcome = "fail"
)

// RunTavernSession applies conversation tavern rules for 5 turns.
func RunTavernSession(ctx context.Context, stats game.Stats, outcomes []TavernOutcome) (game.Stats, SessionSummary, error) {
	summary := SessionSummary{Mode: "tavern"}
	before := stats
	expDelta := 0
	goldDelta := 0
	for _, o := range outcomes {
		switch o {
		case OutcomeSuccess:
			expDelta += 5
			goldDelta += 10
			summary.Correct++
		case OutcomeNormal:
			expDelta += 3
			goldDelta += 5
		case OutcomeFail:
			expDelta += 1
		default:
			expDelta += 1
		}
	}
	stats = game.GainExp(stats, expDelta)
	stats = game.AddGold(stats, goldDelta)
	stats, fainted := applyFaint(stats)
	summary.ExpDelta = expDelta
	summary.GoldDelta = goldDelta
	summary.Fainted = fainted
	summary.LeveledUp = leveledUp(before, stats)

	rec := db.SessionRecord{
		Mode:         "tavern",
		CorrectCount: summary.Correct,
		ExpGained:    expDelta,
		GoldDelta:    goldDelta,
		Fainted:      fainted,
		LeveledUp:    summary.LeveledUp,
	}
	_ = db.SaveSession(ctx, rec)
	return stats, summary, nil
}
