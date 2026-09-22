package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dylanca/chess-tui/tui/internal/board"
	"github.com/dylanca/chess-tui/tui/internal/engine"
	"github.com/dylanca/chess-tui/tui/internal/info"
	"github.com/dylanca/chess-tui/tui/internal/path"
)

type phase int

const (
	phaseBoot phase = iota
	phaseHandshakeUCI
	phaseHandshakeReady
	phasePlay
	phaseError
	phaseDone
)

// Config for constructing the model.
type Config struct {
	EnginePath  string
	FakeHandler func(line string) []string
}

// Model is the Bubble Tea app state.
type Model struct {
	cfg            Config
	phase          phase
	errMsg         string
	client         *engine.Client
	board          board.Board
	moves          []string
	logLines       []string
	tele           info.Telemetry
	input          textinput.Model
	logView        viewport.Model
	waiting        bool
	width          int
	height         int
	ready          bool
	inbox          chan tea.Msg
	legalMoves     map[string]bool // current legal move set from engine
	engineLegal    map[string]bool // legal set for the engine's side (fetched before go)
	awaitingLegal  bool            // waiting for legalmoves response
	pendingGo      bool            // need to send go after receiving engine-side legalmoves
	unicodePieces  bool
	selectedSquare string
	cursorSquare   string
	metrics        BoardMetrics
	setupMode      bool
	brush          byte // 0 empty-place ignored; 'x' erase; otherwise piece letter
	whiteToMove    bool
	baseFEN        string
	whiteSlot      playerSlot
	blackSlot      playerSlot
}

type playerSlot int

const (
	slotHuman playerSlot = iota
	slotEngine
)

func New(cfg Config) Model {
	ti := textinput.New()
	ti.Placeholder = "e2e4 / setup / white engine / quit"
	ti.CharLimit = 24
	ti.Width = 24
	vp := viewport.New(60, 8)
	vp.SetContent("")
	return Model{
		cfg:           cfg,
		phase:         phaseBoot,
		board:         board.StartPos(),
		tele:          info.Idle(),
		input:         ti,
		logView:       vp,
		unicodePieces: true,
		cursorSquare:  "",
		metrics:       DefaultBoardMetrics(),
		whiteToMove:   true,
		baseFEN:       board.StartFEN,
		whiteSlot:     slotHuman,
		blackSlot:     slotEngine,
	}
}

type startResultMsg struct {
	client *engine.Client
	err    error
}

type quitDoneMsg struct{}

func (m Model) Init() tea.Cmd { return m.startEngine }

func (m Model) startEngine() tea.Msg {
	if m.cfg.FakeHandler != nil {
		c, err := engine.StartFake(m.cfg.FakeHandler)
		return startResultMsg{client: c, err: err}
	}
	p := m.cfg.EnginePath
	if p == "" {
		p = path.Resolve(".")
	}
	if err := path.Validate(p); err != nil {
		return startResultMsg{err: err}
	}
	c, err := engine.Start(p)
	return startResultMsg{client: c, err: err}
}

