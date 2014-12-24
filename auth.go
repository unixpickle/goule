package goule

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Admin stores the basic administration info for a Goule server.
type Admin struct {
	Assets  string `assets`
	Hash    string `hash`
	Port    int    `port`
	Timeout int    `timeout`
}

// Try checks if a password is correct.
func (a *Admin) Try(password string) bool {
	return Hash(password) == a.Hash
}

// Hash returns the hexadecimal SHA256 hash of a string.
func Hash(password string) string {
	hash := sha256.Sum256([]byte(password))
	hex := hex.EncodeToString(hash[:])
	// The documentation doesn't actually say the hex will be lowercase.
	return strings.ToLower(hex)
}
