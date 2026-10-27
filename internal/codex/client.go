// Package codex drives OpenAI Codex CLI in `app-server` mode over stdio. It
// speaks the v2 app-server protocol (thread/turn/item notifications) and
// reduces the stream to transport.Event.
//
// The app-server protocol omits the "jsonrpc" header, uses thread/start and
// turn/start to begin a turn, and streams replies as notifications:
// item/agentMessage/delta carries the text; turn/completed carries the status.
package codex

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
)

var stderr = os.Stderr

// Options configures a Codex app-server child process.
type Options struct {
	Bin     string // codex executable (name or resolved path)
	Model   string // model id handed to thread/start
	Verbose bool   // dump every JSON-RPC frame to stderr
	Cwd     string // working directory handed to the agent
}

// Client drives one long-lived `codex app-server` process. It implements
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

	threadID string

	// promptMu serialises turns: one codex process runs one turn at a time.
	promptMu sync.Mutex
	// eventMu guards the live-turn state below. Held only briefly, never
	// across network I/O.
	eventMu   sync.Mutex
	curEvents func(transport.Event)
	curTurnID string
	done      chan string // carries the turn status when turn/completed arrives
}

func New(opts Options) *Client {
	return &Client{
		opts:    opts,
		pending: map[string]chan wireMessage{},
	}
}

// Start launches the child and performs the app-server initialize handshake.
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

	return c.initialize()
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

// readLoop consumes the child's stdout and dispatches frames.
func (c *Client) readLoop() {
	for c.scanner.Scan() {
		line := c.scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var m wireMessage
		if err := json.Unmarshal(line, &m); err != nil {
			fmt.Fprintf(stderr, "codex: bad JSON: %v: %s\n", err, line)
			continue
		}
		if c.opts.Verbose {
			fmt.Fprintf(stderr, "[<-] %s\n", line)
		}

		switch {
		case m.hasID() && m.Method == "":
			c.deliver(m) // a response to one of our requests
		case m.hasID():
			c.answer(m) // a request the server sent us (rare in app-server)
		default:
			c.notifyEvent(m) // a notification (streaming event)
		}
	}
}

// initialize performs the app-server handshake: initialize request, then an
// initialized notification.
func (c *Client) initialize() error {
	res, err := c.call("initialize", map[string]any{
		"clientInfo": map[string]any{
			"name":    "1bot_test",
			"version": "0.1",
		},
		"capabilities": map[string]any{},
	})
	if err != nil {
		return err
	}
	var r struct {
		UserAgent     string `json:"userAgent"`
		PlatformFamily string `json:"platformFamily"`
	}
	_ = json.Unmarshal(res.Result, &r)
	fmt.Fprintf(stderr, "codex: connected (%s, %s)\n", r.UserAgent, r.PlatformFamily)
	return c.notify("initialized", map[string]any{})
}

// startThread creates a conversation thread and remembers its id.
func (c *Client) startThread() error {
	params := map[string]any{}
	if c.opts.Model != "" {
		params["model"] = c.opts.Model
	}
	res, err := c.call("thread/start", params)
	if err != nil {
		return err
	}
	var r struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if err := json.Unmarshal(res.Result, &r); err != nil {
		return fmt.Errorf("thread/start result: %w", err)
	}
	c.threadID = r.Thread.ID
	return nil
}

// Turn implements transport.Transport. It starts a turn on the current thread
// and streams the reply through onEvent until turn/completed arrives. Turns
// are serialised: one process, one turn.
func (c *Client) Turn(text string, onEvent func(transport.Event)) (string, error) {
	c.promptMu.Lock()
	defer c.promptMu.Unlock()

	if c.threadID == "" {
		if err := c.startThread(); err != nil {
			return "", err
		}
	}

	res, err := c.call("turn/start", map[string]any{
		"threadId": c.threadID,
		"input": []any{
			map[string]any{"type": "text", "text": text},
		},
		"approvalPolicy": "never", // gateway context: auto-approve, like codebuddy -y
	})
	if err != nil {
		return "", err
	}
	var r struct {
		Turn struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	_ = json.Unmarshal(res.Result, &r)

	c.eventMu.Lock()
	c.curEvents = onEvent
	c.curTurnID = r.Turn.ID
	c.done = make(chan string, 1)
	c.eventMu.Unlock()
	defer func() {
		c.eventMu.Lock()
		c.curEvents = nil
		c.curTurnID = ""
		c.done = nil
		c.eventMu.Unlock()
	}()

	select {
	case status := <-c.done:
		if status == "" {
			status = "completed"
		}
		return status, nil
	case <-time.After(300 * time.Second):
		return "", fmt.Errorf("turn: timed out")
	}
}

// notifyEvent reduces a server notification to a transport.Event and hands it
// to the turn in flight. It runs on the reader goroutine, in order.
func (c *Client) notifyEvent(m wireMessage) {
	c.eventMu.Lock()
	turnID := c.curTurnID
	fn := c.curEvents
	done := c.done
	c.eventMu.Unlock()

	switch m.Method {
	case "item/agentMessage/delta":
		var p struct {
			TurnID string `json:"turnId"`
			Delta  string `json:"delta"`
		}
		_ = json.Unmarshal(m.Params, &p)
		if p.TurnID == turnID && fn != nil && p.Delta != "" {
			fn(transport.Event{Kind: transport.EventText, Text: p.Delta})
		}

	case "turn/completed":
		var p struct {
			Turn struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"turn"`
		}
		_ = json.Unmarshal(m.Params, &p)
		if p.Turn.ID == turnID && done != nil {
			done <- p.Turn.Status
		}

	case "item/started":
		var p struct {
			TurnID string `json:"turnId"`
			Item   struct {
				Type    string `json:"type"`
				Name    string `json:"name"`
				Command string `json:"command"`
			} `json:"item"`
		}
		_ = json.Unmarshal(m.Params, &p)
		if p.TurnID != turnID || fn == nil {
			return
		}
		switch p.Item.Type {
		case "mcpToolCall":
			fn(transport.Event{Kind: transport.EventTool, Tool: p.Item.Name})
		case "commandExecution":
			fn(transport.Event{Kind: transport.EventTool, Tool: p.Item.Command})
		}
	}
}

// answer replies to a request the server sent us. In practice app-server is
// mostly one-directional (client -> server), but a stray request must not hang.
func (c *Client) answer(m wireMessage) {
	_ = c.write(wireMessage{ID: m.ID, Error: &rpcError{Code: -32601, Message: "method not supported by 1bot_test: " + m.Method}})
}
