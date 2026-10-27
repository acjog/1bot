package openai

import (
	"encoding/json"
	"fmt"
	"strings"
)

type chatMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// buildPrompt flattens an OpenAI conversation into one text prompt for the
// transport. The transport (ACP) takes a single user message, so history is
// carried as labelled text rather than as real message roles.
func buildPrompt(msgs []chatMessage) string {
	var b strings.Builder
	for _, m := range msgs {
		content := strings.TrimSpace(contentText(m.Content))
		if content == "" {
			continue
		}
		role := m.Role
		if role == "" {
			role = "user"
		}
		fmt.Fprintf(&b, "%s: %s\n\n", role, content)
	}
	return strings.TrimSpace(b.String())
}

// contentText extracts the text from a content field that is either a plain
// string or an array of {type, text} parts (multimodal).
func contentText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &parts) == nil {
		var b strings.Builder
		for _, p := range parts {
			if p.Type == "text" || p.Type == "" {
				b.WriteString(p.Text)
			}
		}
		return b.String()
	}
	return ""
}
