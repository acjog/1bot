// mcp.go is the plugin-codebuddy "mcp-helper" subcommand: a stdio MCP server
// CodeBuddy launches to reach client-declared function tools. It implements
// only the methods CodeBuddy needs — initialize, tools/list and tools/call —
// and its tools/call POSTs the callback URL and blocks until the client's tool
// result comes back. See CODEBUDDY-CLIENT-TOOLS.md.
package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

type mcpTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type mcpRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type mcpCallbackRequest struct {
	ToolCallID string          `json:"tool_call_id"`
	Name       string          `json:"name"`
	Arguments  json.RawMessage `json:"arguments"`
}

type mcpCallbackResponse struct {
	Content any  `json:"content"`
	IsError bool `json:"is_error,omitempty"`
}

// runMCPHelper runs the stdio MCP server. args are [callbackURL, toolsPath];
// callbackURL may be a file holding the URL (keeps the token off argv).
func runMCPHelper(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("mcp-helper expects callback URL and tools file")
	}
	callbackURL, toolsPath := args[0], args[1]
	if !strings.HasPrefix(callbackURL, "http://") && !strings.HasPrefix(callbackURL, "https://") {
		b, err := os.ReadFile(callbackURL)
		if err != nil {
			return err
		}
		callbackURL = strings.TrimSpace(string(b))
	}
	b, err := os.ReadFile(toolsPath)
	if err != nil {
		return err
	}
	var tools []mcpTool
	if err := json.Unmarshal(b, &tools); err != nil {
		return fmt.Errorf("read MCP tools: %w", err)
	}

	// The callback is the plugin's loopback endpoint; never go through a proxy.
	direct := http.DefaultTransport.(*http.Transport).Clone()
	direct.Proxy = nil
	client := &http.Client{Transport: direct}

	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 64<<10), 16<<20)
	enc := json.NewEncoder(os.Stdout)
	var outMu sync.Mutex
	respond := func(v any) {
		outMu.Lock()
		_ = enc.Encode(v)
		outMu.Unlock()
	}

	for in.Scan() {
		var req mcpRPCRequest
		if json.Unmarshal(in.Bytes(), &req) != nil {
			continue
		}
		// Notifications carry no id and must not be answered.
		if len(req.ID) == 0 {
			continue
		}
		go func(req mcpRPCRequest) {
			result := any(nil)
			var rpcErr any
			switch req.Method {
			case "initialize":
				result = map[string]any{
					"protocolVersion": "2025-06-18",
					"capabilities":    map[string]any{"tools": map[string]any{}},
					"serverInfo":      map[string]any{"name": "1bot_codebuddy", "version": "1"},
				}
			case "tools/list":
				result = map[string]any{"tools": tools}
			case "tools/call":
				var p struct {
					Name      string          `json:"name"`
					Arguments json.RawMessage `json:"arguments"`
				}
				if err := json.Unmarshal(req.Params, &p); err != nil {
					rpcErr = map[string]any{"code": -32602, "message": err.Error()}
					break
				}
				payload, _ := json.Marshal(mcpCallbackRequest{
					ToolCallID: newToolCallID(),
					Name:       p.Name,
					Arguments:  p.Arguments,
				})
				httpReq, _ := http.NewRequest(http.MethodPost, callbackURL, bytes.NewReader(payload))
				httpReq.Header.Set("Content-Type", "application/json")
				res, err := client.Do(httpReq)
				if err != nil {
					rpcErr = map[string]any{"code": -32000, "message": err.Error()}
					break
				}
				body, _ := io.ReadAll(io.LimitReader(res.Body, 32<<20))
				res.Body.Close()
				if res.StatusCode != http.StatusOK {
					rpcErr = map[string]any{"code": -32000, "message": string(body)}
					break
				}
				var cb mcpCallbackResponse
				if err := json.Unmarshal(body, &cb); err != nil {
					rpcErr = map[string]any{"code": -32000, "message": err.Error()}
					break
				}
				result = map[string]any{"content": cb.Content, "isError": cb.IsError}
			default:
				rpcErr = map[string]any{"code": -32601, "message": "method not found"}
			}
			response := map[string]any{"jsonrpc": "2.0", "id": json.RawMessage(req.ID)}
			if rpcErr != nil {
				response["error"] = rpcErr
			} else {
				response["result"] = result
			}
			respond(response)
		}(req)
	}
	return in.Err()
}

func newToolCallID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return "call_" + hex.EncodeToString(b[:])
}
