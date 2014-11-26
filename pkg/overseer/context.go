package overseer

import (
	"net/http"
	"net/url"
)

type Context struct {
	Response http.ResponseWriter
	Request  *http.Request
	URL      url.URL
	Overseer *Overseer
}

func NewContext(res http.ResponseWriter, req *http.Request,
	overseer *Overseer, scheme string) *Context {
	// Reverse-engineer the incoming URL
	url := *req.URL
	url.Scheme = scheme
	url.Host = req.Host
	return &Context{res, req, url, overseer}
}
