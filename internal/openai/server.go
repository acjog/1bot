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
	model   string // startup model; echoed, never switched
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

	if req.Stream {
		s.stream(w, messages, model)
		return
	}
	s.single(w, messages, model)
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
