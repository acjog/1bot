// Package zcode drives Z.AI's ZCode coding plan directly over its
// Anthropic-compatible Messages API, and reduces its streamed reply to
// transport.Event. It reads the plan's key from ZCode's own credential store
// (~/.zcode/v2) — it never launches `zcode app-server`.
//
// One turn is one HTTP request:
//
//	POST {base}/v1/messages   (base = api.z.ai/api/anthropic, or BigModel's)
//	stream SSE text deltas     -> transport.EventText
//	stream SSE tool_use blocks -> transport.EventToolCall
//
// Client tools are declared as Anthropic `tools` and round-trip through the
// conversation: an assistant `tool_use` is answered by a later `tool_result`
// in a user message. The Messages API is stateless, so no thread/session state
// is kept here.
package zcode

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"1bot_test/internal/transport"
	"1bot_test/internal/vlog"
)

var stderr = os.Stderr

// Options configures the ZCode adapter.
type Options struct {
	Model   string // model id; empty = DefaultModel
	Verbose bool   // dump each request and SSE event to stderr
	APIKey  string // optional: override the key read from ZCode's store
	Base    string // optional: override the Anthropic-compatible base URL
}

// Client turns one conversation into Anthropic Messages requests. It
// implements transport.Transport and transport.RequestTransport.
type Client struct {
	opts Options
	cred cred
	http *http.Client

	// promptMu serialises turns: one request at a time.
	promptMu sync.Mutex
}

func New(opts Options) *Client {
	return &Client{opts: opts}
}

// Start reads ZCode's coding plan credentials. There is no child process to
// launch, so readiness is just whether a key could be found.
func (c *Client) Start() error {
	cred, err := readCred()
	if err != nil {
		return err
	}
	if c.opts.APIKey != "" {
		cred.Key = c.opts.APIKey
	}
	if c.opts.Base != "" {
		cred.Base = strings.TrimRight(c.opts.Base, "/")
	}
	if cred.Key == "" {
		return fmt.Errorf("no ZCode coding plan key")
	}
	if cred.Base == "" {
		cred.Base = ZaiBase
	}
	c.cred = cred
	c.http = &http.Client{Timeout: 5 * time.Minute}
	return nil
}

// Close releases nothing: the adapter owns no subprocess or persistent
// connection.
func (c *Client) Close() error { return nil }

// Models returns the coding plan's model catalog: the models ZCode lists for
// its built-in coding plan provider in config.json, else the default model
// alone.
func (c *Client) Models() ([]transport.Model, error) {
	if ms := readConfigModels(); len(ms) > 0 {
		return ms, nil
	}
	return []transport.Model{{ID: DefaultModel, Name: "GLM-5"}}, nil
}

// anthropicTool is a declared client function, Anthropic's tool shape.
type anthropicTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// anthropicContent is one block of a message: text, a tool_use, or a
// tool_result. Only the fields each block type needs are set.
type anthropicContent struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`          // tool_use
	Name      string          `json:"name,omitempty"`        // tool_use
	Input     json.RawMessage `json:"input,omitempty"`       // tool_use
	ToolUseID string          `json:"tool_use_id,omitempty"` // tool_result
	Content   string          `json:"content,omitempty"`     // tool_result
	IsError   bool            `json:"is_error,omitempty"`    // tool_result
}

type anthropicMessage struct {
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Tools     []anthropicTool    `json:"tools,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Stream    bool               `json:"stream"`
}

// Turn implements transport.Transport; it delegates to Complete with no tools.
func (c *Client) Turn(messages []transport.Message, onEvent func(transport.Event)) (string, error) {
	return c.Complete(transport.Request{Messages: messages}, onEvent)
}

// Complete implements transport.RequestTransport: it posts the whole
// conversation (tools declared, tool_use/tool_result blocks preserved) to the
// Messages endpoint and streams the reply back through onEvent.
func (c *Client) Complete(req transport.Request, onEvent func(transport.Event)) (string, error) {
	c.promptMu.Lock()
	defer c.promptMu.Unlock()

	model := req.Model
	if model == "" {
		model = c.opts.Model
	}
	if model == "" {
		model = DefaultModel
	}

	tools := req.Tools
	if string(req.ToolChoice) == `"none"` {
		tools = nil
	}

	system, out := toAnthropic(req.Messages)
	if len(out) == 0 {
		return "", fmt.Errorf("no messages")
	}

	body, _ := json.Marshal(anthropicRequest{
		Model:     model,
		MaxTokens: 8192,
		System:    system,
		Tools:     toAnthropicTools(tools),
		Messages:  out,
		Stream:    true,
	})
	if c.opts.Verbose {
		vlog.Frame(stderr, "zcode ->", string(body))
	}

	url := c.cred.Base + "/v1/messages"
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.cred.Key)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	res, err := c.http.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 64*1024))
		return "", fmt.Errorf("zcode: %s: %s", res.Status, strings.TrimSpace(string(b)))
	}

	return c.stream(res, onEvent)
}

