## Purpose

Defines a Bubble Tea terminal client that shows the board and moves, logs UCI traffic, and surfaces engine telemetry while driving the Rust UCI engine as a subprocess.

## ADDED Requirements

### Requirement: Launch and UCI handshake
The TUI MUST start the configured engine binary, send `uci` and `isready`, and only enter the playable UI after receiving `uciok` and `readyok`. Handshake lines MUST appear in the UCI log.

#### Scenario: Successful engine connect
- **GIVEN** a valid path to the Rust engine binary
- **WHEN** the TUI starts
- **THEN** the UCI log contains an outbound `uci` line and an inbound `uciok` line
- **AND** the UCI log contains an outbound `isready` line and an inbound `readyok` line
- **AND** the board shows the standard starting position

#### Scenario: Missing engine binary fails clearly
- **GIVEN** an engine path that does not exist
- **WHEN** the TUI starts
- **THEN** the TUI shows an error state naming the missing binary
- **AND** it does not present a playable board session

### Requirement: Board and move list
The TUI MUST display the current position as an 8×8 board and a chronological move list in UCI long-algebraic form. After each completed human or engine move, both board and move list MUST update.

#### Scenario: Human move updates board and list
- **GIVEN** a connected session at startpos with the human to move
- **WHEN** the human submits the legal move `e2e4`
- **THEN** the board shows a white pawn on `e4` and an empty `e2`
- **AND** the move list includes `e2e4`

#### Scenario: Engine reply updates board and list
- **GIVEN** a connected session after human move `e2e4`
- **WHEN** the engine returns `bestmove e7e5`
- **THEN** the board reflects black’s pawn on `e5`
- **AND** the move list includes `e7e5` after `e2e4`

### Requirement: UCI transcript log
The TUI MUST show a scrolling log of every UCI line sent to and received from the engine, clearly distinguishing direction (outbound vs inbound).

#### Scenario: Position and go appear in the log
- **GIVEN** a connected session
- **WHEN** the human plays `e2e4` and the TUI requests an engine move
- **THEN** the log contains an outbound line starting with `position`
- **AND** the log contains an outbound `go` line
- **AND** the log contains an inbound line starting with `bestmove`

### Requirement: Telemetry panel
The TUI MUST provide a telemetry panel that displays the latest engine `info` lines (depth, score, nodes, nps, pv when present). When the engine emits no `info` (stub `go`), the panel MUST show an empty/idle state rather than failing.

#### Scenario: Info lines populate telemetry
- **GIVEN** a connected session whose engine emits `info depth 1 score cp 12 pv e2e4`
- **WHEN** that line is received
- **THEN** the telemetry panel shows depth `1` and score `cp 12` (or equivalent clear rendering of those fields)

#### Scenario: Stub engine leaves telemetry idle
- **GIVEN** a connected session using the current stub engine that replies only `bestmove` with no `info`
- **WHEN** `go` completes
- **THEN** the telemetry panel remains in an idle/empty state
- **AND** the session continues normally

### Requirement: Quit cleans up the engine
When the user quits the TUI, it MUST send `quit` to the engine (when connected) and terminate the subprocess.

#### Scenario: Quit sends UCI quit
- **GIVEN** a connected session
- **WHEN** the user quits the TUI
- **THEN** the UCI log includes an outbound `quit` line (or equivalent shutdown that ends the engine process)
- **AND** the TUI process exits successfully
