package game

import (
	"context"

	"tui-english-quest/internal/db"
)

// SaveStats persists the provided stats using the active profile ID.
func SaveStats(ctx context.Context, stats Stats) error {
	if ctx == nil {
		ctx = context.Background()
	}
	profileID := db.CurrentProfileID()
	if profileID == "" {
		return nil
	}
	rec := profileRecordFromStats(stats)
	rec.ID = profileID
	return db.SaveProfile(ctx, rec)
}

func profileRecordFromStats(stats Stats) db.ProfileRecord {
	return db.ProfileRecord{
		Name:            stats.Name,
		Class:           stats.Class,
		Level:           stats.Level,
		Exp:             stats.Exp,
		NextLevelExp:    stats.Next,
		HP:              stats.HP,
		MaxHP:           stats.MaxHP,
		Attack:          stats.Attack,
		Defense:         stats.Defense,
		Combo:           stats.Combo,
		StreakDays:      stats.Streak,
		Gold:            stats.Gold,
		ExpBoost:        stats.ExpBoost,
		DamageReduction: stats.DamageReduction,
	}
}

// StatsFromProfile builds game stats from the persisted profile record.
func StatsFromProfile(rec db.ProfileRecord) Stats {
	stats := DefaultStats()
	stats.Name = rec.Name
	stats.Class = rec.Class
	stats.Level = rec.Level
	stats.Exp = rec.Exp
	stats.Next = rec.NextLevelExp
	stats.HP = rec.HP
	stats.MaxHP = rec.MaxHP
	stats.Attack = rec.Attack
	stats.Defense = rec.Defense
	stats.Combo = rec.Combo
	stats.Streak = rec.StreakDays
	stats.Gold = rec.Gold
	stats.ExpBoost = rec.ExpBoost
	stats.DamageReduction = rec.DamageReduction
	return stats
}