// stream consumes Anthropic's SSE reply: text deltas go out as EventText, and
// each tool_use block is accumulated into an EventToolCall. A reply that ends
// with tool calls reports StopToolUse so the client runs them and continues.
func (c *Client) stream(res *http.Response, onEvent func(transport.Event)) (string, error) {
	type toolUse struct {
		id   string
		name string
		args strings.Builder
	}
	tools := map[int]*toolUse{}
	var used bool

	sc := bufio.NewScanner(res.Body)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		if c.opts.Verbose {
			vlog.Frame(stderr, "zcode <-", data)
		}
		var ev struct {
			Type  string `json:"type"`
			Index *int   `json:"index"`
			ContentBlock struct {
				Type string `json:"type"`
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"content_block"`
			Delta struct {
				Type        string `json:"type"`
				Text        string `json:"text"`
				PartialJSON string `json:"partial_json"`
			} `json:"delta"`
			Error struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal([]byte(data), &ev) != nil {
			continue
		}
		switch ev.Type {
		case "content_block_start":
			if ev.ContentBlock.Type == "tool_use" && ev.Index != nil {
				tools[*ev.Index] = &toolUse{id: ev.ContentBlock.ID, name: ev.ContentBlock.Name}
			}
		case "content_block_delta":
			switch ev.Delta.Type {
			case "text_delta":
				if ev.Delta.Text != "" {
					onEvent(transport.Event{Kind: transport.EventText, Text: ev.Delta.Text})
				}
			case "input_json_delta":
				if ev.Index != nil {
					if t := tools[*ev.Index]; t != nil {
						t.args.WriteString(ev.Delta.PartialJSON)
					}
				}
			}
		case "content_block_stop":
			if ev.Index != nil {
				if t := tools[*ev.Index]; t != nil {
					args := strings.TrimSpace(t.args.String())
					if args == "" {
						args = "{}"
					}
					onEvent(transport.Event{Kind: transport.EventToolCall, Call: &transport.ToolCall{
						ID:        t.id,
						Name:      t.name,
						Arguments: json.RawMessage(args),
					}})
					used = true
				}
			}
		case "error":
			return "", fmt.Errorf("zcode: %s: %s", ev.Error.Type, ev.Error.Message)
		case "message_stop":
			if used {
				return transport.StopToolUse, nil
			}
			return transport.StopEndTurn, nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return transport.StopEndTurn, nil
}

// toAnthropic folds system/developer text into the top-level system field and
// converts every other message into content blocks, preserving tool_use and
// tool_result parts in order. A tool message (role "tool") becomes a user
// message, since Anthropic has no tool role.
func toAnthropic(msgs []transport.Message) (system string, out []anthropicMessage) {
	for _, m := range msgs {
		if m.Role == "system" || m.Role == "developer" {
			var sb strings.Builder
			for _, p := range m.Parts {
				if p.Kind == transport.PartText {
					sb.WriteString(p.Text)
				}
			}
			if text := sb.String(); text != "" {
				if system != "" {
					system += "\n\n"
				}
				system += text
			}
			continue
		}

		role := m.Role
		if role == "tool" {
			role = "user"
		}
		var blocks []anthropicContent
		for _, p := range m.Parts {
			switch p.Kind {
			case transport.PartText:
				blocks = append(blocks, anthropicContent{Type: "text", Text: p.Text})
			case transport.PartToolCall:
				if p.Call != nil {
					input := p.Call.Arguments
					if len(input) == 0 || string(input) == "null" {
						input = json.RawMessage("{}")
					}
					blocks = append(blocks, anthropicContent{Type: "tool_use", ID: p.Call.ID, Name: p.Call.Name, Input: input})
				}
			case transport.PartToolResult:
				if p.Result != nil {
					blocks = append(blocks, anthropicContent{Type: "tool_result", ToolUseID: p.Result.CallID, Content: p.Result.Content, IsError: p.Result.IsError})
				}
			}
		}
		if len(blocks) > 0 {
			out = append(out, anthropicMessage{Role: role, Content: blocks})
		}
	}
	return system, out
}

// toAnthropicTools maps the gateway's neutral tools to Anthropic's tool shape.
func toAnthropicTools(tools []transport.Tool) []anthropicTool {
	out := make([]anthropicTool, 0, len(tools))
	for _, t := range tools {
		schema := t.Parameters
		if len(schema) == 0 || string(schema) == "null" {
			schema = json.RawMessage(`{"type":"object","properties":{}}`)
		}
		out = append(out, anthropicTool{Name: t.Name, Description: t.Description, InputSchema: schema})
	}
	return out
}
