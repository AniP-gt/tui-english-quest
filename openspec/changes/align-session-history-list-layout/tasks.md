# Tasks: align-session-history-list-layout

この変更を実装するための作業項目一覧。

1. 追加: 共通の "aligned list" ヘルパーを `internal/ui/components/list.go` に実装
   - 目的: ラベルと値、箇条書きのインデントを統一して整列する小さなユーティリティを提供する
   - バリデーション: `internal/ui/status.go` と `internal/ui/history.go` が期待通りの表示を行うスクリーンショット的な手動検証

2. 修正: `internal/ui/status.go` を共通ヘルパーで表示するように変更
   - 具体的には Achievements の箇条を `components.RenderBulletList` 形式に変換

3. 修正: `internal/ui/history.go` のセッション一覧出力を `components.RenderTable` か `RenderAlignedRow` ヘルパーで整列
   - 既存の `fmt.Sprintf` 固定幅フォーマットは残すが、バレットやインデントに対する一貫性を持たせる

4. テスト: `internal/ui` 下にテキスト出力スナップショットテストを追加（オプション）
   - 短いユニットテストで `RenderBulletList` の出力が特定のインデントで始まることを確認

5. ドキュメント: `openspec/changes/align-session-history-list-layout/design.md` に設計意図と非機能要件を記載

実行順序

- 1 → 2,3 → 4 → 5

備考

- 変更は UI 表示ロジックのみでデータ書き込みや DB には触れない
- 端末幅による崩れは lipgloss に委ねる。必要なら後続で折り返しルールを追加する
