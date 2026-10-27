// Package openai exposes an OpenAI-compatible Chat Completions endpoint and
// routes each request through a transport.Transport. It owns only the inbound
// protocol and its translation; it knows nothing about how the transport
// reaches an upstream model.
package openai

import (
	"encoding/json"
	"fmt"
	"net/http"

	"1bot_test/internal/transport"
)

// Server is the OpenAI-compatible HTTP front.
type Server struct {
	t       transport.Transport
	model   string // default model when the request omits one
	verbose bool
}

func New(t transport.Transport, model string, verbose bool) *Server {
	return &Server{t: t, model: model, verbose: verbose}
}

// Handler returns the HTTP routes.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", s.chat)
	mux.HandleFunc("/v1/models", s.models)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})
	return mux
}

func (s *Server) chat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	messages := parseMessages(req.Messages)
	if len(messages) == 0 {
		http.Error(w, "no message content", http.StatusBadRequest)
		return
	}
	model := req.Model
	if model == "" {
		model = s.model
	}
	request := transport.Request{Messages: messages, Model: model, ToolChoice: req.ToolChoice}
	if len(req.ToolChoice) > 0 && string(req.ToolChoice) != "null" && string(req.ToolChoice) != `"auto"` && string(req.ToolChoice) != `"none"` {
		http.Error(w, "tool_choice currently supports auto or none", http.StatusBadRequest)
		return
	}
	seenTools := map[string]bool{}
	for _, tool := range req.Tools {
		if tool.Type != "function" || tool.Function.Name == "" {
			http.Error(w, "only named function tools are supported", http.StatusBadRequest)
			return
		}
		if seenTools[tool.Function.Name] {
			http.Error(w, "duplicate function tool name", http.StatusBadRequest)
			return
		}
		seenTools[tool.Function.Name] = true
		if len(tool.Function.Parameters) == 0 {
			tool.Function.Parameters = json.RawMessage(`{"type":"object","properties":{}}`)
		}
		var schema map[string]any
		if json.Unmarshal(tool.Function.Parameters, &schema) != nil || schema == nil {
			http.Error(w, "function parameters must be a JSON schema object", http.StatusBadRequest)
			return
		}
		request.Tools = append(request.Tools, tool.Function)
	}

	if req.Stream {
		s.stream(w, request, model)
		return
	}
	s.single(w, request, model)
}

func (s *Server) models(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"object": "list",
		"data": []any{map[string]any{
			"id": s.model, "object": "model", "owned_by": "codebuddy",
		}},
	})
}
