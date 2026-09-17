# uci-protocol Specification

## Purpose
Defines the minimal UCI stdin/stdout protocol so a GUI or tool can load the engine, set a position, and receive a legal `bestmove`.

## Requirements

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
On `go`, the engine MUST output exactly one `bestmove <move>` line where `<move>` is a legal UCI move chosen by search for the current position, unless the command includes `infinite` (see below). If `depth N` is present (`N` a positive integer), search MUST use maximum depth `N`. If `depth` is omitted, search MUST use maximum depth `4`. Time-control tokens (`wtime`, `btime`, `movetime`) and `nodes` MUST be ignored. If the side to move has no legal moves, the `bestmove` line (when emitted) MUST be `bestmove 0000`.

#### Scenario: Go from startpos returns a legal move
- **GIVEN** position set with `position startpos`
- **WHEN** the command `go` is received
- **THEN** the output includes a line matching `bestmove <uci>` where `<uci>` is one of the legal moves from startpos

#### Scenario: Go in checkmate returns null move
- **GIVEN** position set with `position fen 7k/6Q1/6K1/8/8/8/8/8 b - - 0 1`
- **WHEN** the command `go` is received
- **THEN** the output includes the line `bestmove 0000`

#### Scenario: Go depth 1 still returns a legal move
- **GIVEN** position set with `position startpos`
- **WHEN** the command `go depth 1` is received
- **THEN** the output includes a line matching `bestmove <uci>` where `<uci>` is one of the legal moves from startpos

### Requirement: Go emits search info
On `go` in a position with at least one legal move, the engine MUST emit one or more `info` lines before `bestmove`. Each completed iterative-deepening iteration MUST produce an `info` line that includes `depth`, `score` (`cp <n>` or `mate <n>`), `nodes`, and `pv`. For `go depth N` with `N >= 1` and at least one legal move, the output MUST include an `info` line whose `depth` is `N`.

#### Scenario: Depth-1 go reports info then bestmove
- **GIVEN** position set with `position startpos`
- **WHEN** the command `go depth 1` is received
- **THEN** the output includes a line starting with `info` that contains `depth 1` and a `pv` token
- **AND** that `info` line appears before the `bestmove` line

#### Scenario: Depth-2 go reports both iterations
- **GIVEN** position set with `position startpos`
- **WHEN** the command `go depth 2` is received
- **THEN** the output includes an `info` line containing `depth 1`
- **AND** the output includes an `info` line containing `depth 2`
- **AND** the last `bestmove` line follows those `info` lines

### Requirement: Go bestmove follows search
`bestmove` MUST match the search best move at the requested depth (or default depth `4` when `depth` is omitted).

#### Scenario: Depth 1 captures the hanging queen
- **GIVEN** position set with `position fen 4k3/8/8/8/7q/8/8/4K2R w - - 0 1`
- **WHEN** the command `go depth 1` is received
- **THEN** the output includes the line `bestmove h1h4`

#### Scenario: Depth 1 mates with the rook
- **GIVEN** position set with `position fen 6k1/4R3/6K1/8/8/8/8/8 w - - 0 1`
- **WHEN** the command `go depth 1` is received
- **THEN** the output includes the line `bestmove e7e8`
- **AND** some preceding `info` line contains `score mate 1`

### Requirement: Infinite search waits for stop
When `go` includes the token `infinite`, the engine MUST search and MAY emit `info` lines, but MUST NOT emit `bestmove` until it later receives `stop`. After `stop`, it MUST emit exactly one `bestmove` line for that search. A `stop` with no pending infinite search MUST produce no output. GUIs such as Nibbler use this handshake when the user moves a piece while analysis is running.

#### Scenario: Go infinite withholds bestmove until stop
- **GIVEN** position set with `position startpos`
- **WHEN** the command `go infinite` is received
- **THEN** the output includes a line starting with `info`
- **AND** the output does not include a `bestmove` line
- **WHEN** the command `stop` is then received
- **THEN** the output includes a line matching `bestmove <uci>` where `<uci>` is one of the legal moves from startpos

#### Scenario: Stop with no search is silent
- **GIVEN** a freshly started engine session
- **WHEN** the command `stop` is received
- **THEN** there is no output

### Requirement: Quit ends the session
When the engine receives `quit`, it MUST stop reading further commands and exit successfully.

#### Scenario: Quit terminates the command loop
- **GIVEN** an engine session after `uci`
- **WHEN** the command `quit` is received
- **THEN** the session ends without requiring further input
