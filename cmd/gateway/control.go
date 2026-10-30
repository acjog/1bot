package main

import (
	"crypto/subtle"
	"encoding/json"
	"net"
	"net/http"
	"os"
)

func controlHandler(next http.Handler, key string, stop chan<- struct{}) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", next)
	for _, path := range []string{"/control/status", "/control/stop"} {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			ip := net.ParseIP(host)
			if key == "" || err != nil || ip == nil || !ip.IsLoopback() ||
				subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+key)) != 1 {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			method := http.MethodGet
			if r.URL.Path == "/control/stop" {
				method = http.MethodPost
			}
			if r.Method != method {
				w.Header().Set("Allow", method)
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"running": true, "pid": os.Getpid()})
			if r.URL.Path == "/control/stop" {
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				select {
				case stop <- struct{}{}:
				default:
				}
			}
		})
	}
	return mux
}
