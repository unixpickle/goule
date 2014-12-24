package goule

import (
	"github.com/unixpickle/ezserver"
	"github.com/unixpickle/reverseproxy"
	"net/http"
	"strings"
	"sync"
)

type Goule struct {
	config   *Config
	http     *ezserver.HTTP
	https    *ezserver.HTTPS
	admin    *ezserver.HTTP
	mutex    sync.RWMutex
	services []Service
	sessions *sessions
}

func NewGoule(config *Config) *Goule {
	res := new(Goule)
	res.sessions = newSessions(res)
	res.config = config
	res.http = ezserver.NewHTTP(res)
	res.https = ezserver.NewHTTPS(res, &config.TLS)

	// Create admin server
	admin := func(w http.ResponseWriter, r *http.Request) {
		res.AdminHTTP(w, r)
	}
	res.admin = ezserver.NewHTTP(http.HandlerFunc(admin))

	// Create services (but don't start them)
	res.services = make([]Service, len(config.Services))
	for i, cfg := range config.Services {
		res.services[i] = NewService(&cfg)
	}

	return res
}

func (g *Goule) AdminHTTP(w http.ResponseWriter, r *http.Request) {
	// If the path begins with "/api/", it's an AJAX API call.
	if strings.HasPrefix(r.URL.Path, "/api/") {
		g.APIHandler(w, r)
	} else {
		g.StaticHandler(w, r)
	}
}

func (g *Goule) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mutex.RLock()
	for _, rule := range g.config.Rules {
		if rule.MatchesRequest(r) {
			ws := g.config.Websockets
			g.mutex.RUnlock()
			reverseproxy.Proxy(w, r, &rule, ws)
			return
		}
	}
	g.mutex.RUnlock()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<html><body>No forward rule found.</body></html>"))
}

func (g *Goule) Start() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Start all the autolaunch services
	for _, service := range g.services {
		if service.Config().Autolaunch {
			service.Start()
		}
	}

	// Start all the HTTP or HTTPS servers.
	if err := g.admin.Start(g.config.Admin.Port); err != nil {
		g.stopServices()
		return err
	}
	if g.config.ServeHTTP {
		if err := g.http.Start(g.config.HTTPPort); err != nil {
			g.stopServices()
			g.admin.Stop()
			return err
		}
	}
	if g.config.ServeHTTPS {
		if err := g.https.Start(g.config.HTTPSPort); err != nil {
			g.stopServices()
			g.admin.Stop()
			if g.config.ServeHTTP {
				g.http.Stop()
			}
			return err
		}
	}

	return nil
}

func (g *Goule) Stop() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Ignore all errors
	g.stopServices()
	g.admin.Stop()
	g.http.Stop()
	g.https.Stop()
}

func (g *Goule) stopServices() {
	for _, service := range g.services {
		service.Stop()
	}
}
