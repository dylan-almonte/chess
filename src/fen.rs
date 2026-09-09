//! FEN parse and serialize.

use crate::board::{Position, CASTLE_BK, CASTLE_BQ, CASTLE_WK, CASTLE_WQ};
use crate::piece::{Color, Piece};
use crate::square::Square;

/// Errors produced while parsing FEN.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum FenError {
    MissingFields,
    InvalidPiecePlacement,
    InvalidSideToMove,
    InvalidCastling,
    InvalidEnPassant,
    InvalidHalfmove,
    InvalidFullmove,
}

/// Parse a FEN string into a [`Position`].
pub fn parse_fen(fen: &str) -> Result<Position, FenError> {
    let mut parts = fen.split_whitespace();
    let placement = parts.next().ok_or(FenError::MissingFields)?;
    let side = parts.next().ok_or(FenError::MissingFields)?;
    let castling = parts.next().ok_or(FenError::MissingFields)?;
    let ep = parts.next().ok_or(FenError::MissingFields)?;
    let halfmove = parts.next().ok_or(FenError::MissingFields)?;
    let fullmove = parts.next().ok_or(FenError::MissingFields)?;

    let mut pos = Position::empty();
    let ranks: Vec<&str> = placement.split('/').collect();
    if ranks.len() != 8 {
        return Err(FenError::InvalidPiecePlacement);
    }

    for (rank_from_top, rank_str) in ranks.iter().enumerate() {
        let rank = 7 - rank_from_top as u8;
        let mut file: u8 = 0;
        for ch in rank_str.chars() {
            if file > 8 {
                return Err(FenError::InvalidPiecePlacement);
            }
            if let Some(empty) = ch.to_digit(10) {
                if empty == 0 || empty > 8 {
                    return Err(FenError::InvalidPiecePlacement);
                }
                file = file.saturating_add(empty as u8);
                if file > 8 {
                    return Err(FenError::InvalidPiecePlacement);
                }
            } else {
                let piece = Piece::from_fen_char(ch).ok_or(FenError::InvalidPiecePlacement)?;
                if file >= 8 {
                    return Err(FenError::InvalidPiecePlacement);
                }
                let square = Square::from_file_rank(file, rank).unwrap();
                pos.pieces[piece.bitboard_index()] |= square.bit();
                file += 1;
            }
        }
        if file != 8 {
            return Err(FenError::InvalidPiecePlacement);
        }
    }
    pos.rebuild_occupancy();

    pos.side_to_move = match side {
        "w" => Color::White,
        "b" => Color::Black,
        _ => return Err(FenError::InvalidSideToMove),
    };

    pos.castling = 0;
    if castling == "-" {
        // none
    } else {
        for ch in castling.chars() {
            pos.castling |= match ch {
                'K' => CASTLE_WK,
                'Q' => CASTLE_WQ,
                'k' => CASTLE_BK,
                'q' => CASTLE_BQ,
                _ => return Err(FenError::InvalidCastling),
            };
        }
    }

    pos.en_passant = if ep == "-" {
        None
    } else {
        Some(ep.parse::<Square>().map_err(|_| FenError::InvalidEnPassant)?)
    };

    pos.halfmove_clock = halfmove
        .parse()
        .map_err(|_| FenError::InvalidHalfmove)?;
    pos.fullmove_number = fullmove
        .parse()
        .map_err(|_| FenError::InvalidFullmove)?;
    if pos.fullmove_number == 0 {
        return Err(FenError::InvalidFullmove);
    }

    Ok(pos)
}

