package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// SessionRecord represents a single completed learning session.
type SessionRecord struct {
	ID            string
	PlayerID      string
	Mode          string
	StartedAt     time.Time
	EndedAt       time.Time
	QuestionSetID string
	CorrectCount  int
	BestCombo     int
	ExpGained     int
	ExpLost       int
	HPDelta       int
	GoldDelta     int
	DefenseDelta  float64
	Fainted       bool
	LeveledUp     bool
}

var dbConn *sql.DB

// InitDB initializes the SQLite database connection and creates tables.
func InitDB(dataSourceName string) error {
	var err error
	dbConn, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	schema := `
  CREATE TABLE IF NOT EXISTS profiles (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        class TEXT NOT NULL,
        level INTEGER NOT NULL,
        exp INTEGER NOT NULL,
        next_level_exp INTEGER NOT NULL,
        hp INTEGER NOT NULL,
        max_hp INTEGER NOT NULL,
        attack INTEGER NOT NULL,
        defense REAL NOT NULL,
        combo INTEGER NOT NULL,
        streak_days INTEGER NOT NULL,
        gold INTEGER NOT NULL,
        exp_boost REAL NOT NULL DEFAULT 0,
        damage_reduction REAL NOT NULL DEFAULT 0,
        ui_language TEXT,
        explanation_language TEXT,
        problem_language TEXT,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );


	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		player_id TEXT NOT NULL,
		mode TEXT NOT NULL,
		started_at TIMESTAMP,
		ended_at TIMESTAMP,
		question_set_id TEXT,
		correct_count INTEGER,
		best_combo INTEGER,
		exp_gained INTEGER,
		exp_lost INTEGER,
		hp_delta INTEGER,
		gold_delta INTEGER,
		defense_delta REAL,
		fainted INTEGER,
		leveled_up INTEGER,
		FOREIGN KEY(player_id) REFERENCES profiles(id)
	);

	CREATE TABLE IF NOT EXISTS equipment (
		id TEXT PRIMARY KEY,
		slot TEXT NOT NULL,
		name TEXT NOT NULL,
		effect_type TEXT,
		effect_value REAL,
		target_mode TEXT,
		price INTEGER
	);

	CREATE TABLE IF NOT EXISTS analysis (
		id TEXT PRIMARY KEY,
		player_id TEXT NOT NULL,
		analyzed_range INTEGER,
		weak_points TEXT,
		strength_points TEXT,
		recommendation TEXT,
		generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(player_id) REFERENCES profiles(id)
	);
	`
	_, err = dbConn.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	if err := ensureProfileColumn("exp_boost", "exp_boost REAL NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureProfileColumn("damage_reduction", "damage_reduction REAL NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	return nil
}

// SaveSession persists a session record to the database.
func SaveSession(ctx context.Context, rec SessionRecord) error {
	if dbConn == nil {
		return nil
	}
	if rec.PlayerID == "" {
		return nil
	}
	stmt, err := dbConn.PrepareContext(ctx, `
            INSERT INTO sessions (id, player_id, mode, started_at, ended_at, question_set_id, correct_count, best_combo, exp_gained, exp_lost, hp_delta, gold_delta, defense_delta, fainted, leveled_up)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `)

	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		rec.ID, rec.PlayerID, rec.Mode, rec.StartedAt, rec.EndedAt, rec.QuestionSetID,
		rec.CorrectCount, rec.BestCombo, rec.ExpGained, rec.ExpLost, rec.HPDelta,
		rec.GoldDelta, rec.DefenseDelta, boolToInt(rec.Fainted), boolToInt(rec.LeveledUp),
	)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}
	return nil
}

// ListSessions fetches recent session records for a player.
func ListSessions(ctx context.Context, playerID string, limit int) ([]SessionRecord, error) {
	rows, err := dbConn.QueryContext(ctx, `
        SELECT id, player_id, mode, started_at, ended_at, question_set_id, correct_count, best_combo, exp_gained, exp_lost, hp_delta, gold_delta, defense_delta, fainted, leveled_up
        FROM sessions
        WHERE player_id = ?
        ORDER BY ended_at DESC
        LIMIT ?
    `, playerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []SessionRecord
	for rows.Next() {
		var rec SessionRecord
		var faintedInt, leveledUpInt int
		err := rows.Scan(
			&rec.ID, &rec.PlayerID, &rec.Mode, &rec.StartedAt, &rec.EndedAt, &rec.QuestionSetID,
			&rec.CorrectCount, &rec.BestCombo, &rec.ExpGained, &rec.ExpLost, &rec.HPDelta,
			&rec.GoldDelta, &rec.DefenseDelta, &faintedInt, &leveledUpInt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session row: %w", err)
		}
		rec.Fainted = intToBool(faintedInt)
		rec.LeveledUp = intToBool(leveledUpInt)
		sessions = append(sessions, rec)
	}
	return sessions, nil
}

func ensureProfileColumn(name, definition string) error {
	if dbConn == nil {
		return fmt.Errorf("database not initialized")
	}
	exists, err := profileColumnExists(name)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = dbConn.Exec(fmt.Sprintf("ALTER TABLE profiles ADD COLUMN %s", definition))
	if err != nil {
		return fmt.Errorf("failed to add column %s: %w", name, err)
	}
	return nil
}

func profileColumnExists(name string) (bool, error) {
	if dbConn == nil {
		return false, fmt.Errorf("database not initialized")
	}
	rows, err := dbConn.Query(`PRAGMA table_info(profiles)`)
	if err != nil {
		return false, fmt.Errorf("failed to query table info: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var colName string
		var colType string
		var notnull int
		var dfltValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &colName, &colType, &notnull, &dfltValue, &pk); err != nil {
			return false, fmt.Errorf("failed to scan table info: %w", err)
		}
		if colName == name {
			return true, nil
		}
	}
	return false, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	return i != 0
}

// NewSessionRecord initializes a history record with IDs and timestamps prefilled.
func NewSessionRecord(mode string, startedAt, endedAt time.Time) SessionRecord {
	return SessionRecord{
		ID:        newSessionID(),
		PlayerID:  CurrentProfileID(),
		Mode:      mode,
		StartedAt: startedAt,
		EndedAt:   endedAt,
	}
}

func newSessionID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("session-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf[:])
}
