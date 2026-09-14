## Why

The Rust engine speaks UCI but is hard to develop against without a visible board, move list, and protocol log. A Bubble Tea TUI in-repo gives an immediate feedback loop while keeping the engine a pure UCI process.

## What Changes

- Add a Go module under `tui/` using Bubble Tea
- Spawn/configure the Rust engine binary and speak UCI over stdin/stdout
- Show a live chess board, move list, scrolling UCI log, and a telemetry panel for `info` lines
- Allow the human to enter moves (UCI long-algebraic) and request an engine reply via `go`

## Capabilities

### New Capabilities

- `chess-tui`: Terminal UI client for playing against and inspecting the UCI engine

### Modified Capabilities

- (none)

## Non-goals

- Replacing or rewriting the Rust engine
- Graphical GUI (egui/web)
- Full analysis workstation (multi-PV charts, databases, opening books)
- Implementing evaluation/search (engine still stubs `bestmove` until those layers exist)
- Cross-compiling or packaging installers

## Impact

- New `tui/` Go module and README run instructions
- Depends on a built Rust binary path (e.g. `target/release/chess` or env override)
- OpenSpec `project.md` / config already describe the monorepo; no change to engine behavioral specs
- Engine roadmap (evaluation, search) remains separate OpenSpec changes
