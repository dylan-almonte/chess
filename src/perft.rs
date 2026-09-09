//! Perft — recursive legal-move node counting.

use crate::board::Position;
use crate::makemove::{make_move, unmake_move};
use crate::movegen::generate_legal;

/// Count leaf nodes at the given depth using legal move generation.
pub fn perft(pos: &mut Position, depth: u32) -> u64 {
    if depth == 0 {
        return 1;
    }
    let moves = generate_legal(pos);
    if depth == 1 {
        return moves.len() as u64;
    }
    let mut nodes = 0u64;
    for mv in moves {
        let undo = make_move(pos, mv);
        nodes += perft(pos, depth - 1);
        unmake_move(pos, mv, undo);
    }
    nodes
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::fen::parse_fen;

    const START: &str = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1";

    #[test]
    fn perft_depth_1_from_startpos() {
        let mut pos = parse_fen(START).unwrap();
        assert_eq!(perft(&mut pos, 1), 20);
    }

    #[test]
    fn perft_depth_2_from_startpos() {
        let mut pos = parse_fen(START).unwrap();
        assert_eq!(perft(&mut pos, 2), 400);
    }

    #[test]
    fn perft_depth_3_from_startpos() {
        let mut pos = parse_fen(START).unwrap();
        assert_eq!(perft(&mut pos, 3), 8902);
    }

    #[test]
    fn perft_depth_4_from_startpos() {
        let mut pos = parse_fen(START).unwrap();
        assert_eq!(perft(&mut pos, 4), 197_281);
    }
}
