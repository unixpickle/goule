package server

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"sync"
)

type HTTP struct {
	mutex      sync.RWMutex
	handler    http.Handler
	listener   *net.Listener
	listenPort int
}

// NewHTTP creates a new HTTP with a given handler.
// The newly created HTTP will not be listening.
func NewHTTP(handler http.Handler) *HTTP {
	return &HTTP{sync.RWMutex{}, handler, nil, 0}
}

// Start starts the server on the specified port.
// An error is returned if the server cannot be started or is already running.
// This is thread-safe.
func (self *HTTP) Start(port int) error {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if self.listener != nil {
		return errors.New("Already started.")
	}

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}

	self.listener = &listener

	// Run the server in the background
	go func() {
		if err := http.Serve(listener, self.handler); err != nil {
			self.mutex.Lock()
			if self.listener == &listener {
				(*self.listener).Close()
				self.listener = nil
			}
			self.mutex.Unlock()
		}
	}()

	return nil
}

// Stop stops the server if it was running.
// This is thread-safe.
func (self *HTTP) Stop() {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if self.listener != nil {
		(*self.listener).Close()
		self.listener = nil
	}
}

// Status returns whether or not the server is listening and which port it is
// using.
// This is thread-safe.
func (self *HTTP) Status() (bool, int) {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.listener != nil, self.listenPort
}

// IsRunning returns the first return value of Status.
func (self *HTTP) IsRunning() bool {
	x, _ := self.Status()
	return x
}
