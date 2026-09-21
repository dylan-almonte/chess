package ui_test

import (
	"fmt"
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

func viewHasGlyph(view, glyph string) bool {
	return strings.Contains(view, glyph)
}

func TestToggleShowsASCIILettersWithoutMovingPieces(t *testing.T) {
	m := connectFake(t, engine.StubHandler)
	if !viewHasGlyph(m.View(), "♙") || !viewHasGlyph(m.View(), "♟") {
		t.Fatalf("expected Unicode pieces before toggle, view:\n%s", m.View())
	}
	movesBefore := append([]string(nil), m.Moves()...)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = model.(ui.Model)
	if viewHasGlyph(m.View(), "♙") || viewHasGlyph(m.View(), "♟") {
		t.Fatalf("expected ASCII letters after toggle, view:\n%s", m.View())
	}
	if !viewHasGlyph(m.View(), "P") || !viewHasGlyph(m.View(), "p") {
		t.Fatalf("expected ASCII P/p after toggle, view:\n%s", m.View())
	}
	if got := m.Moves(); len(got) != len(movesBefore) {
		t.Fatalf("toggle must not change move list, before %v after %v", movesBefore, got)
	}
}

func TestToggleBackRestoresUnicode(t *testing.T) {
	m := connectFake(t, engine.StubHandler)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = model.(ui.Model)
	if viewHasGlyph(m.View(), "♙") {
		t.Fatal("expected ASCII after first Tab")
	}
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = model.(ui.Model)
	if !viewHasGlyph(m.View(), "♙") || !viewHasGlyph(m.View(), "♟") {
		t.Fatalf("expected Unicode after second Tab, view:\n%s", m.View())
	}
}

func TestChosenModeSurvivesAMove(t *testing.T) {
	h := engine.NewStatefulHandler()
	h.SetLegalMoves("e2e4", blackAfterE4())
	m := connectStateful(t, h)
	if !viewHasGlyph(m.View(), "♙") {
		t.Fatalf("expected Unicode default before toggle, view:\n%s", m.View())
	}
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = model.(ui.Model)
	m.SetInputValue("e2e4")
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = pump(t, model, cmd, func(u ui.Model) bool {
		return len(u.Moves()) >= 1 && u.Moves()[0] == "e2e4"
	}, 60)
	view := m.View()
	if viewHasGlyph(view, "♙") || viewHasGlyph(view, "♟") {
		t.Fatalf("ASCII mode should survive e2e4, view:\n%s", view)
	}
	if !viewHasGlyph(view, "P") {
		t.Fatalf("expected ASCII P after e2e4, view:\n%s", view)
	}
}

func clickSquare(m ui.Model, name string) (ui.Model, tea.Cmd) {
	x, y := cellFor(name)
	model, cmd := m.Update(tea.MouseMsg{
		X:      x,
		Y:      y,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		Type:   tea.MouseLeft,
	})
	return model.(ui.Model), cmd
}

func outboundCount(m ui.Model) int {
	n := 0
	for _, e := range m.ClientLog() {
		if e.Dir == engine.Outbound {
			n++
		}
	}
	return n
}

func TestClickE2ThenE4PlaysThroughTheTypedPath(t *testing.T) {
	h := engine.NewStatefulHandler()
	h.SetLegalMoves("e2e4", blackAfterE4())
	m := connectStateful(t, h)
	m, cmd := clickSquare(m, "e2")
	m, cmd2 := clickSquare(m, "e4")
	m = pump(t, m, tea.Batch(cmd, cmd2), func(u ui.Model) bool {
		log := u.LogText()
		return len(u.Moves()) >= 1 && u.Moves()[0] == "e2e4" &&
			strings.Contains(log, "> go")
	}, 80)
	if m.Board().PieceAtName("e4") != 'P' || m.Board().PieceAtName("e2") != 0 {
		t.Fatal("expected e2e4 on the board")
	}
	log := m.LogText()
	if !strings.Contains(log, "> position") || !strings.Contains(log, "e2e4") {
		t.Fatalf("expected outbound position with e2e4, log:\n%s", log)
	}
	if !strings.Contains(log, "> go") {
		t.Fatalf("expected outbound go, log:\n%s", log)
	}
}

func TestClickingAnIllegalDestinationIsRejected(t *testing.T) {
	h := engine.NewStatefulHandler()
	m := connectStateful(t, h)
	before := outboundCount(m)
	m, _ = clickSquare(m, "e2")
	m, _ = clickSquare(m, "e5")
	if len(m.Moves()) != 0 {
		t.Fatalf("expected empty move list, got %v", m.Moves())
	}
	if m.Board().PieceAtName("e2") != 'P' || m.Board().PieceAtName("e5") != 0 {
		t.Fatal("board should be unchanged")
	}
	if !strings.Contains(m.LogText(), "e2e5") || !strings.Contains(m.LogText(), "illegal") {
		t.Fatalf("expected illegal e2e5, log:\n%s", m.LogText())
	}
	if outboundCount(m) != before {
		t.Fatalf("expected no new outbound commands")
	}
}

func TestClickingTheSelectedSquareCancelsTheSelection(t *testing.T) {
	h := engine.NewStatefulHandler()
	m := connectStateful(t, h)
	before := outboundCount(m)
	m, _ = clickSquare(m, "e2")
	if m.SelectedSquare() != "e2" {
		t.Fatalf("expected e2 selected, got %q", m.SelectedSquare())
	}
	m, _ = clickSquare(m, "e2")
	if m.SelectedSquare() != "" {
		t.Fatalf("re-click should cancel, got %q", m.SelectedSquare())
	}
	if len(m.Moves()) != 0 {
		t.Fatalf("cancel must not submit, got %v", m.Moves())
	}
	if m.Board().PieceAtName("e2") != 'P' {
		t.Fatal("board should be unchanged")
	}
	if outboundCount(m) != before {
		t.Fatalf("expected no new outbound commands")
	}
}

func TestClicksAreIgnoredWhileWaitingForTheEngine(t *testing.T) {
	h := engine.NewStatefulHandler()
	h.SetLegalMoves("e2e4", blackAfterE4())
	m := connectStateful(t, h)
	m.SetInputValue("e2e4")
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(ui.Model)
	beforeMoves := append([]string(nil), m.Moves()...)
	beforeOut := outboundCount(m)
	m, _ = clickSquare(m, "e7")
	if m.SelectedSquare() != "" {
		t.Fatalf("waiting clicks must not select, got %q", m.SelectedSquare())
	}
	m, _ = clickSquare(m, "e5")
	if got := m.Moves(); len(got) != len(beforeMoves) || (len(got) > 0 && got[0] != "e2e4") {
		t.Fatalf("waiting clicks must not add moves, got %v", got)
	}
	if outboundCount(m) != beforeOut {
		t.Fatalf("waiting clicks must not send more UCI")
	}
}

func TestArrowSelectionPlaysE2E4(t *testing.T) {
	h := engine.NewStatefulHandler()
	h.SetLegalMoves("e2e4", blackAfterE4())
	m := connectStateful(t, h)
	// Default cursor is e2; confirm, move up to e3 then e4, confirm.
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = model.(ui.Model)
	model, cmd2 := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = model.(ui.Model)
	model, cmd3 := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = model.(ui.Model)
	model, cmd4 := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = pump(t, model, tea.Batch(cmd, cmd2, cmd3, cmd4), func(u ui.Model) bool {
		log := u.LogText()
		return len(u.Moves()) >= 1 && u.Moves()[0] == "e2e4" &&
			strings.Contains(log, "> go")
	}, 80)
	if m.Board().PieceAtName("e4") != 'P' || m.Board().PieceAtName("e2") != 0 {
		t.Fatal("expected e2e4 on the board")
	}
	log := m.LogText()
	if !strings.Contains(log, "> position") || !strings.Contains(log, "e2e4") {
		t.Fatalf("expected outbound position with e2e4, log:\n%s", log)
	}
	if !strings.Contains(log, "> go") {
		t.Fatalf("expected outbound go, log:\n%s", log)
	}
}

func resize(m ui.Model, w, h int) ui.Model {
	model, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return model.(ui.Model)
}

func TestLargeWindowUsesBiggerSquares(t *testing.T) {
	m := resize(connectFake(t, engine.StubHandler), 80, 28)
	if m.CellWidth() < 3 || m.CellHeight() < 2 {
		t.Fatalf("expected cell at least 3x2 at 80x28, got %dx%d", m.CellWidth(), m.CellHeight())
	}
	if !viewHasGlyph(m.View(), "♙") && !viewHasGlyph(m.View(), "P") {
		t.Fatalf("e2 pawn missing in large view:\n%s", m.View())
	}
}

func TestNarrowWindowKeepsAReadable8x8Grid(t *testing.T) {
	m := resize(connectFake(t, engine.StubHandler), 60, 20)
	view := m.View()
	for rank := 1; rank <= 8; rank++ {
		if !strings.Contains(view, fmt.Sprintf("%d", rank)) {
			t.Fatalf("missing rank %d in narrow view:\n%s", rank, view)
		}
	}
	for _, file := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		if !strings.Contains(view, file) {
			t.Fatalf("missing file %s in narrow view:\n%s", file, view)
		}
	}
	if !viewHasGlyph(view, "♙") && !viewHasGlyph(view, "P") {
		t.Fatalf("e2 pawn missing in narrow view:\n%s", view)
	}
}

