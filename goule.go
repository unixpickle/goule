package goule

import (
	"github.com/unixpickle/ezserver"
	"github.com/unixpickle/reverseproxy"
	"net/http"
	"strings"
	"sync"
)

// SessionIdCookie is the cookie name for the Goule session ID.
const SessionIdCookie = "goule_id"

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

	// Create the two regular servers
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		res.serviceHTTP(w, r)
	}
	handler := http.HandlerFunc(handlerFunc)
	res.http = ezserver.NewHTTP(handler)
	res.https = ezserver.NewHTTPS(handler, &config.TLS)

	// Create admin server
	admin := func(w http.ResponseWriter, r *http.Request) {
		res.adminHTTP(w, r)
	}
	res.admin = ezserver.NewHTTP(http.HandlerFunc(admin))

	// Create services (but don't start them)
	res.services = make([]Service, len(config.Services))
	for i, cfg := range config.Services {
		res.services[i] = NewService(&cfg)
	}

	return res
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

func (g *Goule) adminHTTP(w http.ResponseWriter, r *http.Request) {
	// Set the session cookie if necessary
	if cookie, err := r.Cookie(SessionIdCookie); err == nil {
		g.mutex.Lock()
		if g.sessions.validate(cookie.Value) {
			http.SetCookie(w, cookie)
		}
		g.mutex.Unlock()
	}

	// If the path begins with "/api/", it's an AJAX API call.
	if strings.HasPrefix(r.URL.Path, "/api/") {
		g.apiHandler(w, r)
	} else {
		g.staticHandler(w, r)
	}
}

func (g *Goule) serviceHTTP(w http.ResponseWriter, r *http.Request) {
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

func (g *Goule) stopServices() {
	for _, service := range g.services {
		service.Stop()
	}
}
