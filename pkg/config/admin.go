package config

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type AdminSettings struct {
	Rules          []SourceURL `json:"rules"`
	PasswordHash   string      `json:"password_hash"`
	SessionTimeout int         `json:"session_timeout"`
}

func (self *AdminSettings) Copy() AdminSettings {
	rules := make([]SourceURL, len(self.Rules))
	copy(rules, self.Rules)
	return AdminSettings{rules, self.PasswordHash, self.SessionTimeout}
}

func (self *AdminSettings) CheckPassword(password string) bool {
	return HashPassword(password) == self.PasswordHash
}

func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	hex := hex.EncodeToString(hash[:])
	return strings.ToLower(hex)
}
