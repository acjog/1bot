package zcode

import (
	"encoding/json"
	"fmt"
	"time"
)

// wireMessage is one ZCode app-server JSON message (newline-delimited).
//
// Like Codex, ZCode's app-server omits the "jsonrpc":"2.0" header, so JSONRPC
// carries omitempty and is left empty when we write.
type wireMessage struct {
	JSONRPC string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (m wireMessage) hasID() bool   { return len(m.ID) > 0 && string(m.ID) != "null" }
func (m wireMessage) idKey() string { return string(m.ID) }

// deliver hands a response to whoever is waiting on its id.
func (c *Client) deliver(m wireMessage) {
	c.pendMu.Lock()
	ch := c.pending[m.idKey()]
	c.pendMu.Unlock()
	if ch != nil {
		ch <- m
	}
}

// call sends a request and waits for its response.
func (c *Client) call(method string, params any) (wireMessage, error) {
	c.pendMu.Lock()
	c.nextID++
	id := fmt.Sprintf("%d", c.nextID)
	ch := make(chan wireMessage, 1)
	c.pending[id] = ch
	c.pendMu.Unlock()
	defer func() {
		c.pendMu.Lock()
		delete(c.pending, id)
		c.pendMu.Unlock()
	}()

	b, _ := json.Marshal(params)
	req := wireMessage{
		ID:     json.RawMessage(id),
		Method: method,
		Params: b,
	}
	if err := c.write(req); err != nil {
		return wireMessage{}, err
	}

	select {
	case m := <-ch:
		if m.Error != nil {
			return m, fmt.Errorf("%s: rpc error %d: %s", method, m.Error.Code, m.Error.Message)
		}
		return m, nil
	case <-time.After(300 * time.Second):
		return wireMessage{}, fmt.Errorf("%s: timed out", method)
	case <-c.closed:
		return wireMessage{}, fmt.Errorf("%s: zcode app-server disconnected", method)
	}
}

// write serialises one frame to the child's stdin. A mutex keeps concurrent
// writers from interleaving bytes.
func (c *Client) write(m wireMessage) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if c.opts.Verbose {
		fmt.Fprintf(stderr, "[->] %s\n", b)
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_, err = c.stdin.Write(append(b, '\n'))
	return err
}
