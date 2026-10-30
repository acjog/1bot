// Command plugin-zcode is the ZCode plugin: a child process that speaks the
// plugin protocol (proto) on stdio, and calls Z.AI's ZCode coding plan
// directly over its Anthropic-compatible Messages API. It implements the
// plugin side of DESIGN.md section 4, like plugin-codebuddy / plugin-codex but
// backed by a direct HTTP call instead of a subprocess.
//
// The gateway never reaches Z.ai directly; it launches this plugin and talks
// JSON-RPC. This plugin owns all ZCode-specific knowledge.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"1bot_test/internal/proto"
	"1bot_test/internal/transport"
	"1bot_test/plugin/zcode"
)

func main() {
	bin := flag.String("bin", "zcode", "accepted for gateway parity; unused (direct API)")
	model := flag.String("model", "", "model id; empty = glm-5")
	cwd := flag.String("cwd", "", "accepted for gateway parity; unused (direct API)")
	y := flag.Bool("y", false, "accepted for gateway parity; unused (direct API)")
	v := flag.Bool("v", false, "verbose: dump request and SSE events to stderr")
	flag.Parse()

	// Accepted for gateway parity: a direct HTTP call has no subprocess to
	// hand a bin, cwd or auto-approval to, so these flags are parsed but
	// deliberately unused here.
	_ = *bin
	_ = *cwd
	_ = *y

	c := zcode.New(zcode.Options{
		Model:   *model,
		Verbose: *v,
	})
	if err := c.Start(); err != nil {
		fatal("%v", err)
	}
	defer c.Close()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	write := func(f proto.Frame) {
		b, _ := json.Marshal(f)
		_, _ = out.Write(append(b, '\n'))
		_ = out.Flush()
	}

	// Process turns one at a time: one zcode process runs one turn at a time
	// (serialised inside zcode.Client), so we drain stdin in order.
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var f proto.Frame
		if err := json.Unmarshal(line, &f); err != nil {
			continue
		}
		if f.HasID() && f.Method == proto.MethodModels {
			write(proto.ModelsResponse(f.ID, nil, transport.ErrModelsUnsupported))
			continue
		}
		if f.Method != proto.MethodTurn || !f.HasID() {
			continue
		}

		var p proto.TurnParams
		_ = json.Unmarshal(f.Params, &p)

		var turnID int64
		_ = json.Unmarshal(f.ID, &turnID)

		onEvent := func(ev transport.Event) {
			ep := proto.EventParams{ID: turnID}
			switch ev.Kind {
			case transport.EventText:
				ep.Kind, ep.Text = "text", ev.Text
			case transport.EventThought:
				ep.Kind, ep.Text = "thought", ev.Text
			case transport.EventToolCall:
				ep.Kind, ep.Call = "tool_call", ev.Call
			}
			b, _ := json.Marshal(ep)
			write(proto.Frame{JSONRPC: "2.0", Method: proto.MethodEvent, Params: b})
		}

		// transport.Complete falls back to Turn (zcode has no RequestTransport);
		// it rejects client-declared tools, which ZCode's adapter does not support.
		stop, err := transport.Complete(c, p, onEvent)
		if err != nil {
			write(proto.Frame{JSONRPC: "2.0", ID: f.ID, Error: &proto.Error{Code: -32000, Message: err.Error()}})
			continue
		}
		res, _ := json.Marshal(proto.TurnResult{StopReason: stop})
		write(proto.Frame{JSONRPC: "2.0", ID: f.ID, Result: res})
	}
}

func fatal(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "plugin-zcode: "+format+"\n", a...)
	os.Exit(1)
}
