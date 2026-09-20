## Context

See `proposal.md` for motivation and `specs/chess-tui/spec.md` for behavior. The TUI already keeps an ASCII display board, gates moves on `legalmoves`, and always plays Human-White then Engine-Black from startpos. Squares are a fixed 2×1 cell; `SquareAtCell` assumes that geometry. The engine already accepts `position fen <6 fields>` and silently ignores a malformed FEN.

This change stays in `tui/` except for using that existing FEN command.

## Goals / Non-Goals

**Goals:**

- Compute square width/height from `WindowSizeMsg` and keep hit-testing in lockstep
- Add a setup mode that edits the display board, then commits via `position fen` + `legalmoves`
- Drive `go` only when the side to move is assigned Engine
- Keep the four-panel layout; shrink the log first when the board needs room

**Non-Goals:**

- Local legal-move generation or a second rules engine
- Engine `info string` for invalid FEN (TUI only sends well-formed six-field FENs)
- Board flip, FEN paste, or per-slot named brains

## Decisions

### Scale squares from leftover viewport, then shrink the log

After borders, move-list width (~28), rank/file labels, and a minimum log/telemetry/input strip, divide leftover width by 8 and leftover height by 8. Clamp each square to at least 2 columns × 1 row (today’s cell) and at most 5 columns × 3 rows. Center the piece glyph in the cell.

`WindowSizeMsg` updates `cellW`/`cellH` and the board origin used by `SquareAtCell`. Click mapping must use the same numbers as `View`.

At ≥80×28 the leftover budget must yield at least 3×2 cells (spec). Prefer cutting log viewport height over keeping a compact board.

**Why:** The user asked for a larger board that still fits a small terminal.

**Alternative:** A fixed 3×2 board always. Rejected because 60×20 would overflow or clip ranks.

**Alternative:** A user-typed scale factor. Rejected; terminal resize is the control.

### Setup mode is a session phase, not a second program

`setup` (typed command) copies the current display board into an editable buffer and stops the play loop (`go`, click-to-move, cursor-submit). Keys `KQRBNP` / `kqrbnp` choose the brush; `x` is eraser. A board click or Space on the cursor places the brush or clears the square. `side` toggles White/Black to move. `clear` empties pieces. `startpos` restores the standard array. `play` attempts to leave setup.

Setup never sends `go` and never appends the play move list.

**Why:** Matches the chosen visual editor without adding FEN paste.

**Alternative:** Always-on edit while playing. Rejected because it would fight click-to-move.

### Commit setup as a constructed FEN, then reset move history

`play` requires exactly one `K` and one `k`. On success the TUI:

1. Builds `placement side castle ep half full` from the buffer
2. Sends `ucinewgame` then `position fen <fen>` then `legalmoves`
3. Replaces the display board, clears the move list, and enters play

FEN extras:

- Castling: add `K`/`Q`/`k`/`q` only if that king and rook still sit on `e1`/`h1`/`a1` or `e8`/`h8`/`a8`
- En passant: always `-`
- Clocks: `0 1`

**Why:** The engine already parses six-field FEN. Inferring castling from home squares covers the common edited-startpos case without a rights editor.

**Alternative:** Send `position startpos` plus a move list. Rejected; arbitrary setups are not reachable that way.

**Alternative:** Teach Go `ApplyUCI` to be a rules engine. Rejected by project convention.

### Player slots gate `go`, not the engine binary

`white` and `black` are each `human` or `engine` (default Human / Engine). Typed commands: `white human`, `white engine`, `black human`, `black engine`. Shown on the input/status line.

After legal moves are ready, if the side to move’s slot is Engine, send the existing engine-side `legalmoves` + `go depth 4` path. If it is Human, wait for typed / click / cursor submit. After an accepted ply, repeat. Engine vs engine is that loop with no human turn. Changing a slot mid-play takes effect on the next turn.

**Why:** Issue #12 step 1: one UCI process, independently assignable sides.

**Alternative:** Two engine processes. Rejected until a second binary/brain exists.

### Layout keeps four regions; status is a fifth text line

```text
┌──────────────┬─────────────┐
│ Scaled board │ Move list   │
├──────────────┴─────────────┤
│ UCI log                    │
├────────────────────────────┤
│ Telemetry                  │
│ White: Human  Black: Engine│
│ setup Q  side: w  | input  │
└────────────────────────────┘
```

Hit-testing still treats only the board pane as squares. Palette/status clicks are non-goals; keys and typed commands are enough.

## Risks / Trade-offs

- [Scaled cells desync click mapping] → One `boardMetrics` struct shared by `View` and `SquareAtCell`; tests cover 2×1 and 3×2
- [Engine silently drops a bad FEN] → Refuse `play` without both kings; emit only 8-rank placement + inferred rights
- [Inferred castling is wrong for a crafted puzzle] → Document the home-square rule; a later FEN-paste change can override
- [Self-play floods `go`] → Keep the current depth-limited `go`; human can quit

## Migration Plan

1. Add failing scale, setup, and slot tests.
2. Implement metrics + render/hit-test.
3. Add setup phase and FEN commit.
4. Gate `go` on player slots.
5. TUI-only revert; engine FEN handling is unchanged.

## Open Questions

- Exact status-line wording and piece-brush highlight color can be chosen during apply.
