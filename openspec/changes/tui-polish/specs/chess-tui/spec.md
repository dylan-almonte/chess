## ADDED Requirements

### Requirement: Board renders Unicode pieces
The TUI MUST render occupied squares with Unicode chess glyphs by default. White `KQRBNP` MUST map to `♔♕♖♗♘♙`. Black `kqrbnp` MUST map to `♚♛♜♝♞♟`. Empty squares MUST remain non-letter placeholders so the 8×8 grid stays readable. After each accepted human or engine move, the board MUST keep showing Unicode glyphs for the updated position until the user selects ASCII rendering.

#### Scenario: Startpos shows Unicode white pawn
- **GIVEN** a connected session at the standard starting position
- **WHEN** the board is displayed
- **THEN** square `e2` shows `♙` rather than `P`
- **AND** square `e7` shows `♟` rather than `p`
- **AND** square `e4` remains an empty placeholder

#### Scenario: Unicode board updates after a human move
- **GIVEN** a connected session at startpos with Unicode rendering
- **WHEN** the human submits the legal move `e2e4`
- **THEN** square `e4` shows `♙`
- **AND** square `e2` is empty

### Requirement: ASCII piece toggle
The TUI MUST provide a user action that switches board rendering between Unicode glyphs and ASCII letters (`KQRBNP` / `kqrbnp`). Toggling MUST NOT change the position, move list, UCI log contents, or engine session. After a later accepted move, the board MUST keep using the currently selected rendering mode.

#### Scenario: Toggle shows ASCII letters without moving pieces
- **GIVEN** a connected session at startpos showing Unicode pieces
- **WHEN** the user toggles piece rendering
- **THEN** square `e2` shows `P`
- **AND** square `e7` shows `p`
- **AND** the move list is unchanged

#### Scenario: Toggle back restores Unicode
- **GIVEN** a connected session at startpos in ASCII rendering
- **WHEN** the user toggles piece rendering
- **THEN** square `e2` shows `♙`
- **AND** square `e7` shows `♟`

#### Scenario: Chosen mode survives a move
- **GIVEN** a connected session at startpos in ASCII rendering
- **WHEN** the human submits `e2e4`
- **THEN** square `e4` shows `P`
- **AND** square `e2` is empty

### Requirement: Mouse click-to-move uses the typed-input play path
While the session is playable and accepting human input, a mouse click on an occupied board square MUST select that square as the origin. A second click on a different square MUST build a UCI long-algebraic token from origin to destination and submit it through the same acceptance and rejection path as typing that token. Typed UCI entry MUST remain available. A click on an empty square MUST NOT start a selection. A second click on the selected origin MUST cancel the selection without submitting a move. Clicks MUST NOT submit a move while the TUI is waiting for legal moves or an engine reply.

When the two-square token is not legal but the same origin and destination with queen promotion is legal, the TUI MUST submit the queen-promotion token (for example `e7e8q`). Other promotion pieces remain available only via typed UCI.

#### Scenario: Click e2 then e4 plays through the typed path
- **GIVEN** a connected session at startpos that is accepting human input
- **WHEN** the user clicks square `e2` and then square `e4`
- **THEN** the board shows a white pawn on `e4` and an empty `e2`
- **AND** the move list includes `e2e4`
- **AND** the UCI log contains an outbound `position` line that includes `e2e4`
- **AND** the UCI log contains an outbound `go` line

#### Scenario: Clicking an illegal destination is rejected
- **GIVEN** a connected session at startpos that is accepting human input
- **WHEN** the user clicks square `e2` and then square `e5`
- **THEN** the TUI displays an error identifying `e2e5` as illegal
- **AND** a white pawn remains on `e2` and `e5` remains empty
- **AND** the move list remains empty
- **AND** no new outbound `position` or `go` command is sent

#### Scenario: Clicking the selected square cancels the selection
- **GIVEN** a connected session at startpos with square `e2` selected
- **WHEN** the user clicks `e2` again
- **THEN** no move is submitted
- **AND** the board and move list remain unchanged

#### Scenario: Clicks are ignored while waiting for the engine
- **GIVEN** a connected session that has accepted `e2e4` and is waiting for `bestmove`
- **WHEN** the user clicks square `e7` and then square `e5`
- **THEN** the board and move list still contain only `e2e4`
- **AND** no additional outbound `position` or `go` command is sent

### Requirement: Keyboard cursor can select origin and destination
The TUI MUST show a visible board cursor that the user can move with the arrow keys without submitting a move. Confirming an occupied origin square and then confirming a different destination square MUST build a UCI token and submit it through the same acceptance and rejection path as typing that token, including the queen-promotion rule used for click-to-move. Typed UCI entry MUST remain available.

#### Scenario: Arrow selection plays e2e4
- **GIVEN** a connected session at startpos that is accepting human input
- **WHEN** the user moves the board cursor to `e2`, confirms that square, moves the cursor to `e4`, and confirms that square
- **THEN** the board shows a white pawn on `e4` and an empty `e2`
- **AND** the move list includes `e2e4`
- **AND** the UCI log contains an outbound `position` line that includes `e2e4`
- **AND** the UCI log contains an outbound `go` line
