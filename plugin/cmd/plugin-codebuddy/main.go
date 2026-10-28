// Command plugin-codebuddy is the CodeBuddy plugin: a child process that
// speaks the plugin protocol (proto) on stdio, and drives `codebuddy --acp`
// underneath. It implements the plugin side of DESIGN.md section 4.
//
// The gateway never launches codebuddy directly; it launches this plugin and
// talks JSON-RPC. This plugin owns all CodeBuddy-specific knowledge.
//
// Client-declared function tools reach CodeBuddy through an MCP bridge: the
// plugin hosts a loopback HTTP MCP server and points session/new.mcpServers at
// its URL (CodeBuddy's ACP SDK has no stdio MCP variant), then round-trips each
// tools/call and its result through the proto tool channel. See
// CODEBUDDY-CLIENT-TOOLS.md.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	"1bot_test/plugin/acp"
	"1bot_test/internal/proto"
	"1bot_test/internal/transport"
)

func main() {
	bin := flag.String("bin", "codebuddy", "codebuddy executable (name or path)")
	model := flag.String("model", "", "official model id for the child; empty = settings.json's model")
	cwd := flag.String("cwd", "", "working directory handed to the agent")
	y := flag.Bool("y", false, "pass -y to codebuddy (skip permission prompts)")
	v := flag.Bool("v", false, "verbose: dump ACP frames to stderr")
	noFS := flag.Bool("no-fs", false, "don't declare fs/terminal client capabilities (tools run on the API client)")
	flag.Parse()

	if *cwd == "" {
		// The gateway host's directory is not the API client's environment, so
		// default to a fresh, empty directory rather than the plugin's own cwd.
		dir, err := os.MkdirTemp("", "1bot-cwd-")
		if err != nil {
			fatal("tmp cwd: %v", err)
		}
		defer os.RemoveAll(dir)
		*cwd = dir
	}

	c := acp.New(acp.Options{
		Bin:           *bin,
		Model:         *model,
		SkipY:         *y,
		Verbose:       *v,
		Cwd:           *cwd,
		NoFSTerminal:  *noFS,
	})
	if err := c.Start(); err != nil {
		fatal("%v", err)
	}
	defer c.Close()

	bridge := newToolbridge(c, *v)
	if err := bridge.start(); err != nil {
		fatal("tool bridge: %v", err)
	}
	defer bridge.close()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	write := func(f proto.Frame) {
		b, _ := json.Marshal(f)
		_, _ = out.Write(append(b, '\n'))
		_ = out.Flush()
	}

	// configured is the tool set the ACP session was last created with, so a
	// changed set recreates the session (and its MCP helper) with fresh tools.
	var configured []transport.Tool

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

		tools := p.Tools
		if isToolChoiceNone(p.ToolChoice) {
			tools = nil
		}

		if !toolsEqual(tools, configured) {
			servers, err := bridge.setTools(tools)
			if err != nil {
				write(proto.Frame{JSONRPC: "2.0", ID: f.ID, Error: &proto.Error{Code: -32000, Message: err.Error()}})
				continue
			}
			c.SetMCPServers(servers)
			c.ResetSession()
			configured = tools
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
			case transport.EventToolCall:
				ep.Kind, ep.Call = "tool_call", ev.Call
			}
			b, _ := json.Marshal(ep)
			write(proto.Frame{JSONRPC: "2.0", Method: proto.MethodEvent, Params: b})
		}

		stop, err := c.Turn(p.Messages, onEvent)
		if err != nil {
			write(proto.Frame{JSONRPC: "2.0", ID: f.ID, Error: &proto.Error{Code: -32000, Message: err.Error()}})
			continue
		}
		res, _ := json.Marshal(proto.TurnResult{StopReason: stop})
		write(proto.Frame{JSONRPC: "2.0", ID: f.ID, Result: res})
	}
}

// isToolChoiceNone reports whether tool_choice forbids tools.
func isToolChoiceNone(tc json.RawMessage) bool {
	return len(tc) > 0 && string(tc) == `"none"`
}

// toolsEqual compares two tool sets order-insensitively by name, so a client
// that reorders its tools does not needlessly recreate the ACP session.
func toolsEqual(a, b []transport.Tool) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 {
		return true
	}
	as := make(map[string]transport.Tool, len(a))
	for _, t := range a {
		as[t.Name] = t
	}
	for _, t := range b {
		other, ok := as[t.Name]
		if !ok || !reflect.DeepEqual(other, t) {
			return false
		}
	}
	return true
}

func fatal(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "plugin-codebuddy: "+format+"\n", a...)
	os.Exit(1)
}
