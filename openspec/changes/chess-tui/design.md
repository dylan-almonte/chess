## Context

Rust engine with archived `board-representation`, `move-generation`, and `uci-protocol` lives at repo root. See proposal.md for motivation. Monorepo conventions are in `openspec/project.md`. The TUI is a new Go client under `tui/`.

## Goals / Non-Goals

**Goals:**

- Bubble Tea app: board, move list, UCI log, telemetry
- Drive engine via real UCI subprocess
- Testable without a human terminal when using a fake engine script
- Sensible defaults: engine path `../target/release/chess` or `CHESS_ENGINE` env var

**Non-Goals:**

- Mouse-only UX polish, clocks, PGN import/export (can follow later)
- Parsing every UCI extension; ignore unknown inbound lines except logging them
- Embedding Rust via cgo/FFI

## Decisions

### Process boundary: UCI subprocess

- TUI owns stdin/stdout pipes to `chess`
- On start: `uci` → wait `uciok`; `isready` → wait `readyok`; optional `ucinewgame` + `position startpos`
- Human move: append to move list → send `position startpos moves …` → `go` → apply `bestmove`
- **Why:** Matches how real GUIs work; keeps engine language-agnostic
- **Alternative:** call Rust as a library — rejected (wrong language, couples releases)

### Local board state for display

- Maintain FEN / board in Go for rendering and validating human input before sending UCI
- Prefer a small dedicated helper (hand-rolled minimal apply for legal typed moves, or a lightweight Go chess package used **only** in `tui/` for UI state)
- Engine remains source of truth for engine moves via `bestmove`; if apply fails, show error and keep log
- **Why:** Bubble Tea needs immediate UI state; round-tripping board graphics through the engine would require extra non-UCI commands

### Layout (four regions)

```text
┌──────────────┬─────────────┐
│ Board        │ Move list   │
├──────────────┴─────────────┤
│ UCI log (scroll)           │
├────────────────────────────┤
│ Telemetry (info)           │
│ input: e2e4 / quit         │
└────────────────────────────┘
```

- Lip Gloss for styling; Bubbles viewport/textarea for log and input

### Fake engine for tests

- Provide `tui/testdata/fake_engine.go` or a scripted stub that answers `uci`/`isready`/`go` with fixed lines (including optional `info`)
- Model/unit tests drive the Bubble Tea `Update` path with synthetic engine messages
- **Why:** CI should not require a built Rust binary for pure TUI logic tests; one integration smoke test can use the real binary when present

### Telemetry parsing

- Parse common `info` tokens (`depth`, `score cp|mate`, `nodes`, `nps`, `pv`) into a struct for the panel
- Unparsed remainder still appears in the UCI log
- Stub engine → idle panel (required)

## Risks / Trade-offs

- [Go board rules drift from Rust] → Mitigation: only trust engine for its moves; keep human-move validation strict; add integration smoke against real engine
- [Blocking I/O stalls Bubble Tea] → Mitigation: read engine stdout in a goroutine; send Msg events into the program
- [Path / cwd confusion] → Mitigation: document `CHESS_ENGINE`; resolve relative to module/repo root in design tasks

## Open Questions

- Exact Go chess helper vs hand-rolled display state — choose during apply; does not change specs if board/move observables hold.

## Follow-ups (not this change)

Tracked in `tasks.md` §6: Unicode piece glyphs, mouse click-to-move, optional cursor selection. Spin a new OpenSpec change (e.g. `tui-polish`) when implementing.
