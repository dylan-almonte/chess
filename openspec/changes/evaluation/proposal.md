## Why

The engine can load positions and return a legal `bestmove`, but it has no notion of which side is better. Static evaluation (material plus piece-square tables) is the next engine layer: it scores positions so search can compare leaves, and it unlocks deterministic, testable “quality” before alpha-beta exists.

## What Changes

- Add a static evaluation API that scores a `Position` in centipawns
- Score = material balance + piece-square table (PST) bonuses for each piece on its square
- Cover evaluation with named Rust tests using concrete FEN → score expectations
- Keep UCI `go` as the existing stub (first legal move); search will consume eval later

## Capabilities

### New Capabilities

- `evaluation`: Static position scoring from material and piece-square tables, expressed in centipawns

### Modified Capabilities

- (none)

## Non-goals

- Alpha-beta, iterative deepening, quiescence, or any search (later `search` layer)
- Changing UCI `bestmove` selection to use evaluation
- Pawn structure, mobility, king safety, tapered eval / endgame tables beyond a single PST set
- Third-party chess crates or externally loaded weight files

## Impact

- New module(s) under `src/` (e.g. `eval`) used by future search
- Reuses existing `Position` / piece bitboards and square indexing (a1=0 … h8=63)
- No requirement changes to board-representation, move-generation, or uci-protocol
- TUI unaffected
