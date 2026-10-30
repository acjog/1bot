// Package zcode drives Z.AI's ZCode coding plan directly over its
// Anthropic-compatible Messages API, and reduces its streamed reply to
// transport.Event. It reads the plan's key from ZCode's own credential store
// (~/.zcode/v2) — it never launches `zcode app-server`.
//
// One turn is one HTTP request:
//
//	POST {base}/v1/messages   (base = api.z.ai/api/anthropic, or BigModel's)
//	stream SSE text deltas     -> transport.EventText
package zcode

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"1bot_test/internal/transport"
	"1bot_test/internal/vlog"
)

var stderr = os.Stderr

// Options configures the ZCode adapter.
type Options struct {
	Model   string // model id; empty = DefaultModel
	Verbose bool   // dump each request and SSE event to stderr
	APIKey  string // optional: override the key read from ZCode's store
	Base    string // optional: override the Anthropic-compatible base URL
}

// Client turns one conversation into Anthropic Messages requests. It
// implements transport.Transport.
type Client struct {
	opts Options
	cred cred
	http *http.Client

	// promptMu serialises turns: one request at a time.
	promptMu sync.Mutex
}

func New(opts Options) *Client {
	return &Client{opts: opts}
}

// Start reads ZCode's coding plan credentials. There is no child process to
// launch, so readiness is just whether a key could be found.
func (c *Client) Start() error {
	cred, err := readCred()
	if err != nil {
		return err
	}
	if c.opts.APIKey != "" {
		cred.Key = c.opts.APIKey
	}
	if c.opts.Base != "" {
		cred.Base = strings.TrimRight(c.opts.Base, "/")
	}
	if cred.Key == "" {
		return fmt.Errorf("no ZCode coding plan key")
	}
	if cred.Base == "" {
		cred.Base = ZaiBase
	}
	c.cred = cred
	c.http = &http.Client{Timeout: 5 * time.Minute}
	return nil
}

// Close releases nothing: the adapter owns no subprocess or persistent
// connection.
func (c *Client) Close() error { return nil }

// Models returns the coding plan's model catalog: the models ZCode lists for
// its built-in coding plan provider in config.json, else the default model
// alone.
func (c *Client) Models() ([]transport.Model, error) {
	if ms := readConfigModels(); len(ms) > 0 {
		return ms, nil
	}
	return []transport.Model{{ID: DefaultModel, Name: "GLM-5"}}, nil
}

// anthropicMessage is one Messages-API message. Content is sent as a plain
// string; tool parts are dropped (this adapter declares no client tools).
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Stream    bool               `json:"stream"`
}

// Turn implements transport.Transport: it posts the whole conversation to the
// Messages endpoint and streams text deltas back through onEvent.
func (c *Client) Turn(messages []transport.Message, onEvent func(transport.Event)) (string, error) {
	c.promptMu.Lock()
	defer c.promptMu.Unlock()

	model := c.opts.Model
	if model == "" {
		model = DefaultModel
	}
	system, out := toAnthropic(messages)
	if len(out) == 0 {
		return "", fmt.Errorf("no messages")
	}

	body, _ := json.Marshal(anthropicRequest{
		Model:     model,
		MaxTokens: 8192,
		System:    system,
		Messages:  out,
		Stream:    true,
	})
	if c.opts.Verbose {
		vlog.Frame(stderr, "zcode ->", string(body))
	}

	url := c.cred.Base + "/v1/messages"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.cred.Key)
	req.Header.Set("anthropic-version", "2023-06-01")

	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 64*1024))
		return "", fmt.Errorf("zcode: %s: %s", res.Status, strings.TrimSpace(string(b)))
	}

	return c.stream(res, onEvent)
}

// stream consumes Anthropic's SSE reply and forwards text deltas.
func (c *Client) stream(res *http.Response, onEvent func(transport.Event)) (string, error) {
	sc := bufio.NewScanner(res.Body)
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		if c.opts.Verbose {
			vlog.Frame(stderr, "zcode <-", data)
		}
		var ev struct {
			Type  string `json:"type"`
			Delta struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"delta"`
			Error struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal([]byte(data), &ev) != nil {
			continue
		}
		switch ev.Type {
		case "content_block_delta":
			if ev.Delta.Type == "text_delta" && ev.Delta.Text != "" {
				onEvent(transport.Event{Kind: transport.EventText, Text: ev.Delta.Text})
			}
		case "error":
			return "", fmt.Errorf("zcode: %s: %s", ev.Error.Type, ev.Error.Message)
		case "message_stop":
			return transport.StopEndTurn, nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return transport.StopEndTurn, nil
}

// toAnthropic folds system/developer text into the top-level system field and
// user/assistant text into messages, in order.
func toAnthropic(msgs []transport.Message) (system string, out []anthropicMessage) {
	for _, m := range msgs {
		var sb strings.Builder
		for _, p := range m.Parts {
			if p.Kind == transport.PartText {
				sb.WriteString(p.Text)
			}
		}
		text := sb.String()
		switch m.Role {
		case "system", "developer":
			if text != "" {
				if system != "" {
					system += "\n\n"
				}
				system += text
			}
		case "user", "assistant":
			out = append(out, anthropicMessage{Role: m.Role, Content: text})
		}
	}
	return system, out
}
