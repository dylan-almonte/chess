## Context

Board representation, move generation, and UCI stub are archived. See proposal.md for motivation. `Position` already exposes 12 piece bitboards and LERF square indexing (a1=0 … h8=63). Evaluation is a pure library concern for this change; UCI still returns the first legal move.

## Goals / Non-Goals

**Goals:**

- `evaluate(&Position) -> i32` (centipawns, White-positive)
- Material + one middlegame PST set per piece kind
- Named Rust tests matching the evaluation spec scenarios

**Non-Goals:**

- Tapered mg/eg eval, game-phase blending, or endgame-only tables
- Mobility, pawn structure, king safety, tempo, bishop pair beyond what a fixed PST already encodes
- Wiring eval into UCI `go` or any search

## Decisions

### White-relative score, ignore side to move

- Sum White material+PST, subtract Black material+PST; do not flip by `side_to_move`
- **Why:** Search will negate at each ply (`-evaluate` after make); a fixed White-centric leaf keeps PST tables simple
- **Alternative:** side-to-move relative eval — deferred; would require flipping every call site

### Material constants

| Piece  | CP  |
|--------|-----|
| Pawn   | 100 |
| Knight | 320 |
| Bishop | 330 |
| Rook   | 500 |
| Queen  | 900 |
| King   | 0   |

- King value is 0 because checkmate is a search concern; king placement still gets a PST
- **Why:** Classic textbook values; matches the evaluation spec
- **Alternative:** S-like or PeSTO material — unnecessary until tapered eval

### Mirrored piece-square tables

- Store one `[i32; 64]` per `PieceKind` from White's view (a1=0 … h8=63)
- Black piece on square `sq` uses `table[sq ^ 56]` (vertical flip) and contributes negatively
- Choose tables so startpos and mirrored king-only positions score `0`, and White pawn e4 > e2
- Concrete tables: use a simplified classic set (pawn center preference, knights toward center, bishops slightly center-preferring, rooks prefer 7th mildly, queen mild center, king middlegame toward back rank/center files). Exact arrays live in `src/eval` (or similar) and are the source of truth for golden tests beyond the inequality scenarios
- **Why:** One table per kind, no runtime color branch beyond XOR; satisfies symmetry requirements
- **Alternative:** 12 color-specific tables — doubles data with no gain at this layer

### Module layout

- New `src/eval.rs` (or `eval/mod.rs`) exporting `evaluate(pos: &Position) -> i32`
- Wire through `lib.rs`; no UCI changes
- **Why:** Matches existing one-concern-per-module layout (`fen`, `movegen`, `uci`)

### Iteration over bitboards

- For each of the 12 piece bitboards, scan set bits and accumulate `±(material[kind] + pst[kind][sq])`
- **Why:** Fits current board representation; clear and fast enough until search needs incremental eval
- **Alternative:** incremental eval on make/unmake — defer until search hotspots demand it

## Risks / Trade-offs

- [Fixed middlegame PST hurts endgames] → Mitigation: acceptable until tapered eval; kings stay on back-rank tables for now
- [Eval not visible via UCI] → Mitigation: library tests are the contract; search change will consume the API
- [PST choice is arbitrary] → Mitigation: lock tables in code; spec uses symmetry + inequalities so tables can be tuned without rewriting every scenario

## Migration Plan

N/A for users. After archive, main spec lives at `openspec/specs/evaluation/spec.md`.
