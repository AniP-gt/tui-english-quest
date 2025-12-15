# Issue: TUI index out of range when advancing past last item

## Summary
Several TUI screens (e.g. Listening Cave) can panic with `runtime error: index out of range` when a model advances past the last item and a `View()` call accesses the slice without bounds checks. This leads to application crash.

## Reproduction
1. Start the app and navigate to a mode with sequential items (e.g. Listening Cave).
2. Let the screen fetch 5 items and answer them sequentially.
3. After answering the final item, press Enter to continue.

Observed: the app sometimes panics with `runtime error: index out of range` (stacktrace shows panic in `internal/ui/*_tui.go` View methods).

## Root Cause (investigated)
`View()` methods of some TUI models index into slices (e.g. `items[currentIndex]`) without verifying `currentIndex < len(items)`. Bubble Tea can call `View()` between `Update()` changing the index and the state transition, causing `View()` to access an out-of-range index.

## Affected files (observed during investigation)
- `internal/ui/listening_tui.go` (fixed in branch `feature/listening-cave`)
- Potentially other screens with similar patterns: `internal/ui/dungeon.go`, `internal/ui/battle.go`, `internal/ui/tavern_tui.go`, etc.

## Suggested Fix
- Add defensive checks in `View()` so it never reads `items[currentIndex]` when `currentIndex >= len(items)`.
- Alternatively, maintain an explicit `sessionComplete` boolean and use it to guard access in `View()`.
- Add unit tests for model lifecycle to ensure `View()` does not panic after advancing to the end.

## Notes
- A fix for `internal/ui/listening_tui.go` has been implemented on branch `feature/listening-cave`.
- Please triage and apply similar fixes to other TUI screens as needed.
