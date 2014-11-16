package goule

import "net/http"

type Router struct {
	Config       *Configuration
	Server       *HTTPServer
	SecureServer *HTTPSServer
}

func (self *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/plain")
	res.Write([]byte("No forward rule found."))
}

func (self *Router) Run() error {
	self.Config.Lock()
	runHTTP := self.Config.ServeHTTP
	runHTTPS := self.Config.ServeHTTPS
	httpPort := self.Config.HTTPPort
	httpsPort := self.Config.HTTPSPort
	self.Config.Unlock()
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
	server := NewHTTPServer(nil)
	secureServer := NewHTTPSServer(nil, config)
	router := &Router{config, server, secureServer}
	server.handler = router
	secureServer.handler = router
	return router
}
