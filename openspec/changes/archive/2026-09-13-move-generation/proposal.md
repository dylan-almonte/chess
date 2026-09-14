## Why

Board representation and FEN are in place, but the engine cannot yet play or search: it has no legal moves. Move generation (with make/unmake and perft) is the next layer every later capability — UCI stub, evaluation, and search — depends on.

## What Changes

- Encode moves (from/to, promotion, special flags) and expose them in UCI long-algebraic form
- Generate only legal moves for a position (including castling, en passant, and promotions)
- Make and unmake moves so positions can be explored and restored
- Provide a perft function whose node counts match known reference values from the starting position

## Capabilities

### New Capabilities

- `move-generation`: Legal move listing, make/unmake, and perft correctness

### Modified Capabilities

- (none)

## Non-goals

- UCI protocol I/O or GUI integration
- Evaluation or search (alpha-beta, TT, quiescence)
- Magic bitboards / BMI2 attack tables (ray loops are enough for correctness)
- Opening books or endgame tablebases

## Impact

- New modules under `src/` (e.g. `moves`, `movegen`, `perft`) built on existing `Position` / FEN APIs
- `Position` gains make/unmake (and possibly undo state); no FEN behavioral changes
- `main.rs` remains a stub until the `uci-protocol` change
- Library API gains legal-move and perft entry points used by future UCI/search layers
