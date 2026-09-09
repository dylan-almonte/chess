//! Piece types and colors.

use std::fmt;

/// Side to move / piece color.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash)]
pub enum Color {
    White,
    Black,
}

impl Color {
    pub const fn opposite(self) -> Self {
        match self {
            Color::White => Color::Black,
            Color::Black => Color::White,
        }
    }
}

/// Kind of chess piece (ignoring color).
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash)]
pub enum PieceKind {
    Pawn,
    Knight,
    Bishop,
    Rook,
    Queen,
    King,
}

impl PieceKind {
    pub const ALL: [PieceKind; 6] = [
        PieceKind::Pawn,
        PieceKind::Knight,
        PieceKind::Bishop,
        PieceKind::Rook,
        PieceKind::Queen,
        PieceKind::King,
    ];

    pub const fn index(self) -> usize {
        match self {
            PieceKind::Pawn => 0,
            PieceKind::Knight => 1,
            PieceKind::Bishop => 2,
            PieceKind::Rook => 3,
            PieceKind::Queen => 4,
            PieceKind::King => 5,
        }
    }
}

/// A colored piece.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash)]
pub struct Piece {
    pub color: Color,
    pub kind: PieceKind,
}

impl Piece {
    pub const fn new(color: Color, kind: PieceKind) -> Self {
        Self { color, kind }
    }

    pub fn from_fen_char(c: char) -> Option<Self> {
        let (color, kind) = match c {
            'P' => (Color::White, PieceKind::Pawn),
            'N' => (Color::White, PieceKind::Knight),
            'B' => (Color::White, PieceKind::Bishop),
            'R' => (Color::White, PieceKind::Rook),
            'Q' => (Color::White, PieceKind::Queen),
            'K' => (Color::White, PieceKind::King),
            'p' => (Color::Black, PieceKind::Pawn),
            'n' => (Color::Black, PieceKind::Knight),
            'b' => (Color::Black, PieceKind::Bishop),
            'r' => (Color::Black, PieceKind::Rook),
            'q' => (Color::Black, PieceKind::Queen),
            'k' => (Color::Black, PieceKind::King),
            _ => return None,
        };
        Some(Self { color, kind })
    }

    pub fn to_fen_char(self) -> char {
        let c = match self.kind {
            PieceKind::Pawn => 'p',
            PieceKind::Knight => 'n',
            PieceKind::Bishop => 'b',
            PieceKind::Rook => 'r',
            PieceKind::Queen => 'q',
            PieceKind::King => 'k',
        };
        match self.color {
            Color::White => c.to_ascii_uppercase(),
            Color::Black => c,
        }
    }

    /// Index into the 12 piece bitboards: white 0..6, black 6..12.
    pub const fn bitboard_index(self) -> usize {
        self.color as usize * 6 + self.kind.index()
    }
}

impl fmt::Display for Piece {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.to_fen_char())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn fen_char_round_trip() {
        for c in "PNBRQKpnbrqk".chars() {
            let piece = Piece::from_fen_char(c).unwrap();
            assert_eq!(piece.to_fen_char(), c);
        }
        assert!(Piece::from_fen_char('x').is_none());
    }
}
