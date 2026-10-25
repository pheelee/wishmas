package app

import (
	"context"
	data "github.com/pheelee/wishmas/internal/household"
	"log"
	"net"
	"net/http/httptest"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

// This HTTPS harness exists only in go test -c binaries and uses production New.
func TestMain(m *testing.M) {
	mode := os.Getenv("WISHMAS_BROWSER_TEST_MODE")
	if mode == "" {
		os.Exit(m.Run())
	}
	dir := os.Getenv("WISHMAS_DATA_DIR")
	addr := os.Getenv("WISHMAS_ADDR")
	switch mode {
	case "household":
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatal(err)
		}
		config, err := data.StorageConfigFromEnv(os.Getenv)
		if err != nil {
			log.Fatal(err)
		}
		options := Options{}
		if config.Backend == "s3" {
			options.Storage, err = data.NewS3Backend(dir, config)
			if err != nil {
				log.Fatal(err)
			}
		}
		handler := NewWithOptions(dir, options)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := StartBackground(ctx, handler)
		defer func() { cancel(); <-done }()
		server := httptest.NewUnstartedServer(handler)
		server.Listener = listener
		server.StartTLS()
		defer server.Close()
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(signals)
		<-signals
	default:
		log.Fatal("unknown browser test mode")
	}
}
