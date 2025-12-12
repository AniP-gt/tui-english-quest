# Data Model - TUI English Quest v2.0

## Entities

### PlayerProfile
- id (string, unique)
- name (string)
- class (enum: Vocabulary Warrior / Grammar Mage / Conversation Bard)
- level (int)
- exp (int)
- next_level_exp (int)
- hp (int), max_hp (int)
- attack (int)
- defense (float)
- combo (int)
- streak_days (int)
- gold (int)
- equipment_ids (weapon/armor/ring/charm slot refs)
- unlocked_modes (array of mode keys)
- ui_language (enum: ja/en)
- explanation_language (enum: ja+hint/en)
- problem_language (enum: en)
- updated_at (timestamp)

### SessionRecord
- id (string, unique)
- player_id (string, fk PlayerProfile)
- mode (enum: vocab/grammar/tavern/spelling/listening)
- started_at / ended_at (timestamp)
- question_set_id (string)
- correct_count (int) — for tavern, success/normal/fail counts
- best_combo (int)
- exp_gained (int)
- exp_lost (int) — e.g., faint penalty
- hp_delta (int)
- gold_delta (int)
- defense_delta (float)
- fainted (bool)
- leveled_up (bool)

### QuestionSet (per mode)
- id (string, unique)
- mode (enum)
- fetched_at (timestamp)
- payload (jsonb):
  - vocab: questions[5]{enemy_name, word, options[4], answer_index, explanation}
  - grammar: traps[5]{trap_name, question, options[4], answer_index, explanation}
  - tavern: npc_name, npc_opening, evaluation_rubric (3-level), turns[5]{player_prompt, npc_reply}
  - spelling: prompts[5]{ja_hint, correct_spelling, explanation}
  - listening: audio[5]{prompt, options[4], answer_index, transcript}

### EquipmentItem
- id (string, unique)
- slot (enum: weapon/armor/ring/charm)
- name (string)
- effect_type (enum: exp_boost/damage_reduction)
- effect_value (float)
- target_mode (enum or all)
- price (int gold)

### WeaknessAnalysis
- id (string, unique)
- player_id (fk PlayerProfile)
- analyzed_range (int, number of recent questions)
- weak_points (string[])
- strength_points (string[])
- recommendation (string)
- generated_at (timestamp)

## Relationships
- PlayerProfile 1 - N SessionRecord
- PlayerProfile 1 - N WeaknessAnalysis (latest surfaced on town)
- PlayerProfile has 0..1 EquipmentItem per slot (weapon/armor/ring/charm)
- SessionRecord references one QuestionSet

## Validation & Rules
- QuestionSet must contain5問; 不足・不正JSONなら再取得または中断。
- Level-up table: 30/50/80/120 exp thresholds; level-up applies MaxHP+10, full heal, Attack+2, Defense+1.
- Faint: HP<=0 → EXP-5, HP=MaxHP/2, return to town.
- Tavern evaluation: Gemini rubric 3段階評価で成功/普通/失敗を決定。
- Language settings: UI/解説/問題言語はプロファイルに保持し、再描画1回以内に反映。
