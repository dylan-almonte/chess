package engine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type Direction int

const (
	Outbound Direction = iota
	Inbound
)

func (d Direction) String() string {
	if d == Outbound {
		return ">"
	}
	return "<"
}

// LogEntry is one UCI transcript line with direction.
type LogEntry struct {
	Dir  Direction
	Text string
}

// LineMsg is delivered for each inbound engine line (for Bubble Tea).
type LineMsg struct {
	Text string
}

// ErrMsg is delivered when the engine reader fails.
type ErrMsg struct {
	Err error
}

// ExitMsg is delivered when the engine process exits.
type ExitMsg struct {
	Err error
}

// Client owns a UCI engine subprocess or an in-process scripted stub.
type Client struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	handler func(string) []string // scripted mode
	mu      sync.Mutex
	log     []LogEntry
	onLog   func(LogEntry)
	emit    func(any) // optional; scripted Send invokes for each reply
}

// Start launches the engine binary.
func Start(path string, args ...string) (*Client, error) {
	cmd := exec.Command(path, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, err
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return nil, fmt.Errorf("start engine %s: %w", path, err)
	}
	return &Client{cmd: cmd, stdin: stdin, stdout: stdout}, nil
}

// StartFake starts an in-process scripted engine (no OS process / pipes).
// Send() invokes handler synchronously and records inbound replies.
func StartFake(handler func(line string) []string) (*Client, error) {
	if handler == nil {
		handler = StubHandler
	}
	return &Client{handler: handler}, nil
}

// SetLogHook registers a callback for every sent/received line.
func (c *Client) SetLogHook(fn func(LogEntry)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onLog = fn
}

// SetEmit sets the callback used by scripted Send for inbound lines.
func (c *Client) SetEmit(fn func(any)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.emit = fn
}

func (c *Client) appendLog(e LogEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.log = append(c.log, e)
	if c.onLog != nil {
		c.onLog(e)
	}
}

// Log returns a copy of the transcript.
func (c *Client) Log() []LogEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]LogEntry, len(c.log))
	copy(out, c.log)
	return out
}

// Send writes one UCI line to the engine.
func (c *Client) Send(line string) error {
	line = strings.TrimSpace(line)
	c.appendLog(LogEntry{Dir: Outbound, Text: line})

	if c.handler != nil {
		replies := c.handler(line)
		c.mu.Lock()
		emit := c.emit
		c.mu.Unlock()
		for _, r := range replies {
			r = strings.TrimSpace(r)
			c.appendLog(LogEntry{Dir: Inbound, Text: r})
			if emit != nil {
				emit(LineMsg{Text: r})
			}
		}
		return nil
	}

	_, err := fmt.Fprintln(c.stdin, line)
	return err
}

// Scripted reports whether this client is an in-process fake.
func (c *Client) Scripted() bool { return c.handler != nil }

// ReadLoop reads inbound lines from a real subprocess. For scripted fakes it only
// registers emit (replies are delivered synchronously from Send).
func (c *Client) ReadLoop(emit func(any)) {
	if c.handler != nil {
		c.SetEmit(emit)
		return
	}
	sc := bufio.NewScanner(c.stdout)
	for sc.Scan() {
		text := strings.TrimSpace(sc.Text())
		c.appendLog(LogEntry{Dir: Inbound, Text: text})
		emit(LineMsg{Text: text})
	}
	if err := sc.Err(); err != nil {
		emit(ErrMsg{Err: err})
	}
}

// Quit sends quit and waits for the process to exit (if any).
func (c *Client) Quit() error {
	_ = c.Send("quit")
	if c.stdin != nil {
		_ = c.stdin.Close()
	}
	if c.cmd != nil {
		return c.cmd.Wait()
	}
	if c.stdout != nil {
		_ = c.stdout.Close()
	}
	return nil
}

// Handshake sends uci/isready and waits for uciok/readyok.
func (c *Client) Handshake() error {
	if c.handler != nil {
		if err := c.Send("uci"); err != nil {
			return err
		}
		if !c.logContains(Inbound, "uciok") {
			return fmt.Errorf("missing uciok")
		}
		if err := c.Send("isready"); err != nil {
			return err
		}
		if !c.logContains(Inbound, "readyok") {
			return fmt.Errorf("missing readyok")
		}
		return nil
	}

	type result struct {
		line string
		err  error
	}
	ch := make(chan result, 8)
	go func() {
		sc := bufio.NewScanner(c.stdout)
		for sc.Scan() {
			text := strings.TrimSpace(sc.Text())
			c.appendLog(LogEntry{Dir: Inbound, Text: text})
			ch <- result{line: text}
		}
		if err := sc.Err(); err != nil {
			ch <- result{err: err}
		}
		close(ch)
	}()

	waitFor := func(want string) error {
		for r := range ch {
			if r.err != nil {
				return r.err
			}
			if r.line == want {
				return nil
			}
		}
		return fmt.Errorf("engine closed before %s", want)
	}

	if err := c.Send("uci"); err != nil {
		return err
	}
	if err := waitFor("uciok"); err != nil {
		return err
	}
	if err := c.Send("isready"); err != nil {
		return err
	}
	return waitFor("readyok")
}

func (c *Client) logContains(dir Direction, text string) bool {
	for _, e := range c.Log() {
		if e.Dir == dir && e.Text == text {
			return true
		}
	}
	return false
}

// StubHandler is a minimal UCI stub: handshake + bestmove e7e5 on go.
func StubHandler(line string) []string {
	switch {
	case line == "uci":
		return []string{"id name fake", "id author test", "uciok"}
	case line == "isready":
		return []string{"readyok"}
	case line == "ucinewgame":
		return nil
	case strings.HasPrefix(line, "position"):
		return nil
	case strings.HasPrefix(line, "go"):
		return []string{"bestmove e7e5"}
	case line == "quit":
		return nil
	default:
		return nil
	}
}

// InfoStubHandler emits an info line before bestmove.
func InfoStubHandler(line string) []string {
	if strings.HasPrefix(line, "go") {
		return []string{"info depth 1 score cp 12 pv e2e4", "bestmove e7e5"}
	}
	return StubHandler(line)
}
