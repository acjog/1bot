// Package transport defines the boundary between the gateway's core and the
// concrete way it reaches an upstream model — a child process speaking ACP
// today, an HTTP endpoint later. Routing and protocol translation only see
// Transport, Message and Event; they never see a specific implementation.
//
// The package also owns the gateway's one neutral shape for tool use, so the
// OpenAI front, the plugin wire protocol and the CodeBuddy/Codex backends all
// translate into the same ToolCall and ToolResult instead of each other's
// formats.
package transport

import (
	"encoding/json"
	"fmt"
)

// Tool is a function declared and executed by the API client.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters"`
}

// Request preserves request-scoped model and tool declarations across plugins.
// Session is an opaque conversation key: when non-empty, a plugin may reuse the
// backend conversation it opened for the same key (so tools/context injected on
// an earlier turn survive). It is derived by the gateway, not the API client.
type Request struct {
	Session    string          `json:"session,omitempty"`
	Messages   []Message       `json:"messages"`
	Model      string          `json:"model,omitempty"`
	Tools      []Tool          `json:"tools,omitempty"`
	ToolChoice json.RawMessage `json:"toolChoice,omitempty"`
}

type RequestTransport interface {
	Complete(Request, func(Event)) (string, error)
}

func Complete(t Transport, req Request, onEvent func(Event)) (string, error) {
	if rt, ok := t.(RequestTransport); ok {
		return rt.Complete(req, onEvent)
	}
	if len(req.Tools) > 0 {
		return "", fmt.Errorf("transport does not support client tools")
	}
	return t.Turn(req.Messages, onEvent)
}

// EventKind is one reduced category of a streamed reply. The upstream may
// emit many more, but the gateway only needs these.
type EventKind string

const (
	EventText     EventKind = "text"      // user-visible reply text
	EventThought  EventKind = "thought"   // reasoning, usually hidden
	EventToolCall EventKind = "tool_call" // a tool the model asks the client to run
)

// Event is one streamed piece of an upstream reply.
type Event struct {
	Kind EventKind
	Text string    // set when Kind is EventText or EventThought
	Call *ToolCall // set when Kind is EventToolCall
}

// ToolCall is the gateway's neutral tool invocation: an id (mapped back and
// forth to each backend's own id), a name and JSON arguments.
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// ToolResult is the gateway's neutral tool result: the call it answers and
// the text it produced.
type ToolResult struct {
	CallID  string `json:"callId"`
	Content string `json:"content"`
	IsError bool   `json:"isError,omitempty"`
}

// PartKind is one kind of message part.
type PartKind string

const (
	PartText       PartKind = "text"
	PartToolCall   PartKind = "tool_call"
	PartToolResult PartKind = "tool_result"
)

// Part is one item of a Message: plain text, a tool call, or a tool result.
type Part struct {
	Kind   PartKind    `json:"kind"`
	Text   string      `json:"text,omitempty"`
	Call   *ToolCall   `json:"call,omitempty"`
	Result *ToolResult `json:"result,omitempty"`
}

// Message preserves the client's system, developer, user, assistant, and tool
// roles. Legacy adapters may also accept tool-result parts on user messages.
type Message struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

// Stop reasons a Turn may return. A turn either ends with a plain reply
// (StopEndTurn) or with tool calls the client must run (StopToolUse).
const (
	StopEndTurn = "end_turn"
	StopToolUse = "tool_use"
)

// Transport turns one conversation (the whole history, tool calls and results
// included) into a stream of Events plus a stop reason. Implementations own
// concurrency and continuation state; independent conversations must not share
// hidden history. RequestTransport additionally carries model/tool settings.
type Transport interface {
	// Start launches the transport and performs any handshake.
	Start() error
	// Turn advances the conversation one step. It streams the reply through
	// onEvent (called sequentially) and
	// returns StopEndTurn when the reply is final, or StopToolUse when the
	// model has asked for tools whose results the next Turn will carry.
	Turn(messages []Message, onEvent func(Event)) (stopReason string, err error)
	// Close releases the transport's resources.
	Close() error
}
