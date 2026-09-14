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

    #[test]
    fn material_constants_match_spec() {
        assert_eq!(material_value(PieceKind::Pawn), 100);
        assert_eq!(material_value(PieceKind::Knight), 320);
        assert_eq!(material_value(PieceKind::Bishop), 330);
        assert_eq!(material_value(PieceKind::Rook), 500);
        assert_eq!(material_value(PieceKind::Queen), 900);
        assert_eq!(material_value(PieceKind::King), 0);
    }
}
