## 1. Move encoding and wiring

- [ ] 1.1 Add `moves`, `movegen`, `makemove`, and `perft` modules and re-export from `lib.rs` — verify with `cargo check`
- [ ] 1.2 Implement `Move` (from/to/promotion/flags) and UCI long-algebraic display — verify a unit test formats `e2e4` and `e7e8q`

## 2. Attacks and legality helpers

- [ ] 2.1 Implement attack generation (pawn/knight/king tables or functions; ray loops for sliders) and `is_square_attacked` — verify king on e1 is not attacked in startpos and is attacked in a crafted check FEN
- [ ] 2.2 Implement pseudo-legal generation for all piece types including castling, en passant, and promotions — verify raw lists compile and cover special-move FENs before filtering

## 3. Make/unmake and legal moves

- [ ] 3.1 Implement `make_move` / `unmake_move` with `Undo` restoring bitboards, side, castling, EP, and clocks — verify scenario `make_then_unmake_restores_startpos_fen`
- [ ] 3.2 Filter pseudo-legal moves to legal moves (king not left in check; castling path rules) — verify scenario `startpos_has_twenty_legal_moves`
- [ ] 3.3 Cover special-move scenarios: legal castling, castling through check omitted, en passant, and four promotions — verify the four named scenario tests

## 4. Perft gate

- [ ] 4.1 Implement recursive legal `perft(depth)` — verify scenarios `perft_depth_1` through `perft_depth_4` from startpos (20, 400, 8902, 197281)
- [ ] 4.2 Ensure every delta-spec scenario has a named `#[test]` and `cargo test` passes
- [ ] 4.3 Run `openspec validate move-generation --strict` and fix any issues
