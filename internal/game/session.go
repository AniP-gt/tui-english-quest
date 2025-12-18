package game

import (
	"context"
	"log"

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
	// Use spec defaults: BaseExp for vocab = 4
	baseExp := 4
	// Ensure MaxHP is in sync with level
	stats.MaxHP = MaxHPForLevel(stats.Level)
	// Compute allowed misses and damage per miss
	N := len(answers)
	M := AllowedMisses(N)
	dmg := DamagePerMiss(stats.MaxHP, M)

	sumCorrectExp := 0
	hpDelta := 0
	fainted := false
	for i, a := range answers {
		if a.Correct {
			// increment combo
			tmp := Stats{Combo: combo}
			combo = AddCombo(tmp).Combo
			_, tierMul := TierForLevel(stats.Level)
			qexp := QExpFor(baseExp, tierMul, false)
			sumCorrectExp += qexp
			if combo > bestCombo {
				bestCombo = combo
			}
		} else {
			// apply damage immediately
			tmp := Stats{Combo: combo}
			combo = ResetCombo(tmp).Combo
			hpDelta -= dmg
			stats.HP -= dmg
			if stats.HP <= 0 {
				stats.HP = 0
				fainted = true
				// stop processing further questions
				// truncate answers considered to i+1
				answers = answers[:i+1]
				break
			}
		}
	}
	stats.Combo = combo

	// Settlement
	var sessionExp int
	if !fainted && len(answers) == N {
		// clear
		_, tierMul := TierForLevel(stats.Level)
		clearBonus := ClearBonus(N, baseExp, tierMul)
		allCorrect := countVocabCorrect(answers) == N
		sessionExp = SessionExpClear(sumCorrectExp, clearBonus, allCorrect, N, true)
		stats = GainExp(stats, sessionExp)
	} else {
		// fail
		sessionExp = SessionExpFail(sumCorrectExp, 0.40)
		stats = GainExp(stats, sessionExp)
		if fainted {
			stats = ApplyFaintPenalty(stats)
		}
	}

	summary.Correct = countVocabCorrect(answers)
	summary.ExpDelta = sessionExp
	summary.HPDelta = hpDelta
	summary.BestCombo = bestCombo
	summary.Fainted = fainted
	summary.LeveledUp = LeveledUp(before, stats)

	rec := db.SessionRecord{
		PlayerID:     db.CurrentProfileID(),
		Mode:         "vocab",
		CorrectCount: summary.Correct,
		BestCombo:    bestCombo,
		ExpGained:    sessionExp,
		HPDelta:      hpDelta,
		Fainted:      fainted,
		LeveledUp:    summary.LeveledUp,
	}
	_ = db.SaveSession(ctx, rec) // Assuming SaveSession is in db package
	if err := SaveStats(ctx, stats); err != nil {
		log.Printf("failed to persist profile: %v", err)
	}
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
	baseExp := 3
	// Ensure MaxHP is in sync with level
	stats.MaxHP = MaxHPForLevel(stats.Level)
	N := len(answers)
	M := AllowedMisses(N)
	dmg := DamagePerMiss(stats.MaxHP, M)

	sumCorrectExp := 0
	hpDelta := 0
	defDelta := 0.0
	correct := 0
	fainted := false
	for i, a := range answers {
		if a.Correct {
			_, tierMul := TierForLevel(stats.Level)
			qexp := QExpFor(baseExp, tierMul, false)
			sumCorrectExp += qexp
			defDelta += 0.2
			correct++
		} else {
			hpDelta -= dmg
			stats.HP -= dmg
			if stats.HP <= 0 {
				stats.HP = 0
				fainted = true
				answers = answers[:i+1]
				break
			}
		}
	}

	stats = AddDefense(stats, defDelta)

	var sessionExp int
	if !fainted && len(answers) == N {
		// clear
		_, tierMul := TierForLevel(stats.Level)
		clearBonus := ClearBonus(N, baseExp, tierMul)
		allCorrect := correct == N
		sessionExp = SessionExpClear(sumCorrectExp, clearBonus, allCorrect, N, true)
		stats = GainExp(stats, sessionExp)
	} else {
		// fail
		sessionExp = SessionExpFail(sumCorrectExp, 0.40)
		stats = GainExp(stats, sessionExp)
		if fainted {
			stats = ApplyFaintPenalty(stats)
		}
	}

	summary.Correct = correct
	summary.ExpDelta = sessionExp
	summary.HPDelta = hpDelta
	summary.DefenseDelta = defDelta
	summary.Fainted = fainted
	summary.LeveledUp = LeveledUp(before, stats)

	rec := db.SessionRecord{
		PlayerID:     db.CurrentProfileID(),
		Mode:         "grammar",
		CorrectCount: summary.Correct,
		ExpGained:    sessionExp,
		HPDelta:      hpDelta,
		DefenseDelta: defDelta, // Assuming DefenseDelta is added to SessionRecord
		Fainted:      fainted,
		LeveledUp:    summary.LeveledUp,
	}
	_ = db.SaveSession(ctx, rec)
	if err := SaveStats(ctx, stats); err != nil {
		log.Printf("failed to persist profile: %v", err)
	}
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
			var delta int
			stats, delta = applyDamageDelta(stats, 5)
			hpDelta += delta
		case SpellingFail:
			expDelta += 1
			var delta int
			stats, delta = applyDamageDelta(stats, 12)
			hpDelta += delta
		default:
			expDelta += 1
		}
	}

	stats = GainExp(stats, expDelta)
	stats, fainted := applyFaintIfNeeded(stats)

	summary.ExpDelta = expDelta
	summary.HPDelta = hpDelta
	summary.Fainted = fainted
	summary.LeveledUp = LeveledUp(before, stats)

	rec := db.SessionRecord{
		PlayerID:     db.CurrentProfileID(),
		Mode:         "spelling",
		CorrectCount: summary.Correct,
		ExpGained:    expDelta,
		HPDelta:      hpDelta,
		Fainted:      fainted,
		LeveledUp:    summary.LeveledUp,
	}
	_ = db.SaveSession(ctx, rec)
	if err := SaveStats(ctx, stats); err != nil {
		log.Printf("failed to persist profile: %v", err)
	}
	return stats, summary, nil
}

