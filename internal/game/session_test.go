package game

import (
	"context"
	"testing"
)

func TestRunVocabSession_AllCorrect(t *testing.T) {
	stats := DefaultStats()
	stats.Level = 10
	stats.MaxHP = MaxHPForLevel(stats.Level)
	stats.HP = stats.MaxHP
	answers := []VocabAnswer{{true}, {true}, {true}, {true}, {true}}
	updated, summary, err := RunVocabSession(context.Background(), stats, answers)
	if err != nil {
		t.Fatalf("RunVocabSession error: %v", err)
	}
	if summary.Fainted {
		t.Fatalf("expected not fainted on all-correct, got fainted")
	}
	if summary.ExpDelta <= 0 {
		t.Fatalf("expected positive ExpDelta for all-correct, got %d", summary.ExpDelta)
	}
	if updated.Exp == stats.Exp {
		t.Fatalf("expected Exp to increase after session, got same value")
	}
}

func TestRunVocabSession_IncorrectReducesHP(t *testing.T) {
	stats := DefaultStats()
	stats.Level = 10
	stats.MaxHP = MaxHPForLevel(stats.Level)
	stats.HP = stats.MaxHP
	// one incorrect at first
	answers := []VocabAnswer{{false}, {true}, {true}, {true}, {true}}
	updated, summary, err := RunVocabSession(context.Background(), stats, answers)
	if err != nil {
		t.Fatalf("RunVocabSession error: %v", err)
	}
	if summary.Fainted {
		t.Fatalf("did not expect faint on single incorrect")
	}
	if updated.HP >= stats.HP {
		t.Fatalf("expected HP to decrease after incorrect, before: %d, after: %d", stats.HP, updated.HP)
	}
}

func TestRunVocabSession_FailOnTooManyMisses(t *testing.T) {
	stats := DefaultStats()
	stats.Level = 10
	stats.MaxHP = MaxHPForLevel(stats.Level)
	stats.HP = stats.MaxHP
	N := 5
	M := AllowedMisses(N)
	// build answers with M+1 incorrect to ensure death
	answers := make([]VocabAnswer, N)
	for i := 0; i < N; i++ {
		if i <= M { // make first M+1 incorrect
			answers[i] = VocabAnswer{Correct: false}
		} else {
			answers[i] = VocabAnswer{Correct: true}
		}
	}
	updated, summary, err := RunVocabSession(context.Background(), stats, answers)
	if err != nil {
		t.Fatalf("RunVocabSession error: %v", err)
	}
	if !summary.Fainted {
		t.Fatalf("expected fainted true when M+1 misses, got false")
	}
	if updated.HP != stats.MaxHP/2 {
		t.Fatalf("expected updated HP to be half MaxHP after fainting, got %d", updated.HP)
	}
}
