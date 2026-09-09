## 1. Crate skeleton

- [x] 1.1 Add `src/lib.rs` and module stubs (`square`, `piece`, `board`, `fen`); leave `main.rs` as a stub — verify with `cargo check`
- [x] 1.2 Implement `Color`, `PieceKind`, and `Piece` in `piece.rs` with FEN char mapping — verify unit tests for char round-trip

## 2. Squares

- [x] 2.1 Implement `Square` with LERF indexing and algebraic parse/display — verify scenarios `map_corner_and_center_squares` and `reject_invalid_square_names`

## 3. Position and FEN

- [x] 3.1 Implement bitboard `Position` with piece boards, occupancy rebuild, castling, en passant, clocks, and piece-at queries — verify startpos helper places kings/pawns correctly when loaded via FEN
- [x] 3.2 Implement FEN parse into `Position` — verify scenario `load_startpos_pieces_and_side_to_move`
- [x] 3.3 Implement FEN serialize from `Position` — verify round-trip scenarios for startpos, midgame with EP, and reduced castling
- [x] 3.4 Reject malformed FEN (missing fields, invalid piece char, wrong file count) — verify the three reject scenarios

## 4. Gate

- [x] 4.1 Ensure every delta-spec scenario has a named `#[test]` and `cargo test` passes
- [x] 4.2 Run `openspec validate board-representation --strict` and fix any issues
