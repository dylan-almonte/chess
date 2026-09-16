//! Minimal UCI protocol session (testable without a live GUI).

use std::str::FromStr;

use crate::board::Position;
use crate::fen::parse_fen;
use crate::makemove::make_move;
use crate::movegen::generate_legal;
use crate::moves::Move;
use crate::search::{SearchResult, mate_in, search_iter};

const START_FEN: &str = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1";
pub const ENGINE_NAME: &str = "chess";
pub const ENGINE_AUTHOR: &str = "dylanca";
const DEFAULT_GO_DEPTH: u32 = 4;

/// Result of handling one UCI input line.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum UciAction {
    /// Reply lines to write to stdout (without trailing newlines).
    Reply(Vec<String>),
    /// Session should exit successfully.
    Quit,
}

/// Mutable UCI engine state.
#[derive(Clone, Debug)]
pub struct UciSession {
    pub position: Position,
}

impl Default for UciSession {
    fn default() -> Self {
        Self::new()
    }
}

impl UciSession {
    pub fn new() -> Self {
        Self {
            position: parse_fen(START_FEN).expect("startpos FEN is valid"),
        }
    }

    /// Process one command line. Unknown commands yield an empty reply.
    pub fn handle_line(&mut self, line: &str) -> UciAction {
        let line = line.trim();
        if line.is_empty() {
            return UciAction::Reply(vec![]);
        }

        let mut parts = line.split_whitespace();
        let cmd = parts.next().unwrap_or("");

        match cmd {
            "uci" => UciAction::Reply(vec![
                format!("id name {ENGINE_NAME}"),
                format!("id author {ENGINE_AUTHOR}"),
                "uciok".to_string(),
            ]),
            "isready" => UciAction::Reply(vec!["readyok".to_string()]),
            "ucinewgame" => {
                self.position = parse_fen(START_FEN).expect("startpos FEN is valid");
                UciAction::Reply(vec![])
            }
            "position" => self.handle_position(line),
            "go" => UciAction::Reply(self.go_replies(line)),
            "legalmoves" => UciAction::Reply(vec![self.legalmoves_reply()]),
            "quit" => UciAction::Quit,
            _ => UciAction::Reply(vec![]),
        }
    }

    fn handle_position(&mut self, line: &str) -> UciAction {
        let rest = line.strip_prefix("position").unwrap_or("").trim_start();
        let (fen_part, moves_part) = split_moves(rest);

        let fen = if fen_part == "startpos" || fen_part.starts_with("startpos ") {
            START_FEN.to_string()
        } else if let Some(after) = fen_part.strip_prefix("fen ") {
            let fields: Vec<&str> = after.split_whitespace().take(6).collect();
            if fields.len() != 6 {
                return UciAction::Reply(vec![]);
            }
            fields.join(" ")
        } else {
            return UciAction::Reply(vec![]);
        };

        let Ok(mut candidate) = parse_fen(&fen) else {
            return UciAction::Reply(vec![]);
        };

        if let Some(moves_str) = moves_part {
            for token in moves_str.split_whitespace() {
                let Some(mv) = resolve_uci_move(&candidate, token) else {
                    return UciAction::Reply(vec![format!(
                        "info string error illegal move {token}"
                    )]);
                };
                make_move(&mut candidate, mv);
            }
        }

        self.position = candidate;
        UciAction::Reply(vec![])
    }

    fn legalmoves_reply(&self) -> String {
        let moves = generate_legal(&self.position);
        if moves.is_empty() {
            return "legalmoves".to_string();
        }
        let tokens: Vec<String> = moves.iter().map(ToString::to_string).collect();
        format!("legalmoves {}", tokens.join(" "))
    }

    fn go_replies(&self, line: &str) -> Vec<String> {
        let depth = parse_go_depth(line);
        let mut lines = Vec::new();
        let result = search_iter(&self.position, depth, |iter| {
            if iter.best_move.is_some() {
                lines.push(format_info(iter));
            }
        });
        match result.best_move {
            Some(mv) => lines.push(format!("bestmove {mv}")),
            None => lines.push("bestmove 0000".to_string()),
        }
        lines
    }
}

