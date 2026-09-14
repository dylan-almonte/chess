## Context

See proposal.md for motivation. Board representation, legal make/unmake, evaluation (White-positive centipawns), and a UCI stub that replies `bestmove` with the first legal move are in place. `king_square` and `is_square_attacked` already exist in move generation. Specs: `specs/search/spec.md` and `specs/uci-protocol/spec.md`. Square indexing remains LERF (a1=0 … h8=63). Search mutates a position through make/unmake; UCI session state must be restored after `go`.

## Goals / Non-Goals

**Goals:**

- Negamax + alpha-beta with iterative deepening `1..=max_depth`
- Mate / stalemate terminals so depth-1 mate-in-one is visible (eval alone cannot, because king material is 0)
- UCI `go` / `go depth N` consumes search, emits `info`, then `bestmove`
- Named Rust tests mapped to the search and UCI delta scenarios

**Non-Goals:**

- Same exclusions as the proposal (TT, quiescence, time control, `stop` thread)
- Changing evaluation tables or move-gen internals

## Decisions

### Negamax over explicit min/max

- One recursive function; each ply negates the child's score (`-negamax(...)`)
- Leaf eval is side-to-move relative: `evaluate(pos)` if White to move, else `-evaluate(pos)`
- **Why:** Matches the requested algorithm; eval is already White-positive so the flip is a single branch at leaves
- **Alternative:** Separate maximizing/minimizing nodes — equivalent and twice the code

### Make/unmake on a root clone

- `search` clones the input `Position` and make/unmakes on the clone
- **Why:** Project default; UCI `position` is not corrupted if a node forgets to unmake
- **Alternative:** Copy-make per node — simpler debugging, much slower; search on the session position in place — faster but riskier

### Terminals before depth-zero eval

- At every node: if `depth == 0` and the side to move is **not** in check, return STM eval (no move gen)
- Otherwise generate legal moves. Empty + in check → `-MATE + ply`; empty + not in check → `0`; non-empty and `depth == 0` → STM eval
- In check: `king_square(pos, side_to_move)` plus `is_square_attacked(..., opposite)`
- **Why:** Mate in one at depth 1 is a child of the root at depth 0; a pure eval leaf would miss it (king value 0)
- **Alternative:** Always generate at leaves — correct but slower on quiet positions; check extensions — later

### Mate encoding

- `MATE = 30_000` (well above any material+PST total)
- Distance uses ply from the search root so shorter mates score better
- UCI `score mate N`: `N` is moves (not plies) for the side to move; convert with `(MATE - abs(score) + 1) / 2`, sign of `score`
- **Why:** Standard UCI; TUI already renders `score mate N`
- **Alternative:** Report mate as a large `cp` — fails the UCI mate scenario

### Iterative deepening and PV

- Loop `depth = 1..=max_depth`; each iteration is a full alpha-beta from `(-INF, INF)`
- Reuse the previous iteration's first PV move as the first root move (then captures, then quiets)
- Recover PV by returning the best move plus the child PV from the recursive call (no transposition table)
- **Why:** Required by spec (depth 1 vs 2 disagree on the poisoned capture; UCI emits one `info` per completed depth). PV-first ordering is cheap and helps alpha-beta without a TT
- **Alternative:** A single max-depth search — would still find the depth-2 move but would not emit per-iteration `info`; TT-based PV — out of scope

### Capture-first move ordering

- After the ID PV move: captures/en passant (`Move::is_capture`) before non-captures; stable otherwise
- **Why:** Alpha-beta needs a decent first try; flags already exist on `Move`
- **Alternative:** MVV-LVA or history — later; generation order only — enough for spec FENs but worse pruning

### Library API

- New `src/search.rs`: `search(pos: &Position, max_depth: u32) -> SearchResult`
- `SearchResult { best_move: Option<Move>, score: i32, pv: Vec<Move>, nodes: u64 }` — `score` is STM-relative; `best_move` is `None` on terminal positions
- Increment `nodes` once per negamax entry
- **Why:** UCI can format `info` / `bestmove` without knowing negamax; tests assert FENs directly
- **Alternative:** Search writes UCI strings internally — couples the library to the protocol

### UCI `go` parsing and default depth

- Tokenize the `go` line; if `depth` is followed by a positive integer, use it; otherwise `DEFAULT_GO_DEPTH = 4`
- Ignore `wtime` / `btime` / `movetime` / `infinite` / unknown tokens
- After each ID iteration, push `info depth <d> score <cp|mate> nodes <n> pv <uci…>`
- Then `bestmove <uci>` or `bestmove 0000`
- Stay synchronous in `handle_line` (same as the stub)
- **Why:** Spec default is 4 so tests stay fast; TUI already parses those `info` fields; no `stop` thread until time management exists
- **Alternative:** Default depth 1 — weaker play in the TUI; thread + `stop` — needed only for clocks

## Risks / Trade-offs

- [Horizon effect without quiescence] → Mitigation: accepted; poisoned-capture spec is depth 2, not a capture-at-leaf case
- [Move gen at in-check leaves] → Mitigation: quiet depth-0 nodes skip generation; in-check leaves are rare
- [Bare `go` is now depth 4] → Mitigation: startpos depth 4 is still cheap vs perft; existing legality tests stay valid
- [No TT / no time control] → Mitigation: next engine layers; GUIs that send only `go` still get a move
- [PV allocation per node] → Mitigation: fine through depth 4; switch to a triangular table if it shows up in profiling

## Migration Plan

N/A for users. After archive, main specs live at `openspec/specs/search/spec.md` and the updated `openspec/specs/uci-protocol/spec.md`. `chess-tui` telemetry starts filling in from engine `info` with no TUI change.
