package ui

import (
	"tui-english-quest/internal/i18n"
	"tui-english-quest/internal/services"
)

func modeLabel(mode string) string {
	switch mode {
	case services.ModeVocab:
		return i18n.T("town_menu_vocab_battle")
	case services.ModeGrammar:
		return i18n.T("town_menu_grammar_dungeon")
	case services.ModeTavern:
		return i18n.T("town_menu_conversation_tavern")
	case services.ModeSpelling:
		return i18n.T("town_menu_spelling_challenge")
	case services.ModeListening:
		return i18n.T("town_menu_listening_cave")
	default:
		if mode == "" {
			return ""
		}
		return mode
	}
}