fn parse_go_depth(line: &str) -> u32 {
    let mut tokens = line.split_whitespace();
    while let Some(tok) = tokens.next() {
        if tok == "depth" {
            if let Some(n) = tokens.next().and_then(|s| s.parse::<u32>().ok()) {
                if n > 0 {
                    return n;
                }
            }
        }
    }
    DEFAULT_GO_DEPTH
}

fn format_info(result: &SearchResult) -> String {
    let score = match mate_in(result.score) {
        Some(n) => format!("mate {n}"),
        None => format!("cp {}", result.score),
    };
    let pv = result
        .pv
        .iter()
        .map(ToString::to_string)
        .collect::<Vec<_>>()
        .join(" ");
    format!(
        "info depth {} score {} nodes {} pv {pv}",
        result.depth, score, result.nodes
    )
}

fn split_moves(rest: &str) -> (&str, Option<&str>) {
    if let Some(idx) = rest.find(" moves ") {
        (&rest[..idx], Some(&rest[idx + " moves ".len()..]))
    } else if rest.ends_with(" moves") {
        (&rest[..rest.len() - " moves".len()], Some(""))
    } else {
        (rest, None)
    }
}

/// Match a UCI move string to a legal move so flags (castle, EP, etc.) are correct.
fn resolve_uci_move(pos: &Position, uci: &str) -> Option<Move> {
    let parsed = Move::from_str(uci).ok()?;
    generate_legal(pos)
        .into_iter()
        .find(|mv| mv.from == parsed.from && mv.to == parsed.to && mv.promotion == parsed.promotion)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::movegen::generate_legal;

    fn replies(session: &mut UciSession, line: &str) -> Vec<String> {
        match session.handle_line(line) {
            UciAction::Reply(lines) => lines,
            UciAction::Quit => panic!("unexpected quit"),
        }
    }

    #[test]
    fn respond_to_uci_with_identity_and_uciok() {
        let mut session = UciSession::new();
        let out = replies(&mut session, "uci");
        assert!(out.iter().any(|l| l.starts_with("id name ")));
        assert!(out.iter().any(|l| l.starts_with("id author ")));
        assert_eq!(out.last().map(String::as_str), Some("uciok"));
    }

    #[test]
    fn respond_to_isready_with_readyok() {
        let mut session = UciSession::new();
        let _ = replies(&mut session, "uci");
        let out = replies(&mut session, "isready");
        assert!(out.iter().any(|l| l == "readyok"));
    }

    #[test]
    fn position_startpos_sets_the_standard_opening() {
        let mut session = UciSession::new();
        let _ = replies(&mut session, "uci");
        let _ = replies(&mut session, "position startpos");
        let out = replies(&mut session, "go");
        let best = out
            .iter()
            .find(|l| l.starts_with("bestmove "))
            .expect("bestmove line");
        let mv = best.strip_prefix("bestmove ").unwrap();
        let legal: Vec<String> = generate_legal(&parse_fen(START_FEN).unwrap())
            .iter()
            .map(|m| m.to_string())
            .collect();
        assert!(legal.iter().any(|m| m == mv), "got {mv}, legal={legal:?}");
    }

    #[test]
    fn position_fen_plus_moves_applies_the_sequence() {
        let mut session = UciSession::new();
        let _ = replies(&mut session, "uci");
        let _ = replies(
            &mut session,
            "position fen rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1 moves e2e4 e7e5",
        );
        let out = replies(&mut session, "go");
        let best = out
            .iter()
            .find(|l| l.starts_with("bestmove "))
            .expect("bestmove line");
        let mv = best.strip_prefix("bestmove ").unwrap();

        let mut expected = parse_fen(START_FEN).unwrap();
        let m1 = resolve_uci_move(&expected, "e2e4").unwrap();
        make_move(&mut expected, m1);
        let m2 = resolve_uci_move(&expected, "e7e5").unwrap();
        make_move(&mut expected, m2);
        let legal: Vec<String> = generate_legal(&expected)
            .iter()
            .map(|m| m.to_string())
            .collect();
        assert!(legal.iter().any(|m| m == mv), "got {mv}, legal={legal:?}");
    }

    #[test]
    fn go_from_startpos_returns_a_legal_move() {
        let mut session = UciSession::new();
        let _ = replies(&mut session, "position startpos");
        let out = replies(&mut session, "go");
        let best = out
            .iter()
            .find(|l| l.starts_with("bestmove "))
            .expect("bestmove line");
        let mv = best.strip_prefix("bestmove ").unwrap();
        let legal: Vec<String> = generate_legal(&parse_fen(START_FEN).unwrap())
            .iter()
            .map(|m| m.to_string())
            .collect();
        assert!(legal.contains(&mv.to_string()));
    }

    #[test]
    fn go_in_checkmate_returns_null_move() {
        let mut session = UciSession::new();
        let _ = replies(&mut session, "position fen 7k/6Q1/6K1/8/8/8/8/8 b - - 0 1");
        let out = replies(&mut session, "go");
        assert!(out.iter().any(|l| l == "bestmove 0000"));
    }

    #[test]
    fn illegal_first_move_preserves_current_position() {
        let mut session = UciSession::new();
        // Set up a known position: startpos + e2e4
        let _ = replies(&mut session, "position startpos moves e2e4");
        let pos_after_e2e4 = session.position.clone();

        // Now send a position command with an illegal first move (e2e5 is not legal from startpos)
        let out = replies(&mut session, "position startpos moves e2e5");

        // Should emit error
        assert!(
            out.iter()
                .any(|l| l == "info string error illegal move e2e5"),
            "expected illegal move error, got {out:?}"
        );
        // Position should be unchanged (still after e2e4)
        assert_eq!(
            session.position, pos_after_e2e4,
            "position should remain after e2e4, not adopt the illegal sequence"
        );
    }

    #[test]
    fn illegal_later_move_rejects_whole_sequence() {
        let mut session = UciSession::new();
        // Engine starts at default startpos
        let startpos = session.position.clone();

        // Send a sequence where the third move is illegal (e1e3 is not legal after e2e4 e7e5)
        let out = replies(&mut session, "position startpos moves e2e4 e7e5 e1e3");

        // Should emit error
        assert!(
            out.iter()
                .any(|l| l == "info string error illegal move e1e3"),
            "expected illegal move error, got {out:?}"
        );
        // Position should remain the standard starting position (not adopt legal prefix)
        assert_eq!(
            session.position, startpos,
            "position should remain at startpos, not adopt legal prefix e2e4 e7e5"
        );
    }

    #[test]
    fn legalmoves_from_startpos_returns_exactly_twenty_unique_moves() {
        let mut session = UciSession::new();
        let _ = replies(&mut session, "position startpos");
        let out = replies(&mut session, "legalmoves");

        assert_eq!(
            out.len(),
            1,
            "expected exactly one response line, got {out:?}"
        );
        let line = &out[0];
        assert!(
            line.starts_with("legalmoves "),
            "response should start with 'legalmoves ', got {line:?}"
        );
        let moves: Vec<&str> = line
            .strip_prefix("legalmoves ")
            .unwrap()
            .split_whitespace()
            .collect();
        assert_eq!(
            moves.len(),
            20,
            "expected 20 legal moves, got {} : {moves:?}",
            moves.len()
        );

        // Verify no duplicates
        let mut unique = moves.clone();
        unique.sort();
        unique.dedup();
        assert_eq!(unique.len(), 20, "found duplicate moves in {moves:?}");

        // Must include e2e4 and g1f3
        assert!(moves.contains(&"e2e4"), "missing e2e4 in {moves:?}");
        assert!(moves.contains(&"g1f3"), "missing g1f3 in {moves:?}");

        // Must NOT include illegal moves
        assert!(
            !moves.contains(&"e2e5"),
            "e2e5 should not be legal in {moves:?}"
        );
        assert!(
            !moves.contains(&"e1e2"),
            "e1e2 should not be legal in {moves:?}"
        );
    }

    #[test]
    fn legalmoves_in_checkmate_is_empty() {
        let mut session = UciSession::new();
        let _ = replies(&mut session, "position fen 7k/6Q1/6K1/8/8/8/8/8 b - - 0 1");
        let out = replies(&mut session, "legalmoves");

        assert_eq!(
            out.len(),
            1,
            "expected exactly one response line, got {out:?}"
        );
        assert_eq!(
            out[0], "legalmoves",
            "checkmated position should respond with bare 'legalmoves'"
        );
    }

    #[test]
    fn quit_terminates_the_command_loop() {
        let mut session = UciSession::new();
        let _ = replies(&mut session, "uci");
        assert_eq!(session.handle_line("quit"), UciAction::Quit);
    }

    #[test]
    fn go_depth_1_still_returns_a_legal_move() {
        let mut session = UciSession::new();
        let _ = replies(&mut session, "position startpos");
        let out = replies(&mut session, "go depth 1");
        let best = out
            .iter()
            .find(|l| l.starts_with("bestmove "))
            .expect("bestmove line");
        let mv = best.strip_prefix("bestmove ").unwrap();
        let legal: Vec<String> = generate_legal(&parse_fen(START_FEN).unwrap())
            .iter()
            .map(|m| m.to_string())
            .collect();
        assert!(legal.contains(&mv.to_string()), "got {mv}, legal={legal:?}");
    }

    #[test]
    fn depth_1_go_reports_info_then_bestmove() {
        let mut session = UciSession::new();
        let _ = replies(&mut session, "position startpos");
        let out = replies(&mut session, "go depth 1");
        let info_i = out
            .iter()
            .position(|l| l.starts_with("info ") && l.contains("depth 1") && l.contains(" pv "))
            .expect("info depth 1 with pv");
        let best_i = out
            .iter()
            .position(|l| l.starts_with("bestmove "))
            .expect("bestmove");
        assert!(info_i < best_i, "info must precede bestmove: {out:?}");
    }

    #[test]
    fn depth_2_go_reports_both_iterations() {
        let mut session = UciSession::new();
        let _ = replies(&mut session, "position startpos");
        let out = replies(&mut session, "go depth 2");
        assert!(
            out.iter()
                .any(|l| l.starts_with("info ") && l.contains("depth 1")),
            "missing depth 1 info: {out:?}"
        );
        assert!(
            out.iter()
                .any(|l| l.starts_with("info ") && l.contains("depth 2")),
            "missing depth 2 info: {out:?}"
        );
        let last_info = out.iter().rposition(|l| l.starts_with("info ")).unwrap();
        let last_best = out
            .iter()
            .rposition(|l| l.starts_with("bestmove "))
            .unwrap();
        assert!(last_info < last_best, "info must precede bestmove: {out:?}");
    }

    #[test]
    fn depth_1_captures_the_hanging_queen() {
        let mut session = UciSession::new();
        let _ = replies(&mut session, "position fen 4k3/8/8/8/7q/8/8/4K2R w - - 0 1");
        let out = replies(&mut session, "go depth 1");
        assert!(
            out.iter().any(|l| l == "bestmove h1h4"),
            "expected bestmove h1h4, got {out:?}"
        );
    }

    #[test]
    fn depth_1_mates_with_the_rook() {
        let mut session = UciSession::new();
        let _ = replies(&mut session, "position fen 6k1/4R3/6K1/8/8/8/8/8 w - - 0 1");
        let out = replies(&mut session, "go depth 1");
        assert!(
            out.iter().any(|l| l == "bestmove e7e8"),
            "expected bestmove e7e8, got {out:?}"
        );
        assert!(
            out.iter()
                .any(|l| l.starts_with("info ") && l.contains("score mate 1")),
            "expected score mate 1, got {out:?}"
        );
    }
}
