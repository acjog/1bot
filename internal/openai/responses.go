package openai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"1bot_test/internal/transport"
)

// Responses API inbound types (POST /v1/responses). The shape differs from Chat
// Completions: the conversation is an `input` string or an array of typed
// input items, tool declarations are flat (name/description/parameters), and
// tool calls/rounds are separate `function_call` / `function_call_output`
// items rather than assistant.tool_calls / tool messages.
type responsesRequest struct {
	Model        string          `json:"model"`
	Input        json.RawMessage `json:"input"`        // string or []responseInputItem
	Instructions string          `json:"instructions"` // system prompt, optional
	Tools        []responseTool  `json:"tools"`
	ToolChoice   json.RawMessage `json:"tool_choice"`
	Stream       bool            `json:"stream"`
}

type responseTool struct {
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
	Strict      bool            `json:"strict"`
}

type responseInputItem struct {
	Type      string          `json:"type"`
	Role      string          `json:"role"`
	Content   json.RawMessage `json:"content"`
	ID        string          `json:"id"`
	CallID    string          `json:"call_id"`
	Name      string          `json:"name"`
	Arguments string          `json:"arguments"`
	Output    json.RawMessage `json:"output"`
}

// responses serves POST /v1/responses, the OpenAI Responses API.
func (s *Server) responses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.record(body)

	var req responsesRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	messages := parseResponsesMessages(req.Input, req.Instructions)
	if len(messages) == 0 {
		http.Error(w, "no input content", http.StatusBadRequest)
		return
	}
	model := req.Model
	if model == "" {
		model = s.model
	}
	request := transport.Request{Session: sessionKey(messages), Messages: messages, Model: model}

	if len(req.ToolChoice) > 0 && string(req.ToolChoice) != "null" && string(req.ToolChoice) != `"auto"` && string(req.ToolChoice) != `"none"` {
		http.Error(w, "tool_choice currently supports auto or none", http.StatusBadRequest)
		return
	}
	request.ToolChoice = req.ToolChoice

	seenTools := map[string]bool{}
	for _, tool := range req.Tools {
		if tool.Type != "function" || tool.Name == "" {
			http.Error(w, "only named function tools are supported", http.StatusBadRequest)
			return
		}
		if seenTools[tool.Name] {
			http.Error(w, "duplicate function tool name", http.StatusBadRequest)
			return
		}
		seenTools[tool.Name] = true
		if len(tool.Parameters) == 0 {
			tool.Parameters = json.RawMessage(`{"type":"object","properties":{}}`)
		}
		var schema map[string]any
		if json.Unmarshal(tool.Parameters, &schema) != nil || schema == nil {
			http.Error(w, "function parameters must be a JSON schema object", http.StatusBadRequest)
			return
		}
		request.Tools = append(request.Tools, transport.Tool{Name: tool.Name, Description: tool.Description, Parameters: tool.Parameters})
	}

	if req.Stream {
		s.responsesStream(w, request, model)
		return
	}
	s.responsesSingle(w, request, model)
}

// parseResponsesMessages translates a Responses `input` (plus optional
// `instructions`) into the gateway's neutral messages. instructions becomes a
// leading system message; a string input becomes a single user message; an
// array becomes one neutral message per input item. Consecutive function_call
// items are folded into one assistant message so parallel calls survive the
// round trip the same way they do on the Chat Completions path.
func parseResponsesMessages(input json.RawMessage, instructions string) []transport.Message {
	var out []transport.Message
	if instructions != "" {
		out = append(out, transport.Message{Role: "system", Parts: []transport.Part{{Kind: transport.PartText, Text: instructions}}})
	}
	if len(input) == 0 || string(input) == "null" {
		return out
	}

	var s string
	if json.Unmarshal(input, &s) == nil {
		if text := strings.TrimSpace(s); text != "" {
			out = append(out, transport.Message{Role: "user", Parts: []transport.Part{{Kind: transport.PartText, Text: text}}})
		}
		return out
	}

	var items []responseInputItem
	if json.Unmarshal(input, &items) != nil {
		return out
	}
	for _, it := range items {
		switch it.Type {
		case "message":
			role := it.Role
			if role == "" {
				role = "user"
			}
			if text := strings.TrimSpace(responsesContentText(it.Content)); text != "" {
				out = append(out, transport.Message{Role: role, Parts: []transport.Part{{Kind: transport.PartText, Text: text}}})
			}
		case "function_call":
			if it.Name == "" {
				continue
			}
			id := it.CallID
			if id == "" {
				id = it.ID
			}
			part := transport.Part{Kind: transport.PartToolCall, Call: &transport.ToolCall{
				ID:        id,
				Name:      it.Name,
				Arguments: normalizeArgs(json.RawMessage(it.Arguments)),
			}}
			if n := len(out); n > 0 && out[n-1].Role == "assistant" && len(out[n-1].Parts) > 0 && out[n-1].Parts[0].Kind == transport.PartToolCall {
				out[n-1].Parts = append(out[n-1].Parts, part)
			} else {
				out = append(out, transport.Message{Role: "assistant", Parts: []transport.Part{part}})
			}
		case "function_call_output":
			out = append(out, transport.Message{Role: "tool", Parts: []transport.Part{{Kind: transport.PartToolResult, Result: &transport.ToolResult{
				CallID:  it.CallID,
				Content: responsesContentText(it.Output),
			}}}})
		}
	}
	return out
}

