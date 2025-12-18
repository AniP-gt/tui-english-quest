package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Config holds user preferences persisted on disk.
type Config struct {
	LangPref            string `json:"lang_pref"` // "en"/"ja" ("both" removed)
	ApiKey              string `json:"api_key"`
	QuestionsPerSession int    `json:"questions_per_session"`
	ProfileID           string `json:"profile_id"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{LangPref: "en", ApiKey: "", QuestionsPerSession: 5, ProfileID: ""}
}

// ConfigPath returns the platform-appropriate path for the config file.
func ConfigPath() (string, error) {
	var base string
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		base = x
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if runtime.GOOS == "windows" {
			base = filepath.Join(home, "AppData", "Roaming")
		} else {
			base = filepath.Join(home, ".local", "share")
		}
	}
	p := filepath.Join(base, "tui-english-quest", "config.json")
	return p, nil
}

// LoadConfig loads configuration from disk or returns default.
func LoadConfig() (Config, error) {
	p, err := ConfigPath()
	if err != nil {
		return DefaultConfig(), err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		// if file not exist, return default
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return DefaultConfig(), err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return DefaultConfig(), err
	}
	if c.LangPref == "" {
		c.LangPref = "en"
	}
	return c, nil
}

// SaveConfig saves configuration to disk, creating dirs as needed.
func SaveConfig(c Config) error {
	p, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(p, b, 0o644); err != nil {
		return err
	}
	return nil
}
