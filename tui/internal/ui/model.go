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
	cfg      Config
	phase    phase
	errMsg   string
	client   *engine.Client
	board    board.Board
	moves    []string
	logLines []string
	tele     info.Telemetry
	input    textinput.Model
	logView  viewport.Model
	waiting  bool
	width    int
	height   int
	ready    bool
	inbox    chan tea.Msg
}

func New(cfg Config) Model {
	ti := textinput.New()
	ti.Placeholder = "e2e4 / quit"
	ti.CharLimit = 16
	ti.Width = 24
	vp := viewport.New(60, 8)
	vp.SetContent("")
	return Model{
		cfg:     cfg,
		phase:   phaseBoot,
		board:   board.StartPos(),
		tele:    info.Idle(),
		input:   ti,
		logView: vp,
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
		m.logView.Width = max(20, msg.Width-4)
		m.logView.Height = max(4, msg.Height/4)
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
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd

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
		}
	case phasePlay:
		if t, ok := info.ParseInfo(text); ok {
			m.tele = t
		}
		if strings.HasPrefix(text, "bestmove") {
			m.waiting = false
			parts := strings.Fields(text)
			if len(parts) >= 2 && parts[1] != "(none)" {
				mv := parts[1]
				if err := m.board.ApplyUCI(mv); err == nil {
					m.moves = append(m.moves, mv)
				}
			}
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
	if m.waiting {
		return m, nil
	}
	if err := m.board.ApplyUCI(raw); err != nil {
		m.logLines = append(m.logLines, "! "+err.Error())
		m.refreshLogView()
		return m, nil
	}
	m.moves = append(m.moves, raw)
	pos := "position startpos moves " + strings.Join(m.moves, " ")
	_ = m.send(pos)
	_ = m.send("go")
	m.waiting = true
	if m.inbox != nil {
		return m, waitInbox(m.inbox)
	}
	return m, nil
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

func (m Model) View() string {
	if m.phase == phaseError {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(
			fmt.Sprintf("Engine error: %s\n\nPress q or Enter to exit.", m.errMsg),
		)
	}
	if m.phase == phaseBoot || m.phase == phaseHandshakeUCI || m.phase == phaseHandshakeReady {
		return "Connecting to engine…"
	}

	boardStr := m.board.Render()
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
	inputLine := "input: " + m.input.View()
	return lipgloss.JoinVertical(lipgloss.Left, top, logBlock, teleBlock, inputLine)
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
func (m Model) PhaseError() bool          { return m.phase == phaseError }
func (m Model) ErrorText() string         { return m.errMsg }
func (m Model) LogText() string           { return strings.Join(m.logLines, "\n") }
func (m Model) Moves() []string           { return append([]string(nil), m.moves...) }
func (m Model) Board() board.Board        { return m.board }
func (m Model) Telemetry() info.Telemetry { return m.tele }
func (m Model) IsPlayable() bool          { return m.phase == phasePlay }

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