func typeCmd(m ui.Model, raw string) ui.Model {
	m.SetInputValue(raw)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	return model.(ui.Model)
}

func TestPlaceAQueenOnD5(t *testing.T) {
	m := connectStateful(t, engine.NewStatefulHandler())
	beforeGo := strings.Count(m.LogText(), "> go")
	m = typeCmd(m, "setup")
	m = typeCmd(m, "Q")
	m, _ = clickSquare(m, "d5")
	if m.Board().PieceAtName("d5") != 'Q' {
		t.Fatalf("expected white queen on d5, got %c", m.Board().PieceAtName("d5"))
	}
	if len(m.Moves()) != 0 {
		t.Fatalf("setup must not append moves, got %v", m.Moves())
	}
	if strings.Count(m.LogText(), "> go") != beforeGo {
		t.Fatal("setup must not send go")
	}
}

func TestRemoveAPieceFromE2(t *testing.T) {
	m := connectStateful(t, engine.NewStatefulHandler())
	m = typeCmd(m, "setup")
	m = typeCmd(m, "x")
	m, _ = clickSquare(m, "e2")
	if m.Board().PieceAtName("e2") != 0 {
		t.Fatalf("e2 should be empty, got %c", m.Board().PieceAtName("e2"))
	}
	if len(m.Moves()) != 0 {
		t.Fatalf("setup must not append moves, got %v", m.Moves())
	}
}

