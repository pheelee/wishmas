package household

import (
	"log"
	"os"
	"testing"

	"github.com/pheelee/wishmas/internal/testsupport"
)

func TestMain(m *testing.M) {
	if err := testsupport.TrustLocalTLSServers(); err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}
