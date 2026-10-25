package main

import (
	"context"
	"github.com/pheelee/wishmas/internal/app"
	data "github.com/pheelee/wishmas/internal/household"
	"log"
	"net/http"
	"os"
	"time"
)

// @impl CON-009-AC1 CON-009-AC2 WSH-008-AC1 WSH-008-AC2
func main() {
	dataDir := os.Getenv("WISHMAS_DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}
	addr := os.Getenv("WISHMAS_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	proxies, err := app.ParseTrustedProxies(os.Getenv("WISHMAS_TRUSTED_PROXIES"))
	if err != nil {
		log.Fatal(err)
	}
	config, err := data.StorageConfigFromEnv(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	options := app.Options{TrustedProxies: proxies}
	if config.Backend == "s3" {
		options.Storage, err = data.NewS3Backend(dataDir, config)
		if err != nil {
			log.Fatal(err)
		}
	}
	handler := app.NewWithOptions(dataDir, options)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	app.StartBackground(ctx, handler)
	// A response can include a 30-second image fetch and several individually
	// bounded S3 freshness/save operations. Do not truncate its error or draft.
	writeTimeout := time.Minute
	if config.Backend == "s3" {
		const max = time.Duration(1<<63 - 1)
		if config.Timeout > (max-writeTimeout)/16 {
			writeTimeout = max
		} else {
			writeTimeout += 16 * config.Timeout
		}
	}
	server := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: writeTimeout, IdleTimeout: 60 * time.Second}
	log.Printf("Wishmas listening on %s", addr)
	log.Fatal(server.ListenAndServe())
}
