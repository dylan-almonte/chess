//! Attack generation and move generation.

use crate::board::{Position, CASTLE_BK, CASTLE_BQ, CASTLE_WK, CASTLE_WQ};
use crate::makemove::make_move;
use crate::moves::{Move, FLAG_CAPTURE, FLAG_CASTLE, FLAG_DOUBLE_PAWN, FLAG_EN_PASSANT};
use crate::piece::{Color, Piece, PieceKind};
use crate::square::Square;

const KNIGHT_DELTAS: [(i8, i8); 8] = [
    (1, 2),
    (2, 1),
    (2, -1),
    (1, -2),
    (-1, -2),
    (-2, -1),
    (-2, 1),
    (-1, 2),
];

const KING_DELTAS: [(i8, i8); 8] = [
    (1, 0),
    (1, 1),
    (0, 1),
    (-1, 1),
    (-1, 0),
    (-1, -1),
    (0, -1),
    (1, -1),
];

const BISHOP_DIRS: [(i8, i8); 4] = [(1, 1), (1, -1), (-1, 1), (-1, -1)];
const ROOK_DIRS: [(i8, i8); 4] = [(1, 0), (-1, 0), (0, 1), (0, -1)];

fn shift_square(sq: Square, df: i8, dr: i8) -> Option<Square> {
    let file = sq.file() as i8 + df;
    let rank = sq.rank() as i8 + dr;
    if (0..8).contains(&file) && (0..8).contains(&rank) {
        Square::from_file_rank(file as u8, rank as u8)
    } else {
        None
    }
}

fn bits(bb: u64) -> impl Iterator<Item = Square> {
    let mut bb = bb;
    std::iter::from_fn(move || {
        if bb == 0 {
            None
        } else {
            let idx = bb.trailing_zeros() as u8;
            bb &= bb - 1;
            Some(Square::from_index_unchecked(idx))
        }
    })
}

pub fn pawn_attacks(sq: Square, color: Color) -> u64 {
    let mut attacks = 0u64;
    let forward: i8 = match color {
        Color::White => 1,
        Color::Black => -1,
    };
    for df in [-1i8, 1] {
        if let Some(target) = shift_square(sq, df, forward) {
            attacks |= target.bit();
        }
    }
    attacks
}

pub fn knight_attacks(sq: Square) -> u64 {
    let mut attacks = 0u64;
    for (df, dr) in KNIGHT_DELTAS {
        if let Some(target) = shift_square(sq, df, dr) {
            attacks |= target.bit();
        }
    }
    attacks
}

pub fn king_attacks(sq: Square) -> u64 {
    let mut attacks = 0u64;
    for (df, dr) in KING_DELTAS {
        if let Some(target) = shift_square(sq, df, dr) {
            attacks |= target.bit();
        }
    }
    attacks
}

fn sliding_attacks(sq: Square, occ: u64, dirs: &[(i8, i8)]) -> u64 {
    let mut attacks = 0u64;
    for &(df, dr) in dirs {
        let mut cur = sq;
        while let Some(next) = shift_square(cur, df, dr) {
            attacks |= next.bit();
            if occ & next.bit() != 0 {
                break;
            }
            cur = next;
        }
    }
    attacks
}

pub fn bishop_attacks(sq: Square, occ: u64) -> u64 {
    sliding_attacks(sq, occ, &BISHOP_DIRS)
}

pub fn rook_attacks(sq: Square, occ: u64) -> u64 {
    sliding_attacks(sq, occ, &ROOK_DIRS)
}

pub fn queen_attacks(sq: Square, occ: u64) -> u64 {
    bishop_attacks(sq, occ) | rook_attacks(sq, occ)
}

pub fn king_square(pos: &Position, color: Color) -> Option<Square> {
    let bb = pos.pieces[Piece::new(color, PieceKind::King).bitboard_index()];
    if bb == 0 {
        None
    } else {
        Some(Square::from_index_unchecked(bb.trailing_zeros() as u8))
    }
}

