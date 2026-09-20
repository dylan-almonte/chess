## Context

See `proposal.md` for motivation and `specs/chess-tui/spec.md` for the new board-rendering and input requirements.

The Go TUI already keeps a display-only board of ASCII piece bytes, accepts typed UCI, and gates every human ply on the engine `legalmoves` set before sending `position` / `go`. `tea.NewProgram` currently uses only `WithAltScreen()`. Arrow keys go to the focused text input. `Board.Render()` always prints letters plus `.` / `,` empty placeholders.

This change stays inside `tui/`. The Rust UCI process boundary and four-panel layout stay as they are.

## Goals / Non-Goals

**Goals:**

- Keep piece storage as ASCII bytes; treat Unicode vs ASCII as a render-time choice
- Route typed, click-built, and cursor-built moves through one submit function so legality, `ApplyUCI`, `position`, and `go` cannot drift
- Make mouse hit-testing and keyboard cursor selection testable without a real terminal
- Preserve the existing board / move list / UCI log / telemetry / input layout

**Non-Goals:**

- Drag-and-drop, hover-only previews, or a promotion-piece picker
- Recalculating legal moves in Go
- Changing engine path resolution, handshake, or telemetry parsing

## Decisions

### Render-time glyph map, ASCII storage unchanged

`Board` continues to store `KQRBNP` / `kqrbnp`. Rendering accepts the current mode and maps those bytes to `♔♕♖♗♘♙` / `♚♛♜♝♞♟` when Unicode is on. Empty squares keep non-letter placeholders.

The model owns a `unicodePieces bool` defaulting to true and a Tab key (intercepted before the text input) that toggles it. The input placeholder documents `tab` as the glyph switch.

**Why:** Specs require a toggle that does not mutate position or the engine session. Changing stored piece bytes would break `ApplyUCI` and existing board tests.

**Alternative:** Store Unicode in the board. Rejected because move application and tests already speak ASCII letters.

**Alternative:** Typed commands `ascii` / `unicode`. Rejected as secondary; Tab does not consume a legal UCI token and will not be forwarded into the input.

### One `submitMove` for every input path

Extract the current `submitInput` body (legal-set gate, `ApplyUCI`, append history, `position`, `legalmoves`, `go`) into `submitMove(uci string)`. Typed Enter still reads the text input, ignores empty/`quit`, and calls `submitMove`. Mouse and cursor paths never write into the text input; they call `submitMove` with a constructed token.

Before submitting a two-square token `from+to`, resolve promotion:

1. If `from+to` is in the current legal set, use it
2. Else if `from+to+"q"` is in the legal set, use that
3. Else submit `from+to` so the existing illegal-move error path fires

**Why:** Issue #5 requires click-to-move to use the same play path as typed input. Queen-default promotion is the only extra rule; underpromotion stays typed.

**Alternative:** Duplicate `position`/`go` in the mouse handler. Rejected because illegal-move tests would have to cover three copies.

**Alternative:** Prompt for promotion piece. Rejected as out of scope (no picker UI).

### Mouse cell motion plus geometry hit-test

Enable `tea.WithMouseCellMotion()` next to `WithAltScreen()`. Handle only left-button press (`tea.MouseMsg`); ignore motion, wheel, and other buttons.

Keep a stable board geometry:

- Each square occupies 2 columns (glyph + pad) and 1 row
- Rank labels occupy the first 2 columns of the board content
- File labels occupy the row below rank 1
- The board box is the left pane of the top row: lipgloss normal border (1 cell) plus `Padding(0, 1)`, so content origin is `(2, 1)` while the playable view starts at `(0, 0)`

A dedicated helper maps `(mouseX, mouseY)` → square name or miss. Clicks on borders, labels, the move list, log, telemetry, or input are misses. Empty-square first clicks are ignored; a second click on the selected origin clears selection; any other dest click builds UCI and calls `submitMove`. Selection is ignored while `waiting` or `awaitingLegal`.

**Why:** Bubble Tea mouse events are cell coordinates. Fixed 2-column squares match the current ASCII layout and keep hit-testing unit-testable.

**Alternative:** Pixel-accurate or drag gestures. Rejected; terminals report cells, and drag-and-drop is a non-goal.

**Alternative:** A third-party chessboard widget. Rejected; the existing Lip Gloss board is enough.

### Arrow cursor uses Space, not Enter

Intercept Left/Right/Up/Down during `phasePlay` for a board cursor (default `e2` for White at startpos). Space confirms the square under the cursor using the same select / cancel / submit rules as a mouse click. Enter continues to submit the text-input value so typed UCI stays primary.

Highlight the cursor square and the selected origin with Lip Gloss backgrounds during `View`. Cursor movement does not submit a move.

**Why:** Enter is already the typed-submit key. Sending arrows to the text input would make cursor selection impossible.

**Alternative:** Unfocus the input and use Enter to confirm squares. Rejected because it demotes typed UCI, which must stay primary.

### Tests drive `Update` with synthetic mouse and key messages

Board glyph tests cover Unicode default, ASCII toggle, and mode surviving `ApplyUCI`. Model tests reuse the fake engine:

- `MouseMsg` click `e2` then `e4` asserts the same board, move list, and outbound `position`/`go` as typed `e2e4`
- Illegal dest `e2` then `e5` asserts no mutation and no new `position`/`go`
- Re-click origin cancels
- Clicks while waiting after `e2e4` are ignored
- Arrow-to-`e2`, Space, arrow-to-`e4`, Space matches the click happy path

Hit-testing is tested against the documented origin, not against a live terminal. The existing real-engine integration test remains a smoke check that the shared submit path still talks UCI; it does not need to synthesize OS mouse events.

## Risks / Trade-offs

- [Chess glyphs render double-width in some fonts, shifting click columns] → Keep each square in a fixed 2-column slot; ASCII toggle remains the fallback; hit-test tests lock the geometry
- [Arrow interception removes in-field caret editing] → Acceptable: UCI tokens are ≤5 characters; typed entry still uses Backspace
- [Mouse support varies by terminal] → Keyboard typed UCI and cursor selection remain full play paths
- [Layout tweaks could desync click mapping] → Isolate origin/size in one helper and test it; do not scatter magic offsets in `Update`

## Migration Plan

1. Add failing board/render and model input tests.
2. Implement glyph rendering and Tab toggle.
3. Extract `submitMove` / promotion resolve, then mouse and cursor selection.
4. Enable mouse cell motion on the program.
5. Roll back is a TUI-only revert; the engine binary and UCI contract are unchanged.

## Open Questions

- Exact highlight colors for cursor vs selected origin can be chosen during apply as long as both are visible.
