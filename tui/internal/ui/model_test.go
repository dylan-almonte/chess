package ui_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dylanca/chess-tui/tui/internal/engine"
	"github.com/dylanca/chess-tui/tui/internal/ui"
)

func collectMsgs(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if msg == nil {
		return nil
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		var out []tea.Msg
		for _, c := range batch {
			out = append(out, collectMsgs(c)...)
		}
		return out
	}
	return []tea.Msg{msg}
}

func pump(t *testing.T, m tea.Model, cmd tea.Cmd, pred func(ui.Model) bool, limit int) ui.Model {
	t.Helper()
	um := m.(ui.Model)
	pending := collectMsgs(cmd)
	for i := 0; i < limit; i++ {
		if pred(um) {
			return um
		}
		if len(pending) == 0 {
			t.Fatalf("no pending msgs; ready=%v log=%q moves=%v", um.PhaseReady(), um.LogText(), um.Moves())
		}
		msg := pending[0]
		pending = pending[1:]
		if msg == nil {
			continue
		}
		var next tea.Cmd
		m, next = um.Update(msg)
		um = m.(ui.Model)
		if pred(um) {
			return um
		}
		pending = append(pending, collectMsgs(next)...)
	}
	t.Fatalf("limit exceeded; ready=%v log=%q moves=%v", um.PhaseReady(), um.LogText(), um.Moves())
	return um
}

func connectFake(t *testing.T, handler func(string) []string) ui.Model {
	t.Helper()
	m := ui.New(ui.Config{FakeHandler: handler})
	cmd := m.Init()
	return pump(t, m, cmd, func(u ui.Model) bool { return u.LegalReady() }, 60)
}

func TestSuccessfulEngineConnect(t *testing.T) {
	m := connectFake(t, engine.StubHandler)
	log := m.LogText()
	for _, want := range []string{"> uci", "< uciok", "> isready", "< readyok"} {
		if !strings.Contains(log, want) {
			t.Fatalf("log missing %q:\n%s", want, log)
		}
	}
	if m.Board().PieceAtName("e2") != 'P' {
		t.Fatal("board should show startpos")
	}
}

func TestMissingEngineBinaryFailsClearly(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-engine")
	m := ui.New(ui.Config{EnginePath: missing})
	msg := m.Init()()
	model, _ := m.Update(msg)
	um := model.(ui.Model)
	if !um.PhaseError() {
		t.Fatal("expected error phase")
	}
	if !strings.Contains(um.ErrorText(), missing) {
		t.Fatalf("error should name binary, got %q", um.ErrorText())
	}
	if um.IsPlayable() {
		t.Fatal("must not present playable session")
	}
}

func TestHumanMoveUpdatesBoardAndList(t *testing.T) {
	h := engine.NewStatefulHandler()
	h.SetLegalMoves("e2e4", blackAfterE4())
	m := connectStateful(t, h)
	m.SetInputValue("e2e4")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = pump(t, model, cmd, func(u ui.Model) bool {
		return len(u.Moves()) >= 1 && u.Moves()[0] == "e2e4" && u.Board().PieceAtName("e4") == 'P'
	}, 60)
	if m.Board().PieceAtName("e2") != 0 {
		t.Fatal("e2 should be empty")
	}
}

func TestEngineReplyUpdatesBoardAndList(t *testing.T) {
	h := engine.NewStatefulHandler()
	h.SetLegalMoves("e2e4", blackAfterE4())
	h.SetLegalMoves("e2e4 e7e5", whiteAfterE4E5())
	m := connectStateful(t, h)
	m.SetInputValue("e2e4")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = pump(t, model, cmd, func(u ui.Model) bool {
		return len(u.Moves()) >= 2 && u.Moves()[1] == "e7e5"
	}, 80)
	if m.Board().PieceAtName("e5") != 'p' {
		t.Fatal("expected black pawn on e5")
	}
}

