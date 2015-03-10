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
func NewServer(cfg *Config) (*Server, error) {
	cfg.RLock()
	defer cfg.RUnlock()
	
	// Create server-related objects.
	ctrl := ezserver.NewHTTP(Control{cfg})
	proxy := reverseproxy.NewProxy(cfg.Rules)
	http := ezserver.NewHTTP(proxy)
	https := ezserver.NewHTTPS(proxy, cfg.TLS)
	
	// Start admin server.
	if err := ctrl.Start(cfg.AdminPort); err != nil {
		return nil, err
	}
	
	// Start HTTP server.
	if cfg.StartHTTP {
		if err := http.Start(cfg.HTTPPort); err != nil {
			ctrl.Stop()
			return nil, err
		}
	}
	
	// Start HTTPS server.
	if cfg.StartHTTPS {
		if err := https.Start(cfg.HTTPSPort); err != nil {
			if cfg.StartHTTP {
				http.Stop()
			}
			ctrl.Stop()
			return nil, err
		}
	}
	
	return &Server{ctrl, http, https, proxy}, nil
}
