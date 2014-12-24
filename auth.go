package goule

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// Auth stores the basic administration info for a Goule server.
type Auth struct {
	Hash    string
	Port    int
	Timeout time.Duration
}

// Clone clones an Auth object.
func (a *Auth) Clone() *Auth {
	c := *a
	return &c
}

// Try checks if a password is correct.
func (a *Auth) Try(password string) bool {
	return Hash(password) == a.Hash
}

// Hash returns the hexadecimal SHA256 hash of a string.
func Hash(password string) string {
	hash := sha256.Sum256([]byte(password))
	hex := hex.EncodeToString(hash[:])
	// The documentation doesn't actually say the hex will be lowercase.
	return strings.ToLower(hex)
}
