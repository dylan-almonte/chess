## Purpose

Defines depth-limited look-ahead so the engine can choose a move by searching legal lines with iterative deepening, rather than returning the first generated move or a static eval of the current node only.

## ADDED Requirements

### Requirement: Search returns a legal principal variation
Given a position with at least one legal move and a search depth of `1` or more, search MUST return a best move that is legal in that position. When a principal variation (PV) is returned, its first move MUST be that best move, and each subsequent PV move MUST be legal in the position reached by playing the preceding PV moves.

#### Scenario: Depth-1 search from startpos is legal
- **GIVEN** FEN `rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1`
- **WHEN** the position is searched to depth `1`
- **THEN** the best move is one of the 20 legal startpos moves
- **AND** the PV is non-empty and starts with that best move

### Requirement: Depth-1 search takes a hanging piece
At depth `1`, search MUST choose a recapture-free capture of a hanging opponent queen when that capture is available and no stronger depth-1 alternative exists.

#### Scenario: White captures the hanging queen on h4
- **GIVEN** FEN `4k3/8/8/8/7q/8/8/4K2R w - - 0 1` (White rook h1, Black queen h4, kings on e1/e8)
- **WHEN** the position is searched to depth `1`
- **THEN** the best move is `h1h4`

### Requirement: Search finds mate in one
When the side to move has a unique mating move, search at depth `1` MUST return that move, and the result MUST indicate mate in `1` for the side to move.

#### Scenario: White rook mates on the back rank
- **GIVEN** FEN `6k1/4R3/6K1/8/8/8/8/8 w - - 0 1`
- **WHEN** the position is searched to depth `1`
- **THEN** the best move is `e7e8`
- **AND** the result indicates mate in `1`

### Requirement: Terminal positions are mate or stalemate
When the side to move has no legal moves, search MUST NOT return a best move. If that side is in check, the result MUST indicate the side to move is mated. If that side is not in check, the score MUST be `0` (stalemate).

#### Scenario: Checkmate has no move
- **GIVEN** FEN `7k/6Q1/6K1/8/8/8/8/8 b - - 0 1`
- **WHEN** the position is searched to depth `1`
- **THEN** there is no best move
- **AND** the result indicates Black is mated

#### Scenario: Stalemate scores zero
- **GIVEN** FEN `7k/5Q2/6K1/8/8/8/8/8 b - - 0 1`
- **WHEN** the position is searched to depth `1`
- **THEN** there is no best move
- **AND** the score is `0`

### Requirement: Deeper search sees the opponent reply
Iterative deepening MUST search depth `1`, then `2`, up through the requested maximum. A capture that wins material at depth `1` MUST be rejected at depth `2` when the opponent recaptures and the net result is worse than not taking.

#### Scenario: Depth 1 takes a defended pawn
- **GIVEN** FEN `8/8/3k4/3p4/4Q3/8/8/4K3 w - - 0 1` (White queen e4, Black pawn d5 protected by the king on d6)
- **WHEN** the position is searched to depth `1`
- **THEN** the best move is `e4d5`

#### Scenario: Depth 2 refuses the same capture
- **GIVEN** FEN `8/8/3k4/3p4/4Q3/8/8/4K3 w - - 0 1`
- **WHEN** the position is searched to depth `2`
- **THEN** the best move is not `e4d5`
