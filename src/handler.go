package goule

import "net/http"

type Handler func(http.ResponseWriter, *http.Request)

func CreateHandler(config *Configuration, https bool) Handler {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/plain")
		res.Write([]byte("No forward rule found."))
	}
}
