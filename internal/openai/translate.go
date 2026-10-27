package openai

import (
	"encoding/json"
	"strings"

	"1bot_test/internal/transport"
)

type chatToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	} `json:"function"`
}

type chatMessage struct {
	Role       string          `json:"role"`
	Content    json.RawMessage `json:"content"`
	ToolCalls  []chatToolCall  `json:"tool_calls"`
	ToolCallID string          `json:"tool_call_id"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
	Tools    []struct {
		Type     string         `json:"type"`
		Function transport.Tool `json:"function"`
	} `json:"tools"`
	ToolChoice json.RawMessage `json:"tool_choice"`
}

// parseMessages translates an OpenAI conversation into the gateway's neutral
// shape: assistant tool_calls become PartToolCall, tool messages become
// PartToolResult, everything else becomes text. The neutral messages are what
// the transport feeds to the backend, so tool calls and results survive the
// round trip instead of being flattened into prose.
func parseMessages(msgs []chatMessage) []transport.Message {
	var out []transport.Message
	for _, m := range msgs {
		role := m.Role
		if role == "" {
			role = "user"
		}

		var parts []transport.Part
		switch role {
		case "tool":
			parts = append(parts, transport.Part{
				Kind: transport.PartToolResult,
				Result: &transport.ToolResult{
					CallID:  m.ToolCallID,
					Content: contentText(m.Content),
				},
			})
		case "assistant":
			if text := strings.TrimSpace(contentText(m.Content)); text != "" {
				parts = append(parts, transport.Part{Kind: transport.PartText, Text: text})
			}
			for _, tc := range m.ToolCalls {
				if tc.Function.Name == "" {
					continue
				}
				parts = append(parts, transport.Part{
					Kind: transport.PartToolCall,
					Call: &transport.ToolCall{
						ID:        tc.ID,
						Name:      tc.Function.Name,
						Arguments: normalizeArgs(tc.Function.Arguments),
					},
				})
			}
		default: // user, system
			if text := strings.TrimSpace(contentText(m.Content)); text != "" {
				parts = append(parts, transport.Part{Kind: transport.PartText, Text: text})
			}
		}

		if len(parts) == 0 {
			continue
		}
		out = append(out, transport.Message{Role: role, Parts: parts})
	}
	return out
}

// normalizeArgs turns an OpenAI tool call's arguments into a JSON object.
// The wire form is a JSON string holding an object; some clients send the
// object directly. Both are reduced to the object so backends see one shape.
func normalizeArgs(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return json.RawMessage("{}")
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		var obj map[string]any
		if json.Unmarshal([]byte(s), &obj) == nil {
			return json.RawMessage(s)
		}
		b, _ := json.Marshal(map[string]any{"value": s})
		return b
	}
	var obj map[string]any
	if json.Unmarshal(raw, &obj) == nil {
		return raw
	}
	return json.RawMessage("{}")
}

// contentText extracts the text from a content field that is either a plain
// string or an array of {type, text} parts (multimodal).
func contentText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &parts) == nil {
		var b strings.Builder
		for _, p := range parts {
			if p.Type == "text" || p.Type == "" {
				b.WriteString(p.Text)
			}
		}
		return b.String()
	}
	return ""
}
