## Context

See `proposal.md` for motivation and `specs/legal-move-enforcement/spec.md` for behavior. The Rust engine already resolves `position ... moves` tokens against `generate_legal`, but commits the legal prefix when a later token fails and returns no error. The Go board is intentionally a display-only move applier, yet the TUI currently uses it as the human-input validator and mutates it before the engine can reject input.

UCI does not define a move-validation request because conforming GUIs normally know chess rules. This TUI deliberately avoids maintaining a second full rules engine, so a small engine extension is needed at the process boundary.

## Goals / Non-Goals

**Goals:**

- Keep `generate_legal` in Rust as the single authority for legal moves.
- Make `position` updates transactional and invalid tokens observable.
- Validate both human moves and engine replies before mutating visible TUI state.
- Preserve the current board, move-list, log, telemetry, and text-input layout.
- Make synchronization behavior testable with both a fake engine and the real Rust binary.

**Non-Goals:**

- Make `Board.ApplyUCI` a complete Go chess-rules implementation.
- Change LERF square indexing, bitboards, move generation, or search.
- Implement recovery from arbitrary protocol corruption beyond entering a clear error state.

## Decisions

### Add a narrow `legalmoves` engine extension

The Rust session will answer `legalmoves` with one line containing the legal UCI moves for its current position. Ordering is not a contract; consumers treat the response as a set.

The TUI will keep the latest advertised set alongside the move history. On startup it sends `position startpos` followed by `legalmoves`. It will not accept human input until that response arrives.

After a legal human move, the TUI sends:

1. `position startpos moves <complete history>`
2. `legalmoves`
3. `go`

UCI stdout ordering ensures the black legal set is received before `bestmove`, allowing the TUI to validate the engine reply. After accepting `bestmove`, the TUI sends the complete updated position and `legalmoves` again so the next human move is validated against the resulting position.

Alternative: implement legal move generation in Go. Rejected because it creates a second rules engine that can drift from Rust.

Alternative: invoke `go` as a validator. Rejected because it exposes only one selected move, not whether a submitted move is legal.

### Parse and validate a complete position before committing it

Rust `handle_position` will build a candidate `Position` from the supplied base and apply each token only after resolving it against that candidate's legal moves. It commits `self.position = candidate` only after every token succeeds. On the first failure it returns `info string error illegal move <token>` and leaves `self.position` untouched.

The command remains session-safe: malformed FEN or command structure also leaves state unchanged. Existing behavior for those errors can remain silent unless separately specified.

Alternative: retain the legal prefix. Rejected because callers cannot know which position the engine adopted and the TUI history can diverge.

### Keep display application mechanical but gate every call

`Board.ApplyUCI` remains a display helper. The model checks exact UCI input membership in the engine-provided legal set before calling it. This supports castling, en passant, and promotion using the move encoding already produced by Rust without teaching Go how to determine legality.

For `bestmove`, the model checks membership in the legal set fetched immediately before `go`. `0000` is accepted only when that set is empty. Any mismatch is treated as a protocol/engine error, and visible board and move history remain unchanged.

Alternative: optimistically mutate and roll back after rejection. Rejected because rollback adds state complexity and can briefly expose a false position.

### Test at three boundaries in red-green order

- Rust session tests assert transactional rejection and exact `legalmoves` contents.
- Go model tests use a stateful fake handler to assert no mutation/no commands for illegal human input and rejection of an illegal `bestmove`.
- A real-process integration test exercises startup, one human ply, one engine ply, and legal-set refresh to catch drift between fake protocol behavior and the Rust binary.

Tests will assert state and outbound commands, not merely error text. This prevents an agent from satisfying the visible error while still mutating or desynchronizing state.

## Risks / Trade-offs

- [The extension is not standard UCI] → Keep it additive and narrow; normal UCI clients can ignore it, while this repository's TUI explicitly depends on it.
- [Asynchronous responses may use a stale legal set] → Associate legal sets with strict command ordering, disable human submission while waiting, and clear the cached set whenever the position changes.
- [A fake engine may hide integration defects] → Add one real Rust subprocess test for the complete validation/synchronization exchange.
- [The mechanical Go applier may still mishandle a legal special move] → Cover castling, en passant, and promotion display application independently and fail visibly if an advertised move cannot be applied.

## Migration Plan

1. Add the Rust rejection and `legalmoves` tests, then implement the engine behavior.
2. Update the fake engine protocol and add failing Go model tests.
3. Gate TUI mutations on advertised legal moves and add synchronization requests.
4. Add and run the real-process integration test.
5. Roll back the TUI gating and extension together if the interaction causes regressions; transactional Rust rejection is independently safe to retain.