func TestPositionAndGoAppearInTheLog(t *testing.T) {
	h := engine.NewStatefulHandler()
	h.SetLegalMoves("e2e4", blackAfterE4())
	m := connectStateful(t, h)
	m.SetInputValue("e2e4")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = pump(t, model, cmd, func(u ui.Model) bool {
		log := u.LogText()
		return strings.Contains(log, "> position") &&
			strings.Contains(log, "> go") &&
			strings.Contains(log, "< bestmove")
	}, 80)
}

func TestInfoLinesPopulateTelemetry(t *testing.T) {
	h := engine.NewStatefulHandler()
	h.SetLegalMoves("e2e4", blackAfterE4())
	h.SetInfoBeforeBestmove("info depth 1 score cp 12 pv e2e4")
	m := connectStateful(t, h)
	m.SetInputValue("e2e4")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = pump(t, model, cmd, func(u ui.Model) bool {
		tel := u.Telemetry()
		return tel.Depth == "1" && tel.Score == "cp 12"
	}, 80)
}

func TestStubEngineLeavesTelemetryIdle(t *testing.T) {
	h := engine.NewStatefulHandler()
	h.SetLegalMoves("e2e4", blackAfterE4())
	h.SetLegalMoves("e2e4 e7e5", whiteAfterE4E5())
	m := connectStateful(t, h)
	if !m.Telemetry().Idle {
		t.Fatal("expected idle telemetry")
	}
	m.SetInputValue("e2e4")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = pump(t, model, cmd, func(u ui.Model) bool { return len(u.Moves()) >= 2 }, 80)
	if !m.Telemetry().Idle {
		t.Fatal("telemetry should stay idle without info")
	}
}

func TestQuitSendsUCIQuit(t *testing.T) {
	m := connectFake(t, engine.StubHandler)
	model, cmd := m.BeginQuit()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) && cmd != nil {
		for _, msg := range collectMsgs(cmd) {
			if msg == nil {
				continue
			}
			if _, ok := msg.(tea.QuitMsg); ok {
				goto done
			}
			model, cmd = model.Update(msg)
			goto cont
		}
		break
	cont:
	}
done:
	found := false
	for _, e := range model.(ui.Model).ClientLog() {
		if e.Dir == engine.Outbound && e.Text == "quit" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected quit in client log, got %#v", model.(ui.Model).ClientLog())
	}
}

func connectStateful(t *testing.T, h *engine.StatefulHandler) ui.Model {
	t.Helper()
	return connectFake(t, h.Handle)
}

func TestIllegalPawnMoveE2E5Rejected(t *testing.T) {
	h := engine.NewStatefulHandler()
	m := connectStateful(t, h)

	// Count outbound commands before submission
	logBefore := m.ClientLog()
	outboundCountBefore := 0
	for _, e := range logBefore {
		if e.Dir == engine.Outbound {
			outboundCountBefore++
		}
	}

	m.SetInputValue("e2e5")
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(ui.Model)

	// Move list should remain empty
	if len(m.Moves()) != 0 {
		t.Fatalf("expected empty move list, got %v", m.Moves())
	}

	// Board should be unchanged: white pawn still on e2, e5 empty
	if m.Board().PieceAtName("e2") != 'P' {
		t.Fatalf("expected white pawn on e2, got %c", m.Board().PieceAtName("e2"))
	}
	if m.Board().PieceAtName("e5") != 0 {
		t.Fatalf("expected e5 empty, got %c", m.Board().PieceAtName("e5"))
	}

	// Error should mention e2e5
	log := m.LogText()
	if !strings.Contains(log, "e2e5") || !strings.Contains(log, "illegal") {
		t.Fatalf("expected illegal move error mentioning e2e5 in log, got:\n%s", log)
	}

	// No additional outbound position or go commands
	outboundCountAfter := 0
	for _, e := range m.ClientLog() {
		if e.Dir == engine.Outbound {
			outboundCountAfter++
		}
	}
	if outboundCountAfter != outboundCountBefore {
		t.Fatalf("expected no new outbound commands, had %d before, have %d after",
			outboundCountBefore, outboundCountAfter)
	}
}

