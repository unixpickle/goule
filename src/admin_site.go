package goule

import (
	"net/http"
	"path"
	"regexp"
	"runtime"
)

func TryAdminSite(ctx *Context) bool {
	// Forward "/" to index page.
	if ctx.Admin.Path == "/" || ctx.Admin.Path == "" {
		ctx.Admin.Path = "/index.html"
		return TryAdminSite(ctx)
	}

	// Validate the path for a static file request
	charMatch := "[a-zA-Z0-9\\-]"
	htmlMatch := charMatch + "*\\.html"
	cssMatch := "style\\/" + charMatch + "*\\.css"
	scriptMatch := "script\\/" + charMatch + "*\\.js"
	matched, _ := regexp.MatchString("^\\/("+htmlMatch+"|"+cssMatch+"|"+
		scriptMatch+")$", ctx.Admin.Path)

	if matched {
		// Serve static file
		_, filename, _, _ := runtime.Caller(1)
		actualPath := path.Join(path.Dir(filename), "../static"+ctx.Admin.Path)
		http.ServeFile(ctx.Response, ctx.Request, actualPath)
		return true
	} else {
		return false
	}
}
