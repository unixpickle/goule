package main

import (
	"net/http"

	"github.com/gorilla/context"
	"github.com/unixpickle/ezserver"
	"github.com/unixpickle/goule/metrics"
	"github.com/unixpickle/reverseproxy"
)

const hstsPolicy = "max-age=31536000; includeSubDomains"

// A Server contains all the HTTP servers and the proxy object for a Goule instance.
type Server struct {
	Control *ezserver.HTTP
	HTTP    *ezserver.HTTP
	HTTPS   *ezserver.HTTPS
	Tracker *metrics.RequestTracker
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
	res.Tracker = metrics.NewRequestTracker()
	res.Control = ezserver.NewHTTP(context.ClearHandler(Control{cfg, res}))
	res.Proxy = reverseproxy.NewProxy(cfg.Rules)
	res.UpdateMetricsHosts()
	res.HTTP = ezserver.NewHTTP(res.Tracker.Wrap(res.Proxy))
	res.HTTPS = ezserver.NewHTTPS(withHSTS(res.Tracker.Wrap(res.Proxy)), cfg.TLS.TLS)
	res.HTTP.SetSecurityRedirects(cfg.TLS.Redirects)
	res.HTTP.SetAutocertHandler(res.HTTPS.HandleAutocertRequest)

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
			// NOTE: res.HTTP could be running even if StartHTTP was false because the control
			// server is running and someone (theoretically) could have used it to start the server
			// by hand.
			res.HTTP.Stop()
			res.Control.Stop()
			return nil, err
		}
	}

	return res, nil
}

func (s *Server) UpdateMetricsHosts() {
	rules := s.Proxy.RuleTable()
	var hosts []string
	for host := range rules {
		if host != "*" {
			hosts = append(hosts, host)
		}
	}
	s.Tracker.SetHosts(hosts)
}

// withHSTS tells browsers which reached this handler over HTTPS to keep using
// HTTPS for future requests.
func withHSTS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", hstsPolicy)
		next.ServeHTTP(w, r)
	})
}
