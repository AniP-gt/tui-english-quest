# TUI English Quest（日本語版）

<p align="right">
  <a href="../../README.md">English version</a>
</p>

## 概要

TUI English Quest はターミナルで動く RPG 風の英語学習アプリです。タイトル画面から街に出向き、単語バトル・文法ダンジョン・会話タバーン・スペリングチャレンジ・リスニング問題といった学習モードを選択し、装備・AI分析・履歴・ステータス・設定画面も活用することで、短時間で完結するセッションを繰り返します。各セッションでは Gemini（`gemini-2.5-flash`）から 5 問が生成され、オフラインで回答したのち EXP・HP・コンボ・ストリーク・ゴールドが更新され、弱点分析や履歴に記録されます。

<img width="735" height="412" alt="Screenshot 2025-12-19 at 14 20 38" src="https://github.com/user-attachments/assets/fd7856f9-33bb-4af2-b9f1-6664194c89c2" />

<img width="732" height="530" alt="Screenshot 2025-12-19 at 14 21 00" src="https://github.com/user-attachments/assets/e9742ebe-5481-4fe9-8859-78ddf3b31d48" />



## インストールと依存

1. **Go 1.21+ と SQLite** が必要です。Go の開発環境を整え、必要であれば `sqlite3` コマンドもインストールしてください。
2. **ビルド手順**:`specs/001-draft-english-quest-spec/quickstart.md` を参照して:
   - `go mod tidy`
   - `go build ./...`
   - `go run ./cmd/english-quest`
3. **Gemini の準備**: `gemini-2.5-flash` モデルを使うため、Google Cloud にて API キーを作成し、`GEMINI_API_KEY` に設定します。
   - 英語ページ: https://ai.google.dev/gemini-api/docs/api-key?hl=en
   - 日本語ページ: https://ai.google.dev/gemini-api/docs/api-key?hl=ja
4. **環境変数**: `configs/.env.example` をコピーして以下を設定／エクスポートします。
   - `GEMINI_API_KEY`（必須）
   - `DB_PATH`（既定は `./db.sqlite`、任意の場所に変更可能）
   - `LOG_LEVEL`（任意: `info` / `debug`）
   - `SPEAK_CMD`（任意: TTS を自前コマンドに差し替える）
