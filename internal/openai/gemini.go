package openai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"1bot_test/internal/transport"
)

type geminiPart struct {
	Text         *string `json:"text"`
	FunctionCall *struct {
		ID   string          `json:"id"`
		Name string          `json:"name"`
		Args json.RawMessage `json:"args"`
	} `json:"functionCall"`
	FunctionResponse *struct {
		ID       string          `json:"id"`
		Name     string          `json:"name"`
		Response json.RawMessage `json:"response"`
	} `json:"functionResponse"`
}
type geminiContent struct {
	Role  string            `json:"role"`
	Parts []json.RawMessage `json:"parts"`
}
type geminiRequest struct {
	Contents          []geminiContent `json:"contents"`
	SystemInstruction *geminiContent  `json:"systemInstruction"`
	Tools             []struct {
		FunctionDeclarations []struct {
			Name                 string          `json:"name"`
			Description          string          `json:"description"`
			Parameters           json.RawMessage `json:"parameters"`
			ParametersJSONSchema json.RawMessage `json:"parametersJsonSchema"`
		} `json:"functionDeclarations"`
	} `json:"tools"`
	ToolConfig struct {
		FunctionCallingConfig struct {
			Mode    string   `json:"mode"`
			Allowed []string `json:"allowedFunctionNames"`
		} `json:"functionCallingConfig"`
	} `json:"toolConfig"`
	GenerationConfig json.RawMessage `json:"generationConfig"`
}

func parseGemini(in geminiRequest, model string) (transport.Request, error) {
	req := transport.Request{Model: model}
	if len(in.Contents) == 0 {
		return req, fmt.Errorf("contents must not be empty")
	}
	if len(in.GenerationConfig) > 0 && string(in.GenerationConfig) != "null" && string(in.GenerationConfig) != "{}" {
		return req, fmt.Errorf("generationConfig is not supported by the plugin transport")
	}
	contents := in.Contents
	if in.SystemInstruction != nil {
		system := *in.SystemInstruction
		system.Role = "system"
		contents = append([]geminiContent{system}, contents...)
	}
	pending := map[string]string{}
	for mi, c := range contents {
		role := c.Role
		if role == "" {
			role = "user"
		}
		if role == "model" {
			role = "assistant"
		}
		if role != "user" && role != "assistant" && !(mi == 0 && in.SystemInstruction != nil && role == "system") {
			return req, fmt.Errorf("unsupported content role")
		}
		msg := transport.Message{Role: role}
		if len(c.Parts) == 0 {
			return req, fmt.Errorf("parts must not be empty")
		}
		for pi, raw := range c.Parts {
			var fields object
			if json.Unmarshal(raw, &fields) != nil || len(fields) != 1 {
				return req, fmt.Errorf("each part must contain text, functionCall or functionResponse only")
			}
			var p geminiPart
			if json.Unmarshal(raw, &p) != nil {
				return req, fmt.Errorf("invalid content part")
			}
			switch {
			case p.Text != nil:
				msg.Parts = append(msg.Parts, transport.Part{Kind: transport.PartText, Text: *p.Text})
			case p.FunctionCall != nil:
				call := p.FunctionCall
				if role != "assistant" || call.Name == "" {
					return req, fmt.Errorf("invalid functionCall")
				}
				args, err := jsonObject(call.Args)
				if err != nil {
					return req, err
				}
				data, _ := json.Marshal(args)
				id := call.ID
				if id == "" {
					id = fmt.Sprintf("gemini_%d_%d", mi, pi)
				}
				if pending[id] != "" {
					return req, fmt.Errorf("duplicate functionCall id")
				}
				pending[id] = call.Name
				msg.Parts = append(msg.Parts, transport.Part{Kind: transport.PartToolCall, Call: &transport.ToolCall{ID: id, Name: call.Name, Arguments: data}})
			case p.FunctionResponse != nil:
				result := p.FunctionResponse
				if role != "user" || result.Name == "" {
					return req, fmt.Errorf("invalid functionResponse")
				}
				id := result.ID
				if id == "" {
					for candidate, name := range pending {
						if name == result.Name {
							if id != "" {
								return req, fmt.Errorf("ambiguous functionResponse: provide an id")
							}
							id = candidate
						}
					}
				}
				if id == "" || pending[id] != result.Name {
					return req, fmt.Errorf("functionResponse has no matching functionCall")
				}
				value, err := jsonObject(result.Response)
				if err != nil {
					return req, err
				}
				data, _ := json.Marshal(value)
				delete(pending, id)
				msg.Parts = append(msg.Parts, transport.Part{Kind: transport.PartToolResult, Result: &transport.ToolResult{CallID: id, Content: string(data)}})
			default:
				return req, fmt.Errorf("unsupported content part; images, audio and thought signatures are not supported")
			}
		}
		req.Messages = append(req.Messages, msg)
	}
	for _, group := range in.Tools {
		if len(group.FunctionDeclarations) == 0 {
			return req, fmt.Errorf("only functionDeclarations tools are supported")
		}
		for _, t := range group.FunctionDeclarations {
			schema := t.Parameters
			if len(t.ParametersJSONSchema) > 0 {
				if len(schema) > 0 {
					return req, fmt.Errorf("use only one parameters schema")
				}
				schema = t.ParametersJSONSchema
			}
			if len(schema) > 0 {
				var value any
				if json.Unmarshal(schema, &value) != nil {
					return req, fmt.Errorf("invalid tool schema")
				}
				normalizeGeminiSchema(value)
				schema, _ = json.Marshal(value)
			}
			req.Tools = append(req.Tools, transport.Tool{Name: t.Name, Description: t.Description, Parameters: schema})
		}
	}
	config := in.ToolConfig.FunctionCallingConfig
	if len(config.Allowed) > 0 {
		return req, fmt.Errorf("allowedFunctionNames is not supported")
	}
	switch config.Mode {
	case "", "AUTO":
		req.ToolChoice = json.RawMessage(`"auto"`)
	case "NONE":
		req.ToolChoice = json.RawMessage(`"none"`)
	default:
		return req, fmt.Errorf("function calling mode supports AUTO or NONE only")
	}
	return req, validateNative(&req)
}

