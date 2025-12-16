package game

import "math"

// ApplyDamage reduces HP by given value and clamps at zero.
func ApplyDamage(s Stats, dmg int) Stats {
	effectiveDmg := int(math.Round(float64(dmg) * (1 - s.DamageReduction)))
	if effectiveDmg < 0 {
		effectiveDmg = 0
	}
	s.HP -= effectiveDmg
	if s.HP < 0 {
		s.HP = 0
	}
	return s
}

// Fainted reports whether HP has reached zero or below.
func Fainted(s Stats) bool {
	return s.HP <= 0
}

// ApplyFaintPenalty applies faint penalty and restores HP to 50%.
func ApplyFaintPenalty(s Stats) Stats {
	s.Exp -= 5
	if s.Exp < 0 {
		s.Exp = 0
	}
	s.HP = s.MaxHP / 2
	return s
}

// ResetCombo clears combo counter on failure.
func ResetCombo(s Stats) Stats {
	s.Combo = 0
	return s
}

// AddCombo increments combo counter.
func AddCombo(s Stats) Stats {
	s.Combo++
	return s
}
