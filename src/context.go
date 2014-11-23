package goule

import (
	"net/http"
	"net/url"
)

const SessionIdCookie = "goule_id"

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

type AdminContext struct {
	*Context
	Authorized bool
	Path       string
	Rule       SourceURL
}

// NewAdminContext creates an AdminContext based on a Context.
// This may block while it accesses the Overseer's sessions.Manager.
func NewAdminContext(ctx *Context, rule SourceURL) *AdminContext {
	path := rule.SubpathForURL(&ctx.URL)
	result := &AdminContext{ctx, false, path, rule}
	cookie, _ := result.Request.Cookie(SessionIdCookie)
	if cookie != nil {
		if result.Overseer.GetSessions().Validate(cookie.Value) {
			result.Authorized = true
			cookieCopy := *cookie
			cookieCopy.Path = rule.Path
			http.SetCookie(result.Response, &cookieCopy)
		}
	}
	return result
}
