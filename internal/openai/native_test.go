package openai

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"1bot_test/internal/transport"
)

type nativeTransport struct {
	request transport.Request
	events  []transport.Event
	err     error
	calls   int
}

func (*nativeTransport) Start() error { return nil }
func (*nativeTransport) Close() error { return nil }
func (t *nativeTransport) Turn([]transport.Message, func(transport.Event)) (string, error) {
	panic("Complete expected")
}
func (t *nativeTransport) Complete(req transport.Request, emit func(transport.Event)) (string, error) {
	t.request = req
	t.calls++
	for _, e := range t.events {
		emit(e)
	}
	return transport.StopEndTurn, t.err
}
func nativePost(t *nativeTransport, path, body, header, key string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set(header, key)
	w := httptest.NewRecorder()
	New(t, "default", false, nil, "secret", "test").Handler().ServeHTTP(w, req)
	return w
}

func TestFourInboundRoutes(t *testing.T) {
	for _, tc := range []struct{ path, body, header, key, contains string }{
		{"/v1/chat/completions", `{"messages":[{"role":"user","content":"hi"}]}`, "Authorization", "Bearer secret", "chat.completion"},
		{"/v1/responses", `{"input":"hi"}`, "Authorization", "Bearer secret", "output_text"},
		{"/v1/messages", `{"model":"m","max_tokens":100,"messages":[{"role":"user","content":"hi"}]}`, "x-api-key", "secret", "end_turn"},
		{"/v1beta/models/m:generateContent", `{"contents":[{"role":"user","parts":[{"text":"hi"}]}]}`, "x-goog-api-key", "secret", "candidates"},
		{"/v1/models/m:generateContent?key=secret", `{"contents":[{"parts":[{"text":"hi"}]}]}`, "", "", "candidates"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			tr := &nativeTransport{events: []transport.Event{{Kind: transport.EventText, Text: "hello"}}}
			w := nativePost(tr, tc.path, tc.body, tc.header, tc.key)
			if w.Code != 200 || tr.calls != 1 || !strings.Contains(w.Body.String(), tc.contains) {
				t.Fatalf("%d %s", w.Code, w.Body)
			}
			if !json.Valid(w.Body.Bytes()) {
				t.Fatalf("invalid JSON: %s", w.Body)
			}
		})
	}
}

func TestNativeTools(t *testing.T) {
	var a anthropicRequest
	if err := json.Unmarshal([]byte(`{"system":[{"type":"text","text":"system"}],"messages":[{"role":"assistant","content":[{"type":"tool_use","id":"c1","name":"weather","input":{"city":"Paris"}}]},{"role":"user","content":[{"type":"tool_result","tool_use_id":"c1","content":[{"type":"text","text":"sunny"}],"is_error":true}]}],"tools":[{"name":"weather","input_schema":{"type":"object"}}],"tool_choice":{"type":"none"}}`), &a); err != nil {
		t.Fatal(err)
	}
	req, err := parseAnthropic(a)
	if err != nil {
		t.Fatal(err)
	}
	if req.Messages[0].Role != "system" || req.Messages[2].Parts[0].Result.CallID != "c1" || !req.Messages[2].Parts[0].Result.IsError || string(req.ToolChoice) != `"none"` {
		t.Fatalf("%+v", req)
	}
	var g geminiRequest
	if err := json.Unmarshal([]byte(`{"systemInstruction":{"parts":[{"text":"system"}]},"contents":[{"role":"model","parts":[{"functionCall":{"id":"c1","name":"weather","args":{"city":"Paris"}}}]},{"role":"user","parts":[{"functionResponse":{"name":"weather","response":{"result":"sunny"}}}]}],"tools":[{"functionDeclarations":[{"name":"weather","parameters":{"type":"OBJECT","properties":{"city":{"type":"STRING"}}}}]}]}`), &g); err != nil {
		t.Fatal(err)
	}
	req, err = parseGemini(g, "m")
	if err != nil {
		t.Fatal(err)
	}
	if req.Messages[1].Role != "assistant" || req.Messages[2].Parts[0].Result.CallID != "c1" || strings.Contains(string(req.Tools[0].Parameters), "OBJECT") {
		t.Fatalf("%+v", req)
	}
}

