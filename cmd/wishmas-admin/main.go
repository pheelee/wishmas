// wishmas-admin provides explicit offline administration. It never starts or
// mutates the HTTP application.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/pheelee/wishmas/internal/household"
	"golang.org/x/term"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, func() (string, error) {
		raw, err := term.ReadPassword(int(os.Stdin.Fd()))
		defer func() {
			for i := range raw {
				raw[i] = 0
			}
		}()
		return string(raw), err
	}); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// @impl CON-007-AC2 CON-010-AC1
func run(args []string, out, diagnostics io.Writer, readPassword func() (string, error)) error {
	if len(args) == 0 {
		return errors.New("usage: wishmas-admin hash-password")
	}
	switch args[0] {
	case "hash-password":
		if len(args) != 1 {
			return errors.New("hash-password takes no arguments; enter the password privately at the prompts")
		}
		fmt.Fprint(diagnostics, "Password: ")
		password, err := readPassword()
		fmt.Fprintln(diagnostics)
		if err != nil {
			return errors.New("password input requires an interactive terminal")
		}
		fmt.Fprint(diagnostics, "Confirm password: ")
		confirmation, err := readPassword()
		fmt.Fprintln(diagnostics)
		if err != nil {
			return errors.New("password input requires an interactive terminal")
		}
		if password != confirmation {
			return errors.New("passwords do not match")
		}
		encoded, err := household.HashPassword(password)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "admin_password_hash: '%s'\n", encoded)
		return err
	default:
		return errors.New("unknown command; use hash-password")
	}
}
