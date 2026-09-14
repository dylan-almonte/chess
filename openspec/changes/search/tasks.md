## 1. Search module skeleton

- [x] 1.1 Add `src/search.rs` with `SearchResult { best_move, score, pv, nodes }` and `search(pos, max_depth)` stub (empty PV, `best_move: None`, score `0`) and wire `mod search` / `pub use search::{search, SearchResult}` from `lib.rs` — verify with `cargo check`
- [x] 1.2 Add failing named tests for every search delta-spec scenario (startpos depth-1 legal PV, hanging queen `h1h4`, rook mate `e7e8` mate in 1, checkmate no move, stalemate score `0`, depth-1 `e4d5`, depth-2 not `e4d5`) — verify tests compile and fail (`cargo test search_ -- --nocapture` shows failures)

## 2. Negamax, terminals, and iterative deepening

- [x] 2.1 Implement side-to-move eval at quiet depth-0 leaves, legal make/unmake negamax, and mate/stalemate terminals (`MATE = 30_000`, ply distance, in-check via `king_square` + `is_square_attacked`) on a root clone — verify checkmate, stalemate, mate-in-one, hanging-queen, and startpos PV tests pass
- [x] 2.2 Add alpha-beta (`-beta, -alpha`) and capture-first ordering (`Move::is_capture`) — verify the same tests still pass
- [x] 2.3 Add iterative deepening `1..=max_depth`, reconstruct PV from child PVs, and try the previous iteration’s PV move first — verify depth-1 `e4d5`, depth-2 not `e4d5`, and that a depth-2 PV’s moves are legal in sequence

## 3. UCI `go` uses search

- [ ] 3.1 Add failing UCI tests for the delta scenarios (`go depth 1` legal + `info` before `bestmove`, `go depth 2` emits `depth 1` and `depth 2` info, hanging queen `bestmove h1h4`, rook mate `bestmove e7e8` and `score mate 1`) — verify they fail while existing `go` legality / `0000` tests still pass
- [ ] 3.2 Parse `go` tokens (`depth N` or default `4`; ignore time tokens), run search, emit one `info depth … score … nodes … pv …` line per completed iteration, then `bestmove` / `0000` — verify all new and existing UCI tests pass (`cargo test uci`)

## 4. Gate

- [ ] 4.1 Ensure every search and UCI delta-spec scenario has a named `#[test]` and `cargo test` passes
- [ ] 4.2 Run `openspec validate search --strict` and fix any issues
