package goule

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type Sessions struct {
	mutex       sync.Mutex
	sessions    map[string]time.Time
	secret      string
	timeout     time.Duration
	lastCleanup time.Time
}

func NewSessions(secret string, timeout time.Duration) *RotatingKey {
	res := new(Sessions)
	res.secret = secret
	res.timeout = timeout
	return res
}

func (self *Sessions) Validate(key string) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.conditionalCleanup()
	if value, ok := self.sessions[key]; ok {
		if time.Since(value) > self.timeout {
			delete(self.sessions, key)
			return false
		}
		return true
	} else {
		return false
	}
}

func (self *Sessions) Logout(key string) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	delete(self.sessions, key)
}

func (self *Sessions) Login() string {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	info := strconv.Itoa(rand.Int()) + self.secret
	hash := sha256.Sum256(info)
	hex := hex.EncodeToString(hash[:])
	self.sessions[hex] = time.Now()
	return hex
}

func (self *Sessions) cleanup() {
	remainingSessions := make(map[string]time.Time)
	for key, expiration := range self.sessions {
		if time.Since(expiration) < self.timeout {
			remainingSessions[key] = expiration
		}
	}
	self.sessions = remainingSessions
	self.lastCleanup = time.Now()
}

func (self *Sessions) conditionalCleanup() {
	if time.Since(self.lastCleanup) > self.timeout {
		self.cleanup()
	}
}
