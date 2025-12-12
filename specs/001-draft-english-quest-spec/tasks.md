# Tasks: TUI English Quest v2.0

**Input**: Design documents from `/specs/001-draft-english-quest-spec/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: 本仕様ではテスト必須指定なし。必要に応じて追加可。

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 基本的な開発環境と設定ファイルの準備

- [x] T001 Go依存関係を同期しビルド確認を行う（`go.mod` / `go.sum`）
- [x] T002 `.env.example` に `GEMINI_API_KEY` など必要変数を追加し運用手順を記述（`configs/.env.example`）
- [x] T003 [P] Quickstartの手順に沿ってビルド・実行の確認ノートを反映（`specs/001-draft-english-quest-spec/quickstart.md`）

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: すべてのユーザーストーリー実装前に必要な共通基盤

**⚠️ CRITICAL**: このフェーズ完了までユーザーストーリー着手不可

- [x] T004 SQLiteスキーマ初期版を整備（profiles/sessions/equipment/analysis）（`internal/db/schema.sql`）
- [x] T005 [P] 出題JSONバリデーションヘルパーを実装（5件チェック＋スキーマ検証）（`internal/services/gemini.go`）
- [x] T006 [P] ステータス計算ユーティリティ（EXP/HP/レベルアップ/戦闘不能）を共通化（`internal/game/stats.go`）
- [x] T007 ログ・エラーハンドリングの基本方針を設定（`cmd/english-quest/main.go` 入口で初期化）

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - 2分で5問の学習セッションを完走する (Priority: P1) 🎯 MVP

**Goal**: 任意モードで5問を連続実行し、EXP/HP/コンボ/Goldをルール通り反映してリザルトを表示する。

**Independent Test**: 任意モード開始→5問消化→リザルト表示までを単体で実行し、計算結果が仕様（FR-005〜FR-010）と一致すること。

### Implementation for User Story 1

- [x] T029 [US1] New Game初期化（名前/クラス入力と初期ステータス設定）を実装（`cmd/english-quest/main.go`, `internal/ui/top.go`, `internal/game/stats.go`）
- [x] T008 [P] [US1] モード共通の質問セット取得フローを実装（開始時5問まとめて、通信なしで進行）（`internal/services/gemini.go`）
- [x] T009 [P] [US1] 単語バトルUIと結果計算を実装（正解/不正解時のEXP・HP・コンボ処理）（`internal/ui/battle.go`）
- [x] T010 [P] [US1] 文法ダンジョンUIと結果計算を実装（Defense成長、HP減算）（`internal/ui/dungeon.go`）
- [x] T011 [P] [US1] 会話タバーンの5ターン進行とGeminiルーブリック自動判定を実装（`internal/ui/tavern.go`）
- [x] T012 [P] [US1] スペリングチャレンジUIと判定（完全一致/1文字ミス/不正解処理）（`internal/ui/spelling.go`）
- [x] T013 [P] [US1] リスニングモードUIと判定（再生/リプレイ含む）（`internal/ui/listening.go`）
- [x] T014 [US1] リザルト集計とステータス更新を共通関数化し各モードから呼び出し（`internal/game/exp.go` / `internal/game/damage.go`）
- [x] T015 [US1] セッション結果の履歴保存（SessionRecord作成）を実装（`internal/db/history.go`）
- [x] T033 [US1] 出題取得失敗・JSON不正時のリトライ/安全中断を実装（`internal/services/gemini.go`, `internal/ui/*`）
- [x] T034 [US1] 途中離脱(Esc/q)・タイムアウト時にステータス未反映で街へ戻す処理を実装（`internal/ui/*`, `internal/game/stats.go`）
- [x] T035 [US1] リスニングデバイス不可時の代替テキスト提示とスキップ/再試行を実装（`internal/ui/listening.go`）

**Checkpoint**: User Story 1 が単独で完走・記録できること

---

## Phase 4: User Story 2 - 街で状態確認とモード選択ができる (Priority: P2)

**Goal**: 街画面でステータスバーとAIアドバイスを確認し、j/k+Enterで各モードへ遷移できる。

**Independent Test**: アプリ起動→街画面でステータス/アドバイス表示→メニュー遷移→戻るまでが単体で成立すること。

### Implementation for User Story 2

- [x] T016 [P] [US2] ステータスバー表示コンポーネントを実装（LV/EXP/HP/Combo/Streak/Gold）（`internal/ui/components/statusbar.go`）
- [x] T017 [P] [US2] Townメニュー画面のナビゲーション（j/k/Enter）と各モード遷移ハンドラを実装（`internal/ui/town.go`）
- [x] T018 [US2] AIアドバイス表示（弱点・推奨モード）をTown画面に統合（`internal/ui/town.go`）
- [x] T019 [US2] ステータスバーの再描画連携を全モードに配線（`internal/ui/top.go` / `internal/ui/*.go`）
- [x] T032 [US2] 共通キーバインド表示/入力ハンドラの一元化（j/k/Enter/Tab/q/Esc）（`internal/ui/top.go`, `internal/ui/*.go`）

**Checkpoint**: User Story 2 が単独で操作・表示できること

---

## Phase 5: User Story 3 - 成長確認とフィードバックで次の行動を決める (Priority: P3)

**Goal**: 履歴・弱点分析・装備効果を確認し、次の行動を決められる。

**Independent Test**: 直近セッション後に履歴/弱点分析/装備表示を開き、記録と推奨/効果が反映されていることを単体確認。

### Implementation for User Story 3

- [x] T020 [P] [US3] 履歴一覧と詳細表示を実装（正答数・リソース変化・日時）（`internal/ui/history.go`）
- [x] T021 [P] [US3] AI弱点分析の読み込みと表示（直近50〜200問、推奨モード）を実装（`internal/ui/analysis.go`）
- [x] T022 [P] [US3] 装備画面でスロット別装備変更と効果表示を実装（`internal/ui/equipment.go`）
- [x] T030 [US3] 弱点分析生成（直近50〜200問集計→JSON生成）を実装（`internal/services/analysis.go`, `internal/db/history.go`）
- [x] T031 [US3] 弱点に基づく出題優先度反映（モード別選定フック）を実装（`internal/services/gemini.go`）
- [x] T023 [US3] 装備効果の計算を報酬/被ダメに反映するフックを追加（`internal/game/stats.go` / `internal/game/exp.go`）
- [x] T024 [US3] ステータス画面にプレイヤー成長（LV/HP/攻防/バッジなど）を表示（`internal/ui/status.go`）

**Checkpoint**: User Story 3 が単独で閲覧・判断できること

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: 複数ストーリーにまたがる仕上げ

- [x] T025 [P] UIラベル/ヘルプの多言語切替を全画面で確認・整備（`internal/ui/`）
- [x] T026 コード整理とコメント最小限のリファクタ（`internal/` 全体）
- [x] T027 パフォーマンス微調整（描画・入力遅延が体感100ms以内か確認）（`internal/ui/`）
- [x] T028 [P] quickstart検証と更新（`specs/001-draft-english-quest-spec/quickstart.md`）

---

## Dependencies & Execution Order

### Phase Dependencies

- Setup (Phase 1): なし
- Foundational (Phase 2): Setup完了に依存。すべてのユーザーストーリーをブロック。
- User Stories (Phase 3〜5): Foundational完了後に着手。優先度順（P1→P2→P3）推奨だが並行も可。
- Polish (Final): すべてのストーリー完了後。

### User Story Dependencies

- User Story 1 (P1): Foundational後に開始。T029で初期化を満たし、T033〜T035のエラーフローを含め単独完結。
- User Story 2 (P2): Foundational後に開始。US1と独立だがステータス計算が必要。共通キーバインド統合（T032）を含む。
- User Story 3 (P3): Foundational後に開始。履歴・分析・装備はUS1の記録/計算を前提にし、弱点分析生成/反映（T030/T031）を含む。

### Within Each User Story

- モデル/計算 → UI表示 → 保存/連携 の順で実装。
- 各ストーリーは単体で完走・表示・判断が可能な状態で区切る。

### Parallel Opportunities

- Setup: T003 は T001/T002 と並行可。
- Foundational: T005/T006 は T004 と並行可。
- US1: T009〜T013 はモード別で並行可（共通処理T014に依存）。T033/T034/T035は共通エラーフローで直列推奨。
- US2: T016〜T018 は表示コンポーネント単位で並行可。T032は共通キー入力統合で直列推奨。
- US3: T020〜T022 は表示単位で並行可（計算反映T023は後続）。T030/T031は分析生成→優先度反映の順で直列。

## Parallel Example: User Story 1

```bash
# モード別UI実装を並行実行
Task: T009 [US1] internal/ui/battle.go
Task: T010 [US1] internal/ui/dungeon.go
Task: T011 [US1] internal/ui/tavern.go
Task: T012 [US1] internal/ui/spelling.go
Task: T013 [US1] internal/ui/listening.go

# 共通計算と履歴保存で集約
Task: T014 [US1] internal/game/exp.go / damage.go
Task: T015 [US1] internal/db/history.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1 → Phase 2 を完了
2. Phase 3 (US1) を実装し、5問セッション完走とリザルト反映を確認
3. 必要ならデモ/リリース

### Incremental Delivery

1. Setup + Foundational を完了
2. US1 実装・検証（MVP）
3. US2 実装・検証（ナビゲーション/表示）
4. US3 実装・検証（履歴/分析/装備）
5. Polish で多言語とUX調整

### Parallel Team Strategy

1. Setup/Foundational を全員で完了
2. US1/US2/US3 を担当別に並行し、共通計算と保存は同期ポイントで集約

---

## Notes

- [P] タスクはファイル競合がない場合のみ並行可。
- ストーリーごとに単体で動くことを必ず確認してから次へ進める。
- コマンド例: `go test ./...` で計算系の単体テストを随時追加可能（任意）。
