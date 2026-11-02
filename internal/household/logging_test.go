package household

import (
	"bytes"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
)

// @test OPS-001-AC3
func TestFailedSaveLogsWithoutHouseholdOrPath(t *testing.T) {
	var logs bytes.Buffer
	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	defer slog.SetDefault(old)
	s := NewStore(filepath.Join(t.TempDir(), "PrivateHouseholdName", "household.yaml"))
	if err := s.write([]byte("SecretHouseholdContent"), true, ""); err == nil {
		t.Fatal("write should fail")
	}
	if !strings.Contains(logs.String(), "household save failed") || !strings.Contains(logs.String(), "file does not exist") {
		t.Fatal(logs.String())
	}
	for _, secret := range []string{"PrivateHouseholdName", "SecretHouseholdContent", "household.yaml"} {
		if strings.Contains(logs.String(), secret) {
			t.Fatal("private data leaked")
		}
	}
}
