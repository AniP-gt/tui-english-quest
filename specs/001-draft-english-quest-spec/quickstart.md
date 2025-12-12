# Quickstart - TUI English Quest v2.0

1) 前提
- Go 1.21+、SQLite利用可。
- `configs/.env.example` を参考に `GEMINI_API_KEY` と `DB_PATH` を設定。

2) セットアップ/ビルド
- `go mod tidy`
- `go build ./...`
- 起動テスト: `go run ./cmd/english-quest`

3) セッションフロー
- 起動 → Top → Town。
- 任意モード選択時にGeminiへ5問まとめてリクエスト（contracts/gemini-contracts.mdスキーマ）。
- 取得後はオフラインで5問実施し、リザルトでEXP/HP/Gold/Defense/コンボを反映。

4) 装備と弱点分析
- 装備効果をモード別報酬/被ダメに乗算（ExpBoost, DamageReduction）。
- 履歴50〜200問を集計し弱点分析を生成、TownのAIアドバイスに表示。
- 新しい画面: History（セッション履歴）、Equipment（装備変更）、Status（プレイヤー成長）、Analysis（弱点分析詳細）。

5) 多言語設定
- UI言語/解説言語/問題言語を独立設定。切替後は1画面再描画以内に反映。

6) テスト観点
- ステータス計算のテーブルテスト、モード別リザルト計算、Geminiレスポンス検証（5件・スキーマ）を go test でカバー。
- 新機能: 履歴表示、装備効果計算、弱点分析生成。
