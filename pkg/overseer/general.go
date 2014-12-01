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
	self.Set(func() {
		self.configuration.Admin.PasswordHash = newHash
		self.sessions.SetSecret(newHash)
	})
}

// SetHTTPSettings updates the HTTP settings.
// This is thread-safe.
func (self *Overseer) SetHTTPSettings(settings config.ServerSettings) {
	self.Set(func() {
		self.configuration.HTTPSettings = settings
		self.httpServer.Stop()
		if settings.Enabled {
			self.httpServer.Start(settings.Port)
		}
	})
}

// SetHTTPSSettings updates the HTTPS settings.
// This is thread-safe.
func (self *Overseer) SetHTTPSSettings(settings config.ServerSettings) {
	self.Set(func() {
		self.configuration.HTTPSSettings = settings
		self.httpsServer.Stop()
		if settings.Enabled {
			self.httpsServer.Start(settings.Port, self.configuration.TLS)
		}
	})
}

// SetTLS sets the TLS settings for the HTTPS server.
// If the HTTPS server is running, it will be restarted.
// This is thread-safe.
func (self *Overseer) SetTLS(info server.TLSInfo) {
	self.Set(func() {
		self.configuration.TLS = info.Copy()
		if self.configuration.HTTPSSettings.Enabled {
			self.httpsServer.Stop()
			self.httpsServer.Start(self.configuration.HTTPSSettings.Port, info)
		}
	})
}

// SetAdminRules sets the admin rules.
// This is thread-safe.
func (self *Overseer) SetAdminRules(rules []config.SourceURL) {
	self.Set(func() {
		self.configuration.Admin.Rules = make([]config.SourceURL, len(rules))
		copy(self.configuration.Admin.Rules, rules)
	})
}

// SetSessionTimeout sets the session timeout.
// This is thread-safe.
func (self *Overseer) SetSessionTimeout(timeout int) {
	self.Set(func() {
		self.configuration.Admin.SessionTimeout = timeout
		self.sessions.SetTimeout(timeout)
	})
}

// SetProxySettings sets the proxy settings.
// This is thread-safe.
func (self *Overseer) SetProxySettings(settings proxy.Settings) {
	self.Set(func() {
		self.configuration.Proxy = settings
	})
}