func TestIllegalOpponentMoveE7E5Rejected(t *testing.T) {
	h := engine.NewStatefulHandler()
	m := connectStateful(t, h)

	logBefore := m.ClientLog()
	outboundCountBefore := 0
	for _, e := range logBefore {
		if e.Dir == engine.Outbound {
			outboundCountBefore++
		}
	}

	m.SetInputValue("e7e5")
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(ui.Model)

	// Move list should remain empty
	if len(m.Moves()) != 0 {
		t.Fatalf("expected empty move list, got %v", m.Moves())
	}

	// Board unchanged
	if m.Board().PieceAtName("e7") != 'p' {
		t.Fatalf("expected black pawn on e7, got %c", m.Board().PieceAtName("e7"))
	}
	if m.Board().PieceAtName("e5") != 0 {
		t.Fatalf("expected e5 empty, got %c", m.Board().PieceAtName("e5"))
	}

	// Error should mention e7e5
	log := m.LogText()
	if !strings.Contains(log, "e7e5") || !strings.Contains(log, "illegal") {
		t.Fatalf("expected illegal move error mentioning e7e5 in log, got:\n%s", log)
	}

	// No additional outbound commands
	outboundCountAfter := 0
	for _, e := range m.ClientLog() {
		if e.Dir == engine.Outbound {
			outboundCountAfter++
		}
	}
	if outboundCountAfter != outboundCountBefore {
		t.Fatalf("expected no new outbound commands, had %d before, have %d after",
			outboundCountBefore, outboundCountAfter)
	}
}

func TestLegalMoveE2E4AcceptedAndSynchronized(t *testing.T) {
	h := engine.NewStatefulHandler()
	// After e2e4, black has these legal moves (engine side)
	h.SetLegalMoves("e2e4", []string{
		"e7e5", "d7d5", "c7c5", "g8f6", "b8c6",
		"a7a6", "a7a5", "b7b6", "b7b5", "c7c6",
		"d7d6", "e7e6", "f7f6", "f7f5", "g7g6",
		"g7g5", "h7h6", "h7h5", "g8h6", "b8a6",
	})
	// After e2e4 e7e5, white has these (subset for testing)
	h.SetLegalMoves("e2e4 e7e5", []string{
		"g1f3", "d2d4", "f1c4", "b1c3", "d1h5",
		"a2a3", "a2a4", "b2b3", "b2b4", "c2c3",
		"c2c4", "d2d3", "f2f3", "f2f4", "g2g3",
		"g2g4", "h2h3", "h2h4", "b1a3", "g1h3",
	})
	m := connectStateful(t, h)

	m.SetInputValue("e2e4")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = pump(t, model, cmd, func(u ui.Model) bool {
		return len(u.Moves()) >= 2 && u.Moves()[0] == "e2e4" && u.Moves()[1] == "e7e5"
	}, 80)

	// Board should reflect both moves
	if m.Board().PieceAtName("e4") != 'P' {
		t.Fatalf("expected white pawn on e4, got %c", m.Board().PieceAtName("e4"))
	}
	if m.Board().PieceAtName("e5") != 'p' {
		t.Fatalf("expected black pawn on e5, got %c", m.Board().PieceAtName("e5"))
	}

	// Engine should have been synchronized with full history
	clog := m.ClientLog()
	foundSync := false
	for _, e := range clog {
		if e.Dir == engine.Outbound && e.Text == "position startpos moves e2e4 e7e5" {
			foundSync = true
		}
	}
	if !foundSync {
		t.Fatalf("expected outbound 'position startpos moves e2e4 e7e5' in log, got %#v", clog)
	}

	// legalmoves should have been requested for the resulting position
	foundLegal := false
	for _, e := range clog {
		if e.Dir == engine.Outbound && e.Text == "legalmoves" {
			foundLegal = true
		}
	}
	if !foundLegal {
		t.Fatalf("expected outbound 'legalmoves' in log, got %#v", clog)
	}
}