func waitInbox(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return engine.ExitMsg{}
		}
		return msg
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.metrics = ComputeBoardMetrics(msg.Width, msg.Height)
		m.logView.Width = max(20, msg.Width-4)
		boardRows := 8*m.metrics.CellH + fileLabelRows + boardBorderRows
		remain := msg.Height - boardRows - 9
		if remain < 2 {
			remain = 2
		}
		m.logView.Height = remain
		return m, nil

	case startResultMsg:
		if msg.err != nil {
			m.phase = phaseError
			m.errMsg = msg.err.Error()
			return m, nil
		}
		m.client = msg.client
		m.phase = phaseHandshakeUCI
		inbox := make(chan tea.Msg, 64)
		m.inbox = inbox
		emit := func(v any) { inbox <- v.(tea.Msg) }
		if m.client.Scripted() {
			m.client.ReadLoop(emit)
		} else {
			go m.client.ReadLoop(emit)
		}
		if err := m.client.Send("uci"); err != nil {
			m.phase = phaseError
			m.errMsg = err.Error()
			return m, nil
		}
		m.syncLogFromClient()
		return m, waitInbox(inbox)

	case engine.LineMsg:
		nm, cmd := m.handleLine(msg.Text)
		if m.inbox == nil {
			return nm, cmd
		}
		return nm, tea.Batch(cmd, waitInbox(m.inbox))

	case engine.ErrMsg:
		m.phase = phaseError
		m.errMsg = msg.Err.Error()
		return m, nil

	case engine.ExitMsg:
		return m, nil

	case tea.KeyMsg:
		if m.phase == phaseError || m.phase == phaseDone {
			if msg.Type == tea.KeyCtrlC || msg.String() == "q" || msg.Type == tea.KeyEnter {
				return m, tea.Quit
			}
			return m, nil
		}
		if msg.Type == tea.KeyCtrlC {
			return m.beginQuit()
		}
		if m.phase != phasePlay {
			return m, nil
		}
		if msg.Type == tea.KeyEnter {
			return m.submitInput()
		}
		if msg.Type == tea.KeyTab {
			m.unicodePieces = !m.unicodePieces
			return m, nil
		}
		if msg.Type == tea.KeyUp || msg.Type == tea.KeyDown || msg.Type == tea.KeyLeft || msg.Type == tea.KeyRight {
			m.moveCursor(msg.Type)
			return m, nil
		}
		if msg.Type == tea.KeySpace || msg.String() == " " {
			if m.cursorSquare != "" {
				return m.handleSquare(m.cursorSquare)
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd

	case tea.MouseMsg:
		if m.phase != phasePlay {
			return m, nil
		}
		return m.handleMouse(msg)

	case quitDoneMsg:
		m.phase = phaseDone
		return m, tea.Quit
	}

	if m.phase == phasePlay {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) handleLine(text string) (Model, tea.Cmd) {
	m.appendLogLine(engine.Inbound, text)

	switch m.phase {
	case phaseHandshakeUCI:
		if text == "uciok" {
			m.phase = phaseHandshakeReady
			_ = m.send("isready")
		}
	case phaseHandshakeReady:
		if text == "readyok" {
			m.phase = phasePlay
			m.ready = true
			m.input.Focus()
			_ = m.send("ucinewgame")
			_ = m.send("position startpos")
			_ = m.send("legalmoves")
			m.awaitingLegal = true
		}
	case phasePlay:
		if t, ok := info.ParseInfo(text); ok {
			m.tele = t
		}
		if strings.HasPrefix(text, "legalmoves") {
			set := make(map[string]bool)
			parts := strings.Fields(text)
			for _, mv := range parts[1:] {
				set[mv] = true
			}
			if m.pendingGo {
				// This is the engine-side legal set before go.
				// Use an explicit depth so we never rely on `go infinite` (Nibbler analysis).
				m.engineLegal = set
				m.pendingGo = false
				_ = m.send("go depth 4")
			} else {
				// This is the human-side legal set
				m.legalMoves = set
				m.awaitingLegal = false
				nm, cmd := m.requestEngineIfNeeded()
				return nm.(Model), cmd
			}
		}
		if strings.HasPrefix(text, "bestmove") {
			m.waiting = false
			parts := strings.Fields(text)
			if len(parts) >= 2 && parts[1] != "(none)" {
				mv := parts[1]
				// Validate bestmove against the engine-side legal set
				if mv == "0000" {
					if len(m.engineLegal) != 0 {
						m.logLines = append(m.logLines, "! illegal engine move: 0000 (legal moves exist)")
						m.refreshLogView()
					}
					// null move in checkmate/stalemate — no board change
				} else if m.engineLegal != nil && !m.engineLegal[mv] {
					m.logLines = append(m.logLines, fmt.Sprintf("! illegal engine move: %s", mv))
					m.refreshLogView()
				} else {
					candidate := m.board
					if err := candidate.ApplyUCI(mv); err != nil {
						m.logLines = append(m.logLines,
							fmt.Sprintf("! cannot display advertised engine move %s: %v", mv, err))
						m.refreshLogView()
					} else {
						m.board = candidate
						m.moves = append(m.moves, mv)
						m.whiteToMove = !m.whiteToMove
						_ = m.send(m.positionCommand())
						_ = m.send("legalmoves")
						m.legalMoves = nil
						if m.mover() == slotEngine {
							m.pendingGo = true
							m.waiting = true
						} else {
							m.awaitingLegal = true
						}
					}
				}
			}
			m.engineLegal = nil
		}
	}
	return m, nil
}

func (m *Model) send(line string) error {
	if m.client == nil {
		m.appendLogLine(engine.Outbound, line)
		return nil
	}
	err := m.client.Send(line)
	m.syncLogFromClient()
	return err
}

func (m *Model) appendLogLine(dir engine.Direction, text string) {
	m.logLines = append(m.logLines, dir.String()+" "+text)
	m.refreshLogView()
}

func (m *Model) syncLogFromClient() {
	if m.client == nil {
		return
	}
	m.logLines = nil
	for _, e := range m.client.Log() {
		m.logLines = append(m.logLines, e.Dir.String()+" "+e.Text)
	}
	m.refreshLogView()
}

func (m *Model) refreshLogView() {
	m.logView.SetContent(strings.Join(m.logLines, "\n"))
	m.logView.GotoBottom()
}

func (m Model) submitInput() (tea.Model, tea.Cmd) {
	raw := strings.TrimSpace(m.input.Value())
	m.input.SetValue("")
	if raw == "" {
		return m, nil
	}
	if raw == "quit" || raw == "q" {
		return m.beginQuit()
	}
	if cmd, ok := m.handleCommand(raw); ok {
		return cmd()
	}
	if m.setupMode {
		return m, nil
	}
	return m.submitMove(raw)
}

func (m Model) handleCommand(raw string) (func() (tea.Model, tea.Cmd), bool) {
	switch raw {
	case "setup":
		return func() (tea.Model, tea.Cmd) {
			m.setupMode = true
			m.selectedSquare = ""
			m.brush = 0
			return m, nil
		}, true
	case "play":
		return func() (tea.Model, tea.Cmd) { return m.startPlayFromSetup() }, true
	case "side":
		return func() (tea.Model, tea.Cmd) {
			if m.setupMode {
				m.whiteToMove = !m.whiteToMove
			}
			return m, nil
		}, true
	case "clear":
		return func() (tea.Model, tea.Cmd) {
			if m.setupMode {
				m.board.Clear()
			}
			return m, nil
		}, true
	case "startpos":
		return func() (tea.Model, tea.Cmd) {
			if m.setupMode {
				m.board = board.StartPos()
				m.whiteToMove = true
			}
			return m, nil
		}, true
	}
	if slot, kind, ok := parseSlotCommand(raw); ok {
		return func() (tea.Model, tea.Cmd) {
			if slot == "white" {
				m.whiteSlot = kind
			} else {
				m.blackSlot = kind
			}
			return m.requestEngineIfNeeded()
		}, true
	}
	if m.setupMode && isBrushToken(raw) {
		return func() (tea.Model, tea.Cmd) {
			if raw == "x" {
				m.brush = 'x'
			} else {
				m.brush = raw[0]
			}
			return m, nil
		}, true
	}
	return nil, false
}

func parseSlotCommand(raw string) (string, playerSlot, bool) {
	fields := strings.Fields(raw)
	if len(fields) != 2 {
		return "", 0, false
	}
	if fields[0] != "white" && fields[0] != "black" {
		return "", 0, false
	}
	switch fields[1] {
	case "human":
		return fields[0], slotHuman, true
	case "engine":
		return fields[0], slotEngine, true
	}
	return "", 0, false
}

func (m Model) mover() playerSlot {
	if m.whiteToMove {
		return m.whiteSlot
	}
	return m.blackSlot
}

func (m Model) requestEngineIfNeeded() (tea.Model, tea.Cmd) {
	if m.setupMode || m.waiting || m.awaitingLegal || m.legalMoves == nil {
		return m, nil
	}
	if m.mover() != slotEngine {
		return m, nil
	}
	_ = m.send("legalmoves")
	m.pendingGo = true
	m.waiting = true
	if m.inbox != nil {
		return m, waitInbox(m.inbox)
	}
	return m, nil
}

func isBrushToken(raw string) bool {
	if raw == "x" {
		return true
	}
	if len(raw) != 1 {
		return false
	}
	switch raw[0] {
	case 'K', 'Q', 'R', 'B', 'N', 'P', 'k', 'q', 'r', 'b', 'n', 'p':
		return true
	}
	return false
}

func (m Model) startPlayFromSetup() (tea.Model, tea.Cmd) {
	if !m.setupMode {
		return m, nil
	}
	w, b := m.board.CountKings()
	if w != 1 || b != 1 {
		m.logLines = append(m.logLines, "! both kings are required")
		m.refreshLogView()
		return m, nil
	}
	fen := m.board.FEN(m.whiteToMove)
	m.baseFEN = fen
	m.moves = nil
	m.legalMoves = nil
	m.selectedSquare = ""
	m.setupMode = false
	m.brush = 0
	_ = m.send("ucinewgame")
	_ = m.send("position fen " + fen)
	_ = m.send("legalmoves")
	m.awaitingLegal = true
	if m.inbox != nil {
		return m, waitInbox(m.inbox)
	}
	return m, nil
}

func (m Model) resolveSquareMove(from, to string) string {
	token := from + to
	if m.legalMoves != nil && m.legalMoves[token] {
		return token
	}
	if m.legalMoves != nil && m.legalMoves[token+"q"] {
		return token + "q"
	}
	return token
}

func (m Model) submitMove(raw string) (tea.Model, tea.Cmd) {
	if m.waiting || m.awaitingLegal || m.legalMoves == nil {
		return m, nil
	}
	if m.legalMoves != nil && !m.legalMoves[raw] {
		m.logLines = append(m.logLines, fmt.Sprintf("! illegal move: %s", raw))
		m.refreshLogView()
		return m, nil
	}
	candidate := m.board
	if err := candidate.ApplyUCI(raw); err != nil {
		m.logLines = append(m.logLines, "! "+err.Error())
		m.refreshLogView()
		return m, nil
	}
	m.board = candidate
	m.moves = append(m.moves, raw)
	m.legalMoves = nil
	m.selectedSquare = ""
	m.whiteToMove = !m.whiteToMove
	_ = m.send(m.positionCommand())
	_ = m.send("legalmoves")
	if m.mover() == slotEngine {
		m.pendingGo = true
		m.waiting = true
	} else {
		m.awaitingLegal = true
	}
	if m.inbox != nil {
		return m, waitInbox(m.inbox)
	}
	return m, nil
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}
	name, ok := SquareAtCellWith(msg.X, msg.Y, m.metrics)
	if !ok {
		return m, nil
	}
	return m.handleSquare(name)
}

