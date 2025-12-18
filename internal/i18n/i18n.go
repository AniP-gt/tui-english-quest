package i18n

import (
	"fmt"
	"strings"
)

var lang = "en"

// SetLang sets the active language preference: "en" or "ja".
// Any other value will default to "en".
func SetLang(l string) {
	l = strings.TrimSpace(strings.ToLower(l))
	if l == "ja" {
		lang = "ja"
		return
	}
	// default to English
	lang = "en"
}

// T returns the translated string for key based on the active language.
func T(key string) string {
	enVal, okEn := en[key]
	jaVal, okJa := ja[key]
	if !okEn && !okJa {
		return fmt.Sprintf("[%s]", key)
	}
	switch lang {
	case "en":
		if okEn {
			return enVal
		}
		return jaVal
	case "ja":
		if okJa {
			return jaVal
		}
		return enVal
	default:
		// fallback to English value if available
		if okEn {
			return enVal
		}
		return jaVal
	}
}

var en = map[string]string{
	"menu_start":                      "Start Adventure",
	"menu_new":                        "New Game",
	"menu_quit":                       "Quit",
	"note_newgame":                    "Press N to start a new game",
	"note_confirm_newgame":            "Starting a new game resets progress. Proceed? [y/n]",
	"app_title":                       "TUI English Quest",
	"footer_main":                     "[j/k] Move  [Enter] Select  [n] New Game  [q] Quit",
	"settings_title":                  "Settings",
	"settings_prompt":                 "Configure application settings:",
	"settings_menu_api":               "Set Gemini API Key",
	"settings_menu_lang":              "Language (EN/JA)",
	"settings_save":                   "Save and Exit",
	"settings_menu_questions_current": "Questions per session (current: %d)",
	"confirm_save":                    "API key has changed. Do you want to save?",
	"confirm_save_opt1":               "Save Changes",
	"confirm_save_opt2":               "Discard Changes",
	"confirm_save_opt3":               "Cancel",
	"api_label":                       "API Key",
	"fetching_questions":              "Fetching questions...",
	"fetching_tavern":                 "Fetching tavern...",
	"press_enter_return":              "Press Enter to return to Town.",
	"press_enter_continue":            "Press Enter to continue...",
	"your_turn":                       "Your turn",
	"correct_feedback":                "Correct!",
	"incorrect_feedback":              "Incorrect. Answer: %s",
	"exiting_message":                 "Exiting TUI English Quest...",
	"session_complete":                "Session complete",
	"listening_progress":              "Listening %d/%d",
	"press_r_replay":                  "(Press [r] to replay)",
	"footer_listening1":               "[r] Replay  [Enter] Answer/Continue  [Esc/q] Back to Town",
	"footer_listening2":               "[Enter] Continue  [Esc/q] Back to Town",
	"footer_listening3":               "[j/k] Move  [1-4] Quick select  [r] Replay  [Enter] Answer/Continue  [Esc/q] Back to Town",
	"spelling_placeholder":            "Type the spelling...",
	"error_fetching_questions":        "Error fetching questions: %v",
	"spelling_almost_correct":         "Almost! The correct spelling is: %s",
	"spelling_incorrect":              "Incorrect. The correct spelling is: %s",
	"spelling_question_progress":      "Question %d/%d: %s",
	"footer_spelling":                 "[Tab] Toggle MC  [Enter] Submit  [Esc] Back to Town  [q] Quit",
	"town_menu_vocab_battle":          "⚔  Vocabulary Battle",
	"town_menu_grammar_dungeon":       "🏰 Grammar Dungeon",
	"town_menu_conversation_tavern":   "🍺 Conversation Tavern",
	"town_menu_spelling_challenge":    "🪄 Spelling Challenge",
	"town_menu_listening_cave":        "🔊 Listening Cave",
	"town_menu_ai_analysis":           "🧠 AI Analysis",
	"town_menu_history":               "📖 History",
	"town_menu_status":                "🎒 Status",
	"town_menu_settings":              "⚙  Settings",
	"error_ai_advice":                 "Error getting AI advice: %v",
	"town_menu_prompt":                "Where do you want to go?",
	"town_ai_advice_format":           "\nTip / AI Advice\n  Weak points: %s\n  Recommendation: %s",
	"footer_town":                     "[j/k] Move  [Enter] Select  [q] Quit",
	"result_title":                    "Result",
	"result_title_vocab":              "Vocabulary Battle",
	"result_title_grammar":            "Grammar Dungeon",
	"result_title_tavern":             "Conversation Tavern",
	"result_title_spelling":           "Spelling Challenge",
	"result_title_listening":          "Listening Cave",
	"result_exp_gain":                 "EXP: +%d",
	"result_hp_delta":                 "HP: %+d",
	"result_gold_delta":               "Gold: %+d",
	"result_defense_delta":            "Defense: %+0.1f",
	"result_correct":                  "Correct: %d",
	"result_leveled_up":               "Level up! You feel stronger.",
	"result_fainted":                  "Fainted. You lost some EXP.",
	"result_note":                     "Note: %s",
	"result_footer":                   "Press Enter to return to Town.",
}

