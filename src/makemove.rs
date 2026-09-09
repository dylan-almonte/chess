//! Make and unmake moves on a [`Position`].

use crate::board::{Position, CASTLE_BK, CASTLE_BQ, CASTLE_WK, CASTLE_WQ};
use crate::moves::Move;
use crate::piece::{Color, Piece, PieceKind};
use crate::square::Square;

/// State needed to undo a move.
#[derive(Clone, Debug)]
pub struct Undo {
    pub captured: Option<Piece>,
    pub castling: u8,
    pub en_passant: Option<Square>,
    pub halfmove_clock: u32,
    pub fullmove_number: u32,
}

fn clear_bit(bb: &mut u64, sq: Square) {
    *bb &= !sq.bit();
}

fn set_bit(bb: &mut u64, sq: Square) {
    *bb |= sq.bit();
}

fn remove_piece(pos: &mut Position, sq: Square, piece: Piece) {
    clear_bit(&mut pos.pieces[piece.bitboard_index()], sq);
}

fn place_piece(pos: &mut Position, sq: Square, piece: Piece) {
    set_bit(&mut pos.pieces[piece.bitboard_index()], sq);
}

fn update_castling_on_move(pos: &mut Position, from: Square, to: Square, moving: Piece) {
    // King moves clear both rights for that side.
    if moving.kind == PieceKind::King {
        match moving.color {
            Color::White => pos.castling &= !(CASTLE_WK | CASTLE_WQ),
            Color::Black => pos.castling &= !(CASTLE_BK | CASTLE_BQ),
        }
    }
    // Rook moves clear the corresponding right.
    if moving.kind == PieceKind::Rook {
        match (moving.color, from.index()) {
            (Color::White, 0) => pos.castling &= !CASTLE_WQ,
            (Color::White, 7) => pos.castling &= !CASTLE_WK,
            (Color::Black, 56) => pos.castling &= !CASTLE_BQ,
            (Color::Black, 63) => pos.castling &= !CASTLE_BK,
            _ => {}
        }
    }
    // Capturing a rook clears that side's right.
    match to.index() {
        0 => pos.castling &= !CASTLE_WQ,
        7 => pos.castling &= !CASTLE_WK,
        56 => pos.castling &= !CASTLE_BQ,
        63 => pos.castling &= !CASTLE_BK,
        _ => {}
    }
}

/// Apply `mv` to `pos`, returning undo information. Caller must ensure the move
/// is at least pseudo-legal for the current side to move.
pub fn make_move(pos: &mut Position, mv: Move) -> Undo {
    let side = pos.side_to_move;
    let moving = pos
        .piece_at(mv.from)
        .expect("make_move: no piece on from square");

    let undo = Undo {
        captured: if mv.is_en_passant() {
            let cap_sq = en_passant_captured_square(mv, side);
            pos.piece_at(cap_sq)
        } else {
            pos.piece_at(mv.to)
        },
        castling: pos.castling,
        en_passant: pos.en_passant,
        halfmove_clock: pos.halfmove_clock,
        fullmove_number: pos.fullmove_number,
    };

    // Clear EP; may set new below.
    pos.en_passant = None;

    // Halfmove clock
    if moving.kind == PieceKind::Pawn || mv.is_capture() {
        pos.halfmove_clock = 0;
    } else {
        pos.halfmove_clock += 1;
    }

    remove_piece(pos, mv.from, moving);

    if mv.is_en_passant() {
        let cap_sq = en_passant_captured_square(mv, side);
        if let Some(cap) = undo.captured {
            remove_piece(pos, cap_sq, cap);
        }
    } else if let Some(cap) = undo.captured {
        remove_piece(pos, mv.to, cap);
    }

    let placed = if let Some(promo) = mv.promotion {
        Piece::new(side, promo)
    } else {
        moving
    };
    place_piece(pos, mv.to, placed);

    if mv.is_castle() {
        match (side, mv.to.index()) {
            (Color::White, 6) => {
                // e1g1: rook h1 -> f1
                let rook = Piece::new(Color::White, PieceKind::Rook);
                remove_piece(pos, Square::from_index_unchecked(7), rook);
                place_piece(pos, Square::from_index_unchecked(5), rook);
            }
            (Color::White, 2) => {
                // e1c1: rook a1 -> d1
                let rook = Piece::new(Color::White, PieceKind::Rook);
                remove_piece(pos, Square::from_index_unchecked(0), rook);
                place_piece(pos, Square::from_index_unchecked(3), rook);
            }
            (Color::Black, 62) => {
                let rook = Piece::new(Color::Black, PieceKind::Rook);
                remove_piece(pos, Square::from_index_unchecked(63), rook);
                place_piece(pos, Square::from_index_unchecked(61), rook);
            }
            (Color::Black, 58) => {
                let rook = Piece::new(Color::Black, PieceKind::Rook);
                remove_piece(pos, Square::from_index_unchecked(56), rook);
                place_piece(pos, Square::from_index_unchecked(59), rook);
            }
            _ => panic!("invalid castling destination"),
        }
    }

    update_castling_on_move(pos, mv.from, mv.to, moving);

    if mv.is_double_pawn() {
        let ep_rank = match side {
            Color::White => mv.from.rank() + 1,
            Color::Black => mv.from.rank() - 1,
        };
        pos.en_passant = Square::from_file_rank(mv.from.file(), ep_rank);
    }

    if side == Color::Black {
        pos.fullmove_number += 1;
    }
    pos.side_to_move = side.opposite();
    pos.rebuild_occupancy();
    undo
}

