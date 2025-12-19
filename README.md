# TUI English Quest

<p align="right">
  <a href="docs/ja/README.md">日本語版</a>
</p>

> English is the default for this document. Switch to the Japanese version above if you need the localized explanations.

## Overview

TUI English Quest is a terminal-based RPG that keeps English study sessions short, gamified, and data-backed. Navigate from the title screen into the town hub, pick a learning mode (Vocabulary Battle, Grammar Dungeon, Conversation Tavern, Spelling Challenge, Listening Cave), and visit supporting screens for equipment, AI analysis, history, status, and settings. Each session requests five prompts from Gemini (`gemini-2.5-flash`), plays them offline, updates your stats (EXP, HP, Combo, Streak, Gold), logs the run, and feeds the AI weakness report.

## Installation & Dependencies

1. **Go 1.21+ and SQLite** are required. Install the Go toolchain and ensure `sqlite3` is accessible if you plan to inspect the database directly.
2. **Build the app** as documented in `specs/001-draft-english-quest-spec/quickstart.md`:
   - `go mod tidy`
   - `go build ./...`
   - `go run ./cmd/english-quest`
3. **Gemini access**: The project uses the `gemini-2.5-flash` generative model. Get a Google Cloud API key, then set it in `GEMINI_API_KEY` before launching the app.
   - English instructions: https://ai.google.dev/gemini-api/docs/api-key?hl=en
   - 日本語の説明: https://ai.google.dev/gemini-api/docs/api-key?hl=ja
4. **Environment**: Copy `configs/.env.example` beside `./cmd/english-quest` or export the values directly. Configure:
   - `GEMINI_API_KEY` (required)
   - `DB_PATH` (defaults to `./db.sqlite`; change if you need a custom location)
   - `LOG_LEVEL` (optional: `info` or `debug`)
   - `SPEAK_CMD` (optional override for text-to-speech)
