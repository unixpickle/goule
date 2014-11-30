package overseer

import (
	"github.com/unixpickle/goule/pkg/config"
	"github.com/unixpickle/goule/pkg/proxy"
	"github.com/unixpickle/goule/pkg/server"
)

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
func (self *Overseer) SetHTTPSettings(settings config.ServerSettings) {
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
func (self *Overseer) SetHTTPSSettings(settings config.ServerSettings) {
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
func (self *Overseer) SetAdminRules(rules []config.SourceURL) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.configuration.Admin.Rules = make([]config.SourceURL, len(rules))
	copy(self.configuration.Admin.Rules, rules)
	self.configuration.Save()
}

// SetSessionTimeout sets the session timeout.
// This is thread-safe.
func (self *Overseer) SetSessionTimeout(timeout int) {
	self.mutex.Lock()
	self.mutex.Unlock()
	self.configuration.Admin.SessionTimeout = timeout
	self.configuration.Save()
	self.sessions.SetTimeout(timeout)
}

// SetProxySettings sets the proxy settings.
// This is thread-safe.
func (self *Overseer) SetProxySettings(settings proxy.Settings) {
	self.mutex.Lock()
	self.mutex.Unlock()
	self.configuration.Proxy = settings
	self.configuration.Save()
}
