package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// @test OPS-001-AC6: verify the actual binary's startup record, not a formatter.
func TestStartupLogEscapesListenAddress(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "wishmas")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, output)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	command.Env = append(os.Environ(), "WISHMAS_STORAGE=local", "WISHMAS_TRUSTED_PROXIES=", "WISHMAS_DATA_DIR="+t.TempDir(), "WISHMAS_ADDR=127.0.0.1:8080\r\nFORGED\x1b[31m")
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatal("invalid listener did not exit promptly")
	}
	if err == nil {
		t.Fatal("invalid listener unexpectedly accepted")
	}
	if strings.ContainsAny(string(output), "\r\x1b") || strings.Contains(string(output), "\nFORGED") {
		t.Fatalf("startup or listener error contains unescaped controls: %q", output)
	}
	first, _, _ := strings.Cut(string(output), "\n")
	if !strings.Contains(first, "Wishmas listening on") || !strings.Contains(first, "FORGED") {
		t.Fatalf("address split across startup records: %q", output)
	}
	if strings.ContainsAny(first, "\r\x1b") || !strings.Contains(first, `\r\n`) || !strings.Contains(first, `\x1b`) {
		t.Fatalf("startup record contains unescaped controls: %q", first)
	}
}
