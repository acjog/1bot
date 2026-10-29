// Package plugin is the gateway side of the plugin protocol: it launches a
// plugin child process and drives it over stdio JSON-RPC, presenting the
// whole thing as a transport.Transport. The gateway therefore knows nothing
// about what the plugin does internally (ACP, app-server, HTTP, ...).
package plugin

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

	"1bot_test/internal/proto"
	"1bot_test/internal/transport"
	"1bot_test/internal/vlog"
)

var stderr = os.Stderr

// Options configures the plugin child process.
type Options struct {
	Bin     string   // plugin executable (name or path)
	Args    []string // extra args handed to the plugin
	Verbose bool     // dump every plugin-protocol frame to stderr
}

// Host launches and drives one plugin process. It implements transport.Transport.
type Host struct {
	opts Options

	cmd     *exec.Cmd
	stdin   io.WriteCloser
	scanner *bufio.Scanner

	writeMu sync.Mutex
	pendMu  sync.Mutex
	pending map[string]chan proto.Frame
	nextID  int64

	// events maps a turn id to its live callback.
	evMu   sync.Mutex
	events map[string]func(transport.Event)
}

func New(opts Options) *Host {
	return &Host{
		opts:    opts,
		pending: map[string]chan proto.Frame{},
		events:  map[string]func(transport.Event){},
	}
}

// Start launches the plugin child and begins reading its stdout.
func (h *Host) Start() error {
	cmd := exec.Command(h.opts.Bin, h.opts.Args...)
	cmd.Stderr = stderr // plugin's own logs pass through

	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting plugin %s: %w", h.opts.Bin, err)
	}
	h.cmd = cmd
	h.stdin = in
	h.scanner = bufio.NewScanner(out)
	h.scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	go h.readLoop()
	return nil
}

// Close kills the plugin child and waits for it to exit.
func (h *Host) Close() error {
	h.writeMu.Lock()
	if h.stdin != nil {
		_ = h.stdin.Close()
	}
	h.writeMu.Unlock()
	if h.cmd != nil && h.cmd.Process != nil {
		_ = h.cmd.Process.Kill()
	}
	if h.cmd != nil {
		_ = h.cmd.Wait()
	}
	return nil
}

func (h *Host) readLoop() {
	for h.scanner.Scan() {
		line := h.scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var f proto.Frame
		if err := json.Unmarshal(line, &f); err != nil {
			fmt.Fprintf(stderr, "plugin: bad JSON: %v: %s\n", err, line)
			continue
		}
		if h.opts.Verbose {
			vlog.Frame(stderr, "plugin <-", string(line))
		}
		switch {
		case f.Method == proto.MethodEvent:
			h.dispatch(f)
		case f.HasID():
			h.deliver(f)
		}
	}
}

// dispatch routes an event notification to the turn it belongs to.
func (h *Host) dispatch(f proto.Frame) {
	var p proto.EventParams
	if err := json.Unmarshal(f.Params, &p); err != nil {
		return
	}
	h.evMu.Lock()
	fn := h.events[fmt.Sprintf("%d", p.ID)]
	h.evMu.Unlock()
	if fn == nil {
		return
	}
	var ev transport.Event
	switch p.Kind {
	case "text":
		ev = transport.Event{Kind: transport.EventText, Text: p.Text}
	case "thought":
		ev = transport.Event{Kind: transport.EventThought, Text: p.Text}
	case "tool_call":
		ev = transport.Event{Kind: transport.EventToolCall, Call: p.Call}
	default:
		return
	}
	fn(ev)
}

// deliver hands a response to whoever is waiting on its id.
func (h *Host) deliver(f proto.Frame) {
	h.pendMu.Lock()
	ch := h.pending[f.IDKey()]
	h.pendMu.Unlock()
	if ch != nil {
		ch <- f
	}
}

// Turn implements transport.Transport. It sends one turn to the plugin and
// streams events through onEvent until the plugin returns a result.
func (h *Host) Turn(messages []transport.Message, onEvent func(transport.Event)) (string, error) {
	return h.Complete(transport.Request{Messages: messages}, onEvent)
}

func (h *Host) Complete(req transport.Request, onEvent func(transport.Event)) (string, error) {
	h.pendMu.Lock()
	h.nextID++
	id := fmt.Sprintf("%d", h.nextID)
	ch := make(chan proto.Frame, 1)
	h.pending[id] = ch
	h.pendMu.Unlock()
	defer func() {
		h.pendMu.Lock()
		delete(h.pending, id)
		h.pendMu.Unlock()
	}()

	h.evMu.Lock()
	h.events[id] = onEvent
	h.evMu.Unlock()
	defer func() {
		h.evMu.Lock()
		delete(h.events, id)
		h.evMu.Unlock()
	}()

	params, _ := json.Marshal(req)
	frame := proto.Frame{
		JSONRPC: "2.0",
		ID:      json.RawMessage(id),
		Method:  proto.MethodTurn,
		Params:  params,
	}
	if err := h.write(frame); err != nil {
		return "", err
	}

	select {
	case f := <-ch:
		if f.Error != nil {
			return "", fmt.Errorf("turn: plugin error %d: %s", f.Error.Code, f.Error.Message)
		}
		var r proto.TurnResult
		_ = json.Unmarshal(f.Result, &r)
		return r.StopReason, nil
	case <-time.After(300 * time.Second):
		return "", fmt.Errorf("turn: timed out")
	}
}

func (h *Host) write(f proto.Frame) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	if h.opts.Verbose {
		vlog.Frame(stderr, "plugin ->", string(b))
	}
	h.writeMu.Lock()
	defer h.writeMu.Unlock()
	_, err = h.stdin.Write(append(b, '\n'))
	return err
}