fn en_passant_captured_square(mv: Move, _side: Color) -> Square {
    // Captured pawn is on the same file as `to`, on the from-rank of the mover.
    Square::from_file_rank(mv.to.file(), mv.from.rank()).unwrap()
}

/// Reverse a previously applied move using its [`Undo`] snapshot.
pub fn unmake_move(pos: &mut Position, mv: Move, undo: Undo) {
    let side = pos.side_to_move.opposite(); // side that made the move

    pos.side_to_move = side;
    pos.castling = undo.castling;
    pos.en_passant = undo.en_passant;
    pos.halfmove_clock = undo.halfmove_clock;
    pos.fullmove_number = undo.fullmove_number;

    let placed = if let Some(promo) = mv.promotion {
        Piece::new(side, promo)
    } else {
        pos.piece_at(mv.to)
            .expect("unmake_move: no piece on to square")
    };

    // For promotions, remove promoted piece; otherwise remove the moved piece.
    remove_piece(pos, mv.to, placed);

    if mv.is_castle() {
        match (side, mv.to.index()) {
            (Color::White, 6) => {
                let rook = Piece::new(Color::White, PieceKind::Rook);
                remove_piece(pos, Square::from_index_unchecked(5), rook);
                place_piece(pos, Square::from_index_unchecked(7), rook);
            }
            (Color::White, 2) => {
                let rook = Piece::new(Color::White, PieceKind::Rook);
                remove_piece(pos, Square::from_index_unchecked(3), rook);
                place_piece(pos, Square::from_index_unchecked(0), rook);
            }
            (Color::Black, 62) => {
                let rook = Piece::new(Color::Black, PieceKind::Rook);
                remove_piece(pos, Square::from_index_unchecked(61), rook);
                place_piece(pos, Square::from_index_unchecked(63), rook);
            }
            (Color::Black, 58) => {
                let rook = Piece::new(Color::Black, PieceKind::Rook);
                remove_piece(pos, Square::from_index_unchecked(59), rook);
                place_piece(pos, Square::from_index_unchecked(56), rook);
            }
            _ => panic!("invalid castling destination"),
        }
    }

    let original = if mv.promotion.is_some() {
        Piece::new(side, PieceKind::Pawn)
    } else {
        // placed was the moving piece (non-promo)
        placed
    };
    place_piece(pos, mv.from, original);

    if mv.is_en_passant() {
        if let Some(cap) = undo.captured {
            let cap_sq = en_passant_captured_square(mv, side);
            place_piece(pos, cap_sq, cap);
        }
    } else if let Some(cap) = undo.captured {
        place_piece(pos, mv.to, cap);
    }

    pos.rebuild_occupancy();
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::fen::{parse_fen, to_fen};
    use crate::movegen::generate_legal;
    use std::str::FromStr;

    #[test]
    fn make_then_unmake_restores_startpos_fen() {
        let fen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1";
        let mut pos = parse_fen(fen).unwrap();
        let mv = generate_legal(&pos)
            .into_iter()
            .find(|m| m.to_string() == "e2e4")
            .unwrap();
        let undo = make_move(&mut pos, mv);
        unmake_move(&mut pos, mv, undo);
        assert_eq!(to_fen(&pos), fen);
    }

    #[test]
    fn parse_move_string() {
        let mv = Move::from_str("e2e4").unwrap();
        assert_eq!(mv.to_string(), "e2e4");
    }
}
