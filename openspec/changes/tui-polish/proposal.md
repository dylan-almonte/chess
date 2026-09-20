## Why

The TUI already plays against the UCI engine, but the board still uses ASCII letters and the only way to move is typing long-algebraic UCI. That makes interactive play harder than it needs to be. GitHub issue #5 and the original `chess-tui` follow-ups call for Unicode glyphs and click-to-move while keeping typed input primary.

## What Changes

- Render chess pieces as Unicode glyphs by default (♔♕♖♗♘♙ / ♚♛♜♝♞♟), with a toggle back to the current ASCII letters
- Enable mouse click-to-move: click origin, click destination, build a UCI token, and submit it through the same legal-move play path as typed input
- Add keyboard cursor / arrow-key square selection as a keyboard alternative to typing UCI coordinates
- Keep typed UCI entry, engine handshake, move list, UCI log, telemetry, and illegal-move rejection unchanged

## Non-goals

- Changing engine search, evaluation, legal-move generation, or the `legalmoves` UCI extension
- Redesigning the four-panel layout, adding clocks, PGN import/export, or a graphical GUI
- Mouse-only UX, drag-and-drop, or animations
- Teaching the Go display board to generate legal moves; the engine remains the authority
- Changing which side the human plays (tracked separately as choose-colors)

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `chess-tui`: Board rendering gains Unicode pieces plus an ASCII toggle; human moves may be built by mouse clicks or keyboard cursor selection in addition to typed UCI, using the existing play path

## Impact

- Go board rendering and Bubble Tea input handling under `tui/internal/board` and `tui/internal/ui`
- Mouse cell motion enabled on the Bubble Tea program; keyboard text input remains available
- Go model tests against the fake engine, plus a manual or existing real-engine check that click-built moves still drive `position` / `go`
- No Rust engine, UCI protocol, or chassis spec changes
