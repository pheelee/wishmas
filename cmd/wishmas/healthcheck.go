package main

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

// @impl OPS-001-AC2: probe the listener without storage access or proxy discovery.
func healthAddress(addr string) (string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", fmt.Errorf("invalid WISHMAS_ADDR: %w", err)
	}
	switch host {
	case "", "0.0.0.0":
		host = "127.0.0.1"
	case "::":
		host = "::1"
	}
	return net.JoinHostPort(host, port), nil
}
func checkHealth(addr string, timeout time.Duration) error {
	target, err := healthAddress(addr)
	if err != nil {
		return err
	}
	transport := &http.Transport{Proxy: nil}
	defer transport.CloseIdleConnections()
	client := &http.Client{Timeout: timeout, Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	// @impl OPS-001-AC5
	// #nosec G704 -- target derives only from operator-configured WISHMAS_ADDR, never HTTP request input; the fixed /healthz probe disables proxies and redirects.
	resp, err := client.Get("http://" + target + "/healthz")
	if err != nil {
		return fmt.Errorf("healthcheck request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck returned HTTP %d", resp.StatusCode)
	}
	return nil
}
