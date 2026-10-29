package openai

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"1bot_test/internal/transport"
	"1bot_test/internal/vlog"
)

var stderr = os.Stderr

// finishFor maps a transport stop reason to an OpenAI finish_reason.
func finishFor(stop string) string {
	if stop == transport.StopToolUse {
		return "tool_calls"
	}
	return "stop"
}

// toolCallChunk builds the OpenAI chunk for one neutral tool call. Arguments
// are sent as the JSON string OpenAI expects.
func toolCallChunk(call *transport.ToolCall, index int) map[string]any {
	return map[string]any{"tool_calls": []any{map[string]any{
		"index": index,
		"id":    call.ID,
		"type":  "function",
		"function": map[string]any{
			"name":      call.Name,
			"arguments": string(call.Arguments),
		},
	}}}
}

// stream writes the transport's reply as an OpenAI SSE chunk stream.
func (s *Server) stream(w http.ResponseWriter, req transport.Request, model string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	id := "chatcmpl-" + randomID()
	send := func(delta map[string]any, finish any) {
		chunk := map[string]any{
			"id": id, "object": "chat.completion.chunk", "created": time.Now().Unix(), "model": model,
			"choices": []any{map[string]any{
				"index": 0, "delta": delta, "finish_reason": finish,
			}},
		}
		b, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", b)
		flusher.Flush()
	}

	toolIndex := 0
	onEvent := func(ev transport.Event) {
		switch ev.Kind {
		case transport.EventText:
			send(map[string]any{"content": ev.Text}, nil)
		case transport.EventToolCall:
			if ev.Call == nil {
				return
			}
			send(toolCallChunk(ev.Call, toolIndex), nil)
			toolIndex++
		case transport.EventThought:
			if s.verbose {
				vlog.Frame(stderr, "thought", ev.Text)
			}
		}
	}

	stop, err := transport.Complete(s.t, req, onEvent)
	if err != nil {
		fmt.Fprintf(stderr, "turn: %v\n", err)
		send(map[string]any{"content": "\n\n[error: " + err.Error() + "]"}, "stop")
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
		return
	}
	fmt.Fprintf(stderr, "[stop: %s]\n", stop)

	send(map[string]any{}, finishFor(stop))
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

// single accumulates the transport's reply and returns one completion JSON.
func (s *Server) single(w http.ResponseWriter, req transport.Request, model string) {
	var full strings.Builder
	var calls []map[string]any
	onEvent := func(ev transport.Event) {
		switch ev.Kind {
		case transport.EventText:
			full.WriteString(ev.Text)
		case transport.EventToolCall:
			if ev.Call == nil {
				return
			}
			calls = append(calls, map[string]any{
				"id": ev.Call.ID, "type": "function",
				"function": map[string]any{"name": ev.Call.Name, "arguments": string(ev.Call.Arguments)},
			})
		}
	}
	stop, err := transport.Complete(s.t, req, onEvent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	msg := map[string]any{"role": "assistant", "content": full.String()}
	if len(calls) > 0 {
		msg["tool_calls"] = calls
	}
	resp := map[string]any{
		"id": "chatcmpl-" + randomID(), "object": "chat.completion",
		"created": time.Now().Unix(), "model": model,
		"choices": []any{map[string]any{
			"index":         0,
			"message":       msg,
			"finish_reason": finishFor(stop),
		}},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func randomID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
