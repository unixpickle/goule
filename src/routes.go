package goule

import (
	"net/http"
	"net/url"
)

const SessionIdCookie = "goule_id"

type RouteContext struct {
	Response   http.ResponseWriter
	Request    *http.Request
	URL        url.URL
	Overseer   *Overseer
	Authorized bool
	AdminPath  string
	AdminRule  SourceURL
}

func NewRouteContext(res http.ResponseWriter, ctx *http.Request,
	overseer *Overseer, scheme string) *RouteContext {
	// The server won't know if TLS was used, so we need to specify manually.
	url := *ctx.URL
	url.Scheme = scheme
	url.Host = ctx.Host
	return &RouteContext{res, ctx, url, overseer, false, "", SourceURL{}}
}

func Route(ctx *RouteContext) {
	if !ForwardToAdmin(ctx) {
		if !ForwardToService(ctx) {
			// TODO: send a nice 404 page here.
			ctx.Response.Header().Set("Content-Type", "text/plain")
			ctx.Response.Write([]byte("No forward rule found."))
		}
	}
}

func ForwardToService(ctx *RouteContext) bool {
	// TODO: here, check services' forward rules
	return false
}

func ForwardToAdmin(ctx *RouteContext) bool {
	for _, source := range ctx.Overseer.GetAdminSettings().Rules {
		if source.MatchesURL(&ctx.URL) {
			// Configure the administrative fields of the RouteRequest
			ctx.AdminRule = source
			ctx.AdminPath = source.SubpathForURL(&ctx.URL)
			cookie, _ := ctx.Request.Cookie(SessionIdCookie)
			if cookie != nil {
				if ctx.Overseer.GetSessions().Validate(cookie.Value) {
					ctx.Authorized = true
					http.SetCookie(ctx.Response, cookie)
				}
			}
			if !RouteAdminSite(ctx) {
				RouteAdminAPI(ctx)
			}
			return true
		}
	}
	return false
}