func normalizeGeminiSchema(value any) {
	switch v := value.(type) {
	case map[string]any:
		if t, ok := v["type"].(string); ok {
			v["type"] = strings.ToLower(t)
		}
		if props, ok := v["properties"].(map[string]any); ok {
			for _, p := range props {
				normalizeGeminiSchema(p)
			}
		}
		if items, ok := v["items"]; ok {
			normalizeGeminiSchema(items)
		}
	case []any:
		for _, p := range v {
			normalizeGeminiSchema(p)
		}
	}
}

func (s *Server) gemini(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1beta/models/")
	if path == r.URL.Path {
		path = strings.TrimPrefix(r.URL.Path, "/v1/models/")
	}
	colon := strings.LastIndex(path, ":")
	if colon <= 0 {
		nativeError(w, "gemini", 404, "unknown model operation")
		return
	}
	model, operation := path[:colon], path[colon+1:]
	stream := operation == "streamGenerateContent"
	if !stream && operation != "generateContent" {
		nativeError(w, "gemini", 404, "unknown model operation")
		return
	}
	if stream && r.URL.Query().Get("alt") != "sse" {
		nativeError(w, "gemini", 400, "streamGenerateContent requires alt=sse")
		return
	}
	var in geminiRequest
	if !s.nativeBody(w, r, "gemini", &in) {
		return
	}
	req, err := parseGemini(in, model)
	if err != nil {
		nativeError(w, "gemini", 400, err.Error())
		return
	}
	req.Session = nativeSession("gemini", req)
	parts := []object{}
	response := func(p []object, finish bool) object {
		candidate := object{"index": 0, "content": object{"role": "model", "parts": p}}
		if finish {
			candidate["finishReason"] = "STOP"
		}
		return object{"candidates": []object{candidate}, "modelVersion": model}
	}
	if stream {
		startNativeStream(w)
	}
	var eventErr error
	_, err = transport.Complete(s.t, req, func(e transport.Event) {
		if eventErr != nil || r.Context().Err() != nil {
			return
		}
		var part object
		switch e.Kind {
		case transport.EventText:
			if e.Text == "" {
				return
			}
			part = object{"text": e.Text}
		case transport.EventToolCall:
			if e.Call == nil || e.Call.Name == "" || e.Call.ID == "" {
				eventErr = fmt.Errorf("invalid upstream tool call")
				return
			}
			args, err := jsonObject(e.Call.Arguments)
			if err != nil {
				eventErr = err
				return
			}
			part = object{"functionCall": object{"id": e.Call.ID, "name": e.Call.Name, "args": args}}
		default:
			return
		}
		parts = append(parts, part)
		if stream {
			nativeSSE(w, "", response([]object{part}, false))
		}
	})
	if err == nil {
		err = eventErr
	}
	if err != nil {
		if stream {
			nativeSSE(w, "", object{"error": object{"code": 502, "status": "INTERNAL", "message": "upstream request failed"}})
		} else {
			nativeError(w, "gemini", 502, "upstream request failed")
		}
		return
	}
	if r.Context().Err() != nil {
		return
	}
	if stream {
		nativeSSE(w, "", response([]object{}, true))
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response(parts, true))
	}
}
