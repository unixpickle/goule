package goule

import (
	"github.com/unixpickle/executor"
	"github.com/unixpickle/ezserver"
	"github.com/unixpickle/reverseproxy"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Goule struct {
	config   *Config
	http     *ezserver.HTTP
	https    *ezserver.HTTPS
	admin    *ezserver.HTTP
	mutex    sync.RWMutex
	services map[string]executor.Service
	sessions map[string]time.Time
}

func NewGoule(config *Config) *Goule {
	res := new(Goule)
	res.config = config
	res.sessions = map[string]time.Time{}

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
	res.services = map[string]executor.Service{}
	for name, config := range config.Services {
		res.services[name] = config.ToExecutorService()
	}

	return res
}

func (g *Goule) Start() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Start all the autolaunch services
	for name, service := range g.services {
		if g.config.Services[name].Autolaunch {
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
	// If the path begins with "/api/", it's an AJAX API call.
	if strings.HasPrefix(r.URL.Path, "/api/") {
		a := &api{g, w, r, r.URL.Query().Get("id")}
		a.Handle()
	} else {
		g.mutex.RLock()
		path := g.config.Admin.Assets
		g.mutex.RUnlock()
		Asset(path, w, r)
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
