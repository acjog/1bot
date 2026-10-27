// Command plugin-codex is the Codex plugin: a child process that speaks the
// plugin protocol (proto) on stdio, and drives `codex app-server` underneath.
// It implements the plugin side of DESIGN.md section 4, exactly like
// plugin-codebuddy but backed by Codex's app-server protocol.
//
// The gateway never launches codex directly; it launches this plugin and talks
// JSON-RPC. This plugin owns all Codex-specific knowledge.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"1bot_test/internal/codex"
	"1bot_test/internal/proto"
	"1bot_test/internal/transport"
)

func main() {
	bin := flag.String("bin", "codex", "codex executable (name or path)")
	model := flag.String("model", "", "model id for the child")
	cwd := flag.String("cwd", "", "working directory handed to the agent")
	v := flag.Bool("v", false, "verbose: dump app-server frames to stderr")
	flag.Parse()

	if *cwd == "" {
		*cwd, _ = os.Getwd()
	}

	c := codex.New(codex.Options{
		Bin:     *bin,
		Model:   *model,
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

	// Process turns one at a time: one Codex process runs one turn at a time
	// (serialised inside codex.Client), so we drain stdin in order.
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

		stop, err := c.Complete(p, onEvent)
		if err != nil {
			write(proto.Frame{JSONRPC: "2.0", ID: f.ID, Error: &proto.Error{Code: -32000, Message: err.Error()}})
			continue
		}
		res, _ := json.Marshal(proto.TurnResult{StopReason: stop})
		write(proto.Frame{JSONRPC: "2.0", ID: f.ID, Result: res})
	}
}

func fatal(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "plugin-codex: "+format+"\n", a...)
	os.Exit(1)
}
