package main

import "net/http"

// Control is an http.Handler which serves the web control panel.
type Control struct {
	Config *Config
}

// ServeHTTP serves the web control panel.
func (c Control) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}
