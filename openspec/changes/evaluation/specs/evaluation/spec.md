## Purpose

Defines static position evaluation so the engine can score a board in centipawns using material balance and piece-square tables before search exists.

## ADDED Requirements

### Requirement: White-relative centipawn score
The system MUST evaluate a position as a signed integer in centipawns from White's perspective: positive means White is better, negative means Black is better. The score MUST NOT depend on which side is to move.

#### Scenario: Starting position is equal
- **GIVEN** FEN `rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1`
- **WHEN** the position is evaluated
- **THEN** the score is `0`

#### Scenario: Side to move does not change the score
- **GIVEN** FEN `rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1` evaluates to score `S`
- **WHEN** the same placement is evaluated with FEN `rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1`
- **THEN** the score is also `S`

### Requirement: Material values
Material contribution MUST use these centipawn values: Pawn `100`, Knight `320`, Bishop `330`, Rook `500`, Queen `900`, King `0`. Material MUST be White pieces minus Black pieces (each piece counted once at its value).

#### Scenario: White up a queen is material-positive
- **GIVEN** FEN `rnb1kbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1` (startpos without Black's queen)
- **WHEN** the position is evaluated
- **THEN** the score is greater than `800`

#### Scenario: Black up a queen is material-negative
- **GIVEN** FEN `rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNB1KBNR w KQkq - 0 1` (startpos without White's queen)
- **WHEN** the position is evaluated
- **THEN** the score is less than `-800`

### Requirement: Piece-square tables
In addition to material, the evaluation MUST add a piece-square table (PST) bonus for every piece on its square. PST values MUST be mirrored for Black so that a color-symmetric position (including the starting position) still evaluates to `0`. Advancing a White pawn into the center MUST increase the score relative to that pawn on its starting square, all else equal.

#### Scenario: Central White pawn outscores starting pawn
- **GIVEN** FEN `4k3/8/8/8/4P3/8/8/4K3 w - - 0 1` (White pawn on e4)
- **AND** FEN `4k3/8/8/8/8/8/4P3/4K3 w - - 0 1` (White pawn on e2)
- **WHEN** both positions are evaluated
- **THEN** the e4 position's score is strictly greater than the e2 position's score

#### Scenario: Kings alone remain equal when mirrored
- **GIVEN** FEN `4k3/8/8/8/8/8/8/4K3 w - - 0 1`
- **WHEN** the position is evaluated
- **THEN** the score is `0`
