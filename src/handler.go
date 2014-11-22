package goule

import (
	"net/http"
	"net/url"
)

const SessionIdCookie = "goule_id"

type AdminContext struct {
	Authorized bool
	Path       string
	Rule       SourceURL
}

type Context struct {
	Response http.ResponseWriter
	Request  *http.Request
	URL      url.URL
	Overseer *Overseer
	Admin    *AdminContext
}

func NewContext(res http.ResponseWriter, req *http.Request, overseer *Overseer,
	scheme string) *Context {
	// Reverse-engineer the incoming URL
	url := *req.URL
	url.Scheme = scheme
	url.Host = req.Host
	return &Context{res, req, url, overseer, nil}
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
			// Configure the administrative fields of the RouteRequest
			path := source.SubpathForURL(&ctx.URL)
			ctx.Admin = &AdminContext{false, path, source}
			cookie, _ := ctx.Request.Cookie(SessionIdCookie)
			if cookie != nil {
				if ctx.Overseer.GetSessions().Validate(cookie.Value) {
					ctx.Admin.Authorized = true
					cookieCopy := *cookie
					cookieCopy.Path = source.Path
					http.SetCookie(ctx.Response, &cookieCopy)
				}
			}
			if !TryAdminSite(ctx) {
				TryAdminAPI(ctx)
			}
			return true
		}
	}
	return false
}
