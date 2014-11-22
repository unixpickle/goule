package goule

import (
	"net/http"
	"path"
	"regexp"
	"runtime"
)

func TrySite(ctx *AdminContext) bool {
	// The admin control URL should have a "/" after it.
	if ctx.Path == "" {
		http.Redirect(ctx.Response, ctx.Request, ctx.Rule.Path+"/",
			http.StatusMovedPermanently)
		return true
	}
	// Forward "/" to index page.
	if ctx.Path == "/" {
		ctx.Path = "/index.html"
		return TrySite(ctx)
	}

	// Validate the path for a static file request
	charMatch := "[a-zA-Z0-9\\-]*"
	htmlMatch := charMatch + "\\.html"
	cssMatch := "style\\/" + charMatch + "\\.css"
	scriptMatch := "scripts\\/" + charMatch + "\\.js"
	imageMatch := "images\\/" + charMatch + "\\.png"
	matched, _ := regexp.MatchString("^\\/("+htmlMatch+"|"+cssMatch+"|"+
		scriptMatch+"|"+imageMatch+")$", ctx.Path)

	if matched {
		// Serve static file
		_, filename, _, _ := runtime.Caller(1)
		actualPath := path.Join(path.Dir(filename), "../static"+ctx.Path)
		http.ServeFile(ctx.Response, ctx.Request, actualPath)
		return true
	} else {
		return false
	}
}
