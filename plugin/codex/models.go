package codex

import (
	"1bot_test/internal/transport"
	"encoding/json"
	"fmt"
)

func (c *Client) Models() ([]transport.Model, error) {
	return collectModels(func(cursor *string) (json.RawMessage, error) {
		response, err := c.call("model/list", map[string]any{"cursor": cursor, "limit": 100})
		if response.Error != nil && response.Error.Code == -32601 {
			return nil, transport.ErrModelsUnsupported
		}
		return response.Result, err
	})
}

func collectModels(fetch func(*string) (json.RawMessage, error)) ([]transport.Model, error) {
	models := []transport.Model{}
	seen := map[string]bool{}
	var cursor *string
	for page := 0; page < 100; page++ {
		raw, err := fetch(cursor)
		if err != nil {
			return nil, err
		}
		var result struct {
			Data []struct {
				ID    string `json:"id"`
				Model string `json:"model"`
				Name  string `json:"displayName"`
			} `json:"data"`
			NextCursor *string `json:"nextCursor"`
		}
		if err := json.Unmarshal(raw, &result); err != nil {
			return nil, err
		}
		if result.Data == nil {
			return nil, fmt.Errorf("model/list: missing data")
		}
		for _, m := range result.Data {
			id := m.Model
			if id == "" {
				id = m.ID
			}
			models = append(models, transport.Model{ID: id, Name: m.Name})
		}
		cursor = result.NextCursor
		if cursor == nil || *cursor == "" {
			return transport.NormalizeModels(models)
		}
		if seen[*cursor] {
			return nil, fmt.Errorf("model/list: repeated cursor")
		}
		seen[*cursor] = true
	}
	return nil, fmt.Errorf("model/list: pagination limit exceeded")
}
