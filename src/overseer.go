package goule

import (
	"net/http"
	"sync"
)

type Overseer struct {
	httpServer    *HTTPServer
	httpsServer   *HTTPSServer
	adminMutex    sync.RWMutex
	admin         AdminSettings
	servicesMutex sync.RWMutex
	services      []*Service
	sessions      *Sessions
}

type schemeRouter struct {
	scheme   string
	overseer *Overseer
}

type ServiceDescription struct {
	Name            string            `json:"name"`
	ForwardRules    []ForwardRule     `json:"forward_rules"`
	ExecutableInfos []ExecutableInfo  `json:"executable_infos"`
	ExecutableStats []ExecutableStats `json:"executable_stats"`
}

func (self *schemeRouter) ServeHTTP(x http.ResponseWriter, y *http.Request) {
	Route(NewRouteRequest(x, y, self.overseer, self.scheme))
}

// NewOverseer creates a new overseer with a completely disabled configuration.
func NewOverseer() *Overseer {
	httpRouter := schemeRouter{"http", nil}
	httpsRouter := schemeRouter{"https", nil}
	result := &Overseer{NewHTTPServer(&httpRouter),
		NewHTTPSServer(&httpsRouter), sync.RWMutex{}, AdminSettings{},
		sync.RWMutex{}, []*Service{}, NewSessions()}
	httpRouter.overseer = result
	httpsRouter.overseer = result
	return result
}

// Update updates the admin settings, the underlying web servers, and the
// services as needed to match a given configuration.
func (self *Overseer) Update(config *Configuration) {
	self.SetAdminSettings(config.Admin)
	self.httpServer.Update(config.HTTPSettings)
	self.httpsServer.UpdateTLS(config.TLS)
	self.httpsServer.Update(config.HTTPSSettings)
}

// GetConfiguration takes a snapshot of the overall state of the server at the
// given moment and returns a packed Configuration which describes that state.
func (self *Overseer) GetConfiguration() *Configuration {
	config := NewConfiguration()
	config.Admin = self.GetAdminSettings()
	config.HTTPSettings = self.httpServer.GetSettings()
	config.TLS = self.httpsServer.GetTLS()
	config.HTTPSSettings = self.httpsServer.GetSettings()
	return config
}

// SetAdminSettings updates the admin settings that this overseer uses.
func (self *Overseer) SetAdminSettings(s AdminSettings) {
	self.adminMutex.Lock()
	self.admin = s
	self.sessions.SetSecret(s.PasswordHash)
	self.sessions.SetTimeout(s.SessionTimeout)
	self.adminMutex.Unlock()
}

// GetAdminSettings returns the admin settings that this overseer uses.
func (self *Overseer) GetAdminSettings() AdminSettings {
	self.adminMutex.RLock()
	defer self.adminMutex.RUnlock()
	return self.admin
}

// GetSessions is a non-blocking non-locking operation
func (self *Overseer) GetSessions() *Sessions {
	return self.sessions
}

// GetServices takes a snapshot of all the active services.
func (self *Overseer) GetServices() []ServiceDescription {
	self.servicesMutex.RLock()
	defer self.servicesMutex.RUnlock()
	result := make([]ServiceDescription, len(self.services))
	for i, service := range self.services {
		result[i].Name = service.GetName()
		result[i].ForwardRules = service.GetForwardRules()
		info, stats := service.GetExecutables()
		result[i].ExecutableInfos = info
		result[i].ExecutableStats = stats
	}
	return result
}

func (self *Overseer) IsRunning() bool {
	return self.httpServer.IsRunning() || self.httpsServer.IsRunning()
}
