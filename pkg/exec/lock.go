package exec

import (
	"sync"
	"time"
)

// StoppableLock provides extended stoppable mutex functionality with timeouts.
type Lock struct {
	mutex          sync.Mutex
	stopped        bool
	timeoutSkipped bool
	skipTimeout    chan struct{}
}

// NewStoppableLock creates an unlocked unstopped StoppableLock
func NewLock() *Lock {
	return &Lock{sync.Mutex{}, false, false, nil}
}

// Lock seizes the lock or returns false if the lock has been stopped.
func (self *Lock) Lock() bool {
	self.mutex.Lock()
	if self.stopped {
		self.mutex.Unlock()
		return false
	}
	return true
}

// Unlock releases the lock.
func (self *Lock) Unlock() {
	self.mutex.Unlock()
}

// Stop stops the lock.
// The caller must currently own the lock.
// After the call, the owner loses ownership of the lock.
// No further Lock() or Wait() calls will work.
// Any current timeouts will be canceled.
func (self *Lock) Stop() {
	if self.stopped {
		return
	}
	self.stopped = true
	if self.skipTimeout != nil && !self.timeoutSkipped {
		self.skipTimeout <- struct{}{}
		self.timeoutSkipped = true
	}
	self.Unlock()
}

// SkipTimeout stops any current Wait() that's running.
// The caller must currently own the lock.
// The Wait() will return true, but it may do so prematurely.
// Returns true if and only if the lock was currently Wait()ing.
func (self *Lock) SkipWait() bool {
	if self.skipTimeout != nil && !self.timeoutSkipped {
		self.skipTimeout <- struct{}{}
		self.timeoutSkipped = true
		return true
	}
	return false
}

// Wait waits for the StoppableLock to be stopped, the wait to be canceled, or
// for a timeout to elapse.
func (self *Lock) Wait(duration time.Duration) bool {
	if self.stopped {
		return false
	} else if self.skipTimeout != nil {
		panic("StoppableLock already waiting on another thread.")
	} else {
		channel := make(chan struct{}, 1)
		self.skipTimeout = channel
		self.timeoutSkipped = false
		self.Unlock()
		select {
		case <-channel:
		case <-time.After(duration):
		}
		self.mutex.Lock()
		self.skipTimeout = nil
		if self.stopped {
			self.mutex.Unlock()
			return false
		}
		return true
	}
}