5. **実行時設定**: Unicode, LangPref などを含む `config.json` は Unix では `~/.local/share/tui-english-quest/`、Windows なら `%AppData%\tui-english-quest\` に保存されます。言語、API キー、`QuestionsPerSession`、`ProfileID` が格納されます。

## 設定と環境

- `config.Config` のフィールド:
  - `LangPref`: `en` または `ja`。設定画面で UI/ヘルプの切り替えを即時反映します。
  - `ApiKey`: Gemini API キーを保存すると、起動時に環境変数入力を省略できます。
  - `QuestionsPerSession`: モードごとに取得する問題数（デフォルト 5、設定画面で 10/20/30/50 を選択可）。
  - `ProfileID`: 初回起動で生成され、永続的に記録されます。
- データベーススキーマ（`internal/db/schema.sql`）:
  - `profiles`: 名前/クラス/レベル/EXP/HP/攻撃/防御/コンボ/ストリーク/ゴールド/装備バフ/言語
  - `sessions`: モード別の正答数、EXP/HP/Gold 増減、コンボ、戦闘不能、レベルアップフラグ
  - `equipment`・`analysis`: 装備情報と AI 分析レポート
- TTS: `SPEAK_CMD` 未設定で `say` がない場合は音声再生をスキップします。`SPEAK_CMD` に `espeak '%s'` のようなコマンドを与えることもできます。

## ゲームフローとモード

1. **トップと街**: タイトル画面から新しい冒険を始め、街でステータスバーとメニュー（`j/k`, 矢印, Enter で操作）を見ながらモードを選択します。
2. **モードの特徴**（各 5 問を `services.FetchAndValidate` で取得）:
   - **単語バトル**: 正解でコンボと EXP（レベルに応じて増幅）、不正解で `AllowedMisses` から算出したダメージとコンボリセット。
   - **文法ダンジョン**: 似た設計だが正解で防御が増し、ダメージが若干軽減されます。
   - **会話タバーン**: Gemini から NPC の台詞・評価ルーブリックをもらい、5 ターンの会話を `BatchEvaluateTavern` に評価させ、HP を減らさず成功/普通/失敗に対して EXP/Gold を配分。
   - **スペリングチャレンジ**: Tab で記述式と選択式を切り替え。完全一致で +5 EXP、近似一致で +2 EXP（軽微な HP ダメージ）、外しで専用 HP ダメージ。
   - **リスニング問題**: `r` で再生する音声に対して 4 選択肢。誤答で HP ダメージが発生し、他モードと同じく `ApplyDamage` で処理。
3. **補助画面**:
   - **装備**: 武器/防具/指輪/お守りを装備し、`ExpBoost` や `DamageReduction` をモード別に適用。
   - **AI分析**: `services.AnalyzeWeakness` が直近 50〜200 問を集計し、要約・弱点/強み・行動計画を Town/Analysis に表示。
   - **履歴**: `sessions` テーブルから日時・モード・EXP/HP/Gold 変化・最高コンボ・戦闘不能/レベルアップフラグを一覧化。
   - **ステータス**: `game.Stats`（名前/クラス/レベル/EXP/次の閾値/HP/最大HP/コンボなど）＋実績表示。
   - **設定**: 言語切替・API キー変更・問題数変更・保存。
4. **リザルト**: モード終了後、EXP/HP/Gold/防御差分・レベルアップ・戦闘不能のメッセージを表示し、Enter で街に戻る。

## ステータス＆進行

- HUD は `game.Stats` の `Level`, `Exp`, `Next`, `HP`, `MaxHP`, `Attack`, `Defense`, `Combo`, `Streak`, `Gold`, `ExpBoost`, `DamageReduction` を表示します。
- **EXP 曲線**: `ExpToNext(level)` はレベル 99 まで `30 + 5*(level-1)`、それ以降は `500 + 10*(level-100)` を返します。
- **Max HP**: `MaxHPForLevel` で計算され、レベルアップで再計算・HP 全回復、Attack +2、Defense +1。
- **戦闘不能**: HP 0 になると `ApplyFaintPenalty` で EXP −5（最小 0）・HP を MaxHP の 50% に復帰。
- **回復**: レベルアップや Town→モード遷移時に `game.FullHeal` で HP を最大まで回復。
- **コンボ＆ストリーク**: 正解でコンボ継続、ミスでリセット。ストリークは成功日数の連続記録。
- **装備バフ**: `ExpBoost`/`DamageReduction` 付き装備で各モードの報酬/ダメージに乗算。

## 操作

- メニュー移動: `j/k` または ↑↓、Enter で決定。
- Tab はスペリングの記述式⇔選択式の切り替えです。
- `1`〜`4` でスペリング/リスニングの選択肢を選択。
- `r` でリスニングの音声を再生。
- `Esc`, `q`, `Ctrl+C` で画面を閉じたり終了。
- 街メニューで装備・AI分析・履歴・ステータス・設定・終了にアクセス。

## AI分析・履歴・装備

- **Gemini 契約**: 各モードは `specs/.../contracts/gemini-contracts.md` の JSON フォーマットを遵守します。
- **弱点分析**: `services.AnalyzeWeakness` で最近のセッションを集計し、Town と Analysis 画面にまとめます。
- **履歴**: `db.sessions` に日時・モード・正答数・EXP/HP/Gold 差分・コンボ・戦闘不能/レベルアップを記録。
- **装備**: 武器・防具・指輪・お守りの各スロットに `effect_type`（`ExpBoost`/`DamageReduction`）、`effect_value`, `target_mode` を設定し、セッション報酬やダメージ計算に乗算します。

## トラブルシューティング & テスト

- **Gemini 失敗**: 問題取得やタバーン評価が失敗するとエラー表示のみで、ステータスは変更されません。
- **HP 0**: 即座に faint penalty（EXP −5、HP＝MaxHP/2）を適用し、履歴に戦闘不能フラグを残します。
- **途中中断**: 問題中に `Esc`/`q` で離脱するとそのセッションは破棄され、街から再開します。
- **音声なし**: `SPEAK_CMD` 未設定＋`say` 不在のときは音声をスキップし、メッセージを表示。
- **テスト**: `go test ./...` でステータス計算・モード結果・Gemini バリデーションを検証。

## 参考資料

- 仕様: `specs/001-draft-english-quest-spec/spec.md`
- クイックスタート: `specs/001-draft-english-quest-spec/quickstart.md`
- Gemini 契約: `specs/001-draft-english-quest-spec/contracts/gemini-contracts.md`
- データベース: `internal/db/schema.sql`
- AI分析ロジック: `internal/services/analysis.go`
- UI: `internal/ui/` 以下のモデルおよび `internal/ui/components`
- Gemini クライアント: `internal/services/gemini.go`
