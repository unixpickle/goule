package handler

import (
	"github.com/unixpickle/goule/pkg/config"
	"github.com/unixpickle/goule/pkg/overseer"
	"net/http"
)

const SessionIdCookie = "goule_id"

type Context struct {
	*overseer.Context
	Authorized bool
	Path       string
	Rule       config.SourceURL
}

// NewContext creates a Context based on an overseer.Context.
// This may block while it accesses the Overseer's sessions.Manager.
func NewContext(ctx *overseer.Context, rule config.SourceURL) *Context {
	path := rule.SubpathForURL(&ctx.URL)
	result := &Context{ctx, false, path, rule}
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
