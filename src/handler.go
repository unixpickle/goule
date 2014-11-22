package goule

import (
	"./admin"
	"net/http"
	"net/url"
)

type Context struct {
	Response http.ResponseWriter
	Request  *http.Request
	URL      url.URL
	Overseer *Overseer
}

func NewContext(res http.ResponseWriter, req *http.Request, overseer *Overseer,
	scheme string) *Context {
	// Reverse-engineer the incoming URL
	url := *req.URL
	url.Scheme = scheme
	url.Host = req.Host
	return &Context{res, req, url, overseer}
}

func HandleContext(ctx *Context) {
	if !TryAdmin(ctx) {
		if !TryService(ctx) {
			// TODO: send a nice 404 page here.
			ctx.Response.Header().Set("Content-Type", "text/plain")
			ctx.Response.Write([]byte("No forward rule found."))
		}
	}
}

func TryService(ctx *Context) bool {
	// TODO: here, check services' forward rules
	return false
}

func TryAdmin(ctx *Context) bool {
	for _, source := range ctx.Overseer.GetConfiguration().Admin.Rules {
		if source.MatchesURL(&ctx.URL) {
			adminContext := admin.NewContext(ctx, source)
			if !admin.TrySite(adminContext) {
				admin.TryAPI(adminContext)
			}
			return true
		}
	}
	return false
}