/// Returns true if `square` is attacked by `by_color`.
pub fn is_square_attacked(pos: &Position, square: Square, by_color: Color) -> bool {
    let occ = pos.all_occ;
    let them = by_color;

    let pawn_bb = pos.pieces[Piece::new(them, PieceKind::Pawn).bitboard_index()];
    if pawn_attacks(square, them.opposite()) & pawn_bb != 0 {
        return true;
    }

    let knight_bb = pos.pieces[Piece::new(them, PieceKind::Knight).bitboard_index()];
    if knight_attacks(square) & knight_bb != 0 {
        return true;
    }

    let king_bb = pos.pieces[Piece::new(them, PieceKind::King).bitboard_index()];
    if king_attacks(square) & king_bb != 0 {
        return true;
    }

    let bishop_like = pos.pieces[Piece::new(them, PieceKind::Bishop).bitboard_index()]
        | pos.pieces[Piece::new(them, PieceKind::Queen).bitboard_index()];
    if bishop_attacks(square, occ) & bishop_like != 0 {
        return true;
    }

    let rook_like = pos.pieces[Piece::new(them, PieceKind::Rook).bitboard_index()]
        | pos.pieces[Piece::new(them, PieceKind::Queen).bitboard_index()];
    if rook_attacks(square, occ) & rook_like != 0 {
        return true;
    }

    false
}

pub fn in_check(pos: &Position, color: Color) -> bool {
    match king_square(pos, color) {
        Some(k) => is_square_attacked(pos, k, color.opposite()),
        None => false,
    }
}

fn us_occ(pos: &Position, color: Color) -> u64 {
    match color {
        Color::White => pos.white_occ,
        Color::Black => pos.black_occ,
    }
}

fn them_occ(pos: &Position, color: Color) -> u64 {
    match color {
        Color::White => pos.black_occ,
        Color::Black => pos.white_occ,
    }
}

fn push_quiet_or_capture(list: &mut Vec<Move>, from: Square, to: Square, them: u64) {
    let mut mv = Move::new(from, to);
    if them & to.bit() != 0 {
        mv = mv.with_flags(FLAG_CAPTURE);
    }
    list.push(mv);
}

fn add_promotions(list: &mut Vec<Move>, from: Square, to: Square, capture: bool) {
    for kind in [
        PieceKind::Queen,
        PieceKind::Rook,
        PieceKind::Bishop,
        PieceKind::Knight,
    ] {
        let mut mv = Move::new(from, to).with_promotion(kind);
        if capture {
            mv = mv.with_flags(FLAG_CAPTURE);
        }
        list.push(mv);
    }
}

fn gen_pawn_moves(pos: &Position, color: Color, list: &mut Vec<Move>) {
    let pawns = pos.pieces[Piece::new(color, PieceKind::Pawn).bitboard_index()];
    let occ = pos.all_occ;
    let them = them_occ(pos, color);
    let (forward, start_rank, promo_rank, ep_capture_rank): (i8, u8, u8, u8) = match color {
        Color::White => (1, 1, 7, 4),
        Color::Black => (-1, 6, 0, 3),
    };

    for from in bits(pawns) {
        // Single push
        if let Some(one) = shift_square(from, 0, forward) {
            if occ & one.bit() == 0 {
                if one.rank() == promo_rank {
                    add_promotions(list, from, one, false);
                } else {
                    list.push(Move::new(from, one));
                    // Double push
                    if from.rank() == start_rank {
                        if let Some(two) = shift_square(from, 0, forward * 2) {
                            if occ & two.bit() == 0 {
                                list.push(
                                    Move::new(from, two).with_flags(FLAG_DOUBLE_PAWN),
                                );
                            }
                        }
                    }
                }
            }
        }

        // Captures
        for df in [-1i8, 1] {
            if let Some(to) = shift_square(from, df, forward) {
                if them & to.bit() != 0 {
                    if to.rank() == promo_rank {
                        add_promotions(list, from, to, true);
                    } else {
                        list.push(Move::new(from, to).with_flags(FLAG_CAPTURE));
                    }
                }
            }
        }

        // En passant
        if let Some(ep) = pos.en_passant {
            if from.rank() == ep_capture_rank && pawn_attacks(from, color) & ep.bit() != 0 {
                list.push(Move::new(from, ep).with_flags(FLAG_EN_PASSANT | FLAG_CAPTURE));
            }
        }
    }
}

fn gen_knight_moves(pos: &Position, color: Color, list: &mut Vec<Move>) {
    let pieces = pos.pieces[Piece::new(color, PieceKind::Knight).bitboard_index()];
    let ours = us_occ(pos, color);
    let them = them_occ(pos, color);
    for from in bits(pieces) {
        let attacks = knight_attacks(from) & !ours;
        for to in bits(attacks) {
            push_quiet_or_capture(list, from, to, them);
        }
    }
}

