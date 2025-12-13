package game

import (
	"context"

	"tui-english-quest/internal/db"
)

// VocabAnswer represents correctness per question.
type VocabAnswer struct {
	Correct bool
}

// GrammarAnswer represents correctness per floor.
type GrammarAnswer struct {
	Correct bool
}

// SessionSummary summarizes a game session.
type SessionSummary struct {
	Mode         string
	Correct      int
	ExpDelta     int
	HPDelta      int
	GoldDelta    int
	BestCombo    int
	Fainted      bool
	LeveledUp    bool
	Note         string  // For errors or special messages
	DefenseDelta float64 // Added for Grammar Dungeon
}

// ApplyFaint checks if the player has fainted and applies penalties.
func ApplyFaint(s Stats) (Stats, bool) {
	if Fainted(s) { // Assuming Fainted is in game package
		s = ApplyFaintPenalty(s) // Assuming ApplyFaintPenalty is in game package
		return s, true
	}
	return s, false
}

// LeveledUp checks if the player leveled up during the session.
func LeveledUp(before Stats, after Stats) bool {
	return after.Level > before.Level
}

// RunVocabSession applies vocabulary battle rules for 5 questions.
func RunVocabSession(ctx context.Context, stats Stats, answers []VocabAnswer) (Stats, SessionSummary, error) {
	summary := SessionSummary{Mode: "vocab"}
	before := stats
	combo := stats.Combo
	bestCombo := combo
	expDelta := 0
	hpDelta := 0
	for _, a := range answers {
		if a.Correct {
			combo = AddCombo(Stats{Combo: combo}).Combo
			expDelta += 4
			if combo > bestCombo {
				bestCombo = combo
			}
		} else {
			combo = ResetCombo(Stats{Combo: combo}).Combo
			hpDelta -= 10
		}
	}
	stats.Combo = combo
	stats = GainExp(stats, expDelta)
	stats.HP += hpDelta
	if stats.HP < 0 {
		stats.HP = 0
	}
	stats, fainted := ApplyFaint(stats) // Now defined in this file
	summary.Correct = countVocabCorrect(answers)
	summary.ExpDelta = expDelta
	summary.HPDelta = hpDelta
	summary.BestCombo = bestCombo
	summary.Fainted = fainted
	summary.LeveledUp = LeveledUp(before, stats) // Now defined in this file

	rec := db.SessionRecord{
		Mode:         "vocab",
		CorrectCount: summary.Correct,
		BestCombo:    bestCombo,
		ExpGained:    expDelta,
		HPDelta:      hpDelta,
		Fainted:      fainted,
		LeveledUp:    summary.LeveledUp,
	}
	_ = db.SaveSession(ctx, rec) // Assuming SaveSession is in db package
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

// RunGrammarSession applies grammar dungeon rules for 5 floors.
func RunGrammarSession(ctx context.Context, stats Stats, answers []GrammarAnswer) (Stats, SessionSummary, error) {
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
	stats = GainExp(stats, expDelta)
	stats = AddDefense(stats, defDelta)
	stats.HP += hpDelta
	if stats.HP < 0 {
		stats.HP = 0
	}
	stats, fainted := ApplyFaint(stats)
	summary.Correct = correct
	summary.ExpDelta = expDelta
	summary.HPDelta = hpDelta
	summary.DefenseDelta = defDelta // Assuming DefenseDelta is added to SessionSummary
	summary.Fainted = fainted
	summary.LeveledUp = LeveledUp(before, stats)

	rec := db.SessionRecord{
		Mode:         "grammar",
		CorrectCount: summary.Correct,
		ExpGained:    expDelta,
		HPDelta:      hpDelta,
		DefenseDelta: defDelta, // Assuming DefenseDelta is added to SessionRecord
		Fainted:      fainted,
		LeveledUp:    summary.LeveledUp,
	}
	_ = db.SaveSession(ctx, rec)
	return stats, summary, nil
}

func countGrammarCorrect(ans []GrammarAnswer) int {
	c := 0
	for _, a := range ans {
		if a.Correct {
			c++
		}
	}
	return c
}

// Dummy functions for other modes to avoid compilation errors for now
type TavernOutcome int

const (
	OutcomeSuccess TavernOutcome = iota
	OutcomeNormal
	OutcomeFail
)

func RunTavernSession(ctx context.Context, stats Stats, outcomes []TavernOutcome) (Stats, SessionSummary, error) {
	return stats, SessionSummary{}, nil
}

type SpellingOutcome int

const (
	SpellingPerfect SpellingOutcome = iota
	SpellingNear
	SpellingFail
)

func RunSpellingSession(ctx context.Context, stats Stats, outcomes []SpellingOutcome) (Stats, SessionSummary, error) {
	return stats, SessionSummary{}, nil
}

type ListeningAnswer struct{ Correct bool }

func RunListeningSession(ctx context.Context, stats Stats, answers []ListeningAnswer) (Stats, SessionSummary, error) {
	return stats, SessionSummary{}, nil
}
