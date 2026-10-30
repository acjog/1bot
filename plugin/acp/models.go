package acp

import (
	"1bot_test/internal/transport"
	"encoding/json"
)

func (c *Client) Models() ([]transport.Model, error) {
	c.promptMu.Lock()
	defer c.promptMu.Unlock()
	if c.sessionID == "" {
		if err := c.newSession(); err != nil {
			return nil, err
		}
	}
	return append([]transport.Model{}, c.modelCatalog...), c.modelCatalogErr
}

func parseSessionModels(raw json.RawMessage) ([]transport.Model, error) {
	var response struct {
		Models *struct {
			Available []struct {
				ID   string `json:"modelId"`
				Name string `json:"name"`
			} `json:"availableModels"`
		} `json:"models"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, err
	}
	if response.Models == nil || response.Models.Available == nil {
		return nil, transport.ErrModelsUnsupported
	}
	models := make([]transport.Model, 0, len(response.Models.Available))
	for _, m := range response.Models.Available {
		models = append(models, transport.Model{ID: m.ID, Name: m.Name})
	}
	return transport.NormalizeModels(models)
}