func (m Model) positionCommand() string {
	if len(m.moves) == 0 {
		if m.baseFEN == "" || m.baseFEN == board.StartFEN {
			return "position startpos"
		}
		return "position fen " + m.baseFEN
	}
	joined := strings.Join(m.moves, " ")
	if m.baseFEN == "" || m.baseFEN == board.StartFEN {
		return "position startpos moves " + joined
	}
	return "position fen " + m.baseFEN + " moves " + joined
}

func (m Model) handleSquare(name string) (tea.Model, tea.Cmd) {
	if m.setupMode {
		if m.brush == 0 {
			return m, nil
		}
		if m.brush == 'x' {
			m.board.ClearSquare(name)
		} else {
			m.board.SetSquare(name, m.brush)
		}
		return m, nil
	}
	if m.waiting || m.awaitingLegal || m.legalMoves == nil {
		return m, nil
	}
	if m.selectedSquare == "" {
		if m.board.PieceAtName(name) == board.Empty {
			return m, nil
		}
		m.selectedSquare = name
		return m, nil
	}
	if name == m.selectedSquare {
		m.selectedSquare = ""
		return m, nil
	}
	uci := m.resolveSquareMove(m.selectedSquare, name)
	m.selectedSquare = ""
	return m.submitMove(uci)
}