fn gen_king_moves(pos: &Position, color: Color, list: &mut Vec<Move>) {
    let pieces = pos.pieces[Piece::new(color, PieceKind::King).bitboard_index()];
    let ours = us_occ(pos, color);
    let them = them_occ(pos, color);
    for from in bits(pieces) {
        let attacks = king_attacks(from) & !ours;
        for to in bits(attacks) {
            push_quiet_or_capture(list, from, to, them);
        }
    }
}

fn gen_sliding_moves(
    pos: &Position,
    color: Color,
    kind: PieceKind,
    dirs: &[(i8, i8)],
    list: &mut Vec<Move>,
) {
    let pieces = pos.pieces[Piece::new(color, kind).bitboard_index()];
    let ours = us_occ(pos, color);
    let them = them_occ(pos, color);
    let occ = pos.all_occ;
    for from in bits(pieces) {
        let attacks = sliding_attacks(from, occ, dirs) & !ours;
        for to in bits(attacks) {
            push_quiet_or_capture(list, from, to, them);
        }
    }
}

fn gen_castling(pos: &Position, color: Color, list: &mut Vec<Move>) {
    let occ = pos.all_occ;
    let enemy = color.opposite();

    match color {
        Color::White => {
            let e1 = Square::from_index_unchecked(4);
            if pos.has_castling(CASTLE_WK)
                && occ & (Square::from_index_unchecked(5).bit() | Square::from_index_unchecked(6).bit())
                    == 0
                && !is_square_attacked(pos, e1, enemy)
                && !is_square_attacked(pos, Square::from_index_unchecked(5), enemy)
                && !is_square_attacked(pos, Square::from_index_unchecked(6), enemy)
            {
                list.push(
                    Move::new(e1, Square::from_index_unchecked(6)).with_flags(FLAG_CASTLE),
                );
            }
            if pos.has_castling(CASTLE_WQ)
                && occ
                    & (Square::from_index_unchecked(1).bit()
                        | Square::from_index_unchecked(2).bit()
                        | Square::from_index_unchecked(3).bit())
                    == 0
                && !is_square_attacked(pos, e1, enemy)
                && !is_square_attacked(pos, Square::from_index_unchecked(3), enemy)
                && !is_square_attacked(pos, Square::from_index_unchecked(2), enemy)
            {
                list.push(
                    Move::new(e1, Square::from_index_unchecked(2)).with_flags(FLAG_CASTLE),
                );
            }
        }
        Color::Black => {
            let e8 = Square::from_index_unchecked(60);
            if pos.has_castling(CASTLE_BK)
                && occ
                    & (Square::from_index_unchecked(61).bit()
                        | Square::from_index_unchecked(62).bit())
                    == 0
                && !is_square_attacked(pos, e8, enemy)
                && !is_square_attacked(pos, Square::from_index_unchecked(61), enemy)
                && !is_square_attacked(pos, Square::from_index_unchecked(62), enemy)
            {
                list.push(
                    Move::new(e8, Square::from_index_unchecked(62)).with_flags(FLAG_CASTLE),
                );
            }
            if pos.has_castling(CASTLE_BQ)
                && occ
                    & (Square::from_index_unchecked(57).bit()
                        | Square::from_index_unchecked(58).bit()
                        | Square::from_index_unchecked(59).bit())
                    == 0
                && !is_square_attacked(pos, e8, enemy)
                && !is_square_attacked(pos, Square::from_index_unchecked(59), enemy)
                && !is_square_attacked(pos, Square::from_index_unchecked(58), enemy)
            {
                list.push(
                    Move::new(e8, Square::from_index_unchecked(58)).with_flags(FLAG_CASTLE),
                );
            }
        }
    }
}

/// Generate pseudo-legal moves (may leave own king in check). Castling already
/// requires path emptiness and non-attacked king/through/landing squares.
pub fn generate_pseudo_legal(pos: &Position) -> Vec<Move> {
    let color = pos.side_to_move;
    let mut list = Vec::with_capacity(64);
    gen_pawn_moves(pos, color, &mut list);
    gen_knight_moves(pos, color, &mut list);
    gen_sliding_moves(pos, color, PieceKind::Bishop, &BISHOP_DIRS, &mut list);
    gen_sliding_moves(pos, color, PieceKind::Rook, &ROOK_DIRS, &mut list);
    gen_sliding_moves(pos, color, PieceKind::Queen, &BISHOP_DIRS, &mut list);
    gen_sliding_moves(pos, color, PieceKind::Queen, &ROOK_DIRS, &mut list);
    gen_king_moves(pos, color, &mut list);
    gen_castling(pos, color, &mut list);
    list
}

