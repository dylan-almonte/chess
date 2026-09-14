//! Chess engine library — board representation, FEN, move generation, and UCI.

pub mod board;
pub mod eval;
pub mod fen;
pub mod makemove;
pub mod movegen;
pub mod moves;
pub mod perft;
pub mod piece;
pub mod square;
pub mod uci;

pub use board::Position;
pub use eval::evaluate;
pub use fen::{parse_fen, to_fen, FenError};
pub use makemove::{make_move, unmake_move, Undo};
pub use movegen::{generate_legal, generate_pseudo_legal, is_square_attacked};
pub use moves::Move;
pub use perft::perft;
pub use piece::{Color, Piece, PieceKind};
pub use square::Square;
pub use uci::{UciAction, UciSession};
