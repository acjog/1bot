package transport

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrModelsUnsupported = errors.New("plugin does not expose a model catalog")

type Model struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}
type ModelLister interface {
	ListModels(context.Context) ([]Model, error)
}

func NormalizeModels(models []Model) ([]Model, error) {
	result := make([]Model, 0, len(models))
	seen := map[string]bool{}
	for _, model := range models {
		if strings.TrimSpace(model.ID) == "" {
			return nil, fmt.Errorf("model catalog contains an empty ID")
		}
		if !seen[model.ID] {
			result = append(result, model)
			seen[model.ID] = true
		}
	}
	return result, nil
}
