//! Search: iterative-deepening negamax with alpha-beta (stub until wired).

use crate::board::Position;
use crate::moves::Move;

/// Mate score in centipawns; distance from the root is subtracted so shorter mates score higher.
pub const MATE: i32 = 30_000;

/// Result of a search to a given maximum depth.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct SearchResult {
    pub best_move: Option<Move>,
    /// Side-to-move relative score in centipawns (positive = STM is better).
    pub score: i32,
    pub pv: Vec<Move>,
    pub nodes: u64,
}

impl SearchResult {
    fn empty() -> Self {
        Self {
            best_move: None,
            score: 0,
            pv: Vec::new(),
            nodes: 0,
        }
    }
}

/// Search `pos` to `max_depth` plies. Stub: no look-ahead yet.
pub fn search(_pos: &Position, _max_depth: u32) -> SearchResult {
    SearchResult::empty()
}
