package main

import (
	"log"
	"net/http"
	"os"
	"time"
	"wishmas/internal/app"
)

func main() {
	dataDir := os.Getenv("WISHMAS_DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}
	addr := os.Getenv("WISHMAS_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	server := &http.Server{Addr: addr, Handler: app.New(dataDir), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("Wishmas listening on %s", addr)
	log.Fatal(server.ListenAndServe())
}
