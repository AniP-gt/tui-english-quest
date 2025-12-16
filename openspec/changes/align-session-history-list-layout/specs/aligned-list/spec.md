# Spec Delta: aligned-list

Change: align-session-history-list-layout

## MODIFIED Requirements

### Requirement: Consistent Aligned List Rendering

The system SHALL correctly align list items and key-value pairs in the status and history screens.

#### Description
セッション履歴やプレイヤーステータス画面で、ラベル／値の行と箇条書きのインデントが揃うこと。箇条書きのハイフン（`-`）が隣接テキストと同じ行に表示されず、常に項目毎に新しい行で左揃えされること。

#### Rationale
現在、箇条書きや一部ラベルのインデントが不揃いで視認性を損なうため。テキスト UI における一貫した行揃えは可読性とプロダクト品質に直結する。

#### Acceptance Criteria
- `StatusModel.View()` の出力において、"Achievements:" の下の各項目が同じインデントで `- ` を先頭に持つ行として表示される。
- `HistoryModel.View()` の出力で、セッション一覧の各行の先頭にカーソル位置（`> `）が挿入され、それ以外の列は等幅で整列している。

#### Scenarios
- Scenario: No sessions
  - Given: user has no session records
  - When: History screen is opened
  - Then: localized message `history_no_sessions` is shown centered under the title (existing behaviour)

- Scenario: Achievements list alignment
  - Given: user has two achievements
  - When: Status screen is opened
  - Then: Each achievement renders on its own line prefixed with `  - ` and aligned to the same indent columns

- Scenario: Session list alignment with cursor
  - Given: History contains 3 session records
  - When: History screen is opened and the second item is selected
  - Then: the second line starts with `> ` and subsequent columns align visually with the other rows

## ADDED Requirements

### Requirement: Component helper APIs

The system SHALL provide new rendering helpers in the components package to standardize UI list rendering.

#### Description
`internal/ui/components` に以下の小さなレンダリングヘルパーを追加すること。
- `RenderBulletList(items []string, indent int) string`
- `RenderAlignedRow(cols []string, widths []int) string`（オプションでパディング合計を受け取る）

#### Acceptance Criteria
- ヘルパーは単体テストで `RenderBulletList([]string{"A","B"}, 2)` が各行先頭に2つ分のスペースと `- ` を含むことを確認できる。
- `HistoryModel`/`StatusModel` がヘルパーを使って正しく表示を出力する（実装で確認）。
