## Purpose

Defines legal move generation, make/unmake, and perft so the engine can enumerate and verify chess moves before UCI or search.

## ADDED Requirements

### Requirement: Legal moves from the starting position
The system MUST generate exactly the legal moves for the side to move. From the standard starting position, there MUST be exactly 20 legal moves. Moves MUST be expressible in UCI long-algebraic form (e.g. `e2e4`, `g1f3`).

#### Scenario: Startpos has twenty legal moves
- **GIVEN** FEN `rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1`
- **WHEN** legal moves are generated
- **THEN** there are exactly 20 moves
- **AND** the set includes `e2e4` and `g1f3`
- **AND** the set does not include illegal king moves such as `e1e2`

### Requirement: Castling, en passant, and promotion
The system MUST generate legal castling, en passant captures, and pawn promotions (including underpromotion), and MUST omit illegal castling (through/into check or when rights are absent).

#### Scenario: Kingside castling is legal when path is clear
- **GIVEN** FEN `r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1`
- **WHEN** legal moves are generated for White
- **THEN** the set includes `e1g1` and `e1c1`

#### Scenario: Castling omitted when king would pass through check
- **GIVEN** FEN `r3k2r/8/8/8/8/8/4q3/R3K2R w KQkq - 0 1`
- **WHEN** legal moves are generated for White
- **THEN** the set does not include `e1g1`

#### Scenario: En passant capture is generated
- **GIVEN** FEN `rnbqkbnr/ppp1p1pp/8/3pPp2/8/8/PPPP1PPP/RNBQKBNR w KQkq f6 0 3`
- **WHEN** legal moves are generated for White
- **THEN** the set includes `e5f6`

#### Scenario: Promotion generates all four promotion pieces
- **GIVEN** FEN `8/4P3/8/8/8/8/8/4K2k w - - 0 1`
- **WHEN** legal moves are generated for White
- **THEN** the set includes `e7e8q`, `e7e8r`, `e7e8b`, and `e7e8n`

### Requirement: Make and unmake restore position
Applying a legal move and then undoing it MUST restore the prior position, including side to move, castling rights, en passant square, and clocks, such that the FEN matches the pre-move FEN.

#### Scenario: Make then unmake restores startpos FEN
- **GIVEN** FEN `rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1`
- **WHEN** the move `e2e4` is made and then unmade
- **THEN** serializing the position yields the original FEN

### Requirement: Perft matches startpos reference counts
The system MUST provide a perft (node-count) function over legal moves. From the standard starting position, perft MUST return the reference leaf counts below.

#### Scenario: Perft depth 1 from startpos
- **GIVEN** FEN `rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1`
- **WHEN** perft is run at depth 1
- **THEN** the node count is 20

#### Scenario: Perft depth 2 from startpos
- **GIVEN** FEN `rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1`
- **WHEN** perft is run at depth 2
- **THEN** the node count is 400

#### Scenario: Perft depth 3 from startpos
- **GIVEN** FEN `rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1`
- **WHEN** perft is run at depth 3
- **THEN** the node count is 8902

#### Scenario: Perft depth 4 from startpos
- **GIVEN** FEN `rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1`
- **WHEN** perft is run at depth 4
- **THEN** the node count is 197281
