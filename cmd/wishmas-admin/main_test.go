package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pheelee/wishmas/internal/household"
)

// @test CON-010-AC1: retired commands cannot read or modify operator data.
func TestMigrationCommandIsRetired(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	destination := filepath.Join(dir, "destination")
	if err := os.Mkdir(source, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(source, "household.yaml")
	const original = "operator data must remain untouched\n"
	if err := os.WriteFile(path, []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	var out, diagnostics bytes.Buffer
	err := run([]string{"migrate", "--source", source, "--destination", destination}, &out, &diagnostics, func() (string, error) {
		t.Fatal("retired command must not prompt")
		return "", nil
	})
	if err == nil || err.Error() != "unknown command; use hash-password" {
		t.Fatalf("unexpected result: %v", err)
	}
	if out.Len() != 0 || diagnostics.Len() != 0 {
		t.Fatal("retired command produced output")
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != original {
		t.Fatal("source changed")
	}
	if _, err := os.Stat(destination); !os.IsNotExist(err) {
		t.Fatal("destination created")
	}
}

// @test CON-007-AC2
func TestPrivatePasswordPromptAndConfirmation(t *testing.T) {
	cases := []struct {
		name    string
		values  []string
		failure bool
	}{{"valid", []string{"PublicFixture1!", "PublicFixture1!"}, false}, {"mismatch", []string{"PublicFixture1!", "DifferentFixture1!"}, true}, {"policy", []string{"weak", "weak"}, true}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out, diagnostics bytes.Buffer
			index := 0
			err := run([]string{"hash-password"}, &out, &diagnostics, func() (string, error) { value := tc.values[index]; index++; return value, nil })
			if (err != nil) != tc.failure {
				t.Fatal("unexpected helper result")
			}
			for _, value := range tc.values {
				if strings.Contains(diagnostics.String(), value) {
					t.Fatal("secret echoed in prompts")
				}
			}
			if tc.failure {
				if out.Len() != 0 {
					t.Fatal("failure emitted hash")
				}
				return
			}
			encoded := strings.TrimSuffix(strings.TrimPrefix(out.String(), "admin_password_hash: '"), "'\n")
			ok, err := household.CheckPassword(encoded, tc.values[0])
			if err != nil || !ok {
				t.Fatal("helper output is not usable YAML-ready hash")
			}
		})
	}
	var out, diagnostics bytes.Buffer
	if err := run([]string{"hash-password", "never-pass-secrets-as-args"}, &out, &diagnostics, func() (string, error) { t.Fatal("must not prompt for invalid args"); return "", nil }); err == nil {
		t.Fatal("password argument accepted")
	}
	if err := run([]string{"hash-password"}, &out, &diagnostics, func() (string, error) { return "", errors.New("not a terminal") }); err == nil {
		t.Fatal("nonterminal password input accepted")
	}
}
