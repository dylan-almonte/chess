//! Static evaluation: material + piece-square tables (centipawns, White-positive).

use crate::board::Position;
use crate::piece::{Color, Piece, PieceKind};

/// Material values in centipawns.
pub const PAWN_VALUE: i32 = 100;
pub const KNIGHT_VALUE: i32 = 320;
pub const BISHOP_VALUE: i32 = 330;
pub const ROOK_VALUE: i32 = 500;
pub const QUEEN_VALUE: i32 = 900;
pub const KING_VALUE: i32 = 0;

const MATERIAL: [i32; 6] = [
    PAWN_VALUE,
    KNIGHT_VALUE,
    BISHOP_VALUE,
    ROOK_VALUE,
    QUEEN_VALUE,
    KING_VALUE,
];

/// Piece-square tables from White's perspective (a1 = 0 … h8 = 63).
/// Adapted from the classic simplified evaluation function.
#[rustfmt::skip]
const PAWN_PST: [i32; 64] = [
     0,  0,  0,  0,  0,  0,  0,  0,
     5, 10, 10,-20,-20, 10, 10,  5,
     5, -5,-10,  0,  0,-10, -5,  5,
     0,  0,  0, 20, 20,  0,  0,  0,
     5,  5, 10, 25, 25, 10,  5,  5,
    10, 10, 20, 30, 30, 20, 10, 10,
    50, 50, 50, 50, 50, 50, 50, 50,
     0,  0,  0,  0,  0,  0,  0,  0,
];

#[rustfmt::skip]
const KNIGHT_PST: [i32; 64] = [
   -50,-40,-30,-30,-30,-30,-40,-50,
   -40,-20,  0,  0,  0,  0,-20,-40,
   -30,  0, 10, 15, 15, 10,  0,-30,
   -30,  5, 15, 20, 20, 15,  5,-30,
   -30,  0, 15, 20, 20, 15,  0,-30,
   -30,  5, 10, 15, 15, 10,  5,-30,
   -40,-20,  0,  5,  5,  0,-20,-40,
   -50,-40,-30,-30,-30,-30,-40,-50,
];

#[rustfmt::skip]
const BISHOP_PST: [i32; 64] = [
   -20,-10,-10,-10,-10,-10,-10,-20,
   -10,  5,  0,  0,  0,  0,  5,-10,
   -10, 10, 10, 10, 10, 10, 10,-10,
   -10,  0, 10, 10, 10, 10,  0,-10,
   -10,  5,  5, 10, 10,  5,  5,-10,
   -10,  0,  5, 10, 10,  5,  0,-10,
   -10,  0,  0,  0,  0,  0,  0,-10,
   -20,-10,-10,-10,-10,-10,-10,-20,
];

#[rustfmt::skip]
const ROOK_PST: [i32; 64] = [
     0,  0,  0,  5,  5,  0,  0,  0,
    -5,  0,  0,  0,  0,  0,  0, -5,
    -5,  0,  0,  0,  0,  0,  0, -5,
    -5,  0,  0,  0,  0,  0,  0, -5,
    -5,  0,  0,  0,  0,  0,  0, -5,
    -5,  0,  0,  0,  0,  0,  0, -5,
     5, 10, 10, 10, 10, 10, 10,  5,
     0,  0,  0,  0,  0,  0,  0,  0,
];

#[rustfmt::skip]
const QUEEN_PST: [i32; 64] = [
   -20,-10,-10, -5, -5,-10,-10,-20,
   -10,  0,  0,  0,  0,  0,  0,-10,
   -10,  0,  5,  5,  5,  5,  0,-10,
    -5,  0,  5,  5,  5,  5,  0, -5,
     0,  0,  5,  5,  5,  5,  0, -5,
   -10,  5,  5,  5,  5,  5,  0,-10,
   -10,  0,  5,  0,  0,  0,  0,-10,
   -20,-10,-10, -5, -5,-10,-10,-20,
];

#[rustfmt::skip]
const KING_PST: [i32; 64] = [
    20, 30, 10,  0,  0, 10, 30, 20,
    20, 20,  0,  0,  0,  0, 20, 20,
   -10,-20,-20,-20,-20,-20,-20,-10,
   -20,-30,-30,-40,-40,-30,-30,-20,
   -30,-40,-40,-50,-50,-40,-40,-30,
   -30,-40,-40,-50,-50,-40,-40,-30,
   -30,-40,-40,-50,-50,-40,-40,-30,
   -30,-40,-40,-50,-50,-40,-40,-30,
];

const PST: [&[i32; 64]; 6] = [
    &PAWN_PST,
    &KNIGHT_PST,
    &BISHOP_PST,
    &ROOK_PST,
    &QUEEN_PST,
    &KING_PST,
];

/// Evaluate `pos` in centipawns from White's perspective (positive = White better).
/// Side to move does not affect the score.
pub fn evaluate(pos: &Position) -> i32 {
    let mut score = 0;
    for color in [Color::White, Color::Black] {
        let sign = if color == Color::White { 1 } else { -1 };
        for kind in PieceKind::ALL {
            let mut bb = pos.pieces[Piece::new(color, kind).bitboard_index()];
            let pst = PST[kind.index()];
            let mat = material_value(kind);
            while bb != 0 {
                let sq = bb.trailing_zeros() as usize;
                bb &= bb - 1;
                let pst_sq = match color {
                    Color::White => sq,
                    Color::Black => sq ^ 56,
                };
                score += sign * (mat + pst[pst_sq]);
            }
        }
    }
    score
}

fn material_value(kind: PieceKind) -> i32 {
    MATERIAL[kind.index()]
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::fen::parse_fen;

    #[test]
    fn material_constants_match_spec() {
        assert_eq!(material_value(PieceKind::Pawn), 100);
        assert_eq!(material_value(PieceKind::Knight), 320);
        assert_eq!(material_value(PieceKind::Bishop), 330);
        assert_eq!(material_value(PieceKind::Rook), 500);
        assert_eq!(material_value(PieceKind::Queen), 900);
        assert_eq!(material_value(PieceKind::King), 0);
    }

    #[test]
    fn starting_position_is_equal() {
        let pos = parse_fen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1").unwrap();
        assert_eq!(evaluate(&pos), 0);
    }

    #[test]
    fn side_to_move_does_not_change_the_score() {
        let white = parse_fen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1").unwrap();
        let black = parse_fen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1").unwrap();
        assert_eq!(evaluate(&white), evaluate(&black));
    }

    #[test]
    fn white_up_a_queen_is_material_positive() {
        let pos = parse_fen("rnb1kbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1").unwrap();
        assert!(evaluate(&pos) > 800, "score={}", evaluate(&pos));
    }

    #[test]
    fn black_up_a_queen_is_material_negative() {
        let pos = parse_fen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNB1KBNR w KQkq - 0 1").unwrap();
        assert!(evaluate(&pos) < -800, "score={}", evaluate(&pos));
    }

    #[test]
    fn central_white_pawn_outscores_starting_pawn() {
        let e4 = parse_fen("4k3/8/8/8/4P3/8/8/4K3 w - - 0 1").unwrap();
        let e2 = parse_fen("4k3/8/8/8/8/8/4P3/4K3 w - - 0 1").unwrap();
        assert!(
            evaluate(&e4) > evaluate(&e2),
            "e4={} e2={}",
            evaluate(&e4),
            evaluate(&e2)
        );
    }

    #[test]
    fn kings_alone_remain_equal_when_mirrored() {
        let pos = parse_fen("4k3/8/8/8/8/8/8/4K3 w - - 0 1").unwrap();
        assert_eq!(evaluate(&pos), 0);
    }
}
