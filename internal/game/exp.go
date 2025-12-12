package game

// GainExp applies experience gain and returns updated stats.
func GainExp(s Stats, gained int) Stats {
	effectiveGain := int(float64(gained) * (1 + s.ExpBoost))
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
	// Simplified thresholds
	s.Next = nextThreshold(s.Level)
	s.MaxHP += 10
	s.HP = s.MaxHP
	s.Attack += 2
	s.Defense += 1
	return s
}

func nextThreshold(level int) int {
	switch level {
	case 1:
		return 30
	case 2:
		return 50
	case 3:
		return 80
	case 4:
		return 120
	default:
		return 150
	}
}
