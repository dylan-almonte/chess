//! Bitboard-backed chess position.

use crate::piece::{Color, Piece, PieceKind};
use crate::square::Square;

/// Castling rights bitmask: White kingside/queenside, Black kingside/queenside.
pub const CASTLE_WK: u8 = 0b0001;
pub const CASTLE_WQ: u8 = 0b0010;
pub const CASTLE_BK: u8 = 0b0100;
pub const CASTLE_BQ: u8 = 0b1000;

/// A chess position stored as 12 piece bitboards plus game-state fields.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Position {
    /// Piece bitboards indexed by [`Piece::bitboard_index`]: white then black, P..K.
    pub pieces: [u64; 12],
    pub white_occ: u64,
    pub black_occ: u64,
    pub all_occ: u64,
    pub side_to_move: Color,
    pub castling: u8,
    pub en_passant: Option<Square>,
    pub halfmove_clock: u32,
    pub fullmove_number: u32,
}

impl Position {
    pub fn empty() -> Self {
        Self {
            pieces: [0; 12],
            white_occ: 0,
            black_occ: 0,
            all_occ: 0,
            side_to_move: Color::White,
            castling: 0,
            en_passant: None,
            halfmove_clock: 0,
            fullmove_number: 1,
        }
    }

    pub fn rebuild_occupancy(&mut self) {
        self.white_occ = 0;
        self.black_occ = 0;
        for kind in PieceKind::ALL {
            let w = Piece::new(Color::White, kind).bitboard_index();
            let b = Piece::new(Color::Black, kind).bitboard_index();
            self.white_occ |= self.pieces[w];
            self.black_occ |= self.pieces[b];
        }
        self.all_occ = self.white_occ | self.black_occ;
    }

    pub fn piece_at(&self, square: Square) -> Option<Piece> {
        let bit = square.bit();
        if self.all_occ & bit == 0 {
            return None;
        }
        for color in [Color::White, Color::Black] {
            for kind in PieceKind::ALL {
                let piece = Piece::new(color, kind);
                if self.pieces[piece.bitboard_index()] & bit != 0 {
                    return Some(piece);
                }
            }
        }
        None
    }

    pub fn set_piece(&mut self, square: Square, piece: Piece) {
        self.clear_square(square);
        self.pieces[piece.bitboard_index()] |= square.bit();
        self.rebuild_occupancy();
    }

    pub fn clear_square(&mut self, square: Square) {
        let bit = square.bit();
        for bb in &mut self.pieces {
            *bb &= !bit;
        }
        self.rebuild_occupancy();
    }

    pub fn has_castling(&self, right: u8) -> bool {
        self.castling & right != 0
    }
}
