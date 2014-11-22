package admin

import (
	"../"
)

const SessionIdCookie = "goule_id"

type Context struct {
	*goule.Context
	Authorized bool
	Path       string
	Rule       SourceURL
}

// NewContext creates a Context based on a goule.Context.
// This may block while it accesses the Overseer's sessions.Manager.
func NewContext(ctx *goule.Context, rule goule.SourceURL) *Context {
	path := rule.SubpathForURL(&ctx.URL)
	result := &Context{ctx, false, path, rule}
	cookie, _ := result.Request.Cookie(SessionIdCookie)
	if cookie != nil {
		if result.Overseer.GetSessions().Validate(cookie.Value) {
			result.Authorized = true
			cookieCopy := *cookie
			cookieCopy.Path = source.Path
			http.SetCookie(result.Response, &cookieCopy)
		}
	}
}
