# Chess Engine — Project Conventions

## Goal

Build a from-scratch UCI chess engine in Rust. No third-party chess crates (`shakmaty`, `chess`, etc.). Learn engine layers by specifying each one before implementing it.

## Tech stack

- Rust edition 2024
- Library + binary: `src/lib.rs` (engine) and `src/main.rs` (UCI / CLI entry)
- Standard library only for chess logic; tooling crates (e.g. clap) allowed later if needed

## Spec-driven rules

- Specs describe **observable** chess/UCI behavior: FEN strings, legal-move lists, perft node counts, UCI replies
- Representation and algorithms (bitboards, magics, search heuristics) belong in `design.md`, not specs
- Every spec scenario MUST map to a named Rust test (`#[test] fn scenario_name_snake_case`)
- One OpenSpec change per engine layer

## Engine layers (build order)

1. `board-representation` — bitboards + FEN
2. `move-generation` — legal moves + perft
3. `uci-protocol` — UCI stub that returns a legal move
4. `evaluation` — material + piece-square tables
5. `search` — iterative deepening + alpha-beta

## Design defaults

- Square indexing: a1 = 0 … h1 = 7, a8 = 56 … h8 = 63 (little-endian rank-file)
- Board: 12 piece bitboards + derived occupancy; no magic bitboards until move-gen needs them
- Prefer make/unmake over copy-make once move generation exists