/// Generate legal moves for the side to move.
pub fn generate_legal(pos: &Position) -> Vec<Move> {
    let color = pos.side_to_move;
    let mut legal = Vec::new();
    for mv in generate_pseudo_legal(pos) {
        let mut trial = pos.clone();
        let _undo = make_move(&mut trial, mv);
        if !in_check(&trial, color) {
            legal.push(mv);
        }
    }
    legal
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::fen::parse_fen;

    #[test]
    fn king_not_attacked_on_startpos_e1() {
        let pos = parse_fen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1").unwrap();
        let e1 = "e1".parse().unwrap();
        assert!(!is_square_attacked(&pos, e1, Color::Black));
    }

    #[test]
    fn king_attacked_by_queen_check() {
        let pos = parse_fen("4k3/8/8/8/8/8/4q3/4K3 w - - 0 1").unwrap();
        let e1 = "e1".parse().unwrap();
        assert!(is_square_attacked(&pos, e1, Color::Black));
    }

    #[test]
    fn startpos_has_twenty_legal_moves() {
        let pos = parse_fen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1").unwrap();
        let moves = generate_legal(&pos);
        assert_eq!(moves.len(), 20);
        let uci: Vec<String> = moves.iter().map(|m| m.to_string()).collect();
        assert!(uci.iter().any(|m| m == "e2e4"));
        assert!(uci.iter().any(|m| m == "g1f3"));
        assert!(!uci.iter().any(|m| m == "e1e2"));
    }

    #[test]
    fn kingside_castling_is_legal_when_path_is_clear() {
        let pos = parse_fen("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1").unwrap();
        let uci: Vec<String> = generate_legal(&pos).iter().map(|m| m.to_string()).collect();
        assert!(uci.iter().any(|m| m == "e1g1"));
        assert!(uci.iter().any(|m| m == "e1c1"));
    }

    #[test]
    fn castling_omitted_when_king_would_pass_through_check() {
        let pos = parse_fen("r3k2r/8/8/8/8/8/4q3/R3K2R w KQkq - 0 1").unwrap();
        let uci: Vec<String> = generate_legal(&pos).iter().map(|m| m.to_string()).collect();
        assert!(!uci.iter().any(|m| m == "e1g1"));
    }

    #[test]
    fn en_passant_capture_is_generated() {
        let pos =
            parse_fen("rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3").unwrap();
        let uci: Vec<String> = generate_legal(&pos).iter().map(|m| m.to_string()).collect();
        assert!(uci.iter().any(|m| m == "e5f6"));
    }

    #[test]
    fn promotion_generates_all_four_promotion_pieces() {
        let pos = parse_fen("8/4P3/8/8/8/8/8/4K2k w - - 0 1").unwrap();
        let uci: Vec<String> = generate_legal(&pos).iter().map(|m| m.to_string()).collect();
        assert!(uci.iter().any(|m| m == "e7e8q"));
        assert!(uci.iter().any(|m| m == "e7e8r"));
        assert!(uci.iter().any(|m| m == "e7e8b"));
        assert!(uci.iter().any(|m| m == "e7e8n"));
    }

    #[test]
    fn pseudo_legal_covers_special_moves() {
        let castle = parse_fen("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1").unwrap();
        let castle_uci: Vec<_> = generate_pseudo_legal(&castle)
            .iter()
            .map(|m| m.to_string())
            .collect();
        assert!(castle_uci.iter().any(|m| m == "e1g1"));

        let promo = parse_fen("8/4P3/8/8/8/8/8/4K2k w - - 0 1").unwrap();
        let promo_uci: Vec<_> = generate_pseudo_legal(&promo)
            .iter()
            .map(|m| m.to_string())
            .collect();
        assert!(promo_uci.iter().any(|m| m == "e7e8q"));

        let ep =
            parse_fen("rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3").unwrap();
        let ep_uci: Vec<_> = generate_pseudo_legal(&ep)
            .iter()
            .map(|m| m.to_string())
            .collect();
        assert!(ep_uci.iter().any(|m| m == "e5f6"));
    }
}
