package ui

import (
	"context"

	"tui-english-quest/internal/db"
	"tui-english-quest/internal/game"
)

// GrammarAnswer represents correctness per floor.
type GrammarAnswer struct {
	Correct bool
}

// RunGrammarSession applies grammar dungeon rules for 5 floors.
func RunGrammarSession(ctx context.Context, stats game.Stats, answers []GrammarAnswer) (game.Stats, SessionSummary, error) {
	summary := SessionSummary{Mode: "grammar"}
	before := stats
	expDelta := 0
	hpDelta := 0
	defDelta := 0.0
	correct := 0
	for _, a := range answers {
		if a.Correct {
			expDelta += 3
			defDelta += 0.2
			correct++
		} else {
			hpDelta -= 6
		}
	}
	stats = game.GainExp(stats, expDelta)
	stats = game.AddDefense(stats, defDelta)
	stats.HP += hpDelta
	if stats.HP < 0 {
		stats.HP = 0
	}
	stats, fainted := applyFaint(stats)
	summary.Correct = correct
	summary.ExpDelta = expDelta
	summary.HPDelta = hpDelta
	summary.DefenseDelta = defDelta
	summary.Fainted = fainted
	summary.LeveledUp = leveledUp(before, stats)

	rec := db.SessionRecord{
		Mode:         "grammar",
		CorrectCount: correct,
		ExpGained:    expDelta,
		HPDelta:      hpDelta,
		DefenseDelta: defDelta,
		Fainted:      fainted,
		LeveledUp:    summary.LeveledUp,
	}
	_ = db.SaveSession(ctx, rec)
	return stats, summary, nil
}
