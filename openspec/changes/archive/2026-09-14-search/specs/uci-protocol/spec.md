## MODIFIED Requirements

### Requirement: Go returns a legal bestmove
On `go`, the engine MUST output exactly one `bestmove <move>` line where `<move>` is a legal UCI move chosen by search for the current position. If `depth N` is present (`N` a positive integer), search MUST use maximum depth `N`. If `depth` is omitted, search MUST use maximum depth `4`. Time-control tokens (`wtime`, `btime`, `movetime`, `infinite`) MUST be ignored. If the side to move has no legal moves, the engine MUST output `bestmove 0000`.

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

## ADDED Requirements

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
