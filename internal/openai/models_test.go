package openai

import (
	"1bot_test/internal/transport"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
)

type modelTransport struct {
	models []transport.Model
	err    error
}

func (*modelTransport) Start() error { return nil }
func (*modelTransport) Close() error { return nil }
func (*modelTransport) Turn([]transport.Message, func(transport.Event)) (string, error) {
	return "", nil
}
func (m *modelTransport) ListModels(context.Context) ([]transport.Model, error) {
	return m.models, m.err
}

func TestModels(t *testing.T) {
	for _, tc := range []struct {
		name, method, key string
		models            []transport.Model
		err               error
		status, count     int
	}{
		{"catalog", "GET", "secret", []transport.Model{{ID: "a"}, {ID: "b"}, {ID: "a"}}, nil, 200, 2},
		{"empty", "GET", "secret", nil, nil, 200, 0},
		{"unsupported", "GET", "secret", nil, transport.ErrModelsUnsupported, 501, 0},
		{"failure", "GET", "secret", nil, errors.New("backend error"), 502, 0},
		{"timeout", "GET", "secret", nil, context.DeadlineExceeded, 504, 0},
		{"invalid", "GET", "secret", []transport.Model{{ID: ""}}, nil, 502, 0},
		{"auth", "GET", "wrong", nil, nil, 401, 0},
		{"method", "POST", "secret", nil, nil, 405, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := New(&modelTransport{tc.models, tc.err}, "configured-not-catalog", false, nil, "secret", "codex")
			r := httptest.NewRequest(tc.method, "/v1/models", nil)
			r.Header.Set("Authorization", "Bearer "+tc.key)
			w := httptest.NewRecorder()
			srv.Handler().ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("status %d: %s", w.Code, w.Body.String())
			}
			if tc.status != 200 {
				return
			}
			var result struct {
				Object string
				Data   []struct {
					ID     string
					Object string
					Owner  string `json:"owned_by"`
				}
			}
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Object != "list" || result.Data == nil || len(result.Data) != tc.count {
				t.Fatalf("bad catalog: %s", w.Body.String())
			}
			for _, m := range result.Data {
				if m.ID == "configured-not-catalog" || m.Object != "model" || m.Owner != "codex" {
					t.Fatalf("bad model: %+v", m)
				}
			}
		})
	}
}
