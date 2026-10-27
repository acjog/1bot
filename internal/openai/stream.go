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
)

var stderr = os.Stderr

// stream writes the transport's reply as an OpenAI SSE chunk stream.
func (s *Server) stream(w http.ResponseWriter, prompt, model string) {
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

	onEvent := func(ev transport.Event) {
		switch ev.Kind {
		case transport.EventText:
			send(map[string]any{"content": ev.Text}, nil)
		case transport.EventTool:
			fmt.Fprintf(stderr, "[tool] %s\n", ev.Tool)
		case transport.EventThought:
			if s.verbose {
				fmt.Fprintf(stderr, "[thought] %s", ev.Text)
			}
		}
	}

	stop, err := s.t.Turn(prompt, onEvent)
	if err != nil {
		fmt.Fprintf(stderr, "turn: %v\n", err)
		send(map[string]any{"content": "\n\n[error: " + err.Error() + "]"}, "stop")
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
		return
	}
	fmt.Fprintf(stderr, "[stop: %s]\n", stop)

	send(map[string]any{}, "stop")
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

// single accumulates the transport's reply and returns one completion JSON.
func (s *Server) single(w http.ResponseWriter, prompt, model string) {
	var full strings.Builder
	onEvent := func(ev transport.Event) {
		if ev.Kind == transport.EventText {
			full.WriteString(ev.Text)
		}
	}
	_, err := s.t.Turn(prompt, onEvent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]any{
		"id": "chatcmpl-" + randomID(), "object": "chat.completion",
		"created": time.Now().Unix(), "model": model,
		"choices": []any{map[string]any{
			"index": 0,
			"message": map[string]any{
				"role":    "assistant",
				"content": full.String(),
			},
			"finish_reason": "stop",
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
