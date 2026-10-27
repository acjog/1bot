// Package codex drives OpenAI Codex CLI in `app-server` mode over stdio. It
// speaks the v2 app-server protocol (thread/turn/item notifications) and
// reduces the stream to transport.Event.
//
// Tool calls are handed back to the caller as transport.EventToolCall: an
// item/started notification whose item is a commandExecution or mcpToolCall
// becomes a ToolCall instead of a bare tool name. The turn then pauses; the
// tool result the next Turn carries is reported back with an item/update on
// the same turn, and Codex carries on. turn/start input takes user content
// only (text, image, audio, skill, mention) — tool results are not input
// items, they are item updates.
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
	// tool calls through curTool while a Turn waits in its select. done and
	// turnID outlive a single Turn so a tool result can resume the same turn.
	eventMu   sync.Mutex
	curEvents func(transport.Event)
	curTool   chan transport.ToolCall
	turnID    string              // the open turn, "" when none
	done      chan string         // receives the turn/completed status
	itemIDs   map[string]string   // tool call id -> item id, to answer results
}

func New(opts Options) *Client {
	return &Client{
		opts:    opts,
		pending: map[string]chan wireMessage{},
		itemIDs: map[string]string{},
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

// Turn implements transport.Transport. It answers any tool results the
// conversation carries, then resumes the open turn or starts a new one. A
// turn ends with StopToolUse when Codex asks for a tool (handed out as
// EventToolCall), or StopEndTurn when it finishes.
func (c *Client) Turn(messages []transport.Message, onEvent func(transport.Event)) (string, error) {
	c.promptMu.Lock()
	defer c.promptMu.Unlock()

	if c.threadID == "" {
		if err := c.startThread(); err != nil {
			return "", err
		}
	}

	c.eventMu.Lock()
	c.curEvents = onEvent
	c.curTool = make(chan transport.ToolCall, 1)
	turnOpen := c.turnID != ""
	c.eventMu.Unlock()
	defer func() {
		c.eventMu.Lock()
		c.curEvents = nil
		c.curTool = nil
		c.eventMu.Unlock()
	}()

	c.feedResults(messages)

	if turnOpen {
		return c.awaitTurn(onEvent)
	}

	input := c.buildInput(messages)
	res, err := c.call("turn/start", map[string]any{
		"threadId":       c.threadID,
		"input":          input,
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
	c.turnID = r.Turn.ID
	c.done = make(chan string, 1)
	c.eventMu.Unlock()

	return c.awaitTurn(onEvent)
}

// awaitTurn waits for the open turn to complete or to ask for a tool. On a
// tool call it reports the call and stops the turn at StopToolUse, leaving
// the turn open for the next Turn's result.
func (c *Client) awaitTurn(onEvent func(transport.Event)) (string, error) {
	c.eventMu.Lock()
	done := c.done
	tool := c.curTool
	c.eventMu.Unlock()

	select {
	case <-done:
		c.eventMu.Lock()
		c.turnID = ""
		c.done = nil
		c.eventMu.Unlock()
		return transport.StopEndTurn, nil
	case tc := <-tool:
		onEvent(transport.Event{Kind: transport.EventToolCall, Call: &tc})
		return transport.StopToolUse, nil
	case <-time.After(300 * time.Second):
		return "", fmt.Errorf("turn: timed out")
	}
}

// buildInput turns the conversation into app-server input items. Only user
// content (text) goes here; tool results are reported back as item updates
// instead, so they are not input items.
func (c *Client) buildInput(messages []transport.Message) []any {
	var input []any
	for _, msg := range messages {
		for _, p := range msg.Parts {
			if p.Kind == transport.PartText && strings.TrimSpace(p.Text) != "" {
				input = append(input, map[string]any{"type": "text", "text": p.Text})
			}
		}
	}
	return input
}

// feedResults reports the tool results the conversation carries back to the
// items they answer.
func (c *Client) feedResults(messages []transport.Message) {
	for _, msg := range messages {
		for _, p := range msg.Parts {
			if p.Kind == transport.PartToolResult && p.Result != nil {
				c.answerResult(p.Result)
			}
		}
	}
}

// answerResult reports a tool result back to Codex with an item/update.
func (c *Client) answerResult(r *transport.ToolResult) {
	c.eventMu.Lock()
	itemID, ok := c.itemIDs[r.CallID]
	if ok {
		delete(c.itemIDs, r.CallID)
	}
	c.eventMu.Unlock()
	if !ok {
		return
	}
	params := map[string]any{"itemId": itemID, "output": r.Content}
	if r.IsError {
		params["isError"] = true
	}
	if err := c.notify("item/update", params); err != nil {
		fmt.Fprintf(stderr, "answerResult: %v\n", err)
	}
}

// notifyEvent reduces a server notification to a transport.Event and hands it
// to the turn in flight. It runs on the reader goroutine, in order.
func (c *Client) notifyEvent(m wireMessage) {
	c.eventMu.Lock()
	turnID := c.turnID
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
		tc, itemID, ok := toolCallOf(p.Item)
		if !ok {
			return
		}
		c.eventMu.Lock()
		c.itemIDs[tc.ID] = itemID
		c.eventMu.Unlock()
		tool <- tc
	}
}

// toolCallOf maps an app-server item to a neutral ToolCall. commandExecution
// and mcpToolCall items are tool calls; the item's own JSON is kept as the
// arguments so no field is lost. The item id is returned so a later result
// can be reported back with item/update.
func toolCallOf(raw json.RawMessage) (transport.ToolCall, string, bool) {
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
		return transport.ToolCall{}, "", false
	}
	if name == "" {
		return transport.ToolCall{}, "", false
	}
	id := item.ID
	if id == "" {
		id = "call_" + name
	}
	return transport.ToolCall{ID: id, Name: name, Arguments: raw}, item.ID, true
}

// answer replies to a request the server sent us. In practice app-server is
// mostly one-directional (client -> server), but a stray request must not hang.
func (c *Client) answer(m wireMessage) {
	_ = c.write(wireMessage{ID: m.ID, Error: &rpcError{Code: -32601, Message: "method not supported by 1bot_test: " + m.Method}})
}
