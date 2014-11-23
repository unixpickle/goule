package server

import (
	"crypto/tls"
	"net"
	"net/http"
	"strconv"
	"sync"
)

type CertificateInfo struct {
	CertificatePath string   `json:"certificate_path"`
	KeyPath         string   `json:"key_path"`
	AuthorityPaths  []string `json:"authority_paths"`
}

type TLSInfo struct {
	Named   map[string]CertificateInfo `json:"named_certificates"`
	Default CertificateInfo            `json:"default_certificates"`
}

type HTTPS struct {
	mutex      sync.RWMutex
	handler    http.Handler
	listener   *net.Listener
	listenPort int
}

// NewHTTPS creates a new HTTPS with a given handler.
// The newly created HTTPS will not be listening.
func NewHTTPS(handler http.Handler) *HTTPS {
	return &HTTPS{sync.RWMutex{}, handler, nil, 0}
}

// Start starts the server on the specified port with the specified TLS info.
// An error is returned if the server cannot be started or is already running.
// This is thread-safe.
func (self *HTTPS) Start(port int, info TLSInfo) error {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	config, err := createTLSConfig(&info)
	if err != nil {
		return err
	}

	tcpListener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
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

// Stop stops the server if it was running.
// This is thread-safe.
func (self *HTTPS) Stop() {
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
func (self *HTTPS) Status() (bool, int) {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.listener != nil, self.listenPort
}

// IsRunning returns the first return value of Status.
func (self *HTTPS) IsRunning() bool {
	x, _ := self.Status()
	return x
}

// createTLSConfig builds a TLS configuration.
func createTLSConfig(info *TLSInfo) (*tls.Config, error) {
	// Build up the tls.Config to have all the certificates we need
	certs := info.Named
	config := &tls.Config{}

	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	// TODO: here, put the CAs into the configuration

	var err error
	config.Certificates = make([]tls.Certificate, 1)

	// Set the default certificate
	config.Certificates[0], err = tls.LoadX509KeyPair(
		info.Default.CertificatePath, info.Default.KeyPath)
	if err != nil {
		return nil, err
	}

	// Add each certificate to the Certificates list and NameToCertificate map.
	certIdx := 1
	for host, cert := range certs {
		certPath := cert.CertificatePath
		keyPath := cert.KeyPath
		pair, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, err
		}
		config.Certificates = append(config.Certificates, pair)
		config.NameToCertificate[host] = &config.Certificates[certIdx]
		certIdx++
	}
	return config, nil
}
