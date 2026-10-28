// bridge.go is the plugin-codebuddy tool bridge: a loopback HTTP endpoint the
// mcp-helper POSTs each tools/call to. It turns the call into a neutral
// transport.ToolCall, injects it into the ACP client's turn in flight (which
// reports it to the gateway and returns tool_use), then blocks until the
// client's tool result arrives on a later turn. See CODEBUDDY-CLIENT-TOOLS.md.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"1bot_test/plugin/acp"
	"1bot_test/internal/transport"
)

// toolResultTimeout is how long a tools/call waits for the client to send its
// result before the helper is told the call timed out.
const toolResultTimeout = 5 * time.Minute

type toolbridge struct {
	acp *acp.Client

	mu        sync.Mutex
	toolsPath string
	token     string
	host      string // "127.0.0.1:PORT"

	srv   *http.Server
	ln    net.Listener
	once  sync.Once
	done  chan struct{}
}

func newToolbridge(c *acp.Client) *toolbridge {
	return &toolbridge{acp: c, done: make(chan struct{})}
}

// start opens the loopback listener, writes an empty tools.json, and begins
// serving the callback endpoint. It must run before setTools.
func (b *toolbridge) start() error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	b.ln = ln
	b.host = ln.Addr().String()

	f, err := os.CreateTemp("", "1bot-tools-*.json")
	if err != nil {
		_ = ln.Close()
		return err
	}
	b.toolsPath = f.Name()
	if err := f.Close(); err != nil {
		_ = ln.Close()
		return err
	}

	b.token = newBridgeToken()
	b.srv = &http.Server{Handler: b}
	go func() {
		_ = b.srv.Serve(ln)
	}()
	return nil
}

// setTools rewrites tools.json with the given tools and returns the stdio MCP
// server entry CodeBuddy should be told to launch in session/new. An empty
// tool set returns an empty server list (no MCP bridge).
func (b *toolbridge) setTools(tools []transport.Tool) ([]any, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err := b.writeToolsFile(tools); err != nil {
		return nil, err
	}
	if len(tools) == 0 {
		return []any{}, nil
	}
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	server := map[string]any{
		"type":    "stdio",
		"command": exe,
		"args":    []string{"mcp-helper", b.callbackURL(), b.toolsPath},
	}
	return []any{server}, nil
}

// writeToolsFile serialises the client's tools as the MCP helper reads them.
func (b *toolbridge) writeToolsFile(tools []transport.Tool) error {
	out := make([]mcpTool, 0, len(tools))
	for _, t := range tools {
		schema := t.Parameters
		if len(schema) == 0 {
			schema = json.RawMessage(`{"type":"object","properties":{}}`)
		}
		out = append(out, mcpTool{Name: t.Name, Description: t.Description, InputSchema: schema})
	}
	data, err := json.Marshal(out)
	if err != nil {
		return err
	}
	return os.WriteFile(b.toolsPath, data, 0o600)
}

func (b *toolbridge) callbackURL() string {
	return "http://" + b.host + "/cb/" + b.token
}

// ServeHTTP handles the helper's blocked tools/call POST.
func (b *toolbridge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const prefix = "/cb/"
	if r.Method != http.MethodPost || !strings.HasPrefix(r.URL.Path, prefix) || r.URL.Path[len(prefix):] != b.token {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 16<<20))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var call struct {
		ToolCallID string          `json:"tool_call_id"`
		Name       string          `json:"name"`
		Arguments  json.RawMessage `json:"arguments"`
	}
	if json.Unmarshal(body, &call) != nil || call.ToolCallID == "" || call.Name == "" {
		http.Error(w, "invalid tool call", http.StatusBadRequest)
		return
	}

	tc := transport.ToolCall{ID: call.ToolCallID, Name: call.Name, Arguments: call.Arguments}
	resultCh := make(chan transport.ToolResult, 1)
	b.acp.ToolCall(tc, func(r transport.ToolResult) { resultCh <- r })

	select {
	case r := <-resultCh:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mcpCallbackResponse{
			Content: []map[string]any{{"type": "text", "text": r.Content}},
			IsError: r.IsError,
		})
	case <-time.After(toolResultTimeout):
		http.Error(w, "tool call timed out", http.StatusGatewayTimeout)
	case <-r.Context().Done():
		// the helper went away; leave the call registered so a late result
		// can still be handed back if one ever arrives
	case <-b.done:
		http.Error(w, "plugin shutting down", http.StatusGone)
	}
}

// close stops serving and removes the tools file. Pending handlers wake via
// the closed done channel.
func (b *toolbridge) close() {
	b.once.Do(func() {
		close(b.done)
		if b.srv != nil {
			_ = b.srv.Close()
		}
		if b.toolsPath != "" {
			_ = os.Remove(b.toolsPath)
		}
	})
}

func newBridgeToken() string {
	var x [16]byte
	_, _ = rand.Read(x[:])
	return hex.EncodeToString(x[:])
}