5. **Run-time config**: The app writes `config.json` under `~/.local/share/tui-english-quest/` (Unix) or `%AppData%\tui-english-quest\` (Windows). This file stores `LangPref`, `ApiKey`, `QuestionsPerSession`, and the generated `ProfileID`.

## Configuration & Environment

- `config.Config` fields:
  - `LangPref`: `en` or `ja`. The settings screen applies the new UI/explanation language immediately.
  - `ApiKey`: Optionally persist the Gemini key so subsequent launches skip manual entry.
  - `QuestionsPerSession`: Controls how many prompts each mode fetches (default 5, adjustable via the settings screen to 10/20/30/50).
  - `ProfileID`: Internal identifier created on first launch and reused for persistence.
- Database schema (`internal/db/schema.sql`) includes:
  - `profiles`: name, class, EXP, HP, attack/defense, combo/streak counters, gold, equipment stats, and language preferences.
  - `sessions`: per-mode history of correct counts, EXP/HP/Gold deltas, best combos, fainted/leveled-up flags.
  - `equipment` and `analysis` tables for gear and generated AI analysis.
- TTS: Without `SPEAK_CMD`, the app falls back to macOS `say`. Define `SPEAK_CMD` as a format string (e.g., `espeak '%s'`) to use another speech engine.

## Gameplay Flow & Modes

1. **Top & Town**: Start at the title screen, confirm new game when needed, then enter Town where the status bar and menu respond to `j/k`, arrow keys, and Enter.
2. **Modes** (each fetches prompts via `services.FetchAndValidate`):
   - **Vocabulary Battle**: Correct answers grant EXP (base + tier + combo boosts) and raise combo counters; misses deal damage based on `AllowedMisses` and reset combo.
   - **Grammar Dungeon**: Similar math to Vocabulary, with additional defense increases and slightly lower damage per miss.
   - **Conversation Tavern**: Gemini returns NPC turns plus an evaluation rubric; player responses are evaluated via `BatchEvaluateTavern`, resulting in success/normal/fail rewards without HP loss.
   - **Spelling Challenge**: Fill-in answers or Tab-triggered multiple choice. Perfects give +5 EXP, near misses +2 EXP with small HP penalties, failures inflict larger HP loss.
   - **Listening Cave**: Audio prompts (replay with `r`) present four options; incorrect answers deal HP damage akin to other combat modes.
3. **Supporting screens**:
   - **Equipment**: Equip weapon, armor, ring, and charm slots; each item modifies `ExpBoost` or `DamageReduction` per mode.
   - **AI Analysis**: Aggregates the last 50–200 questions via `services.AnalyzeWeakness` to produce summaries, weak/strong modes, action plans, and recommendations.
   - **History**: Displays recent sessions with timestamps, mode, EXP/HP/Gold changes, combos, and flags for fainted/leveled-up.
   - **Status**: Shows `game.Stats` (name, class, level, EXP/Next, HP/MaxHP, combo, etc.) plus achievements.
   - **Settings**: Toggle language, edit the API key, adjust question count, and save preferences.
4. **Session Result**: After each mode, `ResultModel` summarizes EXP/HP/Gold changes, leveled-up/fainted notices, and waits for Enter to return to Town.

## Stats, HP & Progression

- HUD-tracked fields: `Level`, `Exp`, `Next`, `HP`, `MaxHP`, `Attack`, `Defense`, `Combo`, `Streak`, `Gold`, `ExpBoost`, and `DamageReduction`.
- **Experience curve**: `ExpToNext(level)` returns `30 + 5*(level-1)` for levels ≤99, then `500 + 10*(level-100)` for higher tiers.
- **Max HP** increases with `MaxHPForLevel`. Leveling up recalculates Max HP, fully heals HP, adds +2 Attack, and +1 Defense.
- **Faint penalty**: When HP reaches zero, `ApplyFaintPenalty` subtracts 5 EXP (floor 0) and restores HP to 50% of Max HP.
- **Recovery**: Level ups and Town→mode transitions (`game.FullHeal`) heal HP to the maximum before each run.
- **Combo & Streak**: Correct answers increase combo, misses reset it; streak counts track consecutive successful days.
- **Equipment buffs**: Gear amplifies EXP (`ExpBoost`) or mitigates damage (`DamageReduction`) on a per-mode basis.

## Controls & Navigation

- Use `j/k` or arrow keys to move between menus; press Enter to confirm.
- Tab toggles between fill-in and multiple-choice in the Spelling Challenge.
- Numeric keys `1`–`4` select MC answers in Spelling and Listening modes.
- Press `r` to replay the current Listening prompt.
- `Esc`, `q`, or `Ctrl+C` backs out of a screen or exits the application.
- Town menus provide direct access to Equipment, AI Analysis, History, Status, Settings, and quit.

## AI Analysis, History & Equipment

- **Gemini contracts**: Each mode complies with the JSON schema documented in `specs/001-draft-english-quest-spec/contracts/gemini-contracts.md`.
- **Weakness analysis**: `services.AnalyzeWeakness` compiles recent sessions into a `WeaknessReport` that exposes summaries, weak/strong insights, action plans, and recommendations in the Town and Analysis screens.
- **History** (`db.sessions`): Stores timestamps, mode names, correct counts, EXP/HP/Gold deltas, combos, and boolean flags for fainted/leveled-up states.
- **Equipment slots**: Weapon, armor, ring, charm stores `effect_type` (`ExpBoost`, `DamageReduction`), `effect_value`, and `target_mode`, and the UI applies their multipliers to session rewards and damage.

## Troubleshooting & Testing

- **Gemini failures**: If fetching questions or Tavern evaluations fails, the UI shows an error message while leaving existing stats untouched.
- **HP zero**: Players immediately receive the faint penalty (−5 EXP, HP set to 50% Max) and the session logs the faint.
- **Mid-session quit**: Press `Esc` or `q` to abandon a session before completion. Pending EXP/HP changes are discarded and Town returns to a fresh state.
- **Missing TTS**: When `SPEAK_CMD` is unset and `say` is unavailable, the app logs a warning and skips speech.
- **Testing**: Run `go test ./...` to cover stat math, mode results, and Gemini payload validation (`services.ValidatePayload`).

## Resources & References

- Feature spec: `specs/001-draft-english-quest-spec/spec.md`
- Quickstart: `specs/001-draft-english-quest-spec/quickstart.md`
- Gemini contracts: `specs/001-draft-english-quest-spec/contracts/gemini-contracts.md`
- Database schema: `internal/db/schema.sql`
- AI analysis logic: `internal/services/analysis.go`
- UI layouts: review the models under `internal/ui/` and `internal/ui/components`
- Gemini client: `internal/services/gemini.go`
