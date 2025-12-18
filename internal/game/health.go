package game

// FullHeal synchronizes MaxHP to the current level and restores HP to the maximum value.
func FullHeal(s Stats) Stats {
	s.MaxHP = MaxHPForLevel(s.Level)
	s.HP = s.MaxHP
	return s
}