func TestPlayAKingsAndQueenSetup(t *testing.T) {
	h := engine.NewStatefulHandler()
	h.SetLegalMoves("", []string{"d1d8", "e1d2", "e1e2", "e1f2", "e1f1", "e1d1"})
	m := connectStateful(t, h)
	m = typeCmd(m, "setup")
	m = typeCmd(m, "clear")
	m = typeCmd(m, "K")
	m, _ = clickSquare(m, "e1")
	m = typeCmd(m, "Q")
	m, _ = clickSquare(m, "d1")
	m = typeCmd(m, "k")
	m, _ = clickSquare(m, "e8")
	m = typeCmd(m, "play")
	log := m.LogText()
	if !strings.Contains(log, "> position fen 4k3/8/8/8/8/8/8/3QK3 w - - 0 1") {
		t.Fatalf("expected custom FEN position, log:\n%s", log)
	}
	if !strings.Contains(log, "> legalmoves") {
		t.Fatalf("expected legalmoves after play, log:\n%s", log)
	}
	if m.Board().PieceAtName("e1") != 'K' || m.Board().PieceAtName("d1") != 'Q' || m.Board().PieceAtName("e8") != 'k' {
		t.Fatal("board should keep the three setup pieces")
	}
}

func TestSetupWithoutBothKingsIsRejected(t *testing.T) {
	m := connectStateful(t, engine.NewStatefulHandler())
	beforePos := strings.Count(m.LogText(), "> position")
	beforeGo := strings.Count(m.LogText(), "> go")
	m = typeCmd(m, "setup")
	m = typeCmd(m, "clear")
	m = typeCmd(m, "K")
	m, _ = clickSquare(m, "e1")
	m = typeCmd(m, "play")
	if !strings.Contains(m.LogText(), "king") {
		t.Fatalf("expected both-kings error, log:\n%s", m.LogText())
	}
	if strings.Count(m.LogText(), "> position") != beforePos {
		t.Fatal("rejected play must not send position")
	}
	if strings.Count(m.LogText(), "> go") != beforeGo {
		t.Fatal("rejected play must not send go")
	}
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
