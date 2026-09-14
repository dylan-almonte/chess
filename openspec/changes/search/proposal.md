## Why

The engine can score a static position and reply to UCI `go` with the first legal move, so it cannot look ahead or choose a stronger line. Evaluation is in place; search is the next engine layer so `bestmove` comes from looking ahead instead of generation order.

## What Changes

- Add a library search API that chooses a move with **negamax**, **alpha-beta pruning**, and **iterative deepening** to a requested depth
- Score checkmate and stalemate as search terminals (eval does not; king material is 0)
- Wire UCI `go` to search instead of `generate_legal(...).first()`
- Honor `go depth N`; a bare `go` uses a documented default max depth
- Emit UCI `info` lines (depth, score, nodes, pv) so the existing TUI telemetry panel can show search progress
- Cover search with named Rust tests on concrete FENs (winning captures, mate-in-one, legal PV / bestmove)

## Capabilities

### New Capabilities

- `search`: Depth-limited iterative-deepening negamax with alpha-beta; returns a principal variation and a side-to-move score; UCI `go` consumes it

### Modified Capabilities

- `uci-protocol`: `go` selects `bestmove` from search rather than the first legal move; `depth` is honored; `info` lines precede `bestmove`

## Non-goals

- Transposition tables, quiescence search, null-move, LMR, PVS, or aspiration windows
- Time management (`wtime` / `btime` / `movetime` / `infinite`) and a parallel `stop` thread — time tokens stay ignored
- Fifty-move and repetition draws
- Opening book, pondering, or UCI options
- TUI changes (the client already parses `info` and stays idle when none are sent)
- Incremental evaluation or move-gen optimizations

## Impact

- New `src/search.rs` (or similar) using existing `generate_legal`, `make_move` / `unmake_move`, and White-positive `evaluate`
- `src/uci.rs` `go` handler calls search and may emit several `info` lines before `bestmove`
- Existing UCI scenarios that only require a legal `bestmove` (or `0000` in mate) remain valid
- TUI is a consumer only; no Go code changes
- Default depth must stay small enough that unit tests remain fast
