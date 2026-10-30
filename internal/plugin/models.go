package plugin

import (
	"1bot_test/internal/proto"
	"1bot_test/internal/transport"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func (h *Host) ListModels(ctx context.Context) ([]transport.Model, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	h.pendMu.Lock()
	h.nextID++
	id := fmt.Sprintf("%d", h.nextID)
	ch := make(chan proto.Frame, 1)
	h.pending[id] = ch
	h.pendMu.Unlock()
	defer func() { h.pendMu.Lock(); delete(h.pending, id); h.pendMu.Unlock() }()
	if err := h.write(proto.Frame{JSONRPC: "2.0", ID: json.RawMessage(id), Method: proto.MethodModels}); err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-h.closed:
		return nil, fmt.Errorf("models: plugin disconnected")
	case f := <-ch:
		if f.Error != nil {
			if f.Error.Code == -32601 {
				return nil, transport.ErrModelsUnsupported
			}
			return nil, fmt.Errorf("models: plugin error %d: %s", f.Error.Code, f.Error.Message)
		}
		var result proto.ModelsResult
		if err := json.Unmarshal(f.Result, &result); err != nil {
			return nil, fmt.Errorf("models: %w", err)
		}
		if result.Models == nil {
			return nil, fmt.Errorf("models: missing catalog")
		}
		return transport.NormalizeModels(result.Models)
	}
}
