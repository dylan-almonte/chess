//! Chess engine library — board representation, FEN, and move generation.

pub mod board;
pub mod fen;
pub mod moves;
pub mod piece;
pub mod square;

pub use board::Position;
pub use fen::{parse_fen, to_fen, FenError};
pub use moves::Move;
pub use piece::{Color, Piece, PieceKind};
pub use square::Square;
