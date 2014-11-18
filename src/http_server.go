package goule

import (
	"net"
	"net/http"
	"strconv"
	"sync"
)

type HTTPServer struct {
	mutex      sync.RWMutex
	handler    http.Handler
	listener   *net.Listener
	listenPort int
	setting    ServerSettings
}

func NewHTTPServer(handler http.Handler) *HTTPServer {
	return &HTTPServer{sync.RWMutex{}, handler, nil, 0, ServerSettings{}}
}

// Update applies a given server setting to an HTTPServer.
// If the setting is enabled but the receiver is not actively serving, it will
// start its server.
// Conversely, if the setting is disabled but the receiver is actively serving,
// it will stop.
// If both the setting and the receiver are serving, the server may still stop
// itself to change port numbers.
// The returned error will be nil unless the server could not start or restart.
func (self *HTTPServer) Update(setting ServerSettings) error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if !setting.Enabled && self.listener != nil {
		// Stop the presses! Jk, just the server.
		self.stop()
		self.setting = setting
		return nil
	} else if setting.Enabled && self.listener == nil {
		// Start the server at the given port.
		self.setting = setting
		return self.start()
	} else if setting.Enabled && setting.Port != self.listenPort {
		// Restart the server to run on the new port
		self.stop()
		self.setting = setting
		return self.start()
	}
	return nil
}

// GetSetting returns the last setting which was passed via Update().
func (self *HTTPServer) GetSettings() ServerSettings {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.setting
}

// IsRunning returns whether or not the server is actively listening for
// incoming connections.
func (self *HTTPServer) IsRunning() bool {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.listener != nil
}

// start starts the server.
// This method assumes that the receiver is already write-locked.
func (self *HTTPServer) start() error {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(self.setting.Port))
	if err != nil {
		return err
	}

	self.listener = &listener

	// Run the server in the background
	go func() {
		if err := http.Serve(listener, self.handler); err != nil {
			self.mutex.Lock()
			if self.listener == &listener {
				self.stop()
			}
			self.mutex.Unlock()
		}
	}()

	return nil
}

// stop stops the listener.
// This method assumes that the receiver is already write-locked.
func (self *HTTPServer) stop() {
	(*self.listener).Close()
	self.listener = nil
}
