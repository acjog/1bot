// Command gateway is the entry point that wires an OpenAI-compatible HTTP
// front to a CodeBuddy plugin child process, per DESIGN.md.
//
//	F:\go\bin\go.exe build -o plugin-codebuddy.exe ./plugin/cmd/plugin-codebuddy
//	F:\go\bin\go.exe run ./cmd/gateway -plugin plugin-codebuddy.exe -model gpt-6.1-sol -y
//	curl -N http://127.0.0.1:8787/v1/chat/completions \
//	  -H "Content-Type: application/json" \
//	  -d '{"model":"gpt-6.1-sol","stream":true,"messages":[{"role":"user","content":"hi"}]}'
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"

	"1bot_test/internal/openai"
	"1bot_test/internal/plugin"
)

func main() {
	pluginArg := flag.String("plugin", "plugin-codebuddy", "plugin executable (name or path)")
	bin := flag.String("bin", "codebuddy", "codebuddy executable handed to the plugin")
	model := flag.String("model", "", "official model id for the child; empty = settings.json's model")
	addr := flag.String("addr", "127.0.0.1:8787", "HTTP listen address")
	cwd := flag.String("cwd", "", "working directory handed to the agent (default: current)")
	y := flag.Bool("y", false, "pass -y to codebuddy (skip permission prompts)")
	v := flag.Bool("v", false, "verbose: dump plugin-protocol frames")
	logPath := flag.String("log", "", "append inbound requests (JSONL) to this file; empty = disabled")
	flag.Parse()

	var logWriter io.Writer
	if *logPath != "" {
		f, err := os.OpenFile(*logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			fatalf("log: %v", err)
		}
		defer f.Close()
		logWriter = f
	}

	pluginPath, err := resolvePlugin(*pluginArg)
	if err != nil {
		fatalf("%v", err)
	}

	// Build the argument vector the plugin will receive.
	args := []string{}
	if *model != "" {
		args = append(args, "--model", *model)
	}
	if *cwd != "" {
		args = append(args, "--cwd", *cwd)
	}
	if *y {
		args = append(args, "-y")
	}
	if *v {
		args = append(args, "-v")
	}
	if *bin != "codebuddy" {
		args = append(args, "--bin", *bin)
	}

	h := plugin.New(plugin.Options{Bin: pluginPath, Args: args, Verbose: *v})
	if err := h.Start(); err != nil {
		fatalf("%v", err)
	}
	defer h.Close()

	srv := &http.Server{
		Addr:    *addr,
		Handler: openai.New(h, *model, *v, logWriter).Handler(),
	}
	fmt.Fprintf(os.Stderr, "gateway: OpenAI endpoint on http://%s/v1/chat/completions (plugin=%s model=%q)\n", *addr, pluginPath, *model)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fatalf("http: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	fmt.Fprintln(os.Stderr, "\nshutting down")
	_ = srv.Close()
}

func fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "gateway: "+format+"\n", a...)
	os.Exit(1)
}

// resolvePlugin locates the plugin executable and returns an absolute path.
//
// A bare name (no path separator) is resolved via PATH first, then the
// current directory, then the gateway's own directory. We return an absolute
// path so exec never trips Go's "cannot run executable found relative to the
// current directory" guard, which fires when a bare name only resolves in ".".
func resolvePlugin(arg string) (string, error) {
	// An explicit path (.\x, ..\x, C:\x) is used as-is, made absolute.
	if strings.ContainsAny(arg, `/\`) {
		if abs, err := filepath.Abs(arg); err == nil {
			return abs, nil
		}
		return arg, nil
	}

	if p, err := exec.LookPath(arg); err == nil {
		return p, nil
	}

	var candidates []string
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Join(cwd, arg+".exe"),
			filepath.Join(cwd, arg),
		)
	}
	if self, err := os.Executable(); err == nil {
		dir := filepath.Dir(self)
		candidates = append(candidates,
			filepath.Join(dir, arg+".exe"),
			filepath.Join(dir, arg),
		)
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("plugin %q not found; build it with: go build -o plugin-codebuddy.exe ./plugin/cmd/plugin-codebuddy", arg)
}
