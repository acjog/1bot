// Package zcode drives Z.AI's ZCode CLI in `app-server` mode over stdio, and
// reduces its event stream to transport.Event.
//
// The ZCode app-server protocol is newline-delimited JSON (no "jsonrpc" header)
// and, unlike ACP/Codex, has no initialize handshake. A turn is:
//
//	session/create  -> sessionId
//	session/send    -> accepted (asynchronous)
//	poll session/events until turn.completed, emitting model.streaming deltas
//
// ZCode reads its model from ~/.zcode/cli/config.json, NOT from a CLI flag:
// passing --model to app-server makes it print help and exit.
package zcode

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"1bot_test/internal/transport"
	"1bot_test/internal/vlog"
)

var stderr = os.Stderr

// Options configures a ZCode app-server child process.
type Options struct {
	Bin     string // zcode executable (name or resolved path)
	Verbose bool   // dump every JSON frame to stderr
	Cwd     string // working directory handed to the agent
}

// Client drives one long-lived `zcode app-server` process. It implements
// transport.Transport.
type Client struct {
	opts Options

	cmd     *exec.Cmd
	stdin   io.WriteCloser
	scanner *bufio.Scanner

	writeMu sync.Mutex
	pendMu  sync.Mutex
	pending map[string]chan wireMessage
	nextID  int64

	closed chan struct{}

	// promptMu serialises turns: one zcode process runs one turn at a time.
	promptMu sync.Mutex
}

func New(opts Options) *Client {
	return &Client{
		opts:    opts,
		pending: map[string]chan wireMessage{},
		closed:  make(chan struct{}),
	}
}

// Start launches the child and confirms it answers session/list.
func (c *Client) Start() error {
	c.cmd = exec.Command(c.opts.Bin, "app-server")
	c.cmd.Dir = c.opts.Cwd
	c.cmd.Stderr = stderr

	in, err := c.cmd.StdinPipe()
	if err != nil {
		return err
	}
	out, err := c.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("starting %s app-server: %w", c.opts.Bin, err)
	}
	c.stdin = in
	c.scanner = bufio.NewScanner(out)
	c.scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	go c.readLoop()

	// A cheap liveness probe: app-server has no initialize, so confirm it
	// answers a request before we claim readiness.
	if _, err := c.call("session/list", map[string]any{}); err != nil {
		_ = c.Close()
		return fmt.Errorf("zcode app-server handshake: %w", err)
	}
	return nil
}

// Close kills the child and waits for it to exit.
func (c *Client) Close() error {
	c.writeMu.Lock()
	if c.stdin != nil {
		_ = c.stdin.Close()
	}
	c.writeMu.Unlock()
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
	if c.cmd != nil {
		_ = c.cmd.Wait()
	}
	return nil
}

// readLoop consumes the child's stdout and dispatches frames. Notifications
// (e.g. state.updated) are ignored; we poll session/events ourselves.
func (c *Client) readLoop() {
	defer close(c.closed)
	for c.scanner.Scan() {
		line := c.scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var m wireMessage
		if err := json.Unmarshal(line, &m); err != nil {
			fmt.Fprintf(stderr, "zcode: bad JSON: %v: %s\n", err, line)
			continue
		}
		if c.opts.Verbose {
			vlog.Frame(stderr, "zcode <-", string(line))
		}

		switch {
		case m.hasID() && m.Method == "":
			c.deliver(m) // a response to one of our requests
		case m.hasID():
			// A request the server sent us; app-server is client-driven so
			// this should not happen. Answer method-not-found so it never hangs.
			_ = c.write(wireMessage{ID: m.ID, Error: &rpcError{Code: -32601, Message: "method not supported by 1bot_test: " + m.Method}})
		default:
			// notification, ignored
		}
	}
}

// Turn implements transport.Transport. It starts a fresh session, sends the
// user text, and polls the event stream until turn/completed.
func (c *Client) Turn(messages []transport.Message, onEvent func(transport.Event)) (string, error) {
	c.promptMu.Lock()
	defer c.promptMu.Unlock()

	text := lastUserText(messages)
	if text == "" {
		return "", fmt.Errorf("no user text")
	}

	sessionID, err := c.createSession()
	if err != nil {
		return "", err
	}

	if err := c.send(sessionID, text); err != nil {
		return "", err
	}

	return c.await(sessionID, onEvent)
}

// lastUserText returns the concatenated text of the last user message.
func lastUserText(messages []transport.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role != "user" {
			continue
		}
		var sb strings.Builder
		for _, p := range messages[i].Parts {
			if p.Kind == transport.PartText {
				sb.WriteString(p.Text)
			}
		}
		if sb.Len() > 0 {
			return sb.String()
		}
	}
	return ""
}

func (c *Client) createSession() (string, error) {
	res, err := c.call("session/create", map[string]any{
		"workspace": map[string]any{
			"workspacePath": c.opts.Cwd,
			"workspaceKey":  c.opts.Cwd,
		},
	})
	if err != nil {
		return "", err
	}
	var r struct {
		SessionID string `json:"sessionId"`
	}
	_ = json.Unmarshal(res.Result, &r)
	if r.SessionID == "" {
		fmt.Fprintf(stderr, "zcode: session/create result: %s\n", res.Result)
		return "", fmt.Errorf("session/create returned no sessionId")
	}
	return r.SessionID, nil
}

func (c *Client) send(sessionID, text string) error {
	_, err := c.call("session/send", map[string]any{
		"sessionId": sessionID,
		"content":   text,
	})
	return err
}

// zEvent is one entry of a session/events result.
type zEvent struct {
	Seq     int64           `json:"seq"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

func (c *Client) await(sessionID string, onEvent func(transport.Event)) (string, error) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	timer := time.NewTimer(300 * time.Second)
	defer timer.Stop()
	lastSeq := int64(0)

	for {
		select {
		case <-ticker.C:
			events, err := c.pollEvents(sessionID)
			if err != nil {
				return "", err
			}
			for _, e := range events {
				if e.Seq <= lastSeq {
					continue
				}
				lastSeq = e.Seq
				switch e.Type {
				case "model.streaming":
					var p struct {
						Delta string `json:"delta"`
					}
					_ = json.Unmarshal(e.Payload, &p)
					if p.Delta != "" {
						onEvent(transport.Event{Kind: transport.EventText, Text: p.Delta})
					}
				case "turn.completed":
					return transport.StopEndTurn, nil
				}
			}
		case <-c.closed:
			return "", fmt.Errorf("zcode app-server disconnected")
		case <-timer.C:
			return "", fmt.Errorf("zcode turn timed out")
		}
	}
}

func (c *Client) pollEvents(sessionID string) ([]zEvent, error) {
	res, err := c.call("session/events", map[string]any{"sessionId": sessionID})
	if err != nil {
		return nil, err
	}
	var r struct {
		Events []zEvent `json:"events"`
	}
	_ = json.Unmarshal(res.Result, &r)
	return r.Events, nil
}
