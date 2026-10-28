// bridge.go is the plugin-codebuddy tool bridge: a loopback MCP server over
// HTTP (streamable HTTP transport). CodeBuddy's ACP SDK (1.7.0) only accepts
// MCP servers of type http/sse/acp in session/new.mcpServers — there is no
// stdio variant — so the plugin hosts the MCP server itself and hands CodeBuddy
// its URL. tools/call injects the call into the ACP client's turn in flight and
// blocks until the client's tool result arrives on a later turn. See
// CODEBUDDY-CLIENT-TOOLS.md.
package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"1bot_test/plugin/acp"
	"1bot_test/internal/transport"
)

// toolResultTimeout is how long a tools/call waits for the client to send its
// result before the MCP client is told the call timed out.
const toolResultTimeout = 5 * time.Minute

// mcpTool is one tool as the MCP tools/list offers it.
type mcpTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// mcpRPCRequest is one JSON-RPC message the MCP client POSTs.
type mcpRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type toolbridge struct {
	acp     *acp.Client
	verbose bool

	mu    sync.Mutex
	tools []mcpTool // current tools, as tools/list returns them

	srv  *http.Server
	ln   net.Listener
	once sync.Once
	done chan struct{}
}

func newToolbridge(c *acp.Client, verbose bool) *toolbridge {
	return &toolbridge{acp: c, verbose: verbose, done: make(chan struct{})}
}

// start opens the loopback listener and begins serving the MCP endpoint.
func (b *toolbridge) start() error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	b.ln = ln

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", b.handleMCP)
	b.srv = &http.Server{Handler: mux}
	go func() {
		_ = b.srv.Serve(ln)
	}()
	return nil
}

// setTools stores the tools and returns the MCP server entry session/new should
// hand CodeBuddy. An empty tool set returns no server (the bridge is unused).
func (b *toolbridge) setTools(tools []transport.Tool) ([]any, error) {
	b.mu.Lock()
	b.tools = toMCPTools(tools)
	b.mu.Unlock()
	if len(tools) == 0 {
		return []any{}, nil
	}
	return []any{map[string]any{
		"type":    "http",
		"name":    "1bot-tools",
		"url":     b.url(),
		"headers": []any{}, // required by the ACP SDK's McpServerHttp schema
	}}, nil
}

func (b *toolbridge) url() string {
	return "http://" + b.ln.Addr().String() + "/mcp"
}

// handleMCP serves the MCP streamable HTTP endpoint. It answers initialize,
// tools/list and tools/call, and acknowledges notifications with 202.
func (b *toolbridge) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 16<<20))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req mcpRPCRequest
	if json.Unmarshal(body, &req) != nil {
		http.Error(w, "invalid JSON-RPC", http.StatusBadRequest)
		return
	}
	if b.verbose {
		fmt.Fprintf(os.Stderr, "[mcp <-] %s\n", body)
	}

	// Notifications (no id) get an acknowledgement, not a JSON-RPC response.
	if len(req.ID) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	result := any(nil)
	var rpcErr any
	switch req.Method {
	case "initialize":
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(req.Params, &p)
		version := p.ProtocolVersion
		if version == "" {
			version = "2024-11-05"
		}
		result = map[string]any{
			"protocolVersion": version,
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "1bot_codebuddy", "version": "1"},
		}
	case "tools/list":
		b.mu.Lock()
		tools := b.tools
		b.mu.Unlock()
		result = map[string]any{"tools": tools}
	case "tools/call":
		var p struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if json.Unmarshal(req.Params, &p) != nil || p.Name == "" {
			rpcErr = map[string]any{"code": -32602, "message": "tools/call requires a name"}
			break
		}
		tc := transport.ToolCall{ID: newToolCallID(), Name: p.Name, Arguments: p.Arguments}
		resultCh := make(chan transport.ToolResult, 1)
		b.acp.ToolCall(tc, func(r transport.ToolResult) { resultCh <- r })

		select {
		case r := <-resultCh:
			result = map[string]any{
				"content": []map[string]any{{"type": "text", "text": r.Content}},
				"isError": r.IsError,
			}
		case <-time.After(toolResultTimeout):
			rpcErr = map[string]any{"code": -32000, "message": "tool call timed out"}
		case <-r.Context().Done():
			rpcErr = map[string]any{"code": -32000, "message": "client disconnected"}
		case <-b.done:
			rpcErr = map[string]any{"code": -32000, "message": "plugin shutting down"}
		}
	default:
		rpcErr = map[string]any{"code": -32601, "message": "method not found"}
	}

	response := map[string]any{"jsonrpc": "2.0", "id": json.RawMessage(req.ID)}
	if rpcErr != nil {
		response["error"] = rpcErr
	} else {
		response["result"] = result
	}
	if b.verbose {
		if rb, err := json.Marshal(response); err == nil {
			fmt.Fprintf(os.Stderr, "[mcp ->] %s\n", rb)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// close stops serving. A pending tools/call wakes via the closed done channel.
func (b *toolbridge) close() {
	b.once.Do(func() {
		close(b.done)
		if b.srv != nil {
			_ = b.srv.Close()
		}
	})
}

func toMCPTools(tools []transport.Tool) []mcpTool {
	out := make([]mcpTool, 0, len(tools))
	for _, t := range tools {
		schema := t.Parameters
		if len(schema) == 0 {
			schema = json.RawMessage(`{"type":"object","properties":{}}`)
		}
		out = append(out, mcpTool{Name: t.Name, Description: t.Description, InputSchema: schema})
	}
	return out
}

func newToolCallID() string {
	var x [12]byte
	_, _ = rand.Read(x[:])
	return "call_" + hex.EncodeToString(x[:])
}
