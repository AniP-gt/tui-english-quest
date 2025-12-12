package game

// Stats represents player status for display and calculations.
type Stats struct {
	Name            string
	Class           string
	Level           int
	Exp             int
	Next            int
	HP              int
	MaxHP           int
	Attack          int
	Defense         float64
	Combo           int
	Streak          int
	Gold            int
	ExpBoost        float64
	DamageReduction float64
}

// DefaultStats returns initial status for a new game.
func DefaultStats() Stats {
	return Stats{
		Name:            "Takuya",
		Class:           "Vocabulary Warrior",
		Level:           1,
		Exp:             0,
		Next:            30,
		HP:              100,
		MaxHP:           100,
		Attack:          10,
		Defense:         0,
		Combo:           0,
		Streak:          0,
		Gold:            0,
		ExpBoost:        0,
		DamageReduction: 0,
	}
}

// AddDefense increments defense.
func AddDefense(s Stats, delta float64) Stats {
	s.Defense += delta
	return s
}

// AddGold adjusts gold but not below zero.
func AddGold(s Stats, delta int) Stats {
	s.Gold += delta
	if s.Gold < 0 {
		s.Gold = 0
	}
	return s
}
