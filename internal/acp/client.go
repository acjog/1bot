// Package acp drives CodeBuddy Code in ACP (Agent Client Protocol) mode over
// stdio. It speaks the JSON-RPC-over-NDJSON dialect the child expects, and
// reduces the child's many sessionUpdate variants to transport.Event.
package acp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"1bot_test/internal/transport"
)

var stderr = os.Stderr

// Options configures a CodeBuddy ACP child process.
type Options struct {
	Bin     string // codebuddy executable (name or resolved path)
	Model   string // official model id; empty = settings.json's model
	SkipY   bool   // pass -y (skip permission prompts)
	Verbose bool   // dump every JSON-RPC frame to stderr
	Cwd     string // working directory handed to the agent
}

// Client drives one long-lived `codebuddy --acp` process. It implements
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

	sessionID string

	// promptMu serialises turns: one CodeBuddy process runs one turn at a time.
	promptMu sync.Mutex
	// eventMu guards curEvents, the callback for the turn in flight. It is held
	// for microseconds only, never across network I/O, so the reader goroutine
	// never blocks on it.
	eventMu   sync.Mutex
	curEvents func(transport.Event)
}

func New(opts Options) *Client {
	return &Client{
		opts:    opts,
		pending: map[string]chan wireMessage{},
	}
}

// Start launches the child and performs the ACP initialize handshake.
func (c *Client) Start() error {
	args := []string{"--acp"}
	if c.opts.Model != "" {
		args = append(args, "--model", c.opts.Model)
	}
	if c.opts.SkipY {
		args = append(args, "-y")
	}
	c.cmd = exec.Command(c.opts.Bin, args...)
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
		return fmt.Errorf("starting %s --acp: %w", c.opts.Bin, err)
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
			fmt.Fprintf(stderr, "bad JSON: %v: %s\n", err, line)
			continue
		}
		if c.opts.Verbose {
			fmt.Fprintf(stderr, "[<-] %s\n", line)
		}

		switch {
		case m.hasID() && m.Method == "":
			c.deliver(m) // a response to one of our requests
		case m.hasID():
			c.answer(m) // a request the agent sent us
		default:
			c.notify(m) // a notification (streaming event)
		}
	}
}

// initialize performs the ACP handshake.
func (c *Client) initialize() error {
	res, err := c.call("initialize", map[string]any{
		"protocolVersion": 1,
		"clientCapabilities": map[string]any{
			"fs":       map[string]any{"readTextFile": false, "writeTextFile": false},
			"terminal": false,
		},
		"clientInfo": map[string]any{"name": "1bot_test", "version": "0.1"},
	})
	if err != nil {
		return err
	}
	var r struct {
		ProtocolVersion int             `json:"protocolVersion"`
		AgentInfo       json.RawMessage `json:"agentInfo"`
	}
	_ = json.Unmarshal(res.Result, &r)
	fmt.Fprintf(stderr, "connected: protocol v%d %s\n", r.ProtocolVersion, string(r.AgentInfo))
	return nil
}

// newSession makes a fresh session. Caller must hold promptMu.
func (c *Client) newSession() error {
	res, err := c.call("session/new", map[string]any{
		"cwd":        c.opts.Cwd,
		"mcpServers": []any{},
	})
	if err != nil {
		return err
	}
	var r struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.Unmarshal(res.Result, &r); err != nil {
		return fmt.Errorf("session/new result: %w", err)
	}
	c.sessionID = r.SessionID
	return nil
}

// Turn implements transport.Transport. It opens a fresh session, sends text
// as the user prompt, and streams the reply through onEvent (called on the
// reader goroutine, in order). Turns are serialised: one process, one turn.
func (c *Client) Turn(text string, onEvent func(transport.Event)) (string, error) {
	c.promptMu.Lock()
	defer c.promptMu.Unlock()

	if err := c.newSession(); err != nil {
		return "", err
	}

	c.eventMu.Lock()
	c.curEvents = onEvent
	c.eventMu.Unlock()
	defer func() {
		c.eventMu.Lock()
		c.curEvents = nil
		c.eventMu.Unlock()
	}()

	res, err := c.call("session/prompt", map[string]any{
		"sessionId": c.sessionID,
		"prompt": []any{
			map[string]any{"type": "text", "text": text},
		},
	})
	if err != nil {
		return "", err
	}
	var r struct {
		StopReason string `json:"stopReason"`
	}
	_ = json.Unmarshal(res.Result, &r)
	return r.StopReason, nil
}

// notify reduces a session/update notification to a transport.Event and hands
// it to the turn in flight. The callback runs on this (reader) goroutine,
// which is fine: events must be handled in order.
func (c *Client) notify(m wireMessage) {
	var p struct {
		Update struct {
			SessionUpdate string `json:"sessionUpdate"`
			Content       struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			ToolCall struct {
				Name string `json:"name"`
			} `json:"toolCall"`
		} `json:"update"`
	}
	_ = json.Unmarshal(m.Params, &p)

	var ev transport.Event
	switch p.Update.SessionUpdate {
	case "agent_message_chunk":
		if p.Update.Content.Type != "text" {
			return
		}
		ev = transport.Event{Kind: transport.EventText, Text: p.Update.Content.Text}
	case "agent_thought_chunk":
		if p.Update.Content.Type != "text" {
			return
		}
		ev = transport.Event{Kind: transport.EventThought, Text: p.Update.Content.Text}
	case "tool_call":
		if p.Update.ToolCall.Name == "" {
			return
		}
		ev = transport.Event{Kind: transport.EventTool, Tool: p.Update.ToolCall.Name}
	default:
		return
	}

	c.eventMu.Lock()
	fn := c.curEvents
	c.eventMu.Unlock()
	if fn != nil {
		fn(ev)
	}
}

// answer replies to a request the agent sent us. We declare no fs/terminal
// capabilities, so CodeBuddy uses its own tools; in practice only
// session/request_permission arrives, which we auto-allow. Anything else gets
// a method-not-found error so the agent never blocks on us.
func (c *Client) answer(m wireMessage) {
	if m.Method == "session/request_permission" {
		var p struct {
			Options []struct {
				OptionID string `json:"optionId"`
				Kind     string `json:"kind"`
			} `json:"options"`
			ToolCall struct {
				Name string `json:"name"`
			} `json:"toolCall"`
		}
		_ = json.Unmarshal(m.Params, &p)

		opt := ""
		for _, o := range p.Options {
			if opt == "" {
				opt = o.OptionID
			}
			if o.Kind == "allow_once" {
				opt = o.OptionID
				break
			}
		}
		if p.ToolCall.Name != "" {
			fmt.Fprintf(stderr, "[permission] %s -> allow_once\n", p.ToolCall.Name)
		}
		_ = c.respond(m.ID, map[string]any{
			"outcome": map[string]any{"outcome": "selected", "optionId": opt},
		})
		return
	}

	_ = c.respondErr(m.ID, -32601, "method not supported by 1bot_test: "+m.Method)
}
