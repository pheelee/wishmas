// wishmas-admin provides explicit offline administration. It never starts or
// mutates the HTTP application, and migration always requires a separate output.
package main

import (
	"encoding/json"
	"errors"
	"flag"
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

// @impl CON-007-AC2 CON-008-AC1 CON-008-AC5
func run(args []string, out, diagnostics io.Writer, readPassword func() (string, error)) error {
	if len(args) == 0 {
		return errors.New("usage: wishmas-admin hash-password | migrate --source HOUSEHOLD_COPY --destination NEW_HOUSEHOLD")
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
	case "migrate":
		flags := flag.NewFlagSet("migrate", flag.ContinueOnError)
		flags.SetOutput(diagnostics)
		source := flags.String("source", "", "copied legacy household directory")
		destination := flags.String("destination", "", "new household directory (must not exist)")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if *source == "" || *destination == "" || flags.NArg() != 0 {
			return errors.New("migrate requires --source and --destination with no positional arguments")
		}
		report, err := household.Migrate(*source, *destination)
		if err != nil {
			return err
		}
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	default:
		return errors.New("unknown command; use hash-password or migrate")
	}
}
