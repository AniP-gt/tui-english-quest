package ui

import (
	"context"

	"tui-english-quest/internal/db"
	"tui-english-quest/internal/game"
)

// VocabAnswer represents correctness per question.
type VocabAnswer struct {
	Correct bool
}

// RunVocabSession applies vocabulary battle rules for 5 questions.
func RunVocabSession(ctx context.Context, stats game.Stats, answers []VocabAnswer) (game.Stats, SessionSummary, error) {
	summary := SessionSummary{Mode: "vocab"}
	before := stats
	combo := stats.Combo
	bestCombo := combo
	expDelta := 0
	hpDelta := 0
	for _, a := range answers {
		if a.Correct {
			combo = game.AddCombo(game.Stats{Combo: combo}).Combo
			expDelta += 4
			if combo > bestCombo {
				bestCombo = combo
			}
		} else {
			combo = game.ResetCombo(game.Stats{Combo: combo}).Combo
			hpDelta -= 10
		}
	}
	stats.Combo = combo
	stats = game.GainExp(stats, expDelta)
	stats.HP += hpDelta
	if stats.HP < 0 {
		stats.HP = 0
	}
	stats, fainted := applyFaint(stats)
	summary.Correct = countVocabCorrect(answers)
	summary.ExpDelta = expDelta
	summary.HPDelta = hpDelta
	summary.BestCombo = bestCombo
	summary.Fainted = fainted
	summary.LeveledUp = leveledUp(before, stats)

	rec := db.SessionRecord{
		Mode:         "vocab",
		CorrectCount: summary.Correct,
		BestCombo:    bestCombo,
		ExpGained:    expDelta,
		HPDelta:      hpDelta,
		Fainted:      fainted,
		LeveledUp:    summary.LeveledUp,
	}
	_ = db.SaveSession(ctx, rec)
	return stats, summary, nil
}

func countVocabCorrect(ans []VocabAnswer) int {
	c := 0
	for _, a := range ans {
		if a.Correct {
			c++
		}
	}
	return c
}
