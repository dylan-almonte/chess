## Context

Greenfield Rust crate with a stub `main`. See proposal.md for motivation and the board-representation delta for behavioral requirements. Conventions live in `openspec/project.md`.

## Goals / Non-Goals

**Goals:**

- Bitboard-backed `Position` that can load/save FEN and answer piece-on-square queries
- Clear module split so move generation can plug in without rewriting state
- Tests named after every delta-spec scenario

**Non-Goals:**

- Attack generation, legality, make/unmake
- Magic bitboards or precomputed attack tables
- UCI I/O in `main.rs`

## Decisions

### Square indexing: little-endian rank-file (LERF)

- a1 = 0, b1 = 1, …, h1 = 7, a2 = 8, …, h8 = 63
- Bit `n` of a bitboard corresponds to square index `n`
- **Why:** Matches common engine literature and UCI/FEN mental models (a1 is bottom-left for White)
- **Alternative considered:** a8 = 0 (mailbox engines) — rejected for bitboard friendliness

### Board representation: 12 piece bitboards

- One `u64` per (color × piece kind): WP, WN, WB, WR, WQ, WK, BP, BN, BB, BR, BQ, BK
- Derived occupancy: white, black, all (recomputed on load; later make/unmake can update incrementally)
- Side to move, castling rights bitmask (`KQkq`), optional en passant square, halfmove clock, fullmove number
- **Why:** Direct path to fast move gen later; occupancy queries are bit ops
- **Alternative considered:** 8×8 mailbox array — simpler for teaching, slower and awkward for later layers

### Module layout

| Module | Responsibility |
|--------|----------------|
| `square` | `Square` newtype, file/rank, algebraic parse/display |
| `piece` | `Color`, `PieceKind`, `Piece` |
| `board` | `Position`, startpos helper, piece queries |
| `fen` | Parse FEN → `Position`, serialize `Position` → FEN |
| `lib` | Re-exports public API |

`main.rs` remains `println!` / empty until `uci-protocol`.

### FEN handling

- Parse six fields; fail fast on structural errors (rank file count ≠ 8, invalid chars, missing fields)
- Do not validate “chess legality” beyond structure (e.g. two kings) in this change — that belongs with move gen / validation later if needed
- Serialize empty runs with digits 1–8; castling order always `KQkq` subset or `-`

### Error type

- Single `FenError` / `ParseError` enum returned by parse APIs (no panics on bad input)

## Risks / Trade-offs

- [No legality on load] → Mitigation: document that FEN load is structural only; later layers filter illegal moves
- [Twelve bitboards + derived state can drift] → Mitigation: recompute occupancy from piece boards after every FEN load; keep a single `rebuild_occupancy` helper for make/unmake later
- [Edition 2024] → Mitigation: stick to stable APIs; run `cargo test` as the gate

## Migration Plan

N/A (greenfield). After archive, main spec lives at `openspec/specs/board-representation/spec.md`.
