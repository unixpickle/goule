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
func NewOverseer(config *Configuration) *Overseer {
	httpRouter := schemeRouter{"http", nil}
	httpsRouter := schemeRouter{"https", nil}
	sessions := sessions.NewManager()
	sessions.SetSecret(config.Admin.PasswordHash)
	sessions.SetTimeout(config.Admin.SessionTimeout)
	groups := exec.NewGroupMap()
	for _, service := range config.Services {
		groups.Add(service.Name, exec.NewGroup(service.Executables))
	}
	result := &Overseer{sync.RWMutex{}, config.Copy(),
		server.NewHTTP(&httpRouter), server.NewHTTPS(&httpsRouter), groups,
		sessions}
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
func (self *Overseer) GetConfiguration() Configuration {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.configuration.Copy()
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

// SetHTTPSettings updates the HTTP settings.
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

// SetHTTPSSettings updates the HTTPS settings.
// This is thread-safe.
func (self *Overseer) SetHTTPSSettings(settings ServerSettings) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.configuration.HTTPSSettings = settings
	self.configuration.Save()
	// Restart the HTTPS server if needed
	self.httpsServer.Stop()
	if settings.Enabled {
		self.httpsServer.Start(settings.Port, self.configuration.TLS)
	}
}

// AddService adds a service.
// Returns false if and only if the new service's name conflicts with an
// existing service.
// This is thread-safe.
func (self *Overseer) AddService(service *Service) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	// Make sure the service does not exist.
	if self.indexOfService(service.Name) >= 0 {
		return false
	}

	// Add the service to the configuration.
	self.configuration.Services = append(self.configuration.Services,
		service.Copy())
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

	// Find the service.
	index := self.indexOfService(name)
	if index < 0 {
		return false
	}

	// Remove the service from the configuration
	last := len(self.configuration.Services) - 1
	self.configuration.Services[index] = self.configuration.Services[last]
	self.configuration.Services = self.configuration.Services[0:last]
	self.configuration.Save()

	// Remove the executables associated with the service
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
	self.configuration.TLS = info.Copy()
	self.configuration.Save()
	// Restart the HTTPS server if needed
	if self.configuration.HTTPSSettings.Enabled {
		self.httpsServer.Stop()
		self.httpsServer.Start(self.configuration.HTTPSSettings.Port, info)
	}
}

// SetAdminRules sets the admin rules.
// This is thread-safe.
func (self *Overseer) SetAdminRules(rules []SourceURL) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.configuration.Admin.Rules = make([]SourceURL, len(rules))
	copy(self.configuration.Admin.Rules, rules)
	self.configuration.Save()
}

// SetServiceRules sets the forward rules for a service.
// Returns false if and only if the named service does not exist.
// This is thread-safe.
func (self *Overseer) SetServiceRules(name string, rules []ForwardRule) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	index := self.indexOfService(name)
	if index < 0 {
		return false
	}

	self.configuration.Services[index].ForwardRules =
		make([]ForwardRule, len(rules))
	copy(self.configuration.Services[index].ForwardRules, rules)
	self.configuration.Save()
	return true
}

// SetServiceExecutables sets the executables for a service.
// If the service had pre-existing executables, they will be stopped and
// replaced.
// The newly added executables will not automatically be executed.
// Returns false if and only if the named service does not exist.
// This is thread-safe.
func (self *Overseer) SetServiceExecutables(name string,
	execs []exec.Settings) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	index := self.indexOfService(name)
	if index < 0 {
		return false
	}

	service := &self.configuration.Services[index]

	// Update the configuration.
	service.Executables = make([]exec.Settings, len(execs))
	for i := range execs {
		service.Executables[i] = execs[i].Copy()
	}
	self.configuration.Save()

	// Update the executable group.
	self.groups.Remove(name)
	self.groups.Add(name, exec.NewGroup(execs))
	return true
}

// RenameService renames a service without stopping its execution.
// Returns false if and only if the named service does not exist.
// This is thread-safe.
func (self *Overseer) RenameService(oldName string, newName string) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	index := self.indexOfService(oldName)
	if index < 0 {
		return false
	}

	// Update the configuration.
	self.configuration.Services[index].Name = newName
	self.configuration.Save()

	// Update the executable group.
	self.groups.Rename(oldName, newName)
	return true
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
			rules := make([]ForwardRule, len(info.ForwardRules))
			copy(rules, info.ForwardRules)
			desc := ServiceInfo{info.Name, rules, group.GetInfos()}
			result = append(result, desc)
		}
	}
	return result
}

func (self *Overseer) indexOfService(name string) int {
	for i, x := range self.configuration.Services {
		if x.Name == name {
			return i
		}
	}
	return -1
}

type schemeRouter struct {
	scheme   string
	overseer *Overseer
}

func (self *schemeRouter) ServeHTTP(x http.ResponseWriter, y *http.Request) {
	HandleContext(NewContext(x, y, self.overseer, self.scheme))
}
