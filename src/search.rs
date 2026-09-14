//! Search: iterative-deepening negamax with alpha-beta.

use crate::board::Position;
use crate::eval::evaluate;
use crate::makemove::{make_move, unmake_move};
use crate::movegen::{generate_legal, in_check};
use crate::moves::Move;
use crate::piece::Color;

/// Mate score in centipawns; distance from the root is subtracted so shorter mates score higher.
pub const MATE: i32 = 30_000;
const INF: i32 = 32_000;

/// Result of a search to a given maximum depth.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct SearchResult {
    pub best_move: Option<Move>,
    /// Side-to-move relative score in centipawns (positive = STM is better).
    pub score: i32,
    pub pv: Vec<Move>,
    pub nodes: u64,
}

/// Search `pos` to `max_depth` plies.
pub fn search(pos: &Position, max_depth: u32) -> SearchResult {
    let mut pos = pos.clone();
    let mut nodes = 0;
    let (score, pv) = negamax(&mut pos, max_depth, 0, &mut nodes);
    SearchResult {
        best_move: pv.first().copied(),
        score,
        pv,
        nodes,
    }
}

fn stm_eval(pos: &Position) -> i32 {
    let score = evaluate(pos);
    match pos.side_to_move {
        Color::White => score,
        Color::Black => -score,
    }
}

fn negamax(pos: &mut Position, depth: u32, ply: i32, nodes: &mut u64) -> (i32, Vec<Move>) {
    *nodes += 1;

    if depth == 0 && !in_check(pos, pos.side_to_move) {
        return (stm_eval(pos), Vec::new());
    }

    let moves = generate_legal(pos);
    if moves.is_empty() {
        if in_check(pos, pos.side_to_move) {
            return (-MATE + ply, Vec::new());
        }
        return (0, Vec::new());
    }

    if depth == 0 {
        return (stm_eval(pos), Vec::new());
    }

    let mut best_score = -INF;
    let mut best_pv = Vec::new();
    for mv in moves {
        let undo = make_move(pos, mv);
        let (child_score, child_pv) = negamax(pos, depth - 1, ply + 1, nodes);
        unmake_move(pos, mv, undo);
        let score = -child_score;
        if score > best_score {
            best_score = score;
            best_pv.clear();
            best_pv.push(mv);
            best_pv.extend(child_pv);
        }
    }
    (best_score, best_pv)
}

/// Convert a mate score to UCI-style mate-in-N moves (signed, STM-relative).
pub fn mate_in(score: i32) -> Option<i32> {
    if score.abs() <= MATE - 128 {
        return None;
    }
    let plies = MATE - score.abs();
    let moves = (plies + 1) / 2;
    Some(if score > 0 { moves } else { -moves })
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::fen::parse_fen;
    use crate::makemove::make_move;
    use crate::movegen::generate_legal;
    use crate::Position;

    const START: &str = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1";
    const HANGING_QUEEN: &str = "4k3/8/8/8/7q/8/8/4K2R w - - 0 1";
    const ROOK_MATE: &str = "6k1/4R3/6K1/8/8/8/8/8 w - - 0 1";
    const CHECKMATE: &str = "7k/6Q1/6K1/8/8/8/8/8 b - - 0 1";
    const STALEMATE: &str = "7k/5Q2/6K1/8/8/8/8/8 b - - 0 1";
    const DEFENDED_PAWN: &str = "8/8/3k4/3p4/4Q3/8/8/4K3 w - - 0 1";

    fn pos(fen: &str) -> Position {
        parse_fen(fen).expect("valid test FEN")
    }

    fn pv_is_legal(start: &Position, pv: &[Move]) -> bool {
        let mut p = start.clone();
        for mv in pv {
            if !generate_legal(&p).iter().any(|legal| legal == mv) {
                return false;
            }
            make_move(&mut p, *mv);
        }
        true
    }

    fn best_uci(result: &SearchResult) -> String {
        result
            .best_move
            .expect("expected a best move")
            .to_string()
    }

    #[test]
    fn search_depth_1_from_startpos_is_legal() {
        let start = pos(START);
        let result = search(&start, 1);
        let legal: Vec<String> = generate_legal(&start).iter().map(ToString::to_string).collect();
        let best = best_uci(&result);
        assert!(legal.contains(&best), "got {best}, legal={legal:?}");
        assert!(!result.pv.is_empty());
        assert_eq!(result.pv[0], result.best_move.unwrap());
        assert!(pv_is_legal(&start, &result.pv));
    }

    #[test]
    fn search_white_captures_the_hanging_queen_on_h4() {
        let result = search(&pos(HANGING_QUEEN), 1);
        assert_eq!(best_uci(&result), "h1h4");
    }

    #[test]
    fn search_white_rook_mates_on_the_back_rank() {
        let result = search(&pos(ROOK_MATE), 1);
        assert_eq!(best_uci(&result), "e7e8");
        assert_eq!(mate_in(result.score), Some(1));
    }

    #[test]
    fn search_checkmate_has_no_move() {
        let result = search(&pos(CHECKMATE), 1);
        assert!(result.best_move.is_none());
        assert!(
            mate_in(result.score).is_some_and(|n| n <= 0),
            "expected Black mated, score={}",
            result.score
        );
    }

    #[test]
    fn search_stalemate_scores_zero() {
        let result = search(&pos(STALEMATE), 1);
        assert!(result.best_move.is_none());
        assert_eq!(result.score, 0);
    }

    #[test]
    fn search_depth_1_takes_a_defended_pawn() {
        let result = search(&pos(DEFENDED_PAWN), 1);
        assert_eq!(best_uci(&result), "e4d5");
    }

    #[test]
    fn search_depth_2_refuses_the_same_capture() {
        let start = pos(DEFENDED_PAWN);
        let result = search(&start, 2);
        assert_ne!(best_uci(&result), "e4d5");
        assert!(pv_is_legal(&start, &result.pv));
    }
}
