package goule

import (
	"sync"
	"time"
)

// StoppableLock provides extended stoppable mutex functionality with timeouts.
type StoppableLock struct {
	mutex          sync.Mutex
	stopped        bool
	timeoutSkipped bool
	skipTimeout    chan bool
}

// NewStoppableLock creates an unlocked unstopped StoppableLock
func NewStoppableLock() *StoppableLock {
	return &StoppableLock{sync.Mutex{}, false, false, nil}
}

// Lock seizes the lock or returns false if the lock has been stopped.
func (self *StoppableLock) Lock() bool {
	self.mutex.Lock()
	if stopped {
		self.mutex.Unlock()
		return false
	}
	return true
}

// Unlock releases the lock.
func (self *StoppableLock) Unlock() {
	self.mutex.Unlock()
}

// Stop stops the lock.
// The caller must currently own the lock.
// No further Lock() or Wait() calls will work.
// Any current timeouts will be canceled.
func (self *StoppableLock) Stop() {
	if self.stopped {
		return
	}
	self.stopped = true
	if self.skipTimeout != nil && !self.timeoutSkipped {
		self.skipTimeout <- true
		self.timeoutSkipped = true
	}
}

// SkipTimeout stops any current Wait() that's running.
// The caller must currently own the lock.
// The Wait() will return true, but it may do so prematurely.
func (self *StoppableLock) SkipWait() {
	if self.timeoutSkipped {
		return
	}
	if self.skipTimeout != nil {
		self.skipTimeout <- true
		self.timeoutSkipped = true
	}
}

// Wait waits for the StoppableLock to be stopped, the wait to be canceled, or
// for a timeout to elapse.
func (self *StoppableLock) Wait(duration time.Duration) bool {
	if self.stopped {
		return false
	} else if self.killTimeout != nil {
		panic("StoppableLock already waiting on another thread.")
	} else {
		channel := make(chan bool, 1)
		self.killTimeout = channel
		self.timoutKilled = false
		self.Unlock()
		select {
		case <-channel:
		case <-time.After(duration):
		}
		self.mutex.Lock()
		self.killTimeout = nil
		if self.stopped {
			self.mutex.Unlock()
			return false
		}
		return true
	}
}
