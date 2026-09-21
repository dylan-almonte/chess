# Chess TUI

Bubble Tea terminal client for the Rust UCI engine in this monorepo.

## Prerequisites

1. Build the engine from the repo root:

```bash
cargo build --release
```

This produces `target/release/chess`.

2. Go 1.22+ (module uses the toolchain available on the machine).

## Run

From `tui/`:

```bash
go run .
```

By default the TUI looks for the engine at `../target/release/chess` (relative to the `tui/` module). Override with:

```bash
export CHESS_ENGINE=/absolute/or/relative/path/to/chess
go run .
```

## Nibbler / other UCI GUIs

The same `target/release/chess` binary speaks standard UCI:

- Analysis (`go infinite` … `stop`) withholds `bestmove` until `stop` (Nibbler’s analysis loop)
- Play / finite searches (`go`, `go depth N`, `go nodes N`) return `bestmove` immediately (this TUI uses `go depth 4`)

Point Nibbler at the release binary. For the engine to move on the board, use Play White / Play Black / Self-play — drag-to-move in analysis only updates the position.

## Controls

- Type a UCI long-algebraic move (e.g. `e2e4`) and press Enter
- Tab toggles Unicode piece glyphs and ASCII letters
- Click a piece, then click a destination, to play through the same path as typed UCI
- Arrow keys move a board cursor; Space selects / confirms a square the same way a click does
- `setup` enters visual edit mode; `Q`/`q`/`K`/`k`/… then Enter chooses a piece, `x` erases
- In setup, click or Space places the brush; `side` toggles who moves; `clear` empties; `startpos` resets; `play` starts from the board
- `white human|engine` and `black human|engine` assign each side (default White human, Black engine)
- `quit` or `ctrl+c` / `q` exits; the TUI sends UCI `quit` and waits for the engine to exit

The board grows with the terminal: larger windows get bigger squares, smaller ones stay a compact 8×8.

## Layout

- Scaled board + move list
- Scrolling UCI transcript (outbound `>` / inbound `<`)
- Telemetry panel for engine `info` lines
- Status line for player slots and setup, then the input line

## Tests

```bash
go test ./...
```

Tests use an in-process fake engine; a built Rust binary is not required.
