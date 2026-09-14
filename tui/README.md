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

## Controls

- Type a UCI long-algebraic move (e.g. `e2e4`) and press Enter
- `quit` or `ctrl+c` / `q` exits; the TUI sends UCI `quit` and waits for the engine to exit

## Layout

- Board + move list
- Scrolling UCI transcript (outbound `>` / inbound `<`)
- Telemetry panel for engine `info` lines
- Input line for moves

## Tests

```bash
go test ./...
```

Tests use an in-process fake engine; a built Rust binary is not required.
