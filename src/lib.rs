//! Chess engine library — board representation and FEN.

pub mod board;
pub mod fen;
pub mod piece;
pub mod square;

pub use piece::{Color, Piece, PieceKind};
