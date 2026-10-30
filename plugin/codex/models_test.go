package codex

import (
	"encoding/json"
	"testing"
)

func TestModelsPagination(t *testing.T) {
	calls := 0
	models, err := collectModels(func(cursor *string) (json.RawMessage, error) {
		calls++
		if calls == 1 {
			if cursor != nil {
				t.Fatal("unexpected initial cursor")
			}
			return json.RawMessage(`{"data":[{"id":"internal","model":"a","displayName":"A"}],"nextCursor":"next"}`), nil
		}
		if cursor == nil || *cursor != "next" {
			t.Fatal("missing cursor")
		}
		return json.RawMessage(`{"data":[{"id":"b","model":"b"},{"model":"a"}],"nextCursor":null}`), nil
	})
	if err != nil || len(models) != 2 || models[0].ID != "a" || models[1].ID != "b" {
		t.Fatalf("%+v %v", models, err)
	}
}
func TestModelsInvalidPage(t *testing.T) {
	for _, raw := range []string{`{}`, `{"data":[{}]}`, `{"data":[],"nextCursor":"repeat"}`} {
		_, err := collectModels(func(*string) (json.RawMessage, error) { return json.RawMessage(raw), nil })
		if err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