func TestNativeStreams(t *testing.T) {
	for _, protocol := range []string{"anthropic", "gemini"} {
		t.Run(protocol, func(t *testing.T) {
			tr := &nativeTransport{events: []transport.Event{
				{Kind: transport.EventToolCall, Call: &transport.ToolCall{ID: "c1", Name: "weather", Arguments: json.RawMessage(`{"city":"Paris"}`)}},
				{Kind: transport.EventText, Text: "hello"},
			}}
			path, body, header := "/v1/messages", `{"stream":true,"messages":[{"role":"user","content":"hi"}]}`, "x-api-key"
			if protocol == "gemini" {
				path = "/v1beta/models/m:streamGenerateContent?alt=sse"
				body = `{"contents":[{"parts":[{"text":"hi"}]}]}`
				header = "x-goog-api-key"
			}
			w := nativePost(tr, path, body, header, "secret")
			if w.Code != 200 || w.Header().Get("Content-Type") != "text/event-stream" {
				t.Fatalf("%d %s", w.Code, w.Body)
			}
			for _, line := range strings.Split(w.Body.String(), "\n") {
				if strings.HasPrefix(line, "data: ") && !json.Valid([]byte(strings.TrimPrefix(line, "data: "))) {
					t.Fatal(line)
				}
			}
			output := w.Body.String()
			if protocol == "anthropic" {
				for _, part := range []string{"event: message_start", "event: content_block_start", `"index":0`, `"index":1`, "input_json_delta", "tool_use", "event: message_stop"} {
					if !strings.Contains(output, part) {
						t.Fatalf("missing %s: %s", part, output)
					}
				}
			} else if !strings.Contains(output, "functionCall") || !strings.Contains(output, `"finishReason":"STOP"`) {
				t.Fatal(output)
			}
		})
	}
}

func TestNativeValidationAndAuth(t *testing.T) {
	for _, tc := range []struct {
		path, body, header, key string
		code                    int
	}{
		{"/v1/messages", `{}`, "x-api-key", "bad", 401},
		{"/v1beta/models/m:generateContent", `{}`, "x-goog-api-key", "bad", 401},
		{"/v1/messages", `{`, "x-api-key", "secret", 400},
		{"/v1/messages", `{"messages":[]}`, "x-api-key", "secret", 400},
		{"/v1/messages", `{"messages":[{"role":"user","content":[{"type":"image"}]}]}`, "x-api-key", "secret", 400},
		{"/v1/messages", `{"messages":[{"role":"user","content":"hi"}],"tool_choice":{"type":"any"}}`, "x-api-key", "secret", 400},
		{"/v1beta/models/m:generateContent", `{"contents":[{"parts":[{"inlineData":{}}]}]}`, "x-goog-api-key", "secret", 400},
		{"/v1beta/models/m:generateContent", `{"contents":[{"parts":[{"functionResponse":{"name":"x","response":{}}}]}]}`, "x-goog-api-key", "secret", 400},
		{"/v1beta/models/m:streamGenerateContent", `{}`, "x-goog-api-key", "secret", 400},
		{"/v1beta/models/m:unknown", `{}`, "x-goog-api-key", "secret", 404},
	} {
		tr := &nativeTransport{}
		w := nativePost(tr, tc.path, tc.body, tc.header, tc.key)
		if w.Code != tc.code || tr.calls != 0 || !json.Valid(w.Body.Bytes()) {
			t.Fatalf("%s: %d %s", tc.path, w.Code, w.Body)
		}
	}
	for _, path := range []string{"/v1/messages", "/v1beta/models/m:generateContent"} {
		r := httptest.NewRequest("GET", path, nil)
		r.Header.Set("Authorization", "Bearer secret")
		w := httptest.NewRecorder()
		New(&nativeTransport{}, "", false, nil, "secret", "").Handler().ServeHTTP(w, r)
		if w.Code != 405 {
			t.Fatal(w.Code)
		}
	}
}

func TestNativeUpstreamFailure(t *testing.T) {
	for _, stream := range []bool{false, true} {
		for _, protocol := range []string{"anthropic", "gemini"} {
			tr := &nativeTransport{err: errors.New("private upstream detail")}
			path, body, header := "/v1/messages", `{"messages":[{"role":"user","content":"hi"}],"stream":`+fmtBool(stream)+`}`, "x-api-key"
			if protocol == "gemini" {
				path = "/v1beta/models/m:generateContent"
				if stream {
					path = "/v1beta/models/m:streamGenerateContent?alt=sse"
				}
				body = `{"contents":[{"parts":[{"text":"hi"}]}]}`
				header = "x-goog-api-key"
			}
			w := nativePost(tr, path, body, header, "secret")
			if (!stream && w.Code != 502) || !strings.Contains(w.Body.String(), "error") || strings.Contains(w.Body.String(), "private upstream") || strings.Contains(w.Body.String(), "message_stop") {
				t.Fatalf("%d %s", w.Code, w.Body)
			}
		}
	}
}
func fmtBool(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
