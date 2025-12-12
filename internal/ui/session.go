package ui

import "tui-english-quest/internal/game"

// SessionSummary captures a mode run result.
type SessionSummary struct {
	Mode         string
	Correct      int
	HPDelta      int
	ExpDelta     int
	GoldDelta    int
	DefenseDelta float64
	BestCombo    int
	Fainted      bool
	LeveledUp    bool
	Note         string
}

// applyFaint applies faint penalty and flags fainted.
func applyFaint(s game.Stats) (game.Stats, bool) {
	if game.Fainted(s) {
		s = game.ApplyFaintPenalty(s)
		return s, true
	}
	return s, false
}

// leveledUp reports level increase.
func leveledUp(before, after game.Stats) bool {
	return after.Level > before.Level
}
