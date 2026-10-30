package proto

import (
	"1bot_test/internal/transport"
	"encoding/json"
	"errors"
)

const MethodModels = "models"

type ModelsResult struct {
	Models []transport.Model `json:"models"`
}

func ModelsResponse(id json.RawMessage, models []transport.Model, err error) Frame {
	f := Frame{JSONRPC: "2.0", ID: id}
	if err == nil {
		models, err = transport.NormalizeModels(models)
	}
	if err != nil {
		code := -32000
		if errors.Is(err, transport.ErrModelsUnsupported) {
			code = -32601
		}
		f.Error = &Error{Code: code, Message: err.Error()}
	} else {
		f.Result, _ = json.Marshal(ModelsResult{Models: models})
	}
	return f
}
