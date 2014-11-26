package sessions

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

// Manager manages a set of empty sessions with expiration timeouts.
type Manager struct {
	mutex       sync.RWMutex
	sessions    map[string]time.Time
	secret      string
	timeout     int
	lastCleanup time.Time
}

// NewManager returns a new, empty Manager.
// The manager's secret will be "default".
// The manager's timeout will be 30 minutes.
func NewManager() *Manager {
	res := new(Manager)
	res.sessions = map[string]time.Time{}
	res.secret = "default"
	res.timeout = 30 * 60
	return res
}

// GetSecret returns the Manager's current secret.
// This is thread-safe.
func (self *Manager) GetSecret() string {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.secret
}

// GetTimeout returns the Manager's current timeout (in seconds).
// This is thread-safe.
func (self *Manager) GetTimeout() int {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.timeout
}

// SetSecret sets the Manager's current secret.
// Note: revealing the session secret may leave your server insecure. You
// should keep it secret.
// This is thread-safe.
func (self *Manager) SetSecret(val string) {
	self.mutex.Lock()
	self.secret = val
	self.mutex.Unlock()
}

// SetTimeout sets the Manager's session timeout.
// Previously created sessions are affected by this change.
// This is thread-safe.
func (self *Manager) SetTimeout(timeout int) {
	self.mutex.Lock()
	if timeout != 0 {
		self.timeout = timeout
	} else {
		self.timeout = 30 * 60
	}
	self.mutex.Unlock()
}

// Validate checks if the Manager contains a non-expired session identifier.
// This is thread-safe.
func (self *Manager) Validate(key string) bool {
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

// Logout explicitly removes a session from the Manager.
// This is thread-safe.
func (self *Manager) Logout(key string) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	delete(self.sessions, key)
}

// Login generates a new session identifier and returns it.
// This is thread-safe.
func (self *Manager) Login() string {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	info := strconv.Itoa(rand.Int()) + self.secret
	hash := sha256.Sum256([]byte(info))
	hex := hex.EncodeToString(hash[:])
	self.sessions[hex] = time.Now()
	return hex
}

func (self *Manager) cleanup() {
	remainingManager := make(map[string]time.Time)
	for key, expiration := range self.sessions {
		if time.Since(expiration) < self.timeoutDuration() {
			remainingManager[key] = expiration
		}
	}
	self.sessions = remainingManager
	self.lastCleanup = time.Now()
}

func (self *Manager) conditionalCleanup() {
	if time.Since(self.lastCleanup) > self.timeoutDuration() {
		self.cleanup()
	}
}

func (self *Manager) timeoutDuration() time.Duration {
	return time.Duration(self.timeout) * time.Second
}