func (m *Model) moveCursor(key tea.KeyType) {
	sq, err := board.ParseSquare(m.cursorSquare)
	if err != nil {
		m.cursorSquare = "e2"
		return
	}
	file, rank := sq%8, sq/8
	switch key {
	case tea.KeyLeft:
		file--
	case tea.KeyRight:
		file++
	case tea.KeyUp:
		rank++
	case tea.KeyDown:
		rank--
	}
	if file < 0 {
		file = 0
	}
	if file > 7 {
		file = 7
	}
	if rank < 0 {
		rank = 0
	}
	if rank > 7 {
		rank = 7
	}
	m.cursorSquare = board.SquareName(rank*8 + file)
}

func (m Model) beginQuit() (tea.Model, tea.Cmd) {
	client := m.client
	m.phase = phaseDone
	return m, func() tea.Msg {
		if client != nil {
			_ = client.Quit()
		} else {
			// still record quit for log-only test clients
		}
		return quitDoneMsg{}
	}
}

func (m Model) renderBoard() string {
	met := m.metrics
	if met.CellW == 0 {
		met = DefaultBoardMetrics()
	}
	lightSquare := lipgloss.NewStyle().Background(lipgloss.Color("#b58863")).Foreground(lipgloss.Color("#000000"))
	darkSquare := lipgloss.NewStyle().Background(lipgloss.Color("#f0d9b5")).Foreground(lipgloss.Color("#000000"))
	cursorStyle := lipgloss.NewStyle().Background(lipgloss.Color("#7fc97f")).Foreground(lipgloss.Color("#000000"))
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("#ffff66")).Foreground(lipgloss.Color("#000000"))
	var lines []string
	for rank := 7; rank >= 0; rank-- {
		rowCells := make([][]string, met.CellH)
		for i := range rowCells {
			rowCells[i] = make([]string, 0, 9)
		}
		label := string(byte('1' + rank))
		for r := 0; r < met.CellH; r++ {
			if r == met.CellH/2 {
				rowCells[r] = append(rowCells[r], padCenter(label, met.LabelW))
			} else {
				rowCells[r] = append(rowCells[r], strings.Repeat(" ", met.LabelW))
			}
		}
		for file := 0; file < 8; file++ {
			name := board.SquareName(rank*8 + file)
			glyph := m.board.GlyphAt(name, m.unicodePieces)
			cellLines := padCell(glyph, met.CellW, met.CellH)
			isLight := (rank+file)%2 == 1
			style := darkSquare
			if isLight {
				style = lightSquare
			}
			switch name {
			case m.selectedSquare:
				style = selectedStyle
			case m.cursorSquare:
				style = cursorStyle
			}
			for r := 0; r < met.CellH; r++ {
				rowCells[r] = append(rowCells[r], style.Render(cellLines[r]))
			}
		}
		for r := 0; r < met.CellH; r++ {
			lines = append(lines, strings.Join(rowCells[r], ""))
		}
	}
	fileRow := strings.Repeat(" ", met.LabelW)
	for file := 0; file < 8; file++ {
		fileRow += padCenter(string(byte('a'+file)), met.CellW)
	}
	lines = append(lines, fileRow)
	return strings.Join(lines, "\n")
}

