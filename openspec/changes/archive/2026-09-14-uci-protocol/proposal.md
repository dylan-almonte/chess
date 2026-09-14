## Why

Legal move generation works in the library, but nothing speaks UCI yet, so GUIs and tools cannot load the engine. A minimal UCI loop that sets positions and replies to `go` with a legal move unlocks playtesting and comes before evaluation and search.

## What Changes

- Implement a stdin/stdout UCI command loop in the binary (`main.rs`)
- Support the core handshake and session commands needed to load the engine in a GUI
- Apply `position` (startpos / fen + optional moves) to the internal board
- On `go`, return `bestmove` as one legal move from the current position (no search yet)

## Capabilities

### New Capabilities

- `uci-protocol`: UCI I/O handshake, position setup, and stub `bestmove` from legal move generation

### Modified Capabilities

- (none)

## Non-goals

- Evaluation or alpha-beta search (later layers)
- Full UCI option surface (`setoption`, ponder, multipv, etc.)
- Time management beyond accepting `go` and returning immediately
- GUI itself or Cute Chess integration scripts (manual/GUI use is enough)

## Impact

- `src/main.rs` becomes the UCI entry point
- New module(s) under `src/` for UCI parsing/dispatch (e.g. `uci`)
- Reuses existing `parse_fen`, `make_move`, `generate_legal`, and UCI move strings
- No change to board-representation or move-generation behavioral requirements
