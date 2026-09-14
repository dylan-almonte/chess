# Chess Engine — Project Conventions

## Goal

Build a from-scratch UCI chess engine in Rust, plus a terminal UI client so development is visible (board, moves, UCI I/O, telemetry). Learn engine layers by specifying each one before implementing it.

## Monorepo layout

- **Rust engine** (repo root): `src/` library + UCI binary (`cargo run` / `target/release/chess`)
- **Go TUI client** (`tui/`): Bubble Tea app that speaks UCI to the Rust binary as a subprocess
- **OpenSpec** (`openspec/`): shared planning for both engine layers and the TUI client

## Tech stack

### Engine (Rust)

- Rust edition 2024
- Library + binary: `src/lib.rs` (engine) and `src/main.rs` (UCI entry)
- Standard library only for chess logic; no third-party chess crates (`shakmaty`, `chess`, etc.)

### TUI (Go)

- Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea) (+ Bubbles/Lip Gloss as needed)
- UCI client only: spawn/configure path to the Rust engine binary
- May use a small Go helper for local board display / move entry; engine moves always come from UCI `bestmove`

## Spec-driven rules

- Specs describe **observable** behavior: FEN, legal moves, perft, UCI replies, TUI panels/actions
- Representation and algorithms belong in `design.md`, not specs
- Engine scenarios map to named Rust tests; TUI scenarios map to Go tests (Bubble Tea model tests and/or scripted fake-engine tests)
- One OpenSpec change per capability (engine layer **or** TUI feature set)

## Engine layers (build order)

1. `board-representation` — bitboards + FEN
2. `move-generation` — legal moves + perft
3. `uci-protocol` — UCI stub that returns a legal move
4. `chess-tui` — Bubble Tea client (board, moves, UCI log, telemetry)
5. `evaluation` — material + piece-square tables
6. `search` — iterative deepening + alpha-beta

## Design defaults

- Square indexing: a1 = 0 … h1 = 7, a8 = 56 … h8 = 63 (little-endian rank-file)
- Board: 12 piece bitboards + derived occupancy; no magic bitboards until move-gen needs them
- Prefer make/unmake over copy-make once move generation exists
- TUI must not embed a second competing engine; it is a viewer/controller over UCI


## Backlog

- TUI: Unicode pieces + mouse click-to-move (`tui-polish`)
- Engine: search (iterative deepening + alpha-beta)
