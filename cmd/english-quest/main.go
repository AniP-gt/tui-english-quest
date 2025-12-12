package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"tui-english-quest/internal/db"
	"tui-english-quest/internal/ui"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("tui-english-quest: ")

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
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

	p := tea.NewProgram(ui.NewRootModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("failed to run program: %v", err)
	}
}
