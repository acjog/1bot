package zcode

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

type testWriter struct{ bytes.Buffer }

func (*testWriter) Close() error { return nil }

func TestRuntimePreferences(t *testing.T) {
	for _, scope := range []string{"runtime-materialization", "user-execution"} {
		for _, id := range []string{`"server-1"`, `1`} {
			t.Run(scope+"/"+id, func(t *testing.T) {
				c := New(Options{})
				output := &testWriter{}
				c.stdin = output
				c.scanner = bufio.NewScanner(strings.NewReader(
					`{"id":` + id + `,"method":"session/requestRuntimePreferences","params":{"sessionId":"test-session","scope":"` + scope + `"}}` + "\n"))
				c.readLoop()
				var reply wireMessage
				if err := json.Unmarshal(output.Bytes(), &reply); err != nil {
					t.Fatal(err)
				}
				if string(reply.ID) != id || reply.Error != nil {
					t.Fatalf("invalid response: %+v", reply)
				}
				var prefs map[string]any
				if err := json.Unmarshal(reply.Result, &prefs); err != nil {
					t.Fatal(err)
				}
				if len(prefs) != 5 || prefs["askUserQuestionAutoResolutionEnabled"] != true ||
					prefs["memoryEnabled"] != false || prefs["modelContextBudgetStrategy"] != "preflight-v1" ||
					prefs["nativeSearchEnhancementsEnabled"] != true || prefs["toolSearchMode"] != "off" {
					t.Fatalf("preferences differ from runtime compatibility defaults: %+v", prefs)
				}
			})
		}
	}
}

func TestReverseRequests(t *testing.T) {
	for _, id := range []string{`"server-6"`, `6`} {
		t.Run(id, func(t *testing.T) {
			c := New(Options{})
			output := &testWriter{}
			c.stdin = output
			c.scanner = bufio.NewScanner(strings.NewReader(
				`{"id":` + id + `,"method":"interaction/requestOfficialMcpAuthHeaders","params":{"targetOrigin":"https://example.invalid"}}` + "\n" +
					`{"id":"unknown","method":"unknown/method"}` + "\n" +
					`{"method":"notification"}` + "\n" +
					`{"id":42,"result":{}}` + "\n"))
			pending := make(chan wireMessage, 1)
			c.pending["42"] = pending
			c.readLoop()
			scanner := bufio.NewScanner(&output.Buffer)
			var replies []wireMessage
			for scanner.Scan() {
				var m wireMessage
				if err := json.Unmarshal(scanner.Bytes(), &m); err != nil {
					t.Fatal(err)
				}
				replies = append(replies, m)
			}
			if len(replies) != 2 {
				t.Fatalf("unexpected responses: %+v", replies)
			}
			if string(replies[0].ID) != id || replies[0].Error != nil || string(replies[0].Result) != `{"ok":false,"reason":"official_auth_unavailable"}` {
				t.Fatalf("invalid auth reply: %+v", replies[0])
			}
			if replies[1].Error == nil || replies[1].Error.Code != -32601 {
				t.Fatalf("unknown method should fail: %+v", replies[1])
			}
			select {
			case <-pending:
			default:
				t.Fatal("normal response was not delivered")
			}
			select {
			case <-c.closed:
			default:
				t.Fatal("read loop did not close")
			}
		})
	}
}
