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
	mutex       sync.RWMutex
	sessions    map[string]time.Time
	secret      string
	timeout     int
	lastCleanup time.Time
}

func NewSessions() *Sessions {
	res := new(Sessions)
	res.secret = "default"
	res.timeout = 30 * 60
	return res
}

func (self *Sessions) GetSecret() string {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.secret
}

func (self *Sessions) GetTimeout() int {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.timeout
}

func (self *Sessions) SetSecret(val string) {
	self.mutex.Lock()
	self.secret = val
	self.mutex.Unlock()
}

func (self *Sessions) SetTimeout(timeout int) {
	self.mutex.Lock()
	if timeout != 0 {
		self.timeout = timeout
	} else {
		self.timeout = 30 * 60
	}
	self.mutex.Unlock()
}

func (self *Sessions) Validate(key string) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.conditionalCleanup()
	if value, ok := self.sessions[key]; ok {
		if time.Since(value) > self.timeoutDuration() {
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
	hash := sha256.Sum256([]byte(info))
	hex := hex.EncodeToString(hash[:])
	self.sessions[hex] = time.Now()
	return hex
}

func (self *Sessions) cleanup() {
	remainingSessions := make(map[string]time.Time)
	for key, expiration := range self.sessions {
		if time.Since(expiration) < self.timeoutDuration() {
			remainingSessions[key] = expiration
		}
	}
	self.sessions = remainingSessions
	self.lastCleanup = time.Now()
}

func (self *Sessions) conditionalCleanup() {
	if time.Since(self.lastCleanup) > self.timeoutDuration() {
		self.cleanup()
	}
}

func (self *Sessions) timeoutDuration() time.Duration {
	return time.Duration(self.timeout) * time.Second
}
