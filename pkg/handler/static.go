package handler

import (
	"net/http"
	"path"
	"regexp"
	"runtime"
)

func TryStatic(ctx *Context) bool {
	// The admin control URL should have a "/" after it.
	if ctx.Path == "" {
		http.Redirect(ctx.Response, ctx.Request, ctx.Rule.Path+"/",
			http.StatusMovedPermanently)
		return true
	}

	// Forward "/" to index page.
	if ctx.Path == "/" {
		ctx.Path = "/index.html"
		return TryStatic(ctx)
	}

	// Validate the path for a static file request
	charMatch := "[a-zA-Z0-9\\-_]*"
	htmlMatch := charMatch + "\\.html"
	cssMatch := "style\\/" + charMatch + "\\.css"
	scriptMatch := "scripts\\/" + charMatch + "\\.js"
	imageMatch := "images\\/" + charMatch + "\\.png"
	expr := "^\\/(" + htmlMatch + "|" + cssMatch + "|" + scriptMatch + "|" +
		imageMatch + ")$"
	if ok, _ := regexp.MatchString(expr, ctx.Path); !ok {
		return false
	}

	// Path is safe; serve static file.
	_, filename, _, _ := runtime.Caller(1)
	actualPath := path.Join(path.Dir(filename), "static"+ctx.Path)
	http.ServeFile(ctx.Response, ctx.Request, actualPath)
	return true
}
