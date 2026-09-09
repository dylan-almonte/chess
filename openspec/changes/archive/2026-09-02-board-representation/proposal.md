## Why

Every later engine layer (move generation, UCI, evaluation, search) needs a single source of position state and a standard way to load and save positions. FEN is the universal interchange format; without reliable parse/print and square addressing, nothing else can be tested or plugged into a GUI.

## What Changes

- Introduce core chess types: color, piece kind, piece, and square (algebraic ↔ index)
- Introduce a bitboard-backed position that holds pieces, side to move, castling rights, en passant, and clocks
- Parse FEN into a position and serialize a position back to FEN
- Expose a library API from `src/lib.rs`; keep `src/main.rs` a stub until the UCI change

## Capabilities

### New Capabilities

- `board-representation`: Position state, square addressing, and FEN parse/serialize behavior

### Modified Capabilities

- (none)

## Non-goals

- Move generation, make/unmake, legality checks
- UCI protocol, evaluation, or search
- Magic bitboards or attack tables
- Third-party chess crates

## Impact

- New modules under `src/` (`square`, `piece`, `board`, `fen`) and `src/lib.rs`
- `Cargo.toml` remains dependency-free for chess logic
- No existing behavior to break (greenfield crate)
