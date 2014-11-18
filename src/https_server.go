package goule

import (
	"crypto/tls"
	"net"
	"net/http"
	"strconv"
	"sync"
)

type HTTPSServer struct {
	mutex      sync.RWMutex
	handler    http.Handler
	listener   *net.Listener
	listenPort int
	setting    ServerSettings
	tls        TLSInfo
}

func NewHTTPSServer(handler http.Handler) *HTTPSServer {
	return &HTTPSServer{sync.RWMutex{}, handler, nil, 0, ServerSettings{},
		TLSInfo{}}
}

// Update applies a given server setting to an HTTPSServer.
// If the setting is enabled but the receiver is not actively serving, it will
// start its server.
// Conversely, if the setting is disabled but the receiver is actively serving,
// it will stop.
// If both the setting and the receiver are serving, the server may still stop
// itself to change port numbers.
// The returned error will be nil unless the server could not start or restart.
func (self *HTTPSServer) Update(setting ServerSettings) error {
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

// UpdateTLS updates the certificates which this server will use via SNI and by
// default.
// If the server is actively running, this may trigger it to restart.
func (self *HTTPSServer) UpdateTLS(info TLSInfo) error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.tls = info
	if self.listener != nil {
		self.stop()
		return self.start()
	}
	return nil
}

// GetSetting returns the last setting which was passed via Update().
func (self *HTTPSServer) GetSettings() ServerSettings {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.setting
}

// GetTLS returns the last TLS info that was passed via UpdateTLS().
func (self *HTTPSServer) GetTLS() TLSInfo {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.tls
}

// IsRunning returns whether or not the server is actively listening for
// incoming connections.
func (self *HTTPSServer) IsRunning() bool {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.listener != nil
}

// start starts the server.
// This method assumes that the receiver is already write-locked.
func (self *HTTPSServer) start() error {
	config, err := self.createTLSConfig()
	if err != nil {
		return err
	}

	tcpListener, err := net.Listen("tcp", ":"+strconv.Itoa(self.setting.Port))
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
				self.stop()
			}
			self.mutex.Unlock()
		}
	}()

	return nil
}

// stop stops the listener.
// This method assumes that the receiver is already write-locked.
func (self *HTTPSServer) stop() {
	(*self.listener).Close()
	self.listener = nil
}

// createTLSConfig builds a TLS configuration that uses the certificates of the
// receiver.
func (self *HTTPSServer) createTLSConfig() (*tls.Config, error) {
	// TODO: in a future release of Go, this will be improved since they are
	// adding a GetCertificate() method to tls.Config!

	// Build up the tls.Config to have all the certificates we need
	certs := self.tls.Named
	config := &tls.Config{}

	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	// TODO: here, put the CAs into the configuration

	var err error
	config.Certificates = make([]tls.Certificate, 1)

	// Set the default certificate
	config.Certificates[0], err = tls.LoadX509KeyPair(
		self.tls.Default.CertificatePath, self.tls.Default.KeyPath)
	if err != nil {
		return nil, err
	}

	// idx is the current certificate index.
	idx := 1

	// Add each certificate to the Certificates list and NameToCertificate map.
	for host, cert := range certs {
		certPath := cert.CertificatePath
		keyPath := cert.KeyPath
		pair, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, err
		}
		config.Certificates = append(config.Certificates, pair)
		config.NameToCertificate[host] = &config.Certificates[idx]
		idx++
	}
	return config, nil
}
