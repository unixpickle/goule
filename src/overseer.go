package goule

import (
	"net/http"
	"sync"
)

type Overseer struct {
	mutex         sync.RWMutex
	configuration Configuration
	httpServer    *HTTPServer
	httpsServer   *HTTPSServer
	groups        ExecutableGroups
	sessions      *Sessions
}

type ServiceDescription struct {
	Name         string                  `json:"name"`
	ForwardRules []ForwardRule           `json:"forward_rules"`
	Executables  []ExecutableDescription `json:"executables"`
}

// NewOverseer creates a new overseer with a given configuration.
func NewOverseer(config Configuration) *Overseer {
	httpRouter := schemeRouter{"http", nil}
	httpsRouter := schemeRouter{"https", nil}
	sessions := NewSessions()
	sessions.SetSecret(config.Admin.PasswordHash)
	sessions.SetTimeout(config.Admin.SessionTimeout)
	result := &Overseer{sync.RWMutex{}, config,
		NewHTTPServer(&httpRouter), NewHTTPSServer(&httpsRouter),
		NewExecutableGroups(config.Services), sessions}
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
	self.groups.Autolaunch()
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
func (self *Overseer) GetConfiguration() Configuration {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.configuration
}

// SetPasswordHash updates the password hash for this overseer.
// This will change the session secret.
// The new configuration will be saved before this returns.
// This is thread-safe.
func (self *Overseer) SetPasswordHash(newHash string) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.configuration.Admin.PasswordHash = newHash
	self.sessions.SetSecret(newHash)
	self.configuration.Save()
}

// SetHTTPSettings updates the HTTP settings and adjusts the server accordingly.
// This is thread-safe.
func (self *Overseer) SetHTTPSettings(settings ServerSettings) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.configuration.HTTPSettings = settings
	self.httpServer.Stop()
	if settings.Enabled {
		self.httpServer.Start(settings.Port)
	}
	self.configuration.Save()
}

// SetHTTPSSettings updates the HTTPS settings and adjusts the server
// accordingly.
// This is thread-safe.
func (self *Overseer) SetHTTPSSettings(settings ServerSettings) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.configuration.HTTPSSettings = settings
	self.httpsServer.Stop()
	if settings.Enabled {
		self.httpsServer.Start(settings.Port, self.configuration.TLS)
	}
	self.configuration.Save()
}

// GetSessions returns the overseer's session manager.
// This is thread-safe.
func (self *Overseer) GetSessions() *Sessions {
	// No need to lock anything; sessions never changes.
	return self.sessions
}

// GetServiceDescriptions returns descriptions for every service.
// This is thread-safe.
func (self *Overseer) GetServiceDescriptions() []ServiceDescription {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	result := []ServiceDescription{}
	for _, info := range self.configuration.Services {
		if group, ok := self.groups[info.Name]; ok {
			desc := ServiceDescription{info.Name, info.ForwardRules,
				group.GetDescriptions()}
			result = append(result, desc)
		}
	}
	return result
}

type schemeRouter struct {
	scheme   string
	overseer *Overseer
}

func (self *schemeRouter) ServeHTTP(x http.ResponseWriter, y *http.Request) {
	HandleContext(NewContext(x, y, self.overseer, self.scheme))
}
