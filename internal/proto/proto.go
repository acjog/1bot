// Package proto defines the wire protocol between the gateway's PluginHost
// and a plugin child process. It is deliberately shaped like ACP: a plugin
// that wraps a CodeBuddy child can pass frames through almost unchanged.
//
// Transport: stdio JSON-RPC, one frame per line (NDJSON).
package proto

import "encoding/json"

// Frame is one JSON-RPC 2.0 message on the line stream.
type Frame struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Error is a JSON-RPC error payload.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// HasID reports whether the frame carries a non-null request id.
func (f Frame) HasID() bool { return len(f.ID) > 0 && string(f.ID) != "null" }

// IDKey returns the raw id bytes as a map key.
func (f Frame) IDKey() string { return string(f.ID) }

// Method names.
const (
	MethodTurn   = "turn"   // core -> plugin: run one user turn
	MethodEvent  = "event"  // plugin -> core: a streamed reply increment
	MethodCancel = "cancel" // core -> plugin: abort a turn (reserved)
)

// Part is one prompt item.
type Part struct {
	Type string `json:"type"` // "text"
	Text string `json:"text"`
}

// TurnParams is the params of MethodTurn.
type TurnParams struct {
	Prompt []Part `json:"prompt"`
}

// TurnResult is the result of MethodTurn.
type TurnResult struct {
	StopReason string `json:"stopReason"`
}

// EventParams is the params of MethodEvent (a notification).
type EventParams struct {
	ID   int64  `json:"id"`             // which turn this event belongs to
	Kind string `json:"kind"`           // text | thought | tool
	Text string `json:"text,omitempty"` // set when kind is text/thought
	Name string `json:"name,omitempty"` // tool name when kind is tool
}