var ja = map[string]string{

	"menu_start":           "冒険を始める",
	"menu_new":             "新しいゲーム",
	"menu_quit":            "終了",
	"note_newgame":         "Nで新しいゲームを開始",
	"note_confirm_newgame": "新しいゲームを始めると進行状況がリセットされます。よろしいですか？ [y/n]",
	"app_title":            "TUI English Quest",
	"footer_main":          "[j/k] 移動  [Enter] 選択  [n] 新しいゲーム  [q] 終了",
	"settings_title":       "設定",
	"settings_prompt":      "アプリケーション設定:",
	"settings_menu_api":    "ジェミニAPIキー設定",
	"settings_menu_lang":   "言語設定 (EN/JA)",

	"settings_menu_lang_current":      "言語設定 (現在: %s)",
	"settings_save":                   "保存して終了",
	"settings_menu_questions_current": "1セッションの出題数 (現在: %d)",
	"footer_history":                  "[j/k] 移動  [Enter/Esc] Townへ戻る",
	"history_title":                   "セッション履歴",
	"history_no_sessions":             "セッションは見つかりませんでした。",
	"analysis_title":                  "AI 分析",
	"analysis_recent_performance":     "Your recent performance (last 200 questions)",
	"analysis_weak_points":            "Weak points:",
	"analysis_strengths":              "Strengths:",
	"analysis_recommendations":        "Recommendations:",
	"footer_analysis":                 "[Enter] OK  [Esc] Back to Town",
	"confirm_save":                    "APIキーが変更されました。保存しますか?",
	"confirm_save_opt1":               "変更を保存",
	"confirm_save_opt2":               "変更を破棄",
	"confirm_save_opt3":               "キャンセル",
	"api_label":                       "APIキー",
	"fetching_questions":              "問題を取得しています...",
	"fetching_tavern":                 "酒場を取得しています...",
	"press_enter_return":              "Townへ戻るにはEnterを押してください。",
	"press_enter_continue":            "続行するにはEnterを押してください...",
	"your_turn":                       "あなたの番",
	"correct_feedback":                "正解！",
	"incorrect_feedback":              "不正解。正解: %s",
	"exiting_message":                 "TUI English Questを終了しています...",
	"session_complete":                "セッション完了",
	"listening_progress":              "リスニング %d/%d",
	"press_r_replay":                  "([r] で再生)",
	"footer_listening1":               "[r] 再生  [Enter] 解答/続行  [Esc/q] Townへ戻る",
	"footer_listening2":               "[Enter] 続行  [Esc/q] Townへ戻る",
	"footer_listening3":               "[j/k] 移動  [1-4] クイック選択  [r] 再生  [Enter] 解答/続行  [Esc/q] Townへ戻る",
	"spelling_placeholder":            "スペルを入力してください...",
	"error_fetching_questions":        "問題の取得中にエラーが発生しました: %v",
	"spelling_almost_correct":         "惜しい！正しいスペルは: %s",
	"spelling_incorrect":              "不正解。正しいスペルは: %s",
	"spelling_question_progress":      "問題 %d/%d: %s",
	"footer_spelling":                 "[Tab] MC切り替え  [Enter] 送信  [Esc] Townへ戻る  [q] 終了",
	"settings_api_placeholder":        "ジェミニAPIキーを入力",
	"footer_settings_confirm":         "[j/k] 移動  [Enter] 選択",
	"footer_settings_main":            "[j/k] 移動  [Enter] 選択  [Esc] Townへ戻る",

	"dungeon_placeholder":       "あなたの解答...",
	"dungeon_incorrect_answer":  "不正解。正解は: %s",
	"dungeon_question_progress": "問題 %d/%d: %s",
	"footer_dungeon":            "[j/k] 移動  [Enter] 選択/解答  [Esc] Townへ戻る  [q/ctrl+c] 終了",
	"battle_placeholder":        "あなたの解答...",
	"battle_incorrect_answer":   "不正解。正解は: %s",
	"battle_question_format":    "問題 %d/%d: '%s' の意味は？",
	"footer_battle":             "[j/k] 移動  [Enter] 選択/解答  [Esc] Townへ戻る  [q/ctrl+c] 終了",
	"tavern_placeholder":        "Say something...",
	"tavern_exiting":            "Exiting...",
	"tavern_evaluations":        "Evaluations:",
	"tavern_finished_format":    "Tavern finished: Exp +%d, Gold +%d. Correct: %d",
	"tavern_eval_default_fail":  "Evaluation failed; defaulted to Normal.",
	"tavern_npc_line":           "NPC (%s): %s",
	"tavern_player_turn":        "あなたの番 (%d/%d):\n%s",
	"tavern_eval_line":          "Turn %d: %s — %s",

	// Town / Menu related translations (added)
	"town_menu_vocab_battle":        "⚔  単語バトル",
	"town_menu_grammar_dungeon":     "🏰 文法ダンジョン",
	"town_menu_conversation_tavern": "🍺 会話の酒場",
	"town_menu_spelling_challenge":  "🪄 スペルチャレンジ",
	"town_menu_listening_cave":      "🔊 リスニングケイブ",
	"town_menu_ai_analysis":         "🧠 AI 分析",
	"town_menu_history":             "📖 履歴",
	"town_menu_status":              "🎒 ステータス",
	"town_menu_settings":            "⚙  設定",
	"town_menu_prompt":              "どこに行きますか？",
	"town_ai_advice_format":         "\nヒント / AIアドバイス\n  弱点: %s\n  推奨: %s",
	"footer_town":                   "[j/k] 移動  [Enter] 選択  [q] 終了",
	"result_title":                  "結果",
	"result_title_vocab":            "単語バトル",
	"result_title_grammar":          "文法ダンジョン",
	"result_title_tavern":           "会話の酒場",
	"result_title_spelling":         "スペルチャレンジ",
	"result_title_listening":        "リスニングケイブ",
	"result_exp_gain":               "経験値: +%d",
	"result_hp_delta":               "HP: %+d",
	"result_gold_delta":             "ゴールド: %+d",
	"result_defense_delta":          "守備: %+0.1f",
	"result_correct":                "正解数: %d",
	"result_leveled_up":             "レベルアップ！強くなった気がする。",
	"result_fainted":                "気絶しました。経験値を少し失いました。",
	"result_note":                   "備考: %s",
	"result_footer":                 "EnterでTownに戻る。",
}

// helper to combine mode for fetching strings
func FetchingFor(mode string) string {
	key := "fetching_questions"
	if mode == "tavern" {
		key = "fetching_tavern"
	} else if mode == "battle" { // Add this for battle mode
		key = "fetching_questions"
	}
	return T(key)
}

// MenuLabel returns a label suitable for compact menus.
// If current language is "both" and both translations exist, it returns Japanese + "\n" + English
// otherwise it returns the same as T(key).
func MenuLabel(key string) string {
	return T(key)
}
