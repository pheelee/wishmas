package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// @test OPS-001-AC2
func TestHealthcheck(t *testing.T) {
	for _, status := range []int{200, 302, 500} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/healthz" {
				t.Error(r.URL.Path)
			}
			w.Header().Set("Location", "/healthz")
			w.WriteHeader(status)
		}))
		err := checkHealth(s.Listener.Addr().String(), time.Second)
		s.Close()
		if (err == nil) != (status == 200) {
			t.Fatalf("status %d: %v", status, err)
		}
	}
	for _, pair := range [][2]string{{":8080", "127.0.0.1:8080"}, {"0.0.0.0:9090", "127.0.0.1:9090"}, {"[::]:8080", "[::1]:8080"}, {"127.0.0.1:1234", "127.0.0.1:1234"}} {
		got, err := healthAddress(pair[0])
		if err != nil || got != pair[1] {
			t.Fatalf("%s: %s %v", pair[0], got, err)
		}
	}
	if checkHealth("bad", time.Second) == nil {
		t.Fatal("invalid address accepted")
	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(40 * time.Millisecond) }))
	defer s.Close()
	if checkHealth(s.Listener.Addr().String(), time.Millisecond) == nil {
		t.Fatal("timeout accepted")
	}
}
