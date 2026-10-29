// Package codex adapts app-server to tools owned and executed by API clients.
package codex

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"time"

	"1bot_test/internal/transport"
	"1bot_test/internal/vlog"
)

var stderr = os.Stderr

type Options struct {
	Bin, Model, Cwd string
	Verbose         bool
}
type session struct {
	threadID, turnID, model string
	sessionKey              string
	events                  chan wireMessage
	history                 []transport.Message
	tools                   []transport.Tool
	call                    *transport.ToolCall
	requestID               json.RawMessage
	updated                 time.Time
}

type Client struct {
	opts                               Options
	cmd                                *exec.Cmd
	stdin                              io.WriteCloser
	scanner                            *bufio.Scanner
	writeMu, pendMu, promptMu, eventMu sync.Mutex
	pending                            map[string]chan wireMessage
	nextID                             int64
	closed                             chan struct{}
	sessions                           map[string]*session // threadID -> session
	toolSessions                       map[string]*session // callID -> session (pending tool result)
	bySession                          map[string]*session // sessionKey -> session (reused across turns)
	serverOverrides                    map[string]any
}

func New(opts Options) *Client {
	return &Client{opts: opts, pending: map[string]chan wireMessage{}, closed: make(chan struct{}),
		sessions: map[string]*session{}, toolSessions: map[string]*session{}, bySession: map[string]*session{}}
}
func (c *Client) Start() error {
	c.cmd = exec.Command(c.opts.Bin, "app-server")
	c.cmd.Dir, c.cmd.Stderr = c.opts.Cwd, stderr
	in, err := c.cmd.StdinPipe()
	if err != nil {
		return err
	}
	out, err := c.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err = c.cmd.Start(); err != nil {
		return fmt.Errorf("starting app-server: %w", err)
	}
	c.stdin = in
	c.scanner = bufio.NewScanner(out)
	c.scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	go c.readLoop()
	if _, err = c.call("initialize", map[string]any{
		"clientInfo":   map[string]any{"name": "1bot_gateway", "version": "0.2"},
		"capabilities": map[string]any{"experimentalApi": true},
	}); err != nil {
		_ = c.Close()
		return err
	}
	if err = c.notify("initialized", map[string]any{}); err != nil {
		_ = c.Close()
		return err
	}
	// Explicitly disable configured server MCPs as well as the automatic app
	// connector. Empty configuration tables alone would merge with host tables.
	res, err := c.call("config/read", map[string]any{"cwd": c.opts.Cwd, "includeLayers": false})
	if err != nil {
		_ = c.Close()
		return err
	}
	var config struct {
		Config map[string]json.RawMessage `json:"config"`
	}
	if err = json.Unmarshal(res.Result, &config); err != nil {
		_ = c.Close()
		return err
	}
	c.serverOverrides = map[string]any{}
	for _, key := range []string{"mcp_servers", "plugins"} {
		var entries map[string]json.RawMessage
		_ = json.Unmarshal(config.Config[key], &entries)
		disabled := map[string]any{}
		for name := range entries {
			disabled[name] = map[string]any{"enabled": false}
		}
		if len(disabled) > 0 {
			c.serverOverrides[key] = disabled
		}
	}
	return nil
}
func (c *Client) Close() error {
	c.writeMu.Lock()
	if c.stdin != nil {
		_ = c.stdin.Close()
	}
	c.writeMu.Unlock()
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		_ = c.cmd.Wait()
	}
	return nil
}
func (c *Client) readLoop() {
	defer close(c.closed)
	for c.scanner.Scan() {
		line := c.scanner.Bytes()
		var m wireMessage
		if json.Unmarshal(line, &m) != nil {
			continue
		}
		if c.opts.Verbose {
			vlog.Frame(stderr, "codex <-", string(line))
		}
		if m.hasID() && m.Method == "" {
			c.deliver(m)
			continue
		}
		var p struct {
			ThreadID string `json:"threadId"`
		}
		_ = json.Unmarshal(m.Params, &p)
		c.eventMu.Lock()
		s := c.sessions[p.ThreadID]
		c.eventMu.Unlock()
		if s != nil && (m.Method == "item/agentMessage/delta" || m.Method == "turn/completed" || m.Method == "item/tool/call") {
			s.events <- m
		} else if m.hasID() {
			_ = c.write(wireMessage{ID: m.ID, Error: &rpcError{Code: -32601, Message: "server-owned tools are disabled"}})
		}
	}
}

