package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestControl(t *testing.T) {
	for _, tc := range []struct {
		path, method, key, remote string
		status                    int
	}{
		{"/control/status", "GET", "secret", "127.0.0.1:123", 200},
		{"/control/stop", "POST", "secret", "[::1]:123", 200},
		{"/control/stop", "GET", "secret", "127.0.0.1:123", 405},
		{"/control/stop", "POST", "wrong", "127.0.0.1:123", 403},
		{"/control/status", "GET", "secret", "192.0.2.1:123", 403},
	} {
		stop := make(chan struct{}, 1)
		h := controlHandler(http.NotFoundHandler(), "secret", stop)
		r := httptest.NewRequest(tc.method, tc.path, nil)
		r.RemoteAddr = tc.remote
		r.Header.Set("Authorization", "Bearer "+tc.key)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Fatalf("%+v: got %d", tc, w.Code)
		}
		if (len(stop) == 1) != (tc.path == "/control/stop" && tc.status == 200) {
			t.Fatal("unexpected shutdown")
		}
	}
}
