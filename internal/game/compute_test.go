package game

import "testing"

func TestMaxHPForLevel(t *testing.T) {
	if v := MaxHPForLevel(10); v < 110 || v > 118 {
		t.Fatalf("unexpected MaxHP for Lv10: %d", v)
	}
	if v := MaxHPForLevel(60); v < 320 || v > 325 {
		t.Fatalf("unexpected MaxHP for Lv60: %d", v)
	}
	if v := MaxHPForLevel(250); v < 455 || v > 465 {
		t.Fatalf("unexpected MaxHP for Lv250: %d", v)
	}
}

func TestAllowedMissesAndDamage(t *testing.T) {
	M := AllowedMisses(5)
	if M != 2 {
		t.Fatalf("expected M=2 for N=5, got %d", M)
	}
	maxHP := MaxHPForLevel(10)
	d := DamagePerMiss(maxHP, M)
	if d <= 0 {
		t.Fatalf("damage must be positive, got %d", d)
	}
}

func TestTierAndQExp(t *testing.T) {
	_, mul := TierForLevel(60)
	if mul != 1.5 {
		t.Fatalf("expected tier mul 1.5 for Lv60, got %v", mul)
	}
	q := QExpFor(4, mul, false)
	if q != 6 {
		t.Fatalf("expected QExp 6 for base 4 tier 1.5, got %d", q)
	}
	qrare := QExpFor(4, mul, true)
	if qrare != 12 {
		t.Fatalf("expected QExp 12 for rare, got %d", qrare)
	}
}

func TestExpToNext(t *testing.T) {
	if v := ExpToNext(1); v != 30 {
		t.Fatalf("ExpToNext(1) expected 30 got %d", v)
	}
	if v := ExpToNext(100); v != 500 {
		t.Fatalf("ExpToNext(100) expected 500 got %d", v)
	}
}
