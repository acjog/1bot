package openai

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"1bot_test/internal/transport"
)

type object = map[string]any

func nativeError(w http.ResponseWriter, protocol string, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if protocol == "anthropic" {
		kind := "invalid_request_error"
		if code == 401 {
			kind = "authentication_error"
		}
		if code >= 500 {
			kind = "api_error"
		}
		json.NewEncoder(w).Encode(object{"type": "error", "error": object{"type": kind, "message": message}})
	} else {
		status := "INVALID_ARGUMENT"
		if code == 401 {
			status = "UNAUTHENTICATED"
		}
		if code == 404 {
			status = "NOT_FOUND"
		}
		if code == 405 {
			status = "METHOD_NOT_ALLOWED"
		}
		if code >= 500 {
			status = "INTERNAL"
		}
		json.NewEncoder(w).Encode(object{"error": object{"code": code, "status": status, "message": message}})
	}
}

func (s *Server) nativeHandler(protocol string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("x-api-key")
		if protocol == "gemini" {
			key = r.Header.Get("x-goog-api-key")
			if key == "" {
				key = r.URL.Query().Get("key")
			}
		}
		if s.apiKey != "" && !s.checkAuth(r) && subtle.ConstantTimeCompare([]byte(key), []byte(s.apiKey)) != 1 {
			nativeError(w, protocol, 401, "invalid or missing API key")
			return
		}
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			nativeError(w, protocol, 405, "POST only")
			return
		}
		next(w, r)
	}
}

func (s *Server) nativeBody(w http.ResponseWriter, r *http.Request, protocol string, dst any) bool {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 8<<20))
	if err != nil {
		nativeError(w, protocol, 400, "request body exceeds limit or cannot be read")
		return false
	}
	if err = json.Unmarshal(body, dst); err != nil {
		nativeError(w, protocol, 400, "invalid JSON request")
		return false
	}
	s.record(body)
	return true
}

func validateNative(req *transport.Request) error {
	if len(req.Messages) == 0 {
		return fmt.Errorf("messages must not be empty")
	}
	seen := map[string]bool{}
	for i := range req.Tools {
		t := &req.Tools[i]
		if t.Name == "" || seen[t.Name] {
			return fmt.Errorf("tool names must be nonempty and unique")
		}
		seen[t.Name] = true
		if len(t.Parameters) == 0 {
			t.Parameters = json.RawMessage(`{"type":"object","properties":{}}`)
		}
		var schema object
		if json.Unmarshal(t.Parameters, &schema) != nil || schema == nil {
			return fmt.Errorf("tool schema must be an object")
		}
	}
	return nil
}

func jsonObject(raw json.RawMessage) (object, error) {
	var value object
	if len(raw) == 0 {
		return object{}, nil
	}
	if err := json.Unmarshal(raw, &value); err != nil || value == nil {
		return nil, fmt.Errorf("tool arguments must be an object")
	}
	return value, nil
}

func nativeSSE(w http.ResponseWriter, event string, value any) {
	data, _ := json.Marshal(value)
	if event != "" {
		fmt.Fprintf(w, "event: %s\n", event)
	}
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func startNativeStream(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
}

func nativeSession(protocol string, req transport.Request) string {
	return protocol + ":" + req.Model + ":" + sessionKey(req.Messages)
}

func nativeText(raw json.RawMessage) (string, error) {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text, nil
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &blocks) != nil || blocks == nil {
		return "", fmt.Errorf("content must be text or text blocks")
	}
	var out strings.Builder
	for _, b := range blocks {
		if b.Type != "text" {
			return "", fmt.Errorf("unsupported content block: %s", b.Type)
		}
		out.WriteString(b.Text)
	}
	return out.String(), nil
}
