package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"tui-english-quest/internal/config"
	"tui-english-quest/internal/db"
	"tui-english-quest/internal/game"
	"tui-english-quest/internal/ui"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("tui-english-quest: ")

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Warning: failed to load config: %v", err)
	}
	if cfg.ProfileID == "" {
		cfg.ProfileID = newProfileID()
		if err := config.SaveConfig(cfg); err != nil {
			log.Printf("Warning: failed to persist profile id: %v", err)
		}
	}

	// Get database path from environment
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./db.sqlite" // Default fallback
	}

	// Initialize database
	if err := db.InitDB(dbPath); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	db.SetProfileID(cfg.ProfileID)
	ctx := context.Background()

	var stats game.Stats
	if rec, err := db.LoadProfile(ctx, cfg.ProfileID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			stats = game.DefaultStats()
			if err := game.SaveStats(ctx, stats); err != nil {
				log.Printf("failed to seed profile: %v", err)
			}
		} else {
			log.Fatalf("failed to load profile: %v", err)
		}
	} else {
		stats = game.StatsFromProfile(rec)
	}

	p := tea.NewProgram(ui.NewRootModel(stats, cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("failed to run program: %v", err)
	}
}

func newProfileID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		log.Printf("Warning: failed to generate profile id: %v", err)
		return fmt.Sprintf("player-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf[:])
}
