//! Static evaluation: material + piece-square tables (centipawns, White-positive).

use crate::board::Position;
use crate::piece::PieceKind;

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

/// Evaluate `pos` in centipawns from White's perspective (positive = White better).
/// Side to move does not affect the score.
pub fn evaluate(_pos: &Position) -> i32 {
    0
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
