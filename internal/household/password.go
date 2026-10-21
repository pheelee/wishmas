package household

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

var ErrPasswordPolicy = errors.New("password requires at least 8 characters, uppercase, lowercase, digit and punctuation/symbol")
var ErrPasswordHash = errors.New("invalid or unsupported Argon2id password hash")

// @impl ACC-005-AC2 CON-007-AC1 CON-007-AC2
func ValidatePassword(password string) error {
	var upper, lower, digit, special bool
	for _, r := range password {
		upper = upper || unicode.IsUpper(r)
		lower = lower || unicode.IsLower(r)
		digit = digit || unicode.IsDigit(r)
		special = special || unicode.IsPunct(r) || unicode.IsSymbol(r)
	}
	if !utf8.ValidString(password) || utf8.RuneCountInString(password) < 8 || !upper || !lower || !digit || !special {
		return ErrPasswordPolicy
	}
	return nil
}

// @impl CON-007-AC1 CON-007-AC2
func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, 2, 19456, 1, 32)
	return "$argon2id$v=19$m=19456,t=2,p=1$" + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(key), nil
}
func parsePasswordHash(encoded string) (salt, key []byte, err error) {
	// Only the approved cost tuple is supported. Parse/bound before allocation/KDF.
	if len(encoded) > 160 {
		return nil, nil, ErrPasswordHash
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != "m=19456,t=2,p=1" {
		return nil, nil, ErrPasswordHash
	}
	salt, err = base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) != 16 {
		return nil, nil, ErrPasswordHash
	}
	key, err = base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(key) != 32 {
		return nil, nil, ErrPasswordHash
	}
	return salt, key, nil
}

// @impl CON-007-AC1: validation never evaluates user-supplied cost parameters.
func ValidatePasswordHash(encoded string) error { _, _, err := parsePasswordHash(encoded); return err }
func CheckPassword(encoded, password string) (bool, error) {
	salt, want, err := parsePasswordHash(encoded)
	if err != nil {
		return false, err
	}
	got := argon2.IDKey([]byte(password), salt, 2, 19456, 1, 32)
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

// NewID is opaque and path-safe. Names never become reference or picture paths.
func NewID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// @impl CON-007-AC3: shared by future sessions and currently generated list links.
func NewToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
