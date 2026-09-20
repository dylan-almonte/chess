package ui

import "github.com/dylanca/chess-tui/tui/internal/board"

// Playable view starts at (0,0). The board box is the left pane: lipgloss
// normal border (1) plus Padding(0, 1), so board content origin is (2, 1).
const (
	boardContentOriginX = 2
	boardContentOriginY = 1
	rankLabelWidth      = 2
	squareWidth         = 2
)

// SquareAtCell maps a terminal cell to a board square name. Clicks on the
// border, rank/file labels, padding, or off-board cells miss.
func SquareAtCell(x, y int) (string, bool) {
	cx := x - boardContentOriginX
	cy := y - boardContentOriginY
	if cy < 0 || cy > 7 {
		return "", false
	}
	if cx < rankLabelWidth {
		return "", false
	}
	file := (cx - rankLabelWidth) / squareWidth
	if file < 0 || file > 7 {
		return "", false
	}
	rank := 7 - cy
	return board.SquareName(rank*8 + file), true
}