func TestIllegalEngineBestmoveRejected(t *testing.T) {
	h := engine.NewStatefulHandler()
	h.SetLegalMoves("e2e4", []string{
		"e7e5", "d7d5", "c7c5", "g8f6", "b8c6",
		"a7a6", "a7a5", "b7b6", "b7b5", "c7c6",
		"d7d6", "e7e6", "f7f6", "f7f5", "g7g6",
		"g7g5", "h7h6", "h7h5", "g8h6", "b8a6",
	})
	m := connectStateful(t, h)

	// Configure illegal bestmove before submitting move
	h.SetBestmove("e2e5")

	m.SetInputValue("e2e4")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Pump until we either see the error or the move is incorrectly accepted
	m = pump(t, model, cmd, func(u ui.Model) bool {
		logText := u.LogText()
		// Wait for the error to appear, or for the model to stop waiting (bestmove processed)
		return strings.Contains(logText, "illegal") ||
			len(u.Moves()) >= 2 ||
			(!u.IsPlayable())
	}, 80)

	// Only e2e4 should be in moves (not e2e5)
	if len(m.Moves()) != 1 || m.Moves()[0] != "e2e4" {
		t.Fatalf("expected moves [e2e4] only, got %v", m.Moves())
	}

	// Board should show e2e4 but NOT e2e5
	if m.Board().PieceAtName("e4") != 'P' {
		t.Fatalf("expected white pawn on e4, got %c", m.Board().PieceAtName("e4"))
	}
	if m.Board().PieceAtName("e5") != 0 {
		t.Fatalf("expected e5 empty after rejecting illegal bestmove, got %c", m.Board().PieceAtName("e5"))
	}

	// Error should mention the illegal engine move
	log2 := m.LogText()
	if !strings.Contains(log2, "illegal") || !strings.Contains(log2, "e2e5") {
		t.Fatalf("expected error about illegal engine move e2e5, got:\n%s", log2)
	}
}

func TestAdvertisedHumanMoveDisplayFailureIsVisibleWithoutMutation(t *testing.T) {
	h := engine.NewStatefulHandler()
	h.SetLegalMoves("", []string{"e3e4"})
	m := connectStateful(t, h)

	m.SetInputValue("e3e4")
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(ui.Model)

	if len(m.Moves()) != 0 {
		t.Fatalf("display failure must not append a move, got %v", m.Moves())
	}
	if m.Board().PieceAtName("e3") != 0 || m.Board().PieceAtName("e4") != 0 {
		t.Fatal("display failure must leave the board unchanged")
	}
	if !strings.Contains(m.LogText(), "empty origin e3") {
		t.Fatalf("expected visible display error, got:\n%s", m.LogText())
	}
}

func TestAdvertisedEngineMoveDisplayFailureIsVisibleWithoutMutation(t *testing.T) {
	h := engine.NewStatefulHandler()
	h.SetLegalMoves("e2e4", []string{"e3e4"})
	m := connectStateful(t, h)

	m.SetInputValue("e2e4")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = pump(t, model, cmd, func(u ui.Model) bool {
		return strings.Contains(u.LogText(), "cannot display advertised engine move")
	}, 80)

	if got := m.Moves(); len(got) != 1 || got[0] != "e2e4" {
		t.Fatalf("display failure must preserve accepted history, got %v", got)
	}
	if m.Board().PieceAtName("e4") != 'P' || m.Board().PieceAtName("e3") != 0 {
		t.Fatal("display failure must preserve the board after e2e4")
	}
	if !strings.Contains(m.LogText(), "e3e4") {
		t.Fatalf("expected display error to identify e3e4, got:\n%s", m.LogText())
	}
}

// Helper: black legal moves after e2e4
func blackAfterE4() []string {
	return []string{
		"e7e5", "d7d5", "c7c5", "g8f6", "b8c6",
		"a7a6", "a7a5", "b7b6", "b7b5", "c7c6",
		"d7d6", "e7e6", "f7f6", "f7f5", "g7g6",
		"g7g5", "h7h6", "h7h5", "g8h6", "b8a6",
	}
}

// Helper: white legal moves after e2e4 e7e5
func whiteAfterE4E5() []string {
	return []string{
		"g1f3", "d2d4", "f1c4", "b1c3", "d1h5",
		"a2a3", "a2a4", "b2b3", "b2b4", "c2c3",
		"c2c4", "d2d3", "f2f3", "f2f4", "g2g3",
		"g2g4", "h2h3", "h2h4", "b1a3", "g1h3",
	}
}
