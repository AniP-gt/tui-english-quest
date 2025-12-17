package game

import "math"

// GainExp applies experience gain and returns updated stats.
func GainExp(s Stats, gained int) Stats {
	effectiveGain := int(math.Round(float64(gained) * (1 + s.ExpBoost)))
	s.Exp += effectiveGain
	for s.Exp >= s.Next {
		s.Exp -= s.Next
		s = LevelUp(s)
	}
	return s
}

// LevelUp increments level and updates thresholds.
func LevelUp(s Stats) Stats {
	s.Level++
	// Use ExpToNext formula from design
	s.Next = ExpToNext(s.Level)
	// Recompute MaxHP from level and fully heal
	s.MaxHP = MaxHPForLevel(s.Level)
	// Ensure HP is capped to MaxHP and fully heal
	s.HP = s.MaxHP
	// Keep existing attack/defense growth for now (attack +2, defense +1)
	s.Attack += 2
	s.Defense += 1
	return s
}

func nextThreshold(level int) int {
	return ExpToNext(level)
}
