## ADDED Requirements

### Requirement: Board squares scale with terminal size
The TUI MUST render each board square larger than the current one-row, two-column glyph cell when the terminal is large enough, and MUST shrink toward that compact cell when the window is small. Rank and file labels MUST remain visible. Piece glyphs MUST stay inside their squares in both Unicode and ASCII modes.

#### Scenario: Large window uses bigger squares
- **GIVEN** a connected session whose window is at least 80 columns by 28 rows
- **WHEN** the board is displayed
- **THEN** each square occupies at least 3 columns and 2 rows
- **AND** square `e2` still shows a white pawn glyph or `P`

#### Scenario: Narrow window keeps a readable 8×8 grid
- **GIVEN** a connected session whose window is 60 columns by 20 rows
- **WHEN** the board is displayed
- **THEN** the board still shows eight ranks and eight files
- **AND** square `e2` still shows a white pawn glyph or `P`

### Requirement: Visual setup can place and remove pieces
The TUI MUST provide a setup mode that starts from the current display position. In setup mode the user MUST be able to choose a piece type, place that piece on a square, and remove a piece from a square. Setup actions MUST NOT send `go` or append to the play move list.

#### Scenario: Place a queen on d5
- **GIVEN** a connected session in setup mode at the standard starting position
- **WHEN** the user selects a white queen and places it on `d5`
- **THEN** square `d5` shows a white queen
- **AND** the move list remains empty
- **AND** no outbound `go` command is sent

#### Scenario: Remove a piece from e2
- **GIVEN** a connected session in setup mode at the standard starting position
- **WHEN** the user removes the piece on `e2`
- **THEN** square `e2` is empty
- **AND** the move list remains empty

### Requirement: Setup starts play from the edited position
The TUI MUST let the user set the side to move in setup mode and then start play. Starting play MUST synchronize the engine with a `position fen` command for the edited placement and MUST refresh legal moves before accepting a human or engine ply. Starting play MUST fail with a visible error and MUST NOT send `position` or `go` when the board does not contain exactly one white king and one black king.

#### Scenario: Play a kings-and-queen setup
- **GIVEN** a setup-mode board whose only pieces are a white king on `e1`, a white queen on `d1`, and a black king on `e8`, with White to move
- **WHEN** the user starts play
- **THEN** the UCI log contains an outbound `position fen 4k3/8/8/8/8/8/8/3QK3 w - - 0 1`
- **AND** the UCI log contains an outbound `legalmoves`
- **AND** the displayed board still shows those three pieces

#### Scenario: Setup without both kings is rejected
- **GIVEN** a setup-mode board that has a white king on `e1` and no black king
- **WHEN** the user starts play
- **THEN** the TUI displays an error that both kings are required
- **AND** no new outbound `position` or `go` command is sent

### Requirement: White and Black player slots are assignable
The TUI MUST assign each of White and Black independently to Human or Engine. The default MUST be White = Human and Black = Engine. Changing a slot MUST NOT require restarting the engine process. When the side to move is Human, the TUI MUST accept the existing typed, click, and cursor play paths and MUST NOT send `go` for that ply. When the side to move is Engine, the TUI MUST send `go` and apply `bestmove` through the existing validation path.

#### Scenario: Human plays Black from startpos
- **GIVEN** a connected session at startpos with White = Engine and Black = Human
- **WHEN** play begins
- **THEN** the UCI log contains an outbound `go` line before any human move
- **AND** after `bestmove e2e4` the board shows a white pawn on `e4`
- **AND** the TUI then waits for a human Black move and does not send another `go` until that move is accepted

#### Scenario: Engine versus engine plays without typed moves
- **GIVEN** a connected session at startpos with White = Engine and Black = Engine
- **WHEN** play begins and the engine returns `bestmove e2e4` then later `bestmove e7e5`
- **THEN** the move list includes `e2e4` and `e7e5`
- **AND** the UCI log contains at least two outbound `go` lines
- **AND** no human move input was required
