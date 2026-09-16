# Chess Engine — Project Conventions

## Goal / motivation

Build a from-scratch **chess engine lab**: a shared rules + UCI chassis so different **move-selection methods** (“brains”) can be tried and compared — classical search/eval, later MCTS, policy nets, RL, etc.

The Go TUI makes development visible (board, moves, UCI I/O, telemetry). Learn each layer by specifying it before implementing. Classical αβ + hand eval is **brain #1**, not the final architecture.

## Scalability model (intentional)

Keep the project scalable for method experiments with two contracts only:

1. **Shared chassis** — board representation, FEN, move generation, make/unmake, perft, UCI session I/O. Every brain reuses this; do not fork rules per method.
2. **Swappable brain** — given a position + search/play limits, return a move (and optional score / PV / node stats). UCI `go` dispatches to the selected brain; the TUI stays brain-agnostic.

Do **not** prematurely build plugin systems, separate crates per brain, or traits until a second brain exists. When brain #2 lands, introduce a thin dispatch (e.g. `choose_move` / `src/brain/`) behind UCI.

## Monorepo layout

- **Rust engine** (repo root): `src/` library + UCI binary (`cargo run` / `target/release/chess`)
- **Go TUI client** (`tui/`): Bubble Tea app that speaks UCI to the Rust binary as a subprocess
- **OpenSpec** (`openspec/`): shared planning for chassis layers, brains, and the TUI

## Tech stack

### Engine (Rust)

- Rust edition 2024
- Library + binary: `src/lib.rs` (engine) and `src/main.rs` (UCI entry)
- Standard library only for chess **chassis** logic; no third-party chess crates (`shakmaty`, `chess`, etc.)
- Future neural / RL brains may need extra deps or a separate binary; that does not relax stdlib-only for board/movegen

### TUI (Go)

- Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea) (+ Bubbles/Lip Gloss as needed)
- UCI client only: spawn/configure path to the Rust engine binary
- May use a small Go helper for local board display / move entry; engine moves always come from UCI `bestmove`

## Spec-driven rules

- Specs describe **observable** behavior: FEN, legal moves, perft, UCI replies, TUI panels/actions
- Representation and algorithms belong in `design.md`, not specs
- Engine scenarios map to named Rust tests; TUI scenarios map to Go tests (Bubble Tea model tests and/or scripted fake-engine tests)
- One OpenSpec change per capability (chassis layer, one brain/method, **or** TUI feature set)

## Engine layers (build order)

Chassis and first brain:

1. `board-representation` — bitboards + FEN
2. `move-generation` — legal moves + perft
3. `uci-protocol` — UCI stub that returns a legal move
4. `chess-tui` — Bubble Tea client (board, moves, UCI log, telemetry)
5. `evaluation` — material + piece-square tables (classical leaf scorer)
6. `search` — iterative deepening + alpha-beta (classical brain #1)

Later brains are separate changes (not extensions that rewrite the chassis).

## Design defaults

- Square indexing: a1 = 0 … h1 = 7, a8 = 56 … h8 = 63 (little-endian rank-file)
- Board: 12 piece bitboards + derived occupancy; no magic bitboards until move-gen needs them
- Prefer make/unmake over copy-make once move generation exists
- TUI must not embed a second competing engine; it is a viewer/controller over UCI
- Prefer one UCI `go` → brain dispatch site; brains return a shared result shape (move, score, PV, nodes) when practical

## Backlog

### Bugs

- **Illegal moves accepted** — Neither the human nor the engine should play or apply illegal moves. Today the TUI can accept typed UCI that isn’t legal for the position, and/or the engine may apply illegal tokens from `position … moves` without rejecting them. Fix so:
  - Human input is rejected (clear error) unless the move is in the legal set for the current position
  - Engine `position … moves` only applies moves that resolve to a legal move; stop/reject on the first illegal token
  - Engine `bestmove` remains a legal move (or `0000` when none) — search must not emit illegal moves
  - OpenSpec change when tackling (e.g. `legal-move-enforcement`) covering engine and/or TUI scenarios + tests

### Lab / multi-brain (tackle when ready)

- Document and optionally introduce a thin `choose_move(pos, limits) → SearchResult` (or equivalent) so UCI does not call classical search by name forever
- When adding brain #2: `src/brain/` (or similar) with classical vs new method + a selection switch (`setoption`, CLI flag, or env)
- Future methods to try (each its own OpenSpec change): MCTS; policy-net move choice; RL / self-play training + inference
- Optional: second UCI binary for a heavy ML stack so the TUI can A/B via `CHESS_ENGINE` path

### Classical / TUI polish

- Richer handcrafted eval (king safety, pawn structure, …) — still classical brain, still `evaluate`
- TUI: Unicode pieces + mouse click-to-move (`tui-polish`)
- Search hardening as needed (quiescence, TT, move ordering) — still classical brain
