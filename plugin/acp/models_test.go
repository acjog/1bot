package acp

import (
	"1bot_test/internal/transport"
	"errors"
	"testing"
)

func TestSessionModels(t *testing.T) {
	for _, tc := range []struct {
		raw         string
		count       int
		unsupported bool
		invalid     bool
	}{
		{`{"models":{"availableModels":[{"modelId":"a","name":"A"},{"modelId":"b"},{"modelId":"a"}]}}`, 2, false, false},
		{`{"models":{"availableModels":[]}}`, 0, false, false},
		{`{"sessionId":"s"}`, 0, true, false},
		{`{"models":{"availableModels":[{"name":"missing id"}]}}`, 0, false, true},
	} {
		models, err := parseSessionModels([]byte(tc.raw))
		if tc.unsupported {
			if !errors.Is(err, transport.ErrModelsUnsupported) {
				t.Fatal(err)
			}
			continue
		}
		if tc.invalid {
			if err == nil {
				t.Fatal("expected error")
			}
			continue
		}
		if err != nil || len(models) != tc.count {
			t.Fatalf("%v %+v", err, models)
		}
	}
}
