// Package openai exposes OpenAI-compatible inbound endpoints (Chat Completions
// and Responses) and routes each request through a transport.Transport. It owns
// only the inbound protocols and their translation; it knows nothing about how
// the transport reaches an upstream model.
package openai

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"1bot_test/internal/transport"
)

// Server is the OpenAI-compatible HTTP front.
type Server struct {
	t       transport.Transport
	model   string // default model when the request omits one
	owner   string
	verbose bool
	log     io.Writer // optional: append inbound requests (JSONL)
	logMu   sync.Mutex
	apiKey  string // empty = no authentication
}

func New(t transport.Transport, model string, verbose bool, log io.Writer, apiKey string, owner string) *Server {
	return &Server{t: t, model: model, verbose: verbose, log: log, apiKey: apiKey, owner: owner}
}

// Handler returns the HTTP routes. When an API key is configured, the chat and
// models routes require Authorization: Bearer <key>; /healthz stays open.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", s.authorize(s.chat))
	//mux.HandleFunc("/v1/responses", s.authorize(s.responses))
	mux.HandleFunc("/v1/models", s.authorize(s.models))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})
	return mux
}

// authorize gates a handler behind the configured API key. With no key
// configured it passes every request through.
func (s *Server) authorize(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.apiKey != "" && !s.checkAuth(r) {
			w.Header().Set("WWW-Authenticate", `Bearer realm="1bot"`)
			http.Error(w, "invalid or missing API key", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// checkAuth verifies an Authorization: Bearer <key> header against the
// configured key in constant time.
func (s *Server) checkAuth(r *http.Request) bool {
	const prefix = "Bearer "
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, prefix) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(auth[len(prefix):]), []byte(s.apiKey)) == 1
}

func (s *Server) chat(w http.ResponseWriter, r *http.Request) {
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

	var req chatRequest
	if err := json.Unmarshal(body, &req); err != nil {
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
	request := transport.Request{Session: sessionKey(messages), Messages: messages, Model: model, ToolChoice: req.ToolChoice}
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

// record appends one inbound request to the log writer (JSONL). It is a no-op
// when no log is configured. The body is compacted so a pretty-printed request
// still occupies exactly one line.
func (s *Server) record(body []byte) {
	if s.log == nil {
		return
	}
	if compact := new(bytes.Buffer); json.Compact(compact, body) == nil {
		body = compact.Bytes()
	}
	line, _ := json.Marshal(map[string]any{
		"ts":      time.Now().Format(time.RFC3339Nano),
		"request": json.RawMessage(body),
	})
	s.logMu.Lock()
	_, _ = fmt.Fprintf(s.log, "%s\n", line)
	s.logMu.Unlock()
}
