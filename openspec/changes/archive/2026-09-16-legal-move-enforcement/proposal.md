## Why

The playable TUI can mutate its display with an illegal human move while the Rust engine silently stops at the last legal move in a `position ... moves` sequence. This desynchronizes the two positions and allows the existing happy-path test suites to pass even though interactive play is incorrect.

## What Changes

- Reject illegal human moves before changing the TUI board, move list, or sending `position` and `go`.
- Make invalid moves in an engine `position ... moves` command observable and preserve the engine's previously accepted position rather than silently adopting a legal prefix.
- Treat the Rust engine's legal move generation as the authority for playable input without duplicating full chess legality rules in the Go display board.
- Add negative and cross-process tests that prove illegal input leaves state unchanged and accepted move sequences remain synchronized.

## Non-goals

- Changing legal move generation algorithms or chess rules.
- Adding a general plugin protocol, Rust/Go FFI, or a third-party chess-rules implementation to the TUI.
- Redesigning search strength, evaluation, or the TUI move-entry experience.
- Expanding standard UCI with a general-purpose legal-move query command beyond the minimum validation contract needed by this client.

## Capabilities

### New Capabilities

- `legal-move-enforcement`: Defines rejection, state preservation, and engine/TUI synchronization for human and UCI move input.

### Modified Capabilities

None.

## Impact

- Rust UCI session parsing and responses in `src/uci.rs`.
- Go engine client/model flow and display state under `tui/internal/`.
- Rust unit tests, Go model tests, and a real engine/TUI integration test.
- The TUI-to-engine boundary gains a small validation interaction; no third-party dependency is required.
