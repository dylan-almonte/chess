## Context

`board-representation` provides bitboard `Position`, squares (LERF a1=0…h8=63), pieces, and FEN. See proposal.md for motivation and the move-generation delta for behavioral requirements. Project conventions: no third-party chess crates; prefer make/unmake once move gen exists.

## Goals / Non-Goals

**Goals:**

- Correct legal move generation for all piece types and special moves
- Make/unmake with full state restore for perft and future search
- Perft matching startpos depths 1–4 as the correctness gate
- UCI long-algebraic move strings for listing/tests

**Non-Goals:**

- Magic bitboards or PEXT (use ray/step loops; can optimize later)
- Move ordering, transposition tables, or search
- Checkmate/stalemate API beyond “zero legal moves” (no dedicated status enum required this change)

## Decisions

### Pseudo-legal generate, then filter by king safety

- Generate attacks/pushes that ignore absolute pins and “moving into check”, then make each candidate and discard if the moving side’s king is attacked
- Castling: generate only when rights, empty path, and king/through/landing squares are not attacked (checked explicitly, not only via post-move filter)
- **Why:** Straightforward and correct; performance is fine through perft-4
- **Alternative:** Fully legal generation with pin masks — faster later, more complex now

### Make/unmake with an `Undo` snapshot

- `make_move` returns/stores `Undo { captured, castling, en_passant, halfmove, … }` and mutates `Position` in place
- `unmake_move(move, undo)` restores bitboards and state fields
- Update derived occupancy after make/unmake (or maintain incrementally in the same functions)
- **Why:** Matches project default and avoids allocating a board copy per node
- **Alternative:** Copy-make — simpler but slower for perft/search

### Move encoding

- Compact `Move`: from, to, optional promotion piece kind, flags (capture, EP, castle, double-pawn, promotion)
- Display as UCI: `from`+`to`+optional lowercase promotion char (`q/r/b/n`)
- **Why:** Enough for gen, make/unmake, and future UCI `bestmove`

### Slider attacks: ray loops (no magics)

- Precompute knight/king/pawn attack tables by square if useful; bishops/rooks/queens walk rays until blocked
- Occupancy from existing `all_occ` / side occupancies
- **Why:** Correctness first; magics are a later optimization if perft speed becomes painful

### Module layout

| Module | Responsibility |
|--------|----------------|
| `moves` | `Move`, flags, UCI string |
| `movegen` | Attack helpers, legal move list |
| `makemove` | `make_move` / `unmake_move` + `Undo` (or methods on `Position`) |
| `perft` | Recursive legal perft |
| `lib` | Re-export public API |

Keep FEN module unchanged behaviorally; extend `Position` only as needed for make/unmake.

### In-check detection

- Find king square of the side to test; if any opponent piece attacks that square, in check
- Reuse attack generators used by movegen

## Risks / Trade-offs

- [Subtle special-move bugs (EP discovered check, castling through check)] → Mitigation: dedicated scenario tests plus startpos perft-4 as integration gate; debug with divide/perft when counts diverge
- [Ray loops slower than magics] → Mitigation: Accept for this layer; profile only if perft-4 is unusably slow in CI
- [Filter-after-make costs] → Mitigation: Fine for perft-4; pin-aware gen deferred

## Migration Plan

N/A. After archive, main spec lives at `openspec/specs/move-generation/spec.md`.
