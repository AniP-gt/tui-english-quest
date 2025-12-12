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
