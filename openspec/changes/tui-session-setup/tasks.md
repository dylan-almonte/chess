## 1. Scaled board (red)

- [ ] 1.1 Add tests `large_window_uses_bigger_squares` (80×28 → cell at least 3×2, `e2` still a pawn) and `narrow_window_keeps_a_readable_8x8_grid` (60×20 → eight ranks/files, `e2` still a pawn); verify they fail on the current 2×1 render
- [ ] 1.2 Add a hit-test test that maps `e2`/`e4` at 3×2 metrics and still maps them at 2×1; verify it fails because `SquareAtCell` is fixed to 2×1

## 2. Scaled board (green)

- [ ] 2.1 Compute `cellW`/`cellH` from window size (clamp 2×1 … 5×3, shrink log first) and render centered glyphs; verify the two window-size scenarios pass
- [ ] 2.2 Drive `SquareAtCell` from the same metrics as `View`; verify the scaled hit-test and existing click `e2e4` tests pass, then commit the scale work

## 3. Setup mode (red)

- [ ] 3.1 Add fake-engine tests `place_a_queen_on_d5` and `remove_a_piece_from_e2`; verify they fail because `setup` is not a command
- [ ] 3.2 Add tests `play_a_kings_and_queen_setup` (`position fen 4k3/8/8/8/8/8/8/3QK3 w - - 0 1` + `legalmoves`) and `setup_without_both_kings_is_rejected`; verify they fail

## 4. Setup mode (green)

- [ ] 4.1 Implement `setup` / brush keys / place-or-erase on click or Space, with `side`, `clear`, and `startpos`; verify the two edit tests pass and no `go` is sent
- [ ] 4.2 Build a six-field FEN (inferred castling, ep `-`, clocks `0 1`), reject `play` without both kings, and on success send `ucinewgame` + `position fen` + `legalmoves` and clear the move list; verify both start-play tests pass, then commit the setup work

## 5. Player slots (red)

- [ ] 5.1 Add fake-engine tests `human_plays_black_from_startpos` and `engine_versus_engine_plays_without_typed_moves`; verify they fail because White is always human and every human ply sends `go`

## 6. Player slots (green)

- [ ] 6.1 Add `white human|engine` and `black human|engine` (default Human / Engine) and send `go` only when the side to move is Engine; verify both slot scenarios pass and existing Human-White tests still pass
- [ ] 6.2 Show slot and setup status on the input line and document `setup` / `play` / slot commands in `tui/README.md`; verify `go test ./...` under `tui/` passes, then commit the slot work

## 7. Gate

- [ ] 7.1 Ensure every new delta-spec scenario has an automated Go test and run `gofmt` plus `go test ./...` under `tui/`
- [ ] 7.2 Run `openspec validate tui-session-setup --strict` and fix any issues