const gatewayInstructions = `You are a model behind an API gateway. The API client owns the conversation and every tool.
Use only supplied client tools. Never execute commands, read files, browse, or inspect the gateway host.
The gateway host operating system, paths, shell, and working directory are not the client's environment.
Use environment information supplied by the client; if missing, ask instead of guessing.
Return tool calls for the client to execute and use its returned results. Do not claim to have executed them yourself.`

func (c *Client) startSession(req transport.Request) (*session, error) {
	tools := make([]any, 0, len(req.Tools))
	for _, t := range req.Tools {
		tools = append(tools, map[string]any{"type": "function", "name": t.Name, "description": t.Description, "inputSchema": t.Parameters})
	}
	base, developer := gatewayInstructions, ""
	for _, m := range req.Messages {
		for _, p := range m.Parts {
			if p.Kind != transport.PartText {
				continue
			}
			if m.Role == "system" {
				base += "\n\n" + p.Text
			}
			if m.Role == "developer" {
				developer += "\n\n" + p.Text
			}
		}
	}
	params := map[string]any{
		"ephemeral": true, "environments": []any{}, "dynamicTools": tools,
		"baseInstructions": base, "developerInstructions": strings.TrimSpace(developer),
		"approvalPolicy": "never", "allowProviderModelFallback": false,
		"config": map[string]any{
			// Tool execution is delegated to the API client through
			// dynamicTools, so Codex's own sandbox never runs the write. The
			// mode is still raised to workspace-write so Codex's built-in
			// "read-only sandbox" prompt does not make it refuse to plan
			// writes as "no write permission".
			"sandbox_mode":       "workspace-write",
			"features.shell_tool": false, "features.unified_exec": false,
			"features.apps": false, "features.plugins": false, "features.multi_agent": false,
			"features.browser_use": false, "features.computer_use": false, "features.image_generation": false,
			"features.skill_search": false, "features.skip_host_skill_discovery": true,
			"features.view_image": false, "features.sleep_tool": false, "features.goals": false,
			"features.hooks": false, "features.memories": false, "features.code_mode": false,
			"web_search": "disabled", "agents.enabled": false, "project_doc_max_bytes": 0,
		},
	}
	for key, value := range c.serverOverrides {
		params["config"].(map[string]any)[key] = value
	}
	if req.Model != "" {
		params["model"] = req.Model
	}
	res, err := c.call("thread/start", params)
	if err != nil {
		return nil, err
	}
	var r struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if err = json.Unmarshal(res.Result, &r); err != nil {
		return nil, err
	}
	if r.Thread.ID == "" {
		return nil, fmt.Errorf("thread/start returned no thread ID")
	}
	s := &session{threadID: r.Thread.ID, events: make(chan wireMessage, 256), history: req.Messages, tools: req.Tools, model: req.Model, updated: time.Now()}
	c.eventMu.Lock()
	c.sessions[s.threadID] = s
	c.eventMu.Unlock()
	return s, nil
}
func (c *Client) Turn(messages []transport.Message, onEvent func(transport.Event)) (string, error) {
	return c.Complete(transport.Request{Messages: messages}, onEvent)
}

