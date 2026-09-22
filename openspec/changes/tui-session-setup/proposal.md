## Why

The TUI always starts from the standard position on a compact board and always treats the human as White against one engine. That blocks training from custom positions, editing a setup by hand, and choosing which side the agent plays (GitHub issue #12).

## What Changes

- Scale board squares with the terminal so the grid is larger than the current one-cell glyph layout when space allows
- Add a visual setup mode: place and remove pieces on the board, set the side to move, then start play from that position
- Assign White and Black independently to Human or Engine (play as Black, human vs engine, or engine self-play)
- Keep typed UCI, click-to-move, cursor selection, Unicode/ASCII toggle, and the existing UCI `position` / `legalmoves` / `go` play path

## Non-goals

- Pasting arbitrary FEN strings (setup is visual only)
- Flipping the board for Black’s perspective
- Named brains / second engine binaries per slot (issue #12 step 2, after brain dispatch)
- Drag-and-drop, clocks, PGN, or a graphical GUI
- Changing search, evaluation, or legal-move generation
- Teaching the Go display board to generate legal moves

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `chess-tui`: Board rendering scales with terminal size; a setup mode can build a custom position; White and Black player slots choose Human or Engine

## Impact

- Go board rendering and Bubble Tea model under `tui/internal/board` and `tui/internal/ui`
- Session sync uses existing UCI `position fen …` plus `legalmoves` / `go`; no second engine in the TUI
- Go model tests with the fake engine for setup, scaled layout, and player-slot turns
- No required Rust chassis change if the TUI only sends well-formed six-field FENs
