package goule

import (
	"./exec"
	"./server"
	"./sessions"
	"net/http"
	"sync"
)

// Overseer manages all goule services and synchronizes access to them.
type Overseer struct {
	mutex         sync.RWMutex
	configuration Configuration
	httpServer    *server.HTTP
	httpsServer   *server.HTTPS
	groups        exec.GroupMap
	sessions      *sessions.Manager
}

// ServiceInfo contains the info contained in Service as well as live info about
// each executable.
type ServiceInfo struct {
	Name         string        `json:"name"`
	ForwardRules []ForwardRule `json:"forward_rules"`
	Executables  []exec.Info   `json:"executables"`
}

// NewOverseer creates a new overseer with a given configuration.
func NewOverseer(config Configuration) *Overseer {
	httpRouter := schemeRouter{"http", nil}
	httpsRouter := schemeRouter{"https", nil}
	sessions := sessions.NewManager()
	sessions.SetSecret(config.Admin.PasswordHash)
	sessions.SetTimeout(config.Admin.SessionTimeout)
	groups := exec.NewGroupMap()
	for _, service := range config.Services {
		groups.Add(service.Name, exec.NewGroup(service.Executables))
	}
	result := &Overseer{sync.RWMutex{}, config, server.NewHTTP(&httpRouter),
		server.NewHTTPS(&httpsRouter), groups, sessions}
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
		self.startHTTPS()
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
	self.configuration.Save()
	self.httpServer.Stop()
	if settings.Enabled {
		self.httpServer.Start(settings.Port)
	}
}

// SetHTTPSSettings updates the HTTPS settings and adjusts the server
// accordingly.
// This is thread-safe.
func (self *Overseer) SetHTTPSSettings(settings ServerSettings) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.configuration.HTTPSSettings = settings
	self.configuration.Save()
	// Restart the HTTPS server if needed
	if settings.Enabled {
		self.httpsServer.Stop()
		self.startHTTPS()
	}
}

// AddService adds a service based on its information.
// Returns false if and only if the new service's name conflicts with an
// existing service.
// This is thread-safe.
func (self *Overseer) AddService(service Service) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	for _, x := range self.configuration.Services {
		if x.Name == service.Name {
			return false
		}
	}

	// Add the service to the configuration.
	self.configuration.Services = append(self.configuration.Services, service)
	self.configuration.Save()

	// Run the new group.
	group := exec.NewGroup(service.Executables)
	self.groups.Add(service.Name, group)
	group.StartAutolaunch()

	return true
}

// RemoveService removes a service by name.
// Returns false if and only if the service could not be found.
// This is thread-safe.
func (self *Overseer) RemoveService(name string) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	found := false
	for i, x := range self.configuration.Services {
		if x.Name == name {
			// Remove the service by replacing it with the last service in the
			// list.
			idx := len(self.configuration.Services) - 1
			self.configuration.Services[i] = self.configuration.Services[idx]
			self.configuration.Services = self.configuration.Services[0:idx]
			found = true
			break
		}
	}
	if !found {
		return false
	}
	self.configuration.Save()
	self.groups.Remove(name)
	return true
}

// SetTLS sets the TLS settings for the HTTPS server.
// If the HTTPS server is running, it will be restarted.
// This is thread-safe.
func (self *Overseer) SetTLS(info server.TLSInfo) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	// Update the configuration
	self.configuration.TLS = info
	self.configuration.Save()
	// Restart the HTTPS server if needed
	if self.configuration.HTTPSSettings.Enabled {
		self.httpsServer.Stop()
		self.startHTTPS()
	}
}

// GetSessions returns the overseer's session manager.
// This is thread-safe.
func (self *Overseer) GetSessions() *sessions.Manager {
	// No need to lock anything; the variable never changes.
	return self.sessions
}

// GetServiceInfos returns ServiceInfo objects for every service.
// This is thread-safe.
func (self *Overseer) GetServiceInfos() []ServiceInfo {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	result := []ServiceInfo{}
	for _, info := range self.configuration.Services {
		if group, ok := self.groups[info.Name]; ok {
			desc := ServiceInfo{info.Name, info.ForwardRules, group.GetInfos()}
			result = append(result, desc)
		}
	}
	return result
}

func (self *Overseer) startHTTPS() {
	self.httpsServer.Start(self.configuration.HTTPSSettings.Port,
		self.configuration.TLS)
}

type schemeRouter struct {
	scheme   string
	overseer *Overseer
}

func (self *schemeRouter) ServeHTTP(x http.ResponseWriter, y *http.Request) {
	HandleContext(NewContext(x, y, self.overseer, self.scheme))
}
