## Purpose

Defines the minimal UCI stdin/stdout protocol so a GUI or tool can load the engine, set a position, and receive a legal `bestmove`.

## ADDED Requirements

### Requirement: UCI handshake
When the engine receives `uci`, it MUST reply with an `id name` line, an `id author` line, and then `uciok`. When it receives `isready`, it MUST reply with `readyok`.

#### Scenario: Respond to uci with identity and uciok
- **GIVEN** a freshly started engine session
- **WHEN** the command `uci` is received
- **THEN** the output includes a line starting with `id name`
- **AND** the output includes a line starting with `id author`
- **AND** the output ends the handshake with a line `uciok`

#### Scenario: Respond to isready with readyok
- **GIVEN** an engine that has completed the UCI handshake
- **WHEN** the command `isready` is received
- **THEN** the output includes a line `readyok`

### Requirement: Position setup
The engine MUST accept `position startpos`, `position startpos moves …`, `position fen <fen>…`, and `position fen <fen> … moves …`, applying the resulting position as the current game state. Moves in the moves list MUST be UCI long-algebraic (e.g. `e2e4`, `e7e8q`).

#### Scenario: Position startpos sets the standard opening
- **GIVEN** an engine session after `uci`
- **WHEN** the command `position startpos` is received
- **AND** then `go` is received
- **THEN** `bestmove` is one of the 20 legal startpos moves

#### Scenario: Position fen plus moves applies the sequence
- **GIVEN** an engine session after `uci`
- **WHEN** the command `position fen rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1 moves e2e4 e7e5` is received
- **AND** then `go` is received
- **THEN** `bestmove` is a legal move in the position after `e2e4 e7e5`

### Requirement: Go returns a legal bestmove
On `go` (with or without ignored time/depth tokens), the engine MUST output exactly one `bestmove <move>` line where `<move>` is a legal UCI move for the current position. If the side to move has no legal moves, the engine MUST output `bestmove 0000`.

#### Scenario: Go from startpos returns a legal move
- **GIVEN** position set with `position startpos`
- **WHEN** the command `go` is received
- **THEN** the output includes a line matching `bestmove <uci>` where `<uci>` is one of the legal moves from startpos

#### Scenario: Go in checkmate returns null move
- **GIVEN** position set with `position fen 7k/6Q1/6K1/8/8/8/8/8 b - - 0 1`
- **WHEN** the command `go` is received
- **THEN** the output includes the line `bestmove 0000`

### Requirement: Quit ends the session
When the engine receives `quit`, it MUST stop reading further commands and exit successfully.

#### Scenario: Quit terminates the command loop
- **GIVEN** an engine session after `uci`
- **WHEN** the command `quit` is received
- **THEN** the session ends without requiring further input
