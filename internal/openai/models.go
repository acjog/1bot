package openai

import (
	"1bot_test/internal/transport"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

func (s *Server) models(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	fail := func(status int, code string) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"message": code, "type": "server_error", "code": code}})
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		fail(http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	lister, ok := s.t.(transport.ModelLister)
	if !ok {
		fail(http.StatusNotImplemented, "models_not_supported")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	models, err := lister.ListModels(ctx)
	if err == nil {
		models, err = transport.NormalizeModels(models)
	}
	if err != nil {
		switch {
		case errors.Is(err, transport.ErrModelsUnsupported):
			fail(http.StatusNotImplemented, "models_not_supported")
		case errors.Is(err, context.DeadlineExceeded):
			fail(http.StatusGatewayTimeout, "models_query_timeout")
		default:
			fail(http.StatusBadGateway, "models_query_failed")
		}
		return
	}
	data := make([]map[string]any, 0, len(models))
	for _, m := range models {
		data = append(data, map[string]any{"id": m.ID, "object": "model", "created": 0, "owned_by": s.owner})
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"object": "list", "data": data})
}
