package goule

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strconv"
	"sync"
)

type HTTPSServer struct {
	handler  http.Handler
	listener *net.Listener
	mutex    *sync.Mutex
	config   *Configuration
	port     int
}

func NewHTTPSServer(handler http.Handler, config *Configuration) *HTTPSServer {
	return &HTTPSServer{handler, nil, &sync.Mutex{}, config, 0}
}

func (self *HTTPSServer) Run(port int) error {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if self.listener != nil {
		return errors.New("Server was already running.")
	}

	self.port = port

	return self.runOnCurrentPort()
}

func (self *HTTPSServer) Stop() error {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if self.listener == nil {
		return errors.New("Server wasn't running.")
	}

	(*self.listener).Close()
	self.listener = nil
	return nil
}

func (self *HTTPSServer) IsRunning() bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	return self.listener != nil
}

func (self *HTTPSServer) CertificatesChanged() error {
	// In the future, this will not need to do anything because we can implement
	// the server's GetCertificate() method. For now, though, that method is
	// not in the stable release.

	self.mutex.Lock()
	defer self.mutex.Unlock()

	if self.listener == nil {
		return nil
	}
	(*self.listener).Close()
	self.listener = nil

	return self.runOnCurrentPort()
}

func (self *HTTPSServer) runOnCurrentPort() error {
	config, err := self.createTLSConfig()
	if err != nil {
		return err
	}

	tcpListener, err := net.Listen("tcp", ":"+strconv.Itoa(self.port))
	if err != nil {
		return err
	}
	tlsListener := tls.NewListener(tcpListener, config)
	self.listener = &tlsListener

	// Run the server in the background
	go func() {
		if err := http.Serve(tlsListener, self.handler); err != nil {
			self.mutex.Lock()
			if self.listener == &tlsListener {
				(*self.listener).Close()
				self.listener = nil
			}
			self.mutex.Unlock()
		}
	}()

	return nil
}

func (self *HTTPSServer) createTLSConfig() (*tls.Config, error) {
	self.config.RLock()
	defer self.config.RUnlock()

	// Build up the tls.Config to have all the certificates we need
	certs := self.config.Certificates
	config := &tls.Config{}
	// TODO: figure out where to put the CA's
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}
	config.Certificates = make([]tls.Certificate, len(certs))
	// TODO: figure out a better way to use the provided hostname in the
	// configuration.
	for i, cert := range certs {
		var err error
		config.Certificates[i], err = tls.LoadX509KeyPair(cert.Certificate,
			cert.Key)
		if err != nil {
			return nil, err
		}
	}
	config.BuildNameToCertificate()
	return config, nil
}