// responsesContentText extracts text from a Responses content field: a plain
// string or an array of {type, text} parts. The Responses API uses
// input_text / output_text part types, unlike Chat Completions' text.
func responsesContentText(raw json.RawMessage) string {
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
			if p.Type == "input_text" || p.Type == "output_text" || p.Type == "text" || p.Type == "" {
				b.WriteString(p.Text)
			}
		}
		return b.String()
	}
	return ""
}

// responsesSingle accumulates the transport's reply into one Responses API
// response: text becomes a message output item, tool calls become function_call
// output items.
func (s *Server) responsesSingle(w http.ResponseWriter, req transport.Request, model string) {
	var text strings.Builder
	var calls []map[string]any
	onEvent := func(ev transport.Event) {
		switch ev.Kind {
		case transport.EventText:
			text.WriteString(ev.Text)
		case transport.EventToolCall:
			if ev.Call == nil {
				return
			}
			calls = append(calls, functionCallItem(ev.Call))
		}
	}
	if _, err := transport.Complete(s.t, req, onEvent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var output []any
	if text.Len() > 0 {
		output = append(output, messageOutputItem(text.String()))
	}
	for _, c := range calls {
		output = append(output, c)
	}

	resp := map[string]any{
		"id":      "resp_" + randomID(),
		"object":  "response",
		"created": time.Now().Unix(),
		"model":   model,
		"status":  "completed",
		"output":  output,
		"usage":   nil,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// responsesStream writes the transport's reply as a Responses API SSE stream.
func (s *Server) responsesStream(w http.ResponseWriter, req transport.Request, model string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	id := "resp_" + randomID()
	created := time.Now().Unix()

	seq := 0
	send := func(typ string, extra map[string]any) {
		seq++
		ev := map[string]any{"type": typ, "sequence_number": seq}
		for k, v := range extra {
			ev[k] = v
		}
		b, _ := json.Marshal(ev)
		fmt.Fprintf(w, "data: %s\n\n", b)
		flusher.Flush()
	}
	response := func(status string, output []any) map[string]any {
		return map[string]any{
			"id": id, "object": "response", "created": created, "model": model,
			"status": status, "output": output, "usage": nil,
		}
	}

	send("response.created", map[string]any{"response": response("in_progress", []any{})})

	var text strings.Builder
	var calls []map[string]any
	msgItemID := "msg_" + randomID()
	textStarted := false

	onEvent := func(ev transport.Event) {
		switch ev.Kind {
		case transport.EventText:
			if !textStarted {
				textStarted = true
				send("response.output_item.added", map[string]any{
					"output_index": 0,
					"item": map[string]any{
						"type": "message", "id": msgItemID, "role": "assistant",
						"status": "in_progress", "content": []any{},
					},
				})
			}
			text.WriteString(ev.Text)
			send("response.output_text.delta", map[string]any{
				"item_id": msgItemID, "output_index": 0, "content_index": 0, "delta": ev.Text,
			})
		case transport.EventToolCall:
			if ev.Call == nil {
				return
			}
			idx := len(calls)
			if textStarted {
				idx++ // the text message holds output_index 0
			}
			arguments := string(ev.Call.Arguments)
			calls = append(calls, functionCallItem(ev.Call))
			send("response.output_item.added", map[string]any{
				"output_index": idx,
				"item": map[string]any{
					"type": "function_call", "id": ev.Call.ID, "call_id": ev.Call.ID,
					"name": ev.Call.Name, "arguments": "", "status": "in_progress",
				},
			})
			send("response.function_call_arguments.delta", map[string]any{
				"item_id": ev.Call.ID, "output_index": idx, "delta": arguments,
			})
			send("response.function_call_arguments.done", map[string]any{
				"item_id": ev.Call.ID, "output_index": idx, "arguments": arguments,
			})
			send("response.output_item.done", map[string]any{
				"output_index": idx,
				"item": map[string]any{
					"type": "function_call", "id": ev.Call.ID, "call_id": ev.Call.ID,
					"name": ev.Call.Name, "arguments": arguments, "status": "completed",
				},
			})
		case transport.EventThought:
			if s.verbose {
				fmt.Fprintf(stderr, "[thought] %s", ev.Text)
			}
		}
	}

	if _, err := transport.Complete(s.t, req, onEvent); err != nil {
		fmt.Fprintf(stderr, "turn: %v\n", err)
		failed := response("failed", []any{})
		failed["error"] = map[string]any{"code": "backend_error", "message": err.Error()}
		send("response.failed", map[string]any{"response": failed})
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
		return
	}

	var output []any
	if textStarted {
		item := messageOutputItem(text.String())
		item["id"] = msgItemID
		output = append(output, item)
		send("response.output_text.done", map[string]any{
			"item_id": msgItemID, "output_index": 0, "content_index": 0, "text": text.String(),
		})
		send("response.output_item.done", map[string]any{
			"output_index": 0, "item": item,
		})
	}
	for _, c := range calls {
		output = append(output, c)
	}

	send("response.completed", map[string]any{"response": response("completed", output)})
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

// messageOutputItem builds a completed assistant message output item holding
// one output_text content part.
func messageOutputItem(text string) map[string]any {
	return map[string]any{
		"type": "message", "id": "msg_" + randomID(), "role": "assistant", "status": "completed",
		"content": []any{map[string]any{"type": "output_text", "text": text, "annotations": []any{}}},
	}
}

// functionCallItem builds a completed function_call output item from a neutral
// tool call. arguments are the JSON string the Responses API expects.
func functionCallItem(call *transport.ToolCall) map[string]any {
	return map[string]any{
		"id": call.ID, "type": "function_call", "call_id": call.ID,
		"name": call.Name, "arguments": string(call.Arguments), "status": "completed",
	}
}
