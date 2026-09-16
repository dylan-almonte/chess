## 1. Rust engine red tests

- [x] 1.1 Add named `src/uci.rs` tests for `illegal_first_move_preserves_current_position` and `illegal_later_move_rejects_whole_sequence`; verify `cargo test illegal_ -- --nocapture` fails because invalid input is silent and/or commits a legal prefix
- [x] 1.2 Add named tests `legalmoves_from_startpos_returns_exactly_twenty_unique_moves` and `legalmoves_in_checkmate_is_empty`; verify `cargo test legalmoves_ -- --nocapture` fails because the command is not implemented

## 2. Rust engine implementation

- [x] 2.1 Make `position` parsing return an error response and commit its candidate position only after every move resolves legally; verify the two transactional rejection tests pass
- [x] 2.2 Implement the additive `legalmoves` response from `generate_legal`, including the empty-position form; verify all `legalmoves_` tests pass
- [x] 2.3 Run `cargo fmt --check` and `cargo test`, then commit the Rust engine tests and implementation as one focused commit without including unrelated working-tree changes

## 3. TUI red tests

- [x] 3.1 Extend the fake engine to model `position`, `legalmoves`, and configurable invalid `bestmove` responses; verify its focused Go tests pass
- [x] 3.2 Add named model tests for rejecting `e2e5` and `e7e5` from startpos; assert error text, unchanged board/list, and no additional outbound `position` or `go`, then verify the tests fail before model changes
- [x] 3.3 Add named model tests for accepting `e2e4`, validating `e7e5`, synchronizing the complete history, refreshing White's legal set, and rejecting engine `bestmove e2e5` without mutation; verify the tests fail before model changes
- [x] 3.4 Add a real-process integration test covering `e2e4`, one legal engine reply, full-history synchronization, and the next `legalmoves` response; build the Rust binary and verify the test fails before TUI protocol changes

## 4. TUI implementation

- [x] 4.1 Parse `legalmoves` responses into a replace-on-receipt set tied to the current position, clear it when position changes, and prevent input while no current set is available; verify focused parsing/state tests pass
- [x] 4.2 Gate human `ApplyUCI`, move-list mutation, `position`, and `go` on exact membership in the current legal set; verify both illegal-human-move tests pass
- [x] 4.3 Request the engine move-side legal set before `go`, validate `bestmove` (including `0000` only for an empty set), then synchronize the accepted complete history and refresh human-side legal moves; verify synchronization and invalid-bestmove tests pass
- [x] 4.4 Add display-helper tests for advertised castling, en passant, and promotion moves and make failures visible without partial model mutation; verify `go test ./internal/board ./internal/ui` passes
- [x] 4.5 Run `gofmt` and `go test ./...` under `tui/`, then commit the TUI tests and implementation as one focused commit without including unrelated working-tree changes

## 5. Cross-layer validation

- [x] 5.1 Build the Rust engine and run the real-process TUI integration test; verify the engine and displayed position remain synchronized after both plies
- [x] 5.2 Run `cargo fmt --check`, `cargo test`, `go test ./...` under `tui/`, and `openspec validate legal-move-enforcement --strict`; fix failures and commit only any resulting change-scoped fixes
