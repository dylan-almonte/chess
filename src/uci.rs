//! Minimal UCI protocol session (testable without a live GUI).

use std::str::FromStr;

use crate::board::Position;
use crate::fen::parse_fen;
use crate::makemove::make_move;
use crate::movegen::generate_legal;
use crate::moves::Move;

const START_FEN: &str = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1";
pub const ENGINE_NAME: &str = "chess";
pub const ENGINE_AUTHOR: &str = "dylanca";

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
            "position" => {
                self.handle_position(line);
                UciAction::Reply(vec![])
            }
            "go" => UciAction::Reply(vec![self.bestmove_line()]),
            "quit" => UciAction::Quit,
            _ => UciAction::Reply(vec![]),
        }
    }

    fn handle_position(&mut self, line: &str) {
        let rest = line.strip_prefix("position").unwrap_or("").trim_start();
        let (fen_part, moves_part) = split_moves(rest);

        let fen = if fen_part == "startpos" || fen_part.starts_with("startpos ") {
            START_FEN.to_string()
        } else if let Some(after) = fen_part.strip_prefix("fen ") {
            // Six FEN fields
            let fields: Vec<&str> = after.split_whitespace().take(6).collect();
            if fields.len() != 6 {
                return;
            }
            fields.join(" ")
        } else {
            return;
        };

        let Ok(mut pos) = parse_fen(&fen) else {
            return;
        };

        if let Some(moves_str) = moves_part {
            for token in moves_str.split_whitespace() {
                let Some(mv) = resolve_uci_move(&pos, token) else {
                    break;
                };
                make_move(&mut pos, mv);
            }
        }

        self.position = pos;
    }

    fn bestmove_line(&self) -> String {
        let moves = generate_legal(&self.position);
        match moves.first() {
            Some(mv) => format!("bestmove {mv}"),
            None => "bestmove 0000".to_string(),
        }
    }
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
        .find(|mv| {
            mv.from == parsed.from
                && mv.to == parsed.to
                && mv.promotion == parsed.promotion
        })
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
        let _ = replies(
            &mut session,
            "position fen 7k/6Q1/6K1/8/8/8/8/8 b - - 0 1",
        );
        let out = replies(&mut session, "go");
        assert!(out.iter().any(|l| l == "bestmove 0000"));
    }

    #[test]
    fn quit_terminates_the_command_loop() {
        let mut session = UciSession::new();
        let _ = replies(&mut session, "uci");
        assert_eq!(session.handle_line("quit"), UciAction::Quit);
    }
}
