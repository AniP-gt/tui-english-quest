# Design Notes: align-session-history-list-layout

目的の再掲

- テキストUI（Bubble Tea + Lip Gloss）におけるラベル行と箇条書きのインデントを一貫させ、見た目の乱れを解消する。

設計方針

- 最小限の API を `internal/ui/components` に追加する。
  - `RenderKeyValue(label string, value string, labelWidth int) string` — ラベル幅で左揃えに揃える
  - `RenderBulletList(items []string, indent int) string` — 箇条書きを指定インデントで出力
  - `RenderAlignedRow(cols []string, widths []int) string` — テーブルライクに列を整列して出力（履歴の行向け）

- 既存の `status.go` と `history.go` は上記 API に置き換えやラップで対応する。可能な限り既存の `fmt.Sprintf` を直接置き換えるのではなく、箇条や可変幅要素のみをラップして互換性を保つ。

- レンダリングは文字列結合ベースで行い、最終的に lipgloss のスタイルで囲む。端末幅やヘッダー幅はこれまで通り `lipgloss.Width(header)` を参照して境界線を描画する。

アクセシビリティ & 国際化

- 箇条書きやラベルは多言語対応（i18n）されている既存文字列をそのまま使う。
- 全角文字が混在する日本語表示を考慮して、揃えはバイト幅ではなく文字幅（`lipgloss.Width`）に依存する実装を推奨するが、まずは `fmt` 固定幅の既存コードに合わせた最小実装とする。

テスト戦略

- `RenderBulletList` の出力を比較するユニットテストを用意し、先頭に `  - ` が付いていること、すべての行が同じ indent であることを確認する。
- 履歴一覧はサンプルデータで手動確認を推奨（スクリーンショットベースの自動化は現状の CI では不要）。

将来の拡張

- 必要であれば、列幅自動計算や折り返しポリシーを `components` に追加することで、より堅牢なテーブル描画を実装できる。