// Complete starts isolated threads, or resumes the thread addressed by a tool result.
func (c *Client) Complete(req transport.Request, onEvent func(transport.Event)) (string, error) {
	c.promptMu.Lock()
	defer c.promptMu.Unlock()
	if req.Model == "" {
		req.Model = c.opts.Model
	}
	if len(req.ToolChoice) > 0 && string(req.ToolChoice) != "null" && string(req.ToolChoice) != `"auto"` && string(req.ToolChoice) != `"none"` {
		return "", fmt.Errorf("tool_choice supports auto or none in the Codex adapter")
	}
	if string(req.ToolChoice) == `"none"` {
		req.Tools = nil
	}
	if len(req.Messages) == 0 {
		return "", fmt.Errorf("no messages")
	}
	c.reapSessions()
	last := req.Messages[len(req.Messages)-1]
	var result *transport.ToolResult
	for _, p := range last.Parts {
		if p.Kind == transport.PartToolResult {
			result = p.Result
		}
	}
	var s *session
	if result != nil {
		c.eventMu.Lock()
		s = c.toolSessions[result.CallID]
		c.eventMu.Unlock()
		if s == nil {
			return "", fmt.Errorf("unknown or expired tool_call_id %q", result.CallID)
		}
		if err := validateContinuation(s, req, result.CallID); err != nil {
			return "", err
		}
		payload, _ := json.Marshal(map[string]any{"contentItems": []any{map[string]any{"type": "inputText", "text": result.Content}}, "success": !result.IsError})
		if err := c.write(wireMessage{ID: s.requestID, Result: payload}); err != nil {
			c.retire(s)
			return "", err
		}
		c.eventMu.Lock()
		delete(c.toolSessions, result.CallID)
		c.eventMu.Unlock()
		s.call = nil
		s.history = req.Messages
		s.updated = time.Now()
	} else {
		if last.Role != "user" {
			return "", fmt.Errorf("new requests must end with a user message; tool results require a live tool_call_id")
		}
		var err error
		// Reuse the backend thread for the same conversation key, so tools and
		// context injected on an earlier turn survive. A model or tool change
		// invalidates the reuse.
		if req.Session != "" {
			c.eventMu.Lock()
			s = c.bySession[req.Session]
			c.eventMu.Unlock()
			if s != nil && (s.model != req.Model || !semanticEqual(s.tools, req.Tools)) {
				c.retire(s)
				s = nil
			}
		}
		if s == nil {
			c.eventMu.Lock()
			count := len(c.sessions)
			c.eventMu.Unlock()
			if count >= 128 {
				return "", fmt.Errorf("too many pending tool conversations")
			}
			s, err = c.startSession(req)
			if err != nil {
				return "", err
			}
			if req.Session != "" {
				s.sessionKey = req.Session
				c.eventMu.Lock()
				c.bySession[req.Session] = s
				c.eventMu.Unlock()
			}
			if items := historyItems(req.Messages[:len(req.Messages)-1]); len(items) > 0 {
				if _, err = c.call("thread/inject_items", map[string]any{"threadId": s.threadID, "items": items}); err != nil {
					c.retire(s)
					return "", err
				}
			}
		}
		// Refresh the history snapshot before starting a new turn on a reused
		// thread, so a later tool result is validated against this turn's
		// conversation instead of a stale snapshot.
		s.history = req.Messages
		s.updated = time.Now()
		input := []any{}
		for _, p := range last.Parts {
			if p.Kind == transport.PartText {
				input = append(input, map[string]any{"type": "text", "text": p.Text})
			}
		}
		res, err := c.call("turn/start", map[string]any{"threadId": s.threadID, "input": input, "environments": []any{}})
		if err != nil {
			c.retire(s)
			return "", err
		}
		var r struct {
			Turn struct {
				ID string `json:"id"`
			} `json:"turn"`
		}
		_ = json.Unmarshal(res.Result, &r)
		s.turnID = r.Turn.ID
	}
	stop, err := c.await(s, onEvent)
	if err != nil {
		c.retire(s)
	}
	return stop, err
}
func validateContinuation(s *session, req transport.Request, callID string) error {
	n := len(s.history)
	if len(req.Messages) != n+2 {
		fmt.Fprintf(stderr, "[validate] message count mismatch: got %d, want %d (history %d + assistant + tool result)\n",
			len(req.Messages), n+2, n)
		dumpRecent("saved history", s.history, 10)
		dumpRecent("request", req.Messages, 10)
		return fmt.Errorf("tool result conversation, model, or tools do not match the original request")
	}
	for i := 0; i < n; i++ {
		if !semanticEqual(req.Messages[i], s.history[i]) {
			sb, _ := json.Marshal(s.history[i])
			rb, _ := json.Marshal(req.Messages[i])
			fmt.Fprintf(stderr, "[validate] history mismatch at index %d:\n  saved:    %s\n  incoming: %s\n", i, sb, rb)
			return fmt.Errorf("tool result conversation, model, or tools do not match the original request")
		}
	}
	if req.Model != s.model {
		fmt.Fprintf(stderr, "[validate] model mismatch: incoming %q, saved %q\n", req.Model, s.model)
		return fmt.Errorf("tool result conversation, model, or tools do not match the original request")
	}
	if !semanticEqual(req.Tools, s.tools) {
		sb, _ := json.Marshal(s.tools)
		rb, _ := json.Marshal(req.Tools)
		fmt.Fprintf(stderr, "[validate] tools mismatch:\n  saved:    %s\n  incoming: %s\n", sb, rb)
		return fmt.Errorf("tool result conversation, model, or tools do not match the original request")
	}
	assistant := req.Messages[n]
	if assistant.Role != "assistant" {
		fmt.Fprintf(stderr, "[validate] messages[%d] role=%q, want assistant\n", n, assistant.Role)
		return fmt.Errorf("tool result must follow its assistant tool call")
	}
	for _, p := range assistant.Parts {
		if p.Call != nil && p.Call.ID == callID && p.Call.Name == s.call.Name  {
			return nil
		}
	}
	sb, _ := json.Marshal(s.call)
	fmt.Fprintf(stderr, "[validate] assistant tool call does not match pending call:\n  callID=%q\n  pending=%s\n  assistant parts=%d\n", callID, sb, len(assistant.Parts))
	return fmt.Errorf("tool result does not match the pending assistant tool call")
}
// dumpRecent prints the last k messages of a slice, one line each, so a
// continuation mismatch can be diagnosed without dumping every full body.
func dumpRecent(label string, msgs []transport.Message, k int) {
	start := len(msgs) - k
	if start < 0 {
		start = 0
	}
	fmt.Fprintf(stderr, "[validate] %s (last %d of %d):\n", label, len(msgs)-start, len(msgs))
	for i := start; i < len(msgs); i++ {
		fmt.Fprintf(stderr, "  [%d] role=%s", i, msgs[i].Role)
		for _, p := range msgs[i].Parts {
			switch p.Kind {
			case transport.PartText:
				fmt.Fprintf(stderr, " text=%q", truncate(p.Text, 80))
			case transport.PartToolCall:
				if p.Call != nil {
					fmt.Fprintf(stderr, " tool_call(id=%s name=%s)", p.Call.ID, p.Call.Name)
				}
			case transport.PartToolResult:
				if p.Result != nil {
					fmt.Fprintf(stderr, " tool_result(callId=%s)", p.Result.CallID)
				}
			}
		}
		fmt.Fprintln(stderr)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func jsonEqual(a, b json.RawMessage) bool {
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}

func semanticEqual(a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return jsonEqual(x, y)
}

// historyItems preserves roles and native function call/output items.
func historyItems(messages []transport.Message) []any {
	var items []any
	for _, m := range messages {
		// Codex may normalize injected instruction messages away. Supply them
		// through baseInstructions/developerInstructions at thread creation.
		if m.Role == "system" || m.Role == "developer" {
			continue
		}
		for _, p := range m.Parts {
			switch p.Kind {
			case transport.PartText:
				kind := "input_text"
				if m.Role == "assistant" {
					kind = "output_text"
				}
				items = append(items, map[string]any{"type": "message", "role": m.Role, "content": []any{map[string]any{"type": kind, "text": p.Text}}})
			case transport.PartToolCall:
				if p.Call != nil {
					items = append(items, map[string]any{"type": "function_call", "call_id": p.Call.ID, "name": p.Call.Name, "arguments": string(p.Call.Arguments)})
				}
			case transport.PartToolResult:
				if p.Result != nil {
					items = append(items, map[string]any{"type": "function_call_output", "call_id": p.Result.CallID, "output": p.Result.Content})
				}
			}
		}
	}
	return items
}
func (c *Client) await(s *session, onEvent func(transport.Event)) (string, error) {
	timer := time.NewTimer(300 * time.Second)
	defer timer.Stop()
	for {
		select {
		case m := <-s.events:
			switch m.Method {
			case "item/agentMessage/delta":
				var p struct {
					Delta  string `json:"delta"`
					TurnID string `json:"turnId"`
				}
				_ = json.Unmarshal(m.Params, &p)
				if p.TurnID == s.turnID && p.Delta != "" {
					onEvent(transport.Event{Kind: transport.EventText, Text: p.Delta})
				}
			case "item/tool/call":
				var p struct {
					Tool      string          `json:"tool"`
					Arguments json.RawMessage `json:"arguments"`
					TurnID    string          `json:"turnId"`
				}
				_ = json.Unmarshal(m.Params, &p)
				allowed := false
				for _, t := range s.tools {
					if t.Name == p.Tool {
						allowed = true
					}
				}
				if !allowed || p.TurnID != s.turnID {
					return "", fmt.Errorf("upstream requested an undeclared client tool %q", p.Tool)
				}
				var token [16]byte
				if _, err := rand.Read(token[:]); err != nil {
					return "", err
				}
				tc := &transport.ToolCall{ID: "call_" + hex.EncodeToString(token[:]), Name: p.Tool, Arguments: p.Arguments}
				s.call, s.requestID, s.updated = tc, m.ID, time.Now()
				c.eventMu.Lock()
				c.toolSessions[tc.ID] = s
				c.eventMu.Unlock()
				onEvent(transport.Event{Kind: transport.EventToolCall, Call: tc})
				return transport.StopToolUse, nil
			case "turn/completed":
				var p struct {
					Turn struct {
						ID     string `json:"id"`
						Status string `json:"status"`
						Error  *struct {
							Message string `json:"message"`
						} `json:"error"`
					} `json:"turn"`
				}
				_ = json.Unmarshal(m.Params, &p)
				if p.Turn.ID != s.turnID {
					continue
				}
				if p.Turn.Status != "completed" {
					message := ""
					if p.Turn.Error != nil {
						message = p.Turn.Error.Message
					}
					return "", fmt.Errorf("Codex turn %s: %s", p.Turn.Status, message)
				}
				return transport.StopEndTurn, nil
			}
		case <-c.closed:
			return "", fmt.Errorf("Codex app-server disconnected")
		case <-timer.C:
			return "", fmt.Errorf("Codex turn timed out")
		}
	}
}
func (c *Client) retire(s *session) {
	c.eventMu.Lock()
	delete(c.sessions, s.threadID)
	if s.sessionKey != "" && c.bySession[s.sessionKey] == s {
		delete(c.bySession, s.sessionKey)
	}
	if s.call != nil {
		delete(c.toolSessions, s.call.ID)
	}
	c.eventMu.Unlock()
	_, _ = c.call("thread/unsubscribe", map[string]any{"threadId": s.threadID})
}
func (c *Client) reapSessions() {
	c.eventMu.Lock()
	var expired []*session
	for _, s := range c.sessions {
		if time.Since(s.updated) > 15*time.Minute {
			expired = append(expired, s)
		}
	}
	c.eventMu.Unlock()
	for _, s := range expired {
		c.retire(s)
	}
}
