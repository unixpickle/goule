package goule

import "net/http"

type Router struct {
	Config       *Configuration
	Server       *HTTPServer
	SecureServer *HTTPSServer
}

type schemeRouter struct {
	scheme string
	router *Router
}

func (self *schemeRouter) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	self.router.HandleRequest(res, req, self.scheme)
}

func (self *Router) HandleRequest(res http.ResponseWriter, req *http.Request,
	scheme string) {
	self.Config.RLock()

	// The server won't know if TLS was used, so we need to specify manually.
	url := *req.URL
	url.Scheme = scheme
	url.Host = req.Host

	// TODO: here, we will search for a regular forward rule.

	// See if an admin forward rule matches the URL
	for _, source := range self.Config.AdminRules {
		if source.MatchesURL(&url) {
			// Get the path that they requested
			subpath := source.SubpathForURL(&url)
			// Unlock the configuration
			self.Config.RUnlock()
			// Handle the request
			AdminHandler(res, req, subpath)
			return
		}
	}

	self.Config.RUnlock()

	// No rules found. In the future, we might consider sending a meaningful 404
	// page.
	res.Header().Set("Content-Type", "text/plain")
	res.Write([]byte("No forward rule found."))
}

func (self *Router) Run() error {
	self.Config.RLock()
	runHTTP := self.Config.ServeHTTP
	runHTTPS := self.Config.ServeHTTPS
	httpPort := self.Config.HTTPPort
	httpsPort := self.Config.HTTPSPort
	self.Config.RUnlock()
	if runHTTP {
		if err := self.Server.Run(httpPort); err != nil {
			return err
		}
	}
	if runHTTPS {
		if err := self.SecureServer.Run(httpsPort); err != nil {
			self.Server.Stop()
			return err
		}
	}
	return nil
}

func NewRouter(config *Configuration) *Router {
	httpRouter := &schemeRouter{"http", nil}
	httpsRouter := &schemeRouter{"https", nil}
	server := NewHTTPServer(httpRouter)
	secureServer := NewHTTPSServer(httpsRouter, config)
	router := &Router{config, server, secureServer}
	httpRouter.router = router
	httpsRouter.router = router
	return router
}
