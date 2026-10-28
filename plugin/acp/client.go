// Package acp drives CodeBuddy Code in ACP (Agent Client Protocol) mode over
// stdio. It speaks the JSON-RPC-over-NDJSON dialect the child expects, and
// reduces the child's many sessionUpdate variants to transport.Event.
//
// Tool execution is handed back to the caller instead of running on the
// gateway's machine: initialize declares the client's fs and terminal
// capabilities, so CodeBuddy turns a file read or a command into a reverse
// request (fs/read_text_file, fs/write_text_file, terminal/create). Those
// reverse requests become transport.EventToolCall, and the tool results the
// next Turn carries are answered back to the waiting CodeBuddy, which then
// carries on. The session stays alive across turns for exactly this reason.
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
	"time"

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

	// MCPServers is the session/new.mcpServers value handed to CodeBuddy.
	// nil is sent as an empty array. The plugin-codebuddy main process sets it
	// to a single http server that bridges client-declared function tools.
	MCPServers []any

	// NoFSTerminal, when true, omits the fs/terminal client capabilities from
	// the initialize handshake. The gateway itself never runs commands — the API
	// client executes tools through the MCP bridge — so those ACP capabilities
	// only make CodeBuddy offer read_text_file &c. that nothing serves.
	NoFSTerminal bool
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

	// Turn state, guarded by eventMu. The reader goroutine answers reverse
	// requests (tool calls) through handleTool while a Turn waits in await,
	// so both sides share this state through channels.
	eventMu   sync.Mutex
	curEvents func(transport.Event)
	curTool   chan transport.ToolCall    // non-nil while a turn waits for tool calls
	open      bool                       // a session/prompt is outstanding
	openCh    chan wireMessage           // its result channel, when open
	// toolCalls maps a tool call id to the answerer that receives its result
	// on a later Turn. An ACP reverse request answers the waiting CodeBuddy;
	// an MCP bridge tool call answers the blocked HTTP callback instead.
	toolCalls map[string]func(transport.ToolResult)
}

func New(opts Options) *Client {
	return &Client{
		opts:      opts,
		pending:   map[string]chan wireMessage{},
		toolCalls: map[string]func(transport.ToolResult){},
	}
}

// Start launches the child and performs the ACP initialize handshake. The
// client declares fs and terminal capabilities so CodeBuddy delegates those
// tools to the gateway instead of running them itself.
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

