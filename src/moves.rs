//! Move encoding and UCI long-algebraic display.

use std::fmt;
use std::str::FromStr;

use crate::piece::PieceKind;
use crate::square::Square;

pub const FLAG_CAPTURE: u8 = 1 << 0;
pub const FLAG_DOUBLE_PAWN: u8 = 1 << 1;
pub const FLAG_EN_PASSANT: u8 = 1 << 2;
pub const FLAG_CASTLE: u8 = 1 << 3;
pub const FLAG_PROMOTION: u8 = 1 << 4;

/// A chess move.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash)]
pub struct Move {
    pub from: Square,
    pub to: Square,
    pub promotion: Option<PieceKind>,
    pub flags: u8,
}

impl Move {
    pub fn new(from: Square, to: Square) -> Self {
        Self {
            from,
            to,
            promotion: None,
            flags: 0,
        }
    }

    pub fn with_flags(mut self, flags: u8) -> Self {
        self.flags |= flags;
        self
    }

    pub fn with_promotion(mut self, kind: PieceKind) -> Self {
        self.promotion = Some(kind);
        self.flags |= FLAG_PROMOTION;
        self
    }

    pub fn is_capture(self) -> bool {
        self.flags & (FLAG_CAPTURE | FLAG_EN_PASSANT) != 0
    }

    pub fn is_en_passant(self) -> bool {
        self.flags & FLAG_EN_PASSANT != 0
    }

    pub fn is_castle(self) -> bool {
        self.flags & FLAG_CASTLE != 0
    }

    pub fn is_double_pawn(self) -> bool {
        self.flags & FLAG_DOUBLE_PAWN != 0
    }

    pub fn is_promotion(self) -> bool {
        self.flags & FLAG_PROMOTION != 0
    }
}

impl fmt::Display for Move {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}{}", self.from, self.to)?;
        if let Some(kind) = self.promotion {
            let c = match kind {
                PieceKind::Queen => 'q',
                PieceKind::Rook => 'r',
                PieceKind::Bishop => 'b',
                PieceKind::Knight => 'n',
                _ => return Err(fmt::Error),
            };
            write!(f, "{c}")?;
        }
        Ok(())
    }
}

impl FromStr for Move {
    type Err = ();

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        let bytes = s.as_bytes();
        if bytes.len() < 4 || bytes.len() > 5 {
            return Err(());
        }
        let from: Square = std::str::from_utf8(&bytes[0..2])
            .map_err(|_| ())?
            .parse()?;
        let to: Square = std::str::from_utf8(&bytes[2..4])
            .map_err(|_| ())?
            .parse()?;
        let mut mv = Move::new(from, to);
        if bytes.len() == 5 {
            let kind = match bytes[4] {
                b'q' => PieceKind::Queen,
                b'r' => PieceKind::Rook,
                b'b' => PieceKind::Bishop,
                b'n' => PieceKind::Knight,
                _ => return Err(()),
            };
            mv = mv.with_promotion(kind);
        }
        Ok(mv)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn uci_formats_quiet_and_promotion() {
        let quiet = Move::new("e2".parse().unwrap(), "e4".parse().unwrap());
        assert_eq!(quiet.to_string(), "e2e4");

        let promo = Move::new("e7".parse().unwrap(), "e8".parse().unwrap())
            .with_promotion(PieceKind::Queen);
        assert_eq!(promo.to_string(), "e7e8q");
    }
}
