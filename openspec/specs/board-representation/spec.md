# board-representation Specification

## Purpose
Defines how the engine stores a chess position and exchanges it via FEN, including square addressing used by every later layer.

## Requirements

### Requirement: Algebraic square addressing
The system MUST map algebraic square names to little-endian rank-file indices where a1 = 0, h1 = 7, a8 = 56, and h8 = 63. Parsing MUST reject invalid square names.

#### Scenario: Map corner and center squares
- **GIVEN** algebraic names `a1`, `h1`, `a8`, `h8`, and `e4`
- **WHEN** each name is parsed to a square index
- **THEN** the indices are 0, 7, 56, 63, and 28 respectively

#### Scenario: Reject invalid square names
- **GIVEN** the strings `i1`, `a9`, and `e`
- **WHEN** each string is parsed as a square
- **THEN** parsing fails for each

### Requirement: Start position from standard FEN
The system MUST load the standard starting position from FEN `rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1` with correct piece placement, White to move, all castling rights, no en passant square, halfmove clock 0, and fullmove number 1.

#### Scenario: Load startpos pieces and side to move
- **GIVEN** the standard startpos FEN
- **WHEN** the position is parsed
- **THEN** White is to move
- **AND** the white king is on `e1` and the black king is on `e8`
- **AND** white pawns occupy rank 2 and black pawns occupy rank 7
- **AND** castling rights are `KQkq`
- **AND** there is no en passant square
- **AND** the halfmove clock is 0 and the fullmove number is 1

### Requirement: FEN round-trip
The system MUST serialize a loaded position back to a FEN string that matches the input for well-formed positions, including castling, en passant, and empty-square runs.

#### Scenario: Round-trip startpos
- **GIVEN** FEN `rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1`
- **WHEN** the FEN is parsed and then serialized
- **THEN** the output FEN equals the input

#### Scenario: Round-trip midgame with castling and en passant
- **GIVEN** FEN `rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2`
- **WHEN** the FEN is parsed and then serialized
- **THEN** the output FEN equals the input

#### Scenario: Round-trip position with reduced castling rights
- **GIVEN** FEN `r3k2r/8/8/8/8/8/8/R3K2R b kq - 5 20`
- **WHEN** the FEN is parsed and then serialized
- **THEN** the output FEN equals the input

### Requirement: Reject malformed FEN
The system MUST reject FEN strings that are missing fields, contain invalid piece characters, or describe an invalid board layout (wrong number of ranks or files).

#### Scenario: Reject missing fields
- **GIVEN** FEN `rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq`
- **WHEN** the FEN is parsed
- **THEN** parsing fails

#### Scenario: Reject invalid piece character
- **GIVEN** FEN `xnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1`
- **WHEN** the FEN is parsed
- **THEN** parsing fails

#### Scenario: Reject wrong number of files on a rank
- **GIVEN** FEN `rnbqkbnr/ppppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1`
- **WHEN** the FEN is parsed
- **THEN** parsing fails
