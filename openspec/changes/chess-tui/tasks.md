## 1. Scaffold

- [ ] 1.1 Create `tui/` Go module with Bubble Tea / Bubbles / Lip Gloss dependencies — verify `go test ./...` succeeds in `tui/`
- [ ] 1.2 Add engine path resolution (`CHESS_ENGINE` or default `target/release/chess`) and document run steps in `tui/README.md` — verify README names the env var and build command

## 2. UCI subprocess

- [ ] 2.1 Implement engine process spawn + line-oriented read/write with outbound/inbound logging hooks — verify unit tests with a fake engine record `uci`/`uciok` and `isready`/`readyok`
- [ ] 2.2 Wire startup handshake and missing-binary error state into the Bubble Tea model — verify scenarios `successful_engine_connect` and `missing_engine_binary_fails_clearly`

## 3. Play UI

- [ ] 3.1 Render board + move list from local game state; accept UCI move input — verify scenario `human_move_updates_board_and_list`
- [ ] 3.2 On human move, send `position` + `go`, apply `bestmove`, update board/list — verify scenario `engine_reply_updates_board_and_list`
- [ ] 3.3 Show scrolling UCI transcript with direction — verify scenario `position_and_go_appear_in_the_log`

## 4. Telemetry and shutdown

- [ ] 4.1 Parse `info` lines into the telemetry panel; idle when absent — verify scenarios `info_lines_populate_telemetry` and `stub_engine_leaves_telemetry_idle`
- [ ] 4.2 On quit, send `quit` and wait for process exit — verify scenario `quit_sends_uci_quit`

## 5. Gate

- [ ] 5.1 Ensure every delta-spec scenario has an automated test and `go test ./...` passes under `tui/`
- [ ] 5.2 Run `openspec validate chess-tui --strict` and fix any issues
