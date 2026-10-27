// Package transport defines the boundary between the gateway's core and the
// concrete way it reaches an upstream model — a child process speaking ACP
// today, an HTTP endpoint later. Routing and protocol translation only see
// Transport and Event; they never see a specific implementation.
package transport

// EventKind is one reduced category of a streamed reply. The upstream may
// emit many more, but the gateway only needs these three.
type EventKind string

const (
	EventText    EventKind = "text"    // user-visible reply text
	EventThought EventKind = "thought" // reasoning, usually hidden
	EventTool    EventKind = "tool"    // an upstream tool was invoked
)

// Event is one streamed piece of an upstream reply.
type Event struct {
	Kind EventKind
	Text string // set when Kind is EventText or EventThought
	Tool string // set when Kind is EventTool
}

// Transport turns one user message into a stream of Events plus a stop
// reason. Implementations own their concurrency: a single Transport instance
// need not support concurrent Turn calls.
type Transport interface {
	// Start launches the transport and performs any handshake.
	Start() error
	// Turn sends text as the user prompt and streams the reply through
	// onEvent (called sequentially, on the transport's reader goroutine).
	// It returns the upstream stop reason.
	Turn(text string, onEvent func(Event)) (stopReason string, err error)
	// Close releases the transport's resources.
	Close() error
}
