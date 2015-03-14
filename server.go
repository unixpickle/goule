package main

import (
	"github.com/unixpickle/ezserver"
	"github.com/unixpickle/reverseproxy"
)

// A Server contains all the HTTP servers and the proxy object for a Goule
// instance.
type Server struct {
	Control *ezserver.HTTP
	HTTP    *ezserver.HTTP
	HTTPS   *ezserver.HTTPS
	Proxy   *reverseproxy.Proxy
}

// NewServer creates a server based on a configuration.
// The configuration musn't be locked; this will lock it read-only.
// This will start the server(s) which are marked to run at startup.
// If a server cannot be started, this returns an error.
func NewServer(cfg *Config, adminPort int) (*Server, error) {
	cfg.RLock()
	defer cfg.RUnlock()

	res := &Server{}

	// Create server-related objects.
	res.Control = ezserver.NewHTTP(Control{cfg, res})
	res.Proxy = reverseproxy.NewProxy(cfg.Rules)
	res.HTTP = ezserver.NewHTTP(res.Proxy)
	res.HTTPS = ezserver.NewHTTPS(res.Proxy, cfg.TLS)

	// Start admin server.
	if err := res.Control.Start(adminPort); err != nil {
		return nil, err
	}

	// Start HTTP server.
	if cfg.StartHTTP {
		if err := res.HTTP.Start(cfg.HTTPPort); err != nil {
			res.Control.Stop()
			return nil, err
		}
	}

	// Start HTTPS server.
	if cfg.StartHTTPS {
		if err := res.HTTPS.Start(cfg.HTTPSPort); err != nil {
			// NOTE: res.HTTP could be running even if StartHTTP was false
			// because the control server is running and someone (theoretically)
			// could have used it to start the server by hand.
			res.HTTP.Stop()
			res.Control.Stop()
			return nil, err
		}
	}
	
	return res, nil
}
