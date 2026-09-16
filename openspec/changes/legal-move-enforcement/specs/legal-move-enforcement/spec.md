## Purpose

Defines authoritative move validation and state-preservation behavior so illegal input cannot mutate or desynchronize the Rust engine and Go TUI positions.

## ADDED Requirements

### Requirement: Engine rejects an illegal position move transactionally
When a `position ... moves ...` command contains a malformed or illegal move, the engine MUST emit `info string error illegal move <token>` and MUST leave its current position unchanged. It MUST NOT adopt the legal prefix preceding the invalid token.

#### Scenario: Illegal first move preserves the current position
- **GIVEN** an engine whose current position was set by `position startpos moves e2e4`
- **WHEN** it receives `position startpos moves e2e5`
- **THEN** it emits `info string error illegal move e2e5`
- **AND** the current position remains the position after `e2e4`

#### Scenario: Illegal later move rejects the whole sequence
- **GIVEN** an engine at the standard starting position
- **WHEN** it receives `position startpos moves e2e4 e7e5 e1e3`
- **THEN** it emits `info string error illegal move e1e3`
- **AND** the current position remains the standard starting position

### Requirement: Engine exposes legal moves for its current position
The engine MUST accept the extension command `legalmoves` and emit exactly one response line beginning with `legalmoves`. Each following token MUST be a legal UCI long-algebraic move for the current position, with no illegal or duplicate tokens. When no legal move exists, the response MUST be exactly `legalmoves`.

#### Scenario: Start position advertises twenty legal moves
- **GIVEN** an engine after `position startpos`
- **WHEN** it receives `legalmoves`
- **THEN** the response contains exactly 20 move tokens
- **AND** the tokens include `e2e4` and `g1f3`
- **AND** the tokens do not include `e2e5` or `e1e2`

#### Scenario: Checkmate advertises no legal moves
- **GIVEN** an engine after `position fen 7k/6Q1/6K1/8/8/8/8/8 b - - 0 1`
- **WHEN** it receives `legalmoves`
- **THEN** the response is exactly `legalmoves`

### Requirement: TUI rejects human moves not authorized by the engine
Before accepting human input, the TUI MUST have the engine's legal-move set for the displayed position. A submitted move outside that set MUST produce a clear illegal-move error and MUST NOT change the display board, move list, engine position, or trigger `go`.

#### Scenario: Illegal pawn move is rejected without side effects
- **GIVEN** a connected TUI displaying the standard starting position and its engine-provided legal moves
- **WHEN** the human submits `e2e5`
- **THEN** the TUI displays an error identifying `e2e5` as illegal
- **AND** a white pawn remains on `e2` and `e5` remains empty
- **AND** the move list remains empty
- **AND** no new outbound `position` or `go` command is sent

#### Scenario: Moving the opponent's piece is rejected
- **GIVEN** a connected TUI displaying the standard starting position and its engine-provided legal moves
- **WHEN** the human submits `e7e5`
- **THEN** the TUI displays an error identifying `e7e5` as illegal
- **AND** the board and move list remain unchanged
- **AND** no new outbound `position` or `go` command is sent

### Requirement: Accepted plies keep engine and TUI synchronized
After each accepted human move and each engine `bestmove`, the TUI MUST synchronize the engine to the complete accepted move history and refresh the legal-move set before accepting another human move. An invalid or unadvertised engine `bestmove` MUST produce an error and MUST NOT mutate the TUI board or move list.

#### Scenario: Legal human and engine plies produce the same next legal set
- **GIVEN** a connected TUI at the standard starting position
- **WHEN** the human submits `e2e4` and the engine returns `bestmove e7e5`
- **THEN** the TUI board and move list reflect `e2e4 e7e5`
- **AND** the engine is synchronized with `position startpos moves e2e4 e7e5`
- **AND** the TUI refreshes legal moves for White before accepting another move

#### Scenario: Illegal engine bestmove is rejected
- **GIVEN** a connected TUI after accepting human move `e2e4`
- **WHEN** the engine returns `bestmove e2e5`
- **THEN** the TUI displays an error identifying the engine move as illegal
- **AND** the board and move list still contain only the accepted move `e2e4`
