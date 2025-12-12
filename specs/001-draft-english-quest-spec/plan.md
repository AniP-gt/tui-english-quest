# Implementation Plan: TUI English Quest v2.0

**Branch**: `001-draft-english-quest-spec` | **Date**: 2025-12-11 | **Spec**: specs/001-draft-english-quest-spec/spec.md
**Input**: Feature specification from `/specs/001-draft-english-quest-spec/spec.md`

## Summary

- 2分・5問固定のTUI英語学習RPGをGo + Bubble Teaで実装し、Geminiから各モード開始時にまとめて5問取得してオフライン進行。
- ステータスバー常時表示、街メニューから各モード遷移、装備効果とAI弱点分析を反映した出題・報酬調整を行う。
- 会話タバーンはGeminiの簡易ルーブリック（流暢さ/関連性/タスク達成度）で自動3段階判定し、報酬を配分する。

## Technical Context

**Language/Version**: Go (go.mod準拠、1.21+想定)  
**Primary Dependencies**: Bubble Tea, Lip Gloss, SQLite driver, HTTPクライアントでのGemini API呼び出し  
**Storage**: SQLite（履歴・装備・設定保持）  
**Testing**: go test ./...、モード別ロジックのテーブルテストとステート計算の単体テスト  
**Target Platform**: ターミナル（macOS/Linux想定）、ローカル実行  
**Project Type**: TUI単一バイナリ（cmd + internal）  
**Performance Goals**: 1セッション平均2分以内、UI描画/入力反応は体感遅延なし（~100ms以内）  
**Constraints**: 各モード開始時のみネットワーク使用、その後はオフライン進行／JSONスキーマ固定で出題取得  
**Scale/Scope**: シングルユーザー、1日数〜数十セッションの軽負荷

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Constitutionファイルはテンプレートのまま空欄のため、明示的な原則・ゲートは定義されていない。現時点で違反なしとみなし、設計完了後も同様に再確認する。

## Project Structure

### Documentation (this feature)

```text
specs/001-draft-english-quest-spec/
├── plan.md              # このファイル (/speckit.plan 出力)
├── research.md          # Phase 0 出力
├── data-model.md        # Phase 1 出力
├── quickstart.md        # Phase 1 出力
├── contracts/           # Phase 1 出力
└── tasks.md             # Phase 2 (/speckit.tasks で生成予定)
```

### Source Code (repository root)

```text
cmd/
└── english-quest/
    └── main.go

internal/
├── ui/           # Bubble Tea画面: top, town, battle, dungeon, tavern, spelling, listening, equipment, analysis, status
├── game/         # stats/exp/damage計算、レベルアップ/戦闘不能処理
├── services/     # geminiクライアント、出題取得
└── db/           # schema.sql, history永続化

tests/            # go test ./...（必要に応じて追加）
```

**Structure Decision**: 単一バイナリ構成（cmd + internal）。Gemini呼び出しはservicesに集約、UIロジックはui配下で画面ごとに分離、ゲーム計算はgame配下に集約。SQLiteはdb配下で管理。

## Complexity Tracking

（現時点で憲法上の違反や追加複雑性の正当化は不要）
