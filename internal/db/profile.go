package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// ProfileRecord represents the persisted player profile.
type ProfileRecord struct {
	ID              string
	Name            string
	Class           string
	Level           int
	Exp             int
	NextLevelExp    int
	HP              int
	MaxHP           int
	Attack          int
	Defense         float64
	Combo           int
	StreakDays      int
	Gold            int
	ExpBoost        float64
	DamageReduction float64
	UpdatedAt       time.Time
}

var currentProfileID string

// SetProfileID configures the player ID used for saving/loading profile data.
func SetProfileID(id string) {
	currentProfileID = id
}

// CurrentProfileID returns the active player ID for persistence operations.
func CurrentProfileID() string {
	return currentProfileID
}

// LoadProfile loads saved stats for the given player ID.
func LoadProfile(ctx context.Context, playerID string) (ProfileRecord, error) {
	var rec ProfileRecord
	if dbConn == nil {
		return rec, fmt.Errorf("database not initialized")
	}
	if playerID == "" {
		return rec, fmt.Errorf("player ID is required")
	}
	row := dbConn.QueryRowContext(ctx, `
        SELECT id, name, class, level, exp, next_level_exp, hp, max_hp, attack, defense, combo, streak_days, gold, exp_boost, damage_reduction, updated_at
        FROM profiles
        WHERE id = ?
    `, playerID)
	err := row.Scan(
		&rec.ID, &rec.Name, &rec.Class, &rec.Level, &rec.Exp, &rec.NextLevelExp, &rec.HP, &rec.MaxHP,
		&rec.Attack, &rec.Defense, &rec.Combo, &rec.StreakDays, &rec.Gold, &rec.ExpBoost,
		&rec.DamageReduction, &rec.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return rec, err
		}
		return rec, fmt.Errorf("failed to scan profile: %w", err)
	}
	return rec, nil
}

// SaveProfile inserts or updates the profile row for the given player.
func SaveProfile(ctx context.Context, rec ProfileRecord) error {
	if dbConn == nil {
		return nil
	}
	if rec.ID == "" {
		return fmt.Errorf("player ID is required")
	}

	_, err := dbConn.ExecContext(ctx, `
        INSERT INTO profiles (id, name, class, level, exp, next_level_exp, hp, max_hp, attack, defense, combo, streak_days, gold, exp_boost, damage_reduction)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(id) DO UPDATE SET
            name = excluded.name,
            class = excluded.class,
            level = excluded.level,
            exp = excluded.exp,
            next_level_exp = excluded.next_level_exp,
            hp = excluded.hp,
            max_hp = excluded.max_hp,
            attack = excluded.attack,
            defense = excluded.defense,
            combo = excluded.combo,
            streak_days = excluded.streak_days,
            gold = excluded.gold,
            exp_boost = excluded.exp_boost,
            damage_reduction = excluded.damage_reduction,
            updated_at = CURRENT_TIMESTAMP
    `,
		rec.ID, rec.Name, rec.Class, rec.Level, rec.Exp, rec.NextLevelExp, rec.HP, rec.MaxHP,
		rec.Attack, rec.Defense, rec.Combo, rec.StreakDays, rec.Gold, rec.ExpBoost, rec.DamageReduction,
	)
	if err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}
	return nil
}