func applyDamageDelta(s Stats, dmg int) (Stats, int) {
	prev := s.HP
	s = ApplyDamage(s, dmg)
	return s, s.HP - prev
}

func applyFaintIfNeeded(s Stats) (Stats, bool) {
	if Fainted(s) {
		s = ApplyFaintPenalty(s)
		return s, true
	}
	return s, false
}

type ListeningAnswer struct{ Correct bool }

func RunListeningSession(ctx context.Context, stats Stats, answers []ListeningAnswer) (Stats, SessionSummary, error) {
	summary := SessionSummary{Mode: "listening"}
	before := stats
	baseExp := 5
	// Ensure MaxHP is in sync with level
	stats.MaxHP = MaxHPForLevel(stats.Level)
	N := len(answers)
	M := AllowedMisses(N)
	dmg := DamagePerMiss(stats.MaxHP, M)

	sumCorrectExp := 0
	hpDelta := 0
	correct := 0
	fainted := false
	for i, a := range answers {
		if a.Correct {
			_, tierMul := TierForLevel(stats.Level)
			qexp := QExpFor(baseExp, tierMul, false)
			sumCorrectExp += qexp
			correct++
		} else {
			hpDelta -= dmg
			stats.HP -= dmg
			if stats.HP <= 0 {
				stats.HP = 0
				fainted = true
				answers = answers[:i+1]
				break
			}
		}
	}

	var sessionExp int
	if !fainted && len(answers) == N {
		_, tierMul := TierForLevel(stats.Level)
		clearBonus := ClearBonus(N, baseExp, tierMul)
		allCorrect := correct == N
		sessionExp = SessionExpClear(sumCorrectExp, clearBonus, allCorrect, N, true)
		stats = GainExp(stats, sessionExp)
	} else {
		sessionExp = SessionExpFail(sumCorrectExp, 0.40)
		stats = GainExp(stats, sessionExp)
		if fainted {
			stats = ApplyFaintPenalty(stats)
		}
	}

	summary.Correct = correct
	summary.ExpDelta = sessionExp
	summary.HPDelta = hpDelta
	summary.Fainted = fainted
	summary.LeveledUp = LeveledUp(before, stats)

	rec := db.SessionRecord{
		PlayerID:     db.CurrentProfileID(),
		Mode:         "listening",
		CorrectCount: summary.Correct,
		ExpGained:    sessionExp,
		HPDelta:      hpDelta,
		Fainted:      fainted,
		LeveledUp:    summary.LeveledUp,
	}
	_ = db.SaveSession(ctx, rec)
	if err := SaveStats(ctx, stats); err != nil {
		log.Printf("failed to persist profile: %v", err)
	}
	return stats, summary, nil
}
