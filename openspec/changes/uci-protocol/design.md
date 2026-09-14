## Context

Board representation and move generation (including legal moves, make/unmake, and UCI move strings) are archived. See proposal.md for motivation. `main.rs` is still a stub. Project conventions: library holds chess logic; binary is the UCI entry; specs stay behavioral.

## Goals / Non-Goals

**Goals:**

- Minimal UCI subset sufficient for GUIs: `uci`, `isready`, `ucinewgame`, `position`, `go`, `quit`
- Stub move choice: any legal move (deterministic preference is fine, e.g. first in generation order)
- Testable command handling without requiring a live GUI

**Non-Goals:**

- Search, evaluation, pondering, or real time control
- `setoption` / UCI options
- Parallel search thread (synchronous `go` is enough for the stub)

## Decisions

### Synchronous stdin loop in the binary

- `main` reads lines from stdin, dispatches, writes replies to stdout, flushes after each response batch
- Unknown commands are ignored (UCI-tolerant) rather than erroring hard
- **Why:** Matches how GUIs talk to engines; simplest correct shell
- **Alternative:** async/threaded stop for `go` — deferred until search needs it

### Library-facing UCI session for tests

- Put protocol state + dispatch in `src/uci.rs` (or similar) with an API like `process_line(session, line) -> Vec<String>` / `handle_command`
- `main` only wires stdin/stdout to that API
- **Why:** Every scenario can drive the engine with strings and assert reply lines without spawning a process
- **Alternative:** only integration tests via `std::process` — slower and harder to debug

### Position command parsing

- Support `position startpos [moves …]` and `position fen <fen-6-fields> [moves …]`
- FEN for `position fen` is the standard six fields; rebuild via existing `parse_fen`, then `make_move` for each UCI move in order
- Reject/ignore malformed position lines without crashing the loop (log nothing required for this change)
- **Why:** Enough for Arena/Cute Chess style loads

### Go stub

- Ignore tokens after `go` (`wtime`, `depth`, etc.)
- Pick `generate_legal(pos).first()` (or equivalent stable choice); format with existing `Move` Display
- No legal moves → `bestmove 0000` (UCI null move)
- **Why:** Satisfies “GUI can load and get a move” without search

### Identity strings

- Fixed `id name` (e.g. project name `chess`) and `id author` (placeholder or repo owner string) — document chosen literals in tasks/tests

### `ucinewgame`

- Reset to startpos (or clear and wait for next `position`); treat as optional hygiene for GUIs that send it
- **Why:** Common in GUI sessions; cheap to support

## Risks / Trade-offs

- [GUIs send extra commands we ignore] → Mitigation: ignore unknown lines; expand only when a GUI blocks on a missing reply
- [Applying invalid moves in `position … moves` can panic] → Mitigation: skip/stop applying on parse failure; keep session alive
- [First legal move is weak play] → Mitigation: acceptable until evaluation/search; documented as stub

## Migration Plan

N/A. After archive, main spec lives at `openspec/specs/uci-protocol/spec.md`. Manual check: load binary in a UCI GUI and play a move.
