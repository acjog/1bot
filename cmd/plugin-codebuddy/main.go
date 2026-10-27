// Command plugin-codebuddy is the CodeBuddy plugin: a child process that
// speaks the plugin protocol (proto) on stdio, and drives `codebuddy --acp`
// underneath. It implements the plugin side of DESIGN.md section 4.
//
// The gateway never launches codebuddy directly; it launches this plugin and
// talks JSON-RPC. This plugin owns all CodeBuddy-specific knowledge.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"1bot_test/internal/acp"
	"1bot_test/internal/proto"
	"1bot_test/internal/transport"
)

func main() {
	bin := flag.String("bin", "codebuddy", "codebuddy executable (name or path)")
	model := flag.String("model", "", "official model id for the child; empty = settings.json's model")
	cwd := flag.String("cwd", "", "working directory handed to the agent")
	y := flag.Bool("y", false, "pass -y to codebuddy (skip permission prompts)")
	v := flag.Bool("v", false, "verbose: dump ACP frames to stderr")
	flag.Parse()

	if *cwd == "" {
		*cwd, _ = os.Getwd()
	}

	c := acp.New(acp.Options{
		Bin:     *bin,
		Model:   *model,
		SkipY:   *y,
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

	// Process turns one at a time: one CodeBuddy process runs one turn at a
	// time (serialised inside acp.Client), so we drain stdin in order.
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
		text := ""
		for _, part := range p.Prompt {
			if part.Type == "text" {
				text += part.Text
			}
		}

		var turnID int64
		_ = json.Unmarshal(f.ID, &turnID)

		onEvent := func(ev transport.Event) {
			ep := proto.EventParams{ID: turnID}
			switch ev.Kind {
			case transport.EventText:
				ep.Kind, ep.Text = "text", ev.Text
			case transport.EventThought:
				ep.Kind, ep.Text = "thought", ev.Text
			case transport.EventTool:
				ep.Kind, ep.Name = "tool", ev.Tool
			}
			b, _ := json.Marshal(ep)
			write(proto.Frame{JSONRPC: "2.0", Method: proto.MethodEvent, Params: b})
		}

		stop, err := c.Turn(text, onEvent)
		if err != nil {
			write(proto.Frame{JSONRPC: "2.0", ID: f.ID, Error: &proto.Error{Code: -32000, Message: err.Error()}})
			continue
		}
		res, _ := json.Marshal(proto.TurnResult{StopReason: stop})
		write(proto.Frame{JSONRPC: "2.0", ID: f.ID, Result: res})
	}
}

func fatal(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "plugin-codebuddy: "+format+"\n", a...)
	os.Exit(1)
}
