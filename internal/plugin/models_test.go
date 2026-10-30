package plugin

import (
	"1bot_test/internal/proto"
	"1bot_test/internal/transport"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type modelWriter struct{ write func([]byte) }

func (w modelWriter) Write(b []byte) (int, error) { w.write(b); return len(b), nil }
func (modelWriter) Close() error                  { return nil }
func TestHostModels(t *testing.T) {
	h := New(Options{})
	h.stdin = modelWriter{func(b []byte) {
		var req proto.Frame
		if err := json.Unmarshal(b, &req); err != nil {
			t.Fatal(err)
		}
		if req.Method != proto.MethodModels {
			t.Fatal(req.Method)
		}
		h.deliver(proto.ModelsResponse(req.ID, []transport.Model{{ID: "one"}, {ID: "two"}}, nil))
	}}
	models, err := h.ListModels(context.Background())
	if err != nil || len(models) != 2 {
		t.Fatalf("%+v %v", models, err)
	}
	if len(h.pending) != 0 {
		t.Fatal("pending request leaked")
	}
}
func TestHostModelsTimeoutAndDisconnect(t *testing.T) {
	for _, disconnect := range []bool{false, true} {
		h := New(Options{})
		h.stdin = modelWriter{func([]byte) {}}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		if disconnect {
			close(h.closed)
		}
		_, err := h.ListModels(ctx)
		cancel()
		if err == nil || (!disconnect && !errors.Is(err, context.DeadlineExceeded)) {
			t.Fatal(err)
		}
		if len(h.pending) != 0 {
			t.Fatal("pending request leaked")
		}
	}
}