// initialize performs the ACP handshake, declaring the fs and terminal
// capabilities so the agent hands those tools back to the client.
func (c *Client) initialize() error {
	params := map[string]any{
		"protocolVersion": 1,
		"clientInfo":      map[string]any{"name": "1bot_test", "version": "0.1"},
	}
	if !c.opts.NoFSTerminal {
		params["clientCapabilities"] = map[string]any{
			"fs":       map[string]any{"readTextFile": true, "writeTextFile": true},
			"terminal": true,
		}
	}
	res, err := c.call("initialize", params)
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
	servers := c.opts.MCPServers
	if servers == nil {
		servers = []any{}
	}
	res, err := c.call("session/new", map[string]any{
		"cwd":        c.opts.Cwd,
		"mcpServers": servers,
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

// SetMCPServers replaces the mcpServers value the next session/new sends. Call
// it between turns (the main loop is the only Turn caller, so this is safe).
func (c *Client) SetMCPServers(servers []any) {
	c.promptMu.Lock()
	c.opts.MCPServers = servers
	c.promptMu.Unlock()
}

// ResetSession drops the current ACP session so the next Turn starts a fresh
// one carrying the MCP servers most recently set. Use it when the client's tool
// set changes; the trade-off is losing CodeBuddy's in-session context.
func (c *Client) ResetSession() {
	c.promptMu.Lock()
	c.sessionID = ""
	c.promptMu.Unlock()
}

// Turn implements transport.Transport. It first answers any tool results the
// conversation carries, then either resumes the outstanding session/prompt or
// starts a new one. A turn ends with StopToolUse when the agent asks for a
// tool (handed out as EventToolCall), or StopEndTurn when it finishes.
func (c *Client) Turn(messages []transport.Message, onEvent func(transport.Event)) (string, error) {
	c.promptMu.Lock()
	defer c.promptMu.Unlock()

	if c.sessionID == "" {
		if err := c.newSession(); err != nil {
			return "", err
		}
	}

	c.feedResults(messages)

	c.eventMu.Lock()
	c.curEvents = onEvent
	c.curTool = make(chan transport.ToolCall, 1)
	open := c.open
	openCh := c.openCh
	c.eventMu.Unlock()
	defer func() {
		c.eventMu.Lock()
		c.curEvents = nil
		c.curTool = nil
		c.eventMu.Unlock()
	}()

	if open {
		return c.await(openCh, onEvent)
	}

	prompt := userText(messages)
	id, ch := c.register()
	req := wireMessage{
		JSONRPC: "2.0",
		ID:      json.RawMessage(id),
		Method:  "session/prompt",
		Params:  mustJSON(map[string]any{"sessionId": c.sessionID, "prompt": []any{map[string]any{"type": "text", "text": prompt}}}),
	}
	c.eventMu.Lock()
	c.open = true
	c.openCh = ch
	c.eventMu.Unlock()
	if err := c.write(req); err != nil {
		c.eventMu.Lock()
		c.open = false
		c.eventMu.Unlock()
		return "", err
	}
	return c.await(ch, onEvent)
}

// await waits for the outstanding session/prompt to resolve, or for the agent
// to ask for a tool. On a tool call it reports the call and stops the turn at
// StopToolUse, leaving the session/prompt outstanding for the next Turn.
func (c *Client) await(ch chan wireMessage, onEvent func(transport.Event)) (string, error) {
	c.eventMu.Lock()
	tool := c.curTool
	c.eventMu.Unlock()

	select {
	case m := <-ch:
		c.eventMu.Lock()
		c.open = false
		c.eventMu.Unlock()
		if m.Error != nil {
			return "", fmt.Errorf("session/prompt: rpc error %d: %s", m.Error.Code, m.Error.Message)
		}
		var r struct {
			StopReason string `json:"stopReason"`
		}
		_ = json.Unmarshal(m.Result, &r)
		if r.StopReason == "" {
			r.StopReason = transport.StopEndTurn
		}
		return r.StopReason, nil
	case tc := <-tool:
		onEvent(transport.Event{Kind: transport.EventToolCall, Call: &tc})
		return transport.StopToolUse, nil
	case <-time.After(300 * time.Second):
		return "", fmt.Errorf("turn: timed out")
	}
}

// feedResults answers the tool results the conversation carries. Each
// PartToolResult is mapped back to the reverse request it answers and sent to
// the waiting agent.
func (c *Client) feedResults(messages []transport.Message) {
	for _, msg := range messages {
		for _, p := range msg.Parts {
			if p.Kind == transport.PartToolResult && p.Result != nil {
				c.answerResult(p.Result)
			}
		}
	}
}

// answerResult hands a tool result to the answerer registered for its call.
func (c *Client) answerResult(r *transport.ToolResult) {
	c.eventMu.Lock()
	answer, ok := c.toolCalls[r.CallID]
	if ok {
		delete(c.toolCalls, r.CallID)
	}
	c.eventMu.Unlock()
	if !ok {
		return
	}
	answer(*r)
}

// userText flattens a conversation's text parts into the single user prompt
// the transport sends for a new turn. Tool calls and results travel through
// feedResults instead.
func userText(messages []transport.Message) string {
	var b strings.Builder
	for _, msg := range messages {
		for _, p := range msg.Parts {
			if p.Kind == transport.PartText && strings.TrimSpace(p.Text) != "" {
				b.WriteString(p.Text)
				b.WriteString("\n\n")
			}
		}
	}
	return strings.TrimSpace(b.String())
}

// register reserves an id and returns it with its result channel.
func (c *Client) register() (string, chan wireMessage) {
	c.pendMu.Lock()
	c.nextID++
	id := fmt.Sprintf("%d", c.nextID)
	ch := make(chan wireMessage, 1)
	c.pending[id] = ch
	c.pendMu.Unlock()
	return id, ch
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
		// The tool's arguments arrive in its own reverse request
		// (fs/read_text_file &c), which handleTool turns into a full
		// EventToolCall. This notification only names it; report nothing here
		// to avoid a duplicate, empty call.
		if p.Update.ToolCall.Name != "" && c.opts.Verbose {
			fmt.Fprintf(stderr, "[tool] %s\n", p.Update.ToolCall.Name)
		}
		return
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

// answer replies to a request the agent sent us. Tool requests (the fs and
// terminal capabilities we declared) become tool calls handed to the caller;
// the result is answered on a later Turn. Everything else is answered as
// before: permissions auto-allowed, unknown methods refused.
func (c *Client) answer(m wireMessage) {
	if isToolMethod(m.Method) {
		c.handleTool(m)
		return
	}
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

// handleTool turns a reverse tool request into a transport.ToolCall and hands
// it to the turn in flight. The reverse request stays unanswered until the
// next Turn carries its result.
func (c *Client) handleTool(m wireMessage) {
	tc := transport.ToolCall{
		ID:        string(m.ID),
		Name:      toolName(m.Method),
		Arguments: m.Params,
	}
	reqID := m.ID
	c.dispatchToolCall(tc, func(r transport.ToolResult) {
		result := map[string]any{
			"content": []any{map[string]any{"type": "text", "text": r.Content}},
		}
		if r.IsError {
			result["isError"] = true
		}
		if err := c.respond(reqID, result); err != nil {
			fmt.Fprintf(stderr, "handleTool: %v\n", err)
		}
	})
}

// ToolCall injects a tool call that did not arrive through ACP — an MCP bridge
// tools/call — into the turn in flight. answer is called with the client's
// result on a later Turn, exactly like an ACP reverse tool request.
func (c *Client) ToolCall(tc transport.ToolCall, answer func(transport.ToolResult)) {
	c.dispatchToolCall(tc, answer)
}

// dispatchToolCall registers the call's answerer and hands the call to the
// turn waiting in await, which reports it as EventToolCall and returns
// StopToolUse. The call stays registered until its result arrives.
func (c *Client) dispatchToolCall(tc transport.ToolCall, answer func(transport.ToolResult)) {
	c.eventMu.Lock()
	c.toolCalls[tc.ID] = answer
	tool := c.curTool
	c.eventMu.Unlock()

	if tool != nil {
		tool <- tc
	}
}

// isToolMethod reports whether a reverse request is one of the fs/terminal
// tools the client declared, i.e. a tool call to hand to the caller.
func isToolMethod(method string) bool {
	switch method {
	case "fs/read_text_file", "fs/write_text_file", "terminal/create":
		return true
	}
	return false
}

// toolName maps an ACP tool method to the neutral name the OpenAI front hands
// back to the client.
func toolName(method string) string {
	switch method {
	case "fs/read_text_file":
		return "read_text_file"
	case "fs/write_text_file":
		return "write_text_file"
	case "terminal/create":
		return "terminal"
	}
	return method
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