/// Serialize a [`Position`] to a FEN string.
pub fn to_fen(pos: &Position) -> String {
    let mut placement = String::new();
    for rank in (0..8).rev() {
        let mut empty = 0u8;
        for file in 0..8 {
            let square = Square::from_file_rank(file, rank).unwrap();
            match pos.piece_at(square) {
                None => empty += 1,
                Some(piece) => {
                    if empty > 0 {
                        placement.push(char::from(b'0' + empty));
                        empty = 0;
                    }
                    placement.push(piece.to_fen_char());
                }
            }
        }
        if empty > 0 {
            placement.push(char::from(b'0' + empty));
        }
        if rank > 0 {
            placement.push('/');
        }
    }

    let side = match pos.side_to_move {
        Color::White => 'w',
        Color::Black => 'b',
    };

    let mut castling = String::new();
    if pos.has_castling(CASTLE_WK) {
        castling.push('K');
    }
    if pos.has_castling(CASTLE_WQ) {
        castling.push('Q');
    }
    if pos.has_castling(CASTLE_BK) {
        castling.push('k');
    }
    if pos.has_castling(CASTLE_BQ) {
        castling.push('q');
    }
    if castling.is_empty() {
        castling.push('-');
    }

    let ep = pos
        .en_passant
        .map(|s| s.to_string())
        .unwrap_or_else(|| "-".to_string());

    format!(
        "{placement} {side} {castling} {ep} {} {}",
        pos.halfmove_clock, pos.fullmove_number
    )
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::board::{CASTLE_BK, CASTLE_BQ, CASTLE_WK, CASTLE_WQ};
    use crate::piece::{Color, Piece, PieceKind};

    const START_FEN: &str = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1";

    #[test]
    fn load_startpos_pieces_and_side_to_move() {
        let pos = parse_fen(START_FEN).unwrap();
        assert_eq!(pos.side_to_move, Color::White);

        let e1 = "e1".parse::<Square>().unwrap();
        let e8 = "e8".parse::<Square>().unwrap();
        assert_eq!(
            pos.piece_at(e1),
            Some(Piece::new(Color::White, PieceKind::King))
        );
        assert_eq!(
            pos.piece_at(e8),
            Some(Piece::new(Color::Black, PieceKind::King))
        );

        for file in 0..8 {
            let w = Square::from_file_rank(file, 1).unwrap();
            let b = Square::from_file_rank(file, 6).unwrap();
            assert_eq!(
                pos.piece_at(w),
                Some(Piece::new(Color::White, PieceKind::Pawn))
            );
            assert_eq!(
                pos.piece_at(b),
                Some(Piece::new(Color::Black, PieceKind::Pawn))
            );
        }

        assert!(pos.has_castling(CASTLE_WK | CASTLE_WQ | CASTLE_BK | CASTLE_BQ));
        assert_eq!(pos.en_passant, None);
        assert_eq!(pos.halfmove_clock, 0);
        assert_eq!(pos.fullmove_number, 1);
    }

    #[test]
    fn round_trip_startpos() {
        let pos = parse_fen(START_FEN).unwrap();
        assert_eq!(to_fen(&pos), START_FEN);
    }

    #[test]
    fn round_trip_midgame_with_castling_and_en_passant() {
        let fen = "rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2";
        let pos = parse_fen(fen).unwrap();
        assert_eq!(to_fen(&pos), fen);
    }

    #[test]
    fn round_trip_position_with_reduced_castling_rights() {
        let fen = "r3k2r/8/8/8/8/8/8/R3K2R b kq - 5 20";
        let pos = parse_fen(fen).unwrap();
        assert_eq!(to_fen(&pos), fen);
    }

    #[test]
    fn reject_missing_fields() {
        let fen = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq";
        assert_eq!(parse_fen(fen), Err(FenError::MissingFields));
    }

    #[test]
    fn reject_invalid_piece_character() {
        let fen = "xnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1";
        assert_eq!(parse_fen(fen), Err(FenError::InvalidPiecePlacement));
    }

    #[test]
    fn reject_wrong_number_of_files_on_a_rank() {
        let fen = "rnbqkbnr/ppppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1";
        assert_eq!(parse_fen(fen), Err(FenError::InvalidPiecePlacement));
    }
}