func padCell(glyph string, w, h int) []string {
	out := make([]string, h)
	for r := 0; r < h; r++ {
		if r == h/2 {
			out[r] = padCenter(glyph, w)
		} else {
			out[r] = strings.Repeat(" ", w)
		}
	}
	return out
}

func padCenter(s string, w int) string {
	n := len([]rune(s))
	if n >= w {
		return s
	}
	left := (w - n) / 2
	right := w - n - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func (m Model) View() string {
	if m.phase == phaseError {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(
			fmt.Sprintf("Engine error: %s\n\nPress q or Enter to exit.", m.errMsg),
		)
	}
	if m.phase == phaseBoot || m.phase == phaseHandshakeUCI || m.phase == phaseHandshakeReady {
		return "Connecting to engine…"
	}

	boardStr := m.renderBoard()
	movesStr := "Moves:\n"
	if len(m.moves) == 0 {
		movesStr += "(none)"
	} else {
		movesStr += strings.Join(m.moves, " ")
	}
	top := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1).Render(boardStr),
		lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1).Width(28).Render(movesStr),
	)
	logBlock := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1).Render(
		"UCI log\n" + m.logView.View(),
	)
	teleBlock := lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1).Render(
		"Telemetry\n" + m.tele.Render(),
	)
	status := fmt.Sprintf("White: %s  Black: %s", slotName(m.whiteSlot), slotName(m.blackSlot))
	if m.setupMode {
		brush := "none"
		if m.brush == 'x' {
			brush = "erase"
		} else if m.brush != 0 {
			brush = string(m.brush)
		}
		side := "b"
		if m.whiteToMove {
			side = "w"
		}
		status += fmt.Sprintf("  setup %s  side: %s", brush, side)
	}
	inputLine := status + "\ninput: " + m.input.View()
	return lipgloss.JoinVertical(lipgloss.Left, top, logBlock, teleBlock, inputLine)
}

func slotName(s playerSlot) string {
	if s == slotEngine {
		return "Engine"
	}
	return "Human"
}

// Test inspectors / helpers

func (m *Model) SetInputValue(v string) { m.input.SetValue(v) }

func (m Model) BeginQuit() (tea.Model, tea.Cmd) { return m.beginQuit() }

func (m Model) ClientLog() []engine.LogEntry {
	if m.client == nil {
		return nil
	}
	return m.client.Log()
}

func (m Model) PhaseReady() bool          { return m.ready && m.phase == phasePlay }
func (m Model) LegalReady() bool          { return m.PhaseReady() && !m.awaitingLegal }
func (m Model) PhaseError() bool          { return m.phase == phaseError }
func (m Model) ErrorText() string         { return m.errMsg }
func (m Model) LogText() string           { return strings.Join(m.logLines, "\n") }
func (m Model) Moves() []string           { return append([]string(nil), m.moves...) }
func (m Model) Board() board.Board        { return m.board }
func (m Model) Telemetry() info.Telemetry { return m.tele }
func (m Model) IsPlayable() bool          { return m.phase == phasePlay }
func (m Model) SelectedSquare() string    { return m.selectedSquare }
func (m Model) CursorSquare() string      { return m.cursorSquare }
func (m Model) CellWidth() int {
	if m.metrics.CellW == 0 {
		return 2
	}
	return m.metrics.CellW
}
func (m Model) CellHeight() int {
	if m.metrics.CellH == 0 {
		return 1
	}
	return m.metrics.CellH
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
