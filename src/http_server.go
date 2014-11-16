package goule

import (
	"net"
	"net/http"
	"errors"
	"sync"
	"strconv"
)

type HTTPServer struct {
	handler  http.Handler
	listener *net.Listener
	mutex    *sync.Mutex
}

func NewHTTPServer(handler http.Handler) *HTTPServer {
	return &HTTPServer{handler, nil, &sync.Mutex{}}
}

func (self *HTTPServer) Run(port int) error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	
	if self.listener != nil {
		return errors.New("Server was already running.")
	}
	
	listener, err := net.Listen("tcp", ":" + strconv.Itoa(port))
	if err != nil {
		return err
	}
	
	self.listener = &listener
	if err := http.Serve(listener, self.handler); err != nil {
		(*self.listener).Close()
		self.listener = nil
		return err
	}
	
	return nil
}

func (self *HTTPServer) Stop() error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	
	if self.listener == nil {
		return errors.New("Server wasn't running.")
	}
	
	(*self.listener).Close()
	self.listener = nil
	return nil
}

func (self *HTTPServer) IsRunning() bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.listener != nil
}
