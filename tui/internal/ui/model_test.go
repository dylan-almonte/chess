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
	return pump(t, m, cmd, func(u ui.Model) bool { return u.PhaseReady() }, 40)
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
	m := connectFake(t, engine.StubHandler)
	m.SetInputValue("e2e4")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = pump(t, model, cmd, func(u ui.Model) bool {
		return len(u.Moves()) >= 1 && u.Moves()[0] == "e2e4" && u.Board().PieceAtName("e4") == 'P'
	}, 40)
	if m.Board().PieceAtName("e2") != 0 {
		t.Fatal("e2 should be empty")
	}
}

func TestEngineReplyUpdatesBoardAndList(t *testing.T) {
	m := connectFake(t, engine.StubHandler)
	m.SetInputValue("e2e4")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = pump(t, model, cmd, func(u ui.Model) bool {
		return len(u.Moves()) >= 2 && u.Moves()[1] == "e7e5"
	}, 40)
	if m.Board().PieceAtName("e5") != 'p' {
		t.Fatal("expected black pawn on e5")
	}
}

func TestPositionAndGoAppearInTheLog(t *testing.T) {
	m := connectFake(t, engine.StubHandler)
	m.SetInputValue("e2e4")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = pump(t, model, cmd, func(u ui.Model) bool {
		log := u.LogText()
		return strings.Contains(log, "> position") &&
			strings.Contains(log, "> go") &&
			strings.Contains(log, "< bestmove")
	}, 40)
}

func TestInfoLinesPopulateTelemetry(t *testing.T) {
	m := connectFake(t, engine.InfoStubHandler)
	m.SetInputValue("e2e4")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = pump(t, model, cmd, func(u ui.Model) bool {
		tel := u.Telemetry()
		return tel.Depth == "1" && tel.Score == "cp 12"
	}, 40)
}

func TestStubEngineLeavesTelemetryIdle(t *testing.T) {
	m := connectFake(t, engine.StubHandler)
	if !m.Telemetry().Idle {
		t.Fatal("expected idle telemetry")
	}
	m.SetInputValue("e2e4")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = pump(t, model, cmd, func(u ui.Model) bool { return len(u.Moves()) >= 2 }, 40)
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
