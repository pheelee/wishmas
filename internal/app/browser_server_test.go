package app

import (
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

// This executable harness exists only in `go test -c` binaries. Production has
// no legacy runtime option and always uses the version-2 New constructor.
func TestMain(m *testing.M) {
	mode := os.Getenv("WISHMAS_BROWSER_TEST_MODE")
	if mode == "" {
		os.Exit(m.Run())
	}
	dir := os.Getenv("WISHMAS_DATA_DIR")
	addr := os.Getenv("WISHMAS_ADDR")
	switch mode {
	case "legacy":
		server := &http.Server{Addr: addr, Handler: newLegacy(dir)}
		log.Fatal(server.ListenAndServe())
	case "household":
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatal(err)
		}
		server := httptest.NewUnstartedServer(New(dir))
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
