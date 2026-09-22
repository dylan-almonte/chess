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

// BoardMetrics describes the on-screen board geometry used by View and hit-testing.
type BoardMetrics struct {
	OriginX int
	OriginY int
	LabelW  int
	CellW   int
	CellH   int
}

// DefaultBoardMetrics is the compact 2×1 cell layout.
func DefaultBoardMetrics() BoardMetrics {
	return BoardMetrics{
		OriginX: boardContentOriginX,
		OriginY: boardContentOriginY,
		LabelW:  rankLabelWidth,
		CellW:   squareWidth,
		CellH:   1,
	}
}

// SquareAtCell maps a terminal cell using compact 2×1 metrics.
func SquareAtCell(x, y int) (string, bool) {
	return SquareAtCellWith(x, y, DefaultBoardMetrics())
}

// SquareAtCellWith maps a terminal cell using the given metrics.
func SquareAtCellWith(x, y int, m BoardMetrics) (string, bool) {
	if m.CellW <= 0 {
		m.CellW = squareWidth
	}
	if m.CellH <= 0 {
		m.CellH = 1
	}
	if m.LabelW <= 0 {
		m.LabelW = rankLabelWidth
	}
	cx := x - m.OriginX
	cy := y - m.OriginY
	if cy < 0 || cy >= 8*m.CellH {
		return "", false
	}
	if cx < m.LabelW {
		return "", false
	}
	file := (cx - m.LabelW) / m.CellW
	if file < 0 || file > 7 {
		return "", false
	}
	rank := 7 - cy/m.CellH
	if rank < 0 || rank > 7 {
		return "", false
	}
	return board.SquareName(rank*8 + file), true
}

const (
	moveListCols    = 32
	minChromeRows   = 8
	fileLabelRows   = 1
	boardBorderRows = 2
	minCellW        = 3
	minCellH        = 2
	maxCellW        = 7
	maxCellH        = 5
)

// ComputeBoardMetrics sizes squares from leftover viewport. Log/chrome is
// reserved first so a large window can grow to at least 3×2 at 80×28.
func ComputeBoardMetrics(width, height int) BoardMetrics {
	m := DefaultBoardMetrics()
	if width <= 0 || height <= 0 {
		return m
	}
	leftoverW := width - moveListCols - m.LabelW - 4
	if leftoverW < 8*minCellW {
		leftoverW = 8 * minCellW
	}
	cellW := leftoverW / 8
	if cellW < minCellW {
		cellW = minCellW
	}
	if cellW > maxCellW {
		cellW = maxCellW
	}
	leftoverH := height - minChromeRows - fileLabelRows - boardBorderRows
	if leftoverH < 8*minCellH {
		leftoverH = 8 * minCellH
	}
	cellH := leftoverH / 8
	if cellH < minCellH {
		cellH = minCellH
	}
	if cellH > maxCellH {
		cellH = maxCellH
	}
	m.CellW = cellW
	m.CellH = cellH
	return m
}
