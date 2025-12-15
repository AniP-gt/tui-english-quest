package services

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Speak speaks the provided text using a platform-appropriate TTS command.
// It prefers an environment-configured command via SPEAK_CMD where "%s" will be
// replaced with the text. If unset, it falls back to macOS `say`.
func Speak(text string) error {
	if text == "" {
		return nil
	}
	// If user provides SPEAK_CMD (e.g. "say -v Alex '%s'" or "espeak '%s'"), use it.
	if cmdTemplate := os.Getenv("SPEAK_CMD"); cmdTemplate != "" {
		cmdStr := fmt.Sprintf(cmdTemplate, escapeArg(text))
		parts := strings.Fields(cmdStr)
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Fallback: macOS `say`
	if _, err := exec.LookPath("say"); err == nil {
		cmd := exec.Command("say", text)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// If no TTS available, return no-op error to inform caller.
	return fmt.Errorf("no TTS available: set SPEAK_CMD or install 'say' on macOS")
}

// escapeArg provides a naive escaping for use inside a single argument.
// For templates where the user includes quotes, they should handle quoting.
func escapeArg(s string) string {
	// Replace newlines with spaces to keep speech command simple
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
