## 1. Unicode rendering (red)

- [x] 1.1 Add board render tests `startpos_shows_unicode_white_pawn` and `unicode_board_updates_after_a_human_move`; verify `go test ./internal/board` fails because `Render` still prints ASCII `P`/`p`
- [x] 1.2 Add model tests `toggle_shows_ascii_letters_without_moving_pieces`, `toggle_back_restores_unicode`, and `chosen_mode_survives_a_move` that drive Tab then inspect `View()`; verify they fail because there is no glyph toggle

## 2. Unicode rendering (green)

- [x] 2.1 Change board rendering to a fixed 2-column square layout with a Unicode/ASCII mode flag, mapping `KQRBNP`/`kqrbnp` to `♔♕♖♗♘♙`/`♚♛♜♝♞♟` while keeping stored pieces as ASCII; verify the two board render tests pass
- [x] 2.2 Default the model to Unicode, intercept Tab to toggle without mutating position/move list/engine session, and keep the selected mode after `ApplyUCI`; verify the three toggle tests pass and `go test ./internal/board ./internal/ui` still passes existing play tests

## 3. Click and cursor input (red)

- [x] 3.1 Add a hit-test helper test that maps cell coordinates at the documented board origin to `e2`/`e4` and treats labels, padding, and off-board cells as misses; verify it fails because the helper does not exist
- [x] 3.2 Add fake-engine model tests `click_e2_then_e4_plays_through_the_typed_path`, `clicking_an_illegal_destination_is_rejected`, `clicking_the_selected_square_cancels_the_selection`, and `clicks_are_ignored_while_waiting_for_the_engine`; verify they fail before mouse handling exists
- [x] 3.3 Add fake-engine model test `arrow_selection_plays_e2e4` (cursor to `e2`, Space, cursor to `e4`, Space); verify it fails because arrows still go to the text input

## 4. Shared submit path (green)

- [x] 4.1 Extract `submitMove` from typed Enter and add two-square promotion resolve (`from+to`, else `from+to+"q"`, else submit `from+to`); verify existing typed `e2e4` / illegal-move tests still pass
- [x] 4.2 Implement left-click select/cancel/submit using the hit-test helper, ignoring empty first clicks and clicks while `waiting` or `awaitingLegal`; verify the four click scenario tests pass
- [x] 4.3 Intercept arrow keys for a visible board cursor (default `e2`) and Space to confirm a square through the same select/submit rules; keep Enter as typed UCI submit; verify `arrow_selection_plays_e2e4` passes
- [x] 4.4 Enable `tea.WithMouseCellMotion()` in `main.go` and document Tab / click / arrows+Space in the TUI README or input placeholder; verify `go test ./...` under `tui/` passes

## 5. Gate

- [x] 5.1 Ensure every new delta-spec scenario has an automated Go test and run `gofmt` plus `go test ./...` under `tui/`
- [x] 5.2 Run `openspec validate tui-polish --strict` and fix any issues
