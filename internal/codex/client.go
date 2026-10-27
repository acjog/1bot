// Package codex drives OpenAI Codex CLI in `app-server` mode over stdio. It
// speaks the v2 app-server protocol (thread/turn/item notifications) and
// reduces the stream to transport.Event.
//
// Tool calls are handed back to the caller as transport.EventToolCall: an
// item/started notification whose item is a commandExecution or mcpToolCall
// becomes a ToolCall instead of a bare tool name, so the OpenAI front can
// surface it as a tool_calls delta and the client can run it. The tool
// results the next Turn carries are sent back as input items on the same
// thread, which stays alive across turns.
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

	// Turn state, guarded by eventMu. The reader goroutine feeds item/started
	// tool calls through curTool while a Turn waits in its select.
	eventMu   sync.Mutex
	curEvents func(transport.Event)
	curTool   chan transport.ToolCall
	curTurnID string
	done      chan string
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

// Turn implements transport.Transport. It sends the conversation's text and
// tool results as one turn's input on the current thread, then streams the
// reply until turn/completed (StopEndTurn) or a tool call (StopToolUse).
func (c *Client) Turn(messages []transport.Message, onEvent func(transport.Event)) (string, error) {
	c.promptMu.Lock()
	defer c.promptMu.Unlock()

	if c.threadID == "" {
		if err := c.startThread(); err != nil {
			return "", err
		}
	}

	input := c.buildInput(messages)

	res, err := c.call("turn/start", map[string]any{
		"threadId": c.threadID,
		"input":    input,
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
	c.curTool = make(chan transport.ToolCall, 1)
	c.eventMu.Unlock()
	defer func() {
		c.eventMu.Lock()
		c.curEvents = nil
		c.curTool = nil
		c.curTurnID = ""
		c.done = nil
		c.eventMu.Unlock()
	}()

	c.eventMu.Lock()
	tool := c.curTool
	done := c.done
	c.eventMu.Unlock()

	select {
	case <-done:
		return transport.StopEndTurn, nil
	case tc := <-tool:
		onEvent(transport.Event{Kind: transport.EventToolCall, Call: &tc})
		return transport.StopToolUse, nil
	case <-time.After(300 * time.Second):
		return "", fmt.Errorf("turn: timed out")
	}
}

// buildInput turns the conversation into app-server input items: text parts
// become text items, tool results become tool-result items so the thread can
// carry on from the tools the caller ran.
func (c *Client) buildInput(messages []transport.Message) []any {
	var input []any
	for _, msg := range messages {
		for _, p := range msg.Parts {
			switch p.Kind {
			case transport.PartText:
				input = append(input, map[string]any{"type": "text", "text": p.Text})
			case transport.PartToolResult:
				if p.Result == nil {
					continue
				}
				input = append(input, map[string]any{
					"type":       "tool_result",
					"toolCallId": p.Result.CallID,
					"content":    p.Result.Content,
				})
			}
		}
	}
	return input
}

// notifyEvent reduces a server notification to a transport.Event and hands it
// to the turn in flight. It runs on the reader goroutine, in order.
func (c *Client) notifyEvent(m wireMessage) {
	c.eventMu.Lock()
	turnID := c.curTurnID
	fn := c.curEvents
	tool := c.curTool
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
			TurnID string          `json:"turnId"`
			Item   json.RawMessage `json:"item"`
		}
		_ = json.Unmarshal(m.Params, &p)
		if p.TurnID != turnID || tool == nil {
			return
		}
		if tc, ok := toolCallOf(p.Item); ok {
			tool <- tc
		}
	}
}

// toolCallOf maps an app-server item to a neutral ToolCall. commandExecution
// and mcpToolCall items are tool calls; the item's own JSON is kept as the
// arguments so no field is lost.
func toolCallOf(raw json.RawMessage) (transport.ToolCall, bool) {
	var item struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Name    string `json:"name"`
		Command string `json:"command"`
	}
	_ = json.Unmarshal(raw, &item)

	var name string
	switch item.Type {
	case "commandExecution":
		name = "bash"
	case "mcpToolCall":
		name = item.Name
	default:
		return transport.ToolCall{}, false
	}
	if name == "" {
		return transport.ToolCall{}, false
	}
	id := item.ID
	if id == "" {
		id = "call_" + name
	}
	return transport.ToolCall{ID: id, Name: name, Arguments: raw}, true
}

// answer replies to a request the server sent us. In practice app-server is
// mostly one-directional (client -> server), but a stray request must not hang.
func (c *Client) answer(m wireMessage) {
	_ = c.write(wireMessage{ID: m.ID, Error: &rpcError{Code: -32601, Message: "method not supported by 1bot_test: " + m.Method}})
}
