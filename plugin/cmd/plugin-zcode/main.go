// Command plugin-zcode is the ZCode plugin: a child process that speaks the
// plugin protocol (proto) on stdio, and drives `zcode app-server` underneath.
// It implements the plugin side of DESIGN.md section 4, exactly like
// plugin-codebuddy / plugin-codex but backed by Z.AI's ZCode app-server.
//
// The gateway never launches zcode directly; it launches this plugin and talks
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
	bin := flag.String("bin", "zcode", "zcode executable (name or path)")
	model := flag.String("model", "", "ignored: zcode reads its model from ~/.zcode/cli/config.json")
	cwd := flag.String("cwd", "", "working directory handed to the agent")
	y := flag.Bool("y", false, "accepted for gateway parity; zcode auto-approval is configured separately")
	v := flag.Bool("v", false, "verbose: dump app-server frames to stderr")
	flag.Parse()

	// Accepted for gateway parity: ZCode ignores a CLI --model (it reads the
	// model from ~/.zcode/cli/config.json) and auto-approval is configured
	// separately, so these two flags are parsed but deliberately unused here.
	_ = *model
	_ = *y

	if *cwd == "" {
		*cwd, _ = os.Getwd()
	}

	c := zcode.New(zcode.Options{
		Bin:     *bin,
		Verbose: *v,
		Cwd:     *cwd,
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
