package overseer

import (
	"github.com/unixpickle/goule/pkg/config"
	"github.com/unixpickle/goule/pkg/exec"
	"github.com/unixpickle/goule/pkg/server"
	"github.com/unixpickle/goule/pkg/sessions"
	"net/http"
	"sync"
)

type Handler func(*Context)

// Overseer manages all goule services and synchronizes access to them.
type Overseer struct {
	handler       Handler
	mutex         sync.RWMutex
	configuration config.Config
	httpServer    *server.HTTP
	httpsServer   *server.HTTPS
	groups        exec.GroupMap
	sessions      *sessions.Manager
}

// NewOverseer creates a new overseer with a given configuration.
func NewOverseer(configuration *config.Config, handler Handler) *Overseer {
	httpRouter := schemeRouter{"http", nil}
	httpsRouter := schemeRouter{"https", nil}
	manager := sessions.NewManager()
	manager.SetSecret(configuration.Admin.PasswordHash)
	manager.SetTimeout(configuration.Admin.SessionTimeout)
	groups := exec.NewGroupMap()
	for _, service := range configuration.Services {
		groups.Add(service.Name, exec.NewGroup(service.Executables))
	}
	result := &Overseer{handler, sync.RWMutex{}, configuration.Copy(),
		server.NewHTTP(&httpRouter), server.NewHTTPS(&httpsRouter), groups,
		manager}
	httpRouter.overseer = result
	httpsRouter.overseer = result
	return result
}

// Start starts the overseer's servers and executables.
// This is thread-safe.
func (self *Overseer) Start() {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	// Start the servers
	if self.configuration.HTTPSettings.Enabled {
		self.httpServer.Start(self.configuration.HTTPSettings.Port)
	}
	if self.configuration.HTTPSSettings.Enabled {
		self.httpsServer.Start(self.configuration.HTTPSSettings.Port,
			self.configuration.TLS)
	}

	// Start executable groups
	self.groups.StartAutolaunch()
}

// Stop stops the overseer's servers and executables.
// This is thread-safe.
func (self *Overseer) Stop() {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	// Stop the servers
	self.httpServer.Stop()
	self.httpsServer.Stop()

	// Stop executable groups
	self.groups.StopAll()
}

// IsRunning returns true if any servers are running.
// This is thread-safe.
func (self *Overseer) IsRunning() bool {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.httpServer.IsRunning() || self.httpsServer.IsRunning()
}

// GetConfiguration returns a copy of the overseer's configuration.
// This is thread-safe.
func (self *Overseer) GetConfiguration() config.Config {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.configuration.Copy()
}

// GetSessions returns the overseer's session manager.
// This is thread-safe.
func (self *Overseer) GetSessions() *sessions.Manager {
	// No need to lock anything; the variable never changes.
	return self.sessions
}

type schemeRouter struct {
	scheme   string
	overseer *Overseer
}

func (self *schemeRouter) ServeHTTP(x http.ResponseWriter, y *http.Request) {
	context := NewContext(x, y, self.overseer, self.scheme)
	self.overseer.handler(context)
}
