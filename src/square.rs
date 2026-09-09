//! Square indexing: little-endian rank-file (a1 = 0 … h8 = 63).

use std::fmt;
use std::str::FromStr;

/// A square on the chessboard (0 = a1 … 63 = h8).
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash)]
pub struct Square(u8);

impl Square {
    pub const fn new(index: u8) -> Option<Self> {
        if index < 64 {
            Some(Self(index))
        } else {
            None
        }
    }

    /// # Safety
    /// `index` must be in 0..64.
    pub const fn from_index_unchecked(index: u8) -> Self {
        Self(index)
    }

    pub const fn index(self) -> u8 {
        self.0
    }

    pub const fn file(self) -> u8 {
        self.0 % 8
    }

    pub const fn rank(self) -> u8 {
        self.0 / 8
    }

    pub const fn from_file_rank(file: u8, rank: u8) -> Option<Self> {
        if file < 8 && rank < 8 {
            Some(Self(rank * 8 + file))
        } else {
            None
        }
    }

    pub const fn bit(self) -> u64 {
        1u64 << self.0
    }
}

impl FromStr for Square {
    type Err = ();

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        let bytes = s.as_bytes();
        if bytes.len() != 2 {
            return Err(());
        }
        let file = bytes[0];
        let rank = bytes[1];
        if !(b'a'..=b'h').contains(&file) || !(b'1'..=b'8').contains(&rank) {
            return Err(());
        }
        Ok(Self((rank - b'1') * 8 + (file - b'a')))
    }
}

impl fmt::Display for Square {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let file = (b'a' + self.file()) as char;
        let rank = (b'1' + self.rank()) as char;
        write!(f, "{file}{rank}")
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn map_corner_and_center_squares() {
        assert_eq!("a1".parse::<Square>().unwrap().index(), 0);
        assert_eq!("h1".parse::<Square>().unwrap().index(), 7);
        assert_eq!("a8".parse::<Square>().unwrap().index(), 56);
        assert_eq!("h8".parse::<Square>().unwrap().index(), 63);
        assert_eq!("e4".parse::<Square>().unwrap().index(), 28);
    }

    #[test]
    fn reject_invalid_square_names() {
        assert!("i1".parse::<Square>().is_err());
        assert!("a9".parse::<Square>().is_err());
        assert!("e".parse::<Square>().is_err());
    }
}
