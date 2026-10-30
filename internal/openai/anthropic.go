package openai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"1bot_test/internal/transport"
)

type anthropicRequest struct {
	Model    string          `json:"model"`
	System   json.RawMessage `json:"system"`
	Messages []struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"messages"`
	Tools []struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Schema      json.RawMessage `json:"input_schema"`
	} `json:"tools"`
	ToolChoice struct {
		Type string `json:"type"`
	} `json:"tool_choice"`
	Stream   bool            `json:"stream"`
	Thinking json.RawMessage `json:"thinking"`
}

func parseAnthropic(in anthropicRequest) (transport.Request, error) {
	req := transport.Request{Model: in.Model}
	if len(in.Thinking) > 0 && string(in.Thinking) != "null" {
		return req, fmt.Errorf("extended thinking is not supported")
	}
	if len(in.System) > 0 {
		text, err := nativeText(in.System)
		if err != nil {
			return req, err
		}
		req.Messages = append(req.Messages, transport.Message{Role: "system", Parts: []transport.Part{{Kind: transport.PartText, Text: text}}})
	}
	if len(in.Messages) == 0 {
		return req, fmt.Errorf("messages must not be empty")
	}
	for _, m := range in.Messages {
		if m.Role != "user" && m.Role != "assistant" {
			return req, fmt.Errorf("unsupported message role")
		}
		msg := transport.Message{Role: m.Role}
		var text string
		if json.Unmarshal(m.Content, &text) == nil {
			msg.Parts = append(msg.Parts, transport.Part{Kind: transport.PartText, Text: text})
		} else {
			var blocks []struct {
				Type      string          `json:"type"`
				Text      string          `json:"text"`
				ID        string          `json:"id"`
				Name      string          `json:"name"`
				Input     json.RawMessage `json:"input"`
				ToolUseID string          `json:"tool_use_id"`
				Content   json.RawMessage `json:"content"`
				IsError   bool            `json:"is_error"`
			}
			if json.Unmarshal(m.Content, &blocks) != nil || len(blocks) == 0 {
				return req, fmt.Errorf("invalid message content")
			}
			for _, b := range blocks {
				switch b.Type {
				case "text":
					msg.Parts = append(msg.Parts, transport.Part{Kind: transport.PartText, Text: b.Text})
				case "tool_use":
					if m.Role != "assistant" || b.ID == "" || b.Name == "" {
						return req, fmt.Errorf("invalid tool_use block")
					}
					args, err := jsonObject(b.Input)
					if err != nil {
						return req, err
					}
					raw, _ := json.Marshal(args)
					msg.Parts = append(msg.Parts, transport.Part{Kind: transport.PartToolCall, Call: &transport.ToolCall{ID: b.ID, Name: b.Name, Arguments: raw}})
				case "tool_result":
					if m.Role != "user" || b.ToolUseID == "" {
						return req, fmt.Errorf("invalid tool_result block")
					}
					result, err := nativeText(b.Content)
					if err != nil {
						return req, err
					}
					msg.Parts = append(msg.Parts, transport.Part{Kind: transport.PartToolResult, Result: &transport.ToolResult{CallID: b.ToolUseID, Content: result, IsError: b.IsError}})
				default:
					return req, fmt.Errorf("unsupported content block: %s", b.Type)
				}
			}
		}
		req.Messages = append(req.Messages, msg)
	}
	for _, t := range in.Tools {
		req.Tools = append(req.Tools, transport.Tool{Name: t.Name, Description: t.Description, Parameters: t.Schema})
	}
	switch in.ToolChoice.Type {
	case "", "auto":
		req.ToolChoice = json.RawMessage(`"auto"`)
	case "none":
		req.ToolChoice = json.RawMessage(`"none"`)
	default:
		return req, fmt.Errorf("tool_choice supports auto or none only")
	}
	return req, validateNative(&req)
}

func (s *Server) anthropic(w http.ResponseWriter, r *http.Request) {
	var in anthropicRequest
	if !s.nativeBody(w, r, "anthropic", &in) {
		return
	}
	req, err := parseAnthropic(in)
	if err != nil {
		nativeError(w, "anthropic", 400, err.Error())
		return
	}
	if req.Model == "" {
		req.Model = s.model
	}
	req.Session = nativeSession("anthropic", req)
	id := fmt.Sprintf("msg_%d", time.Now().UnixNano())
	content := []object{}
	index := 0
	hadTool := false
	if in.Stream {
		startNativeStream(w)
		nativeSSE(w, "message_start", object{"type": "message_start", "message": object{"id": id, "type": "message", "role": "assistant", "model": req.Model, "content": []object{}, "stop_reason": nil, "stop_sequence": nil, "usage": object{"input_tokens": 0, "output_tokens": 0}}})
	}
	var eventErr error
	stop, err := transport.Complete(s.t, req, func(e transport.Event) {
		if eventErr != nil || r.Context().Err() != nil {
			return
		}
		var block object
		var delta object
		switch e.Kind {
		case transport.EventText:
			if e.Text == "" {
				return
			}
			block = object{"type": "text", "text": e.Text}
			delta = object{"type": "text_delta", "text": e.Text}
		case transport.EventToolCall:
			if e.Call == nil || e.Call.Name == "" {
				eventErr = fmt.Errorf("invalid upstream tool call")
				return
			}
			args, err := jsonObject(e.Call.Arguments)
			if err != nil {
				eventErr = err
				return
			}
			callID := e.Call.ID
			if callID == "" {
				eventErr = fmt.Errorf("upstream tool call is missing an ID")
				return
			}
			block = object{"type": "tool_use", "id": callID, "name": e.Call.Name, "input": args}
			raw, _ := json.Marshal(args)
			delta = object{"type": "input_json_delta", "partial_json": string(raw)}
			hadTool = true
		default:
			return
		}
		content = append(content, block)
		if in.Stream {
			initial := object{"type": "text", "text": ""}
			if e.Kind == transport.EventToolCall {
				initial = object{"type": "tool_use", "id": block["id"], "name": block["name"], "input": object{}}
			}
			nativeSSE(w, "content_block_start", object{"type": "content_block_start", "index": index, "content_block": initial})
			nativeSSE(w, "content_block_delta", object{"type": "content_block_delta", "index": index, "delta": delta})
			nativeSSE(w, "content_block_stop", object{"type": "content_block_stop", "index": index})
		}
		index++
	})
	if err == nil {
		err = eventErr
	}
	if err != nil {
		if in.Stream {
			nativeSSE(w, "error", object{"type": "error", "error": object{"type": "api_error", "message": "upstream request failed"}})
		} else {
			nativeError(w, "anthropic", 502, "upstream request failed")
		}
		return
	}
	if r.Context().Err() != nil {
		return
	}
	reason := "end_turn"
	if hadTool || stop == transport.StopToolUse {
		reason = "tool_use"
	}
	if in.Stream {
		nativeSSE(w, "message_delta", object{"type": "message_delta", "delta": object{"stop_reason": reason, "stop_sequence": nil}, "usage": object{"output_tokens": 0}})
		nativeSSE(w, "message_stop", object{"type": "message_stop"})
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(object{"id": id, "type": "message", "role": "assistant", "model": req.Model, "content": content, "stop_reason": reason, "stop_sequence": nil, "usage": object{"input_tokens": 0, "output_tokens": 0}})
	}
}
