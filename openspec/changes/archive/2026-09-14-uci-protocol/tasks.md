## 1. UCI session skeleton

- [x] 1.1 Add `uci` module and wire it from `lib.rs` / `main.rs` stdin loop — verify with `cargo check`
- [x] 1.2 Implement handshake handlers for `uci` and `isready` with fixed `id name` / `id author` — verify scenarios `respond_to_uci_with_identity_and_uciok` and `respond_to_isready_with_readyok`

## 2. Position and go

- [x] 2.1 Parse `position startpos` / `position fen …` with optional `moves` and apply via FEN + make_move — verify scenarios `position_startpos_sets_the_standard_opening` and `position_fen_plus_moves_applies_the_sequence`
- [x] 2.2 Implement `go` stub that emits `bestmove` from a legal move (or `0000` if none) — verify scenarios `go_from_startpos_returns_a_legal_move` and `go_in_checkmate_returns_null_move`
- [x] 2.3 Handle `quit` (and optionally `ucinewgame`) so the loop exits cleanly — verify scenario `quit_terminates_the_command_loop`

## 3. Gate

- [x] 3.1 Ensure every delta-spec scenario has a named `#[test]` and `cargo test` passes
- [x] 3.2 Run `openspec validate uci-protocol --strict` and fix any issues
