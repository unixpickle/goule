package goule

import (
	"net/http"
	"path"
	"regexp"
	"runtime"
)

func RouteAdminSite(req *RouteRequest) bool {
	// Forward "/" to index page.
	if req.AdminPath == "/" || req.AdminPath == "" {
		req.AdminPath = "/index.html"
		return RouteAdminSite(req)
	}

	// Validate the path for a static file request
	charMatch := "[a-zA-Z0-9\\-]"
	htmlMatch := charMatch + "*\\.html"
	cssMatch := "style\\/" + charMatch + "*\\.css"
	scriptMatch := "script\\/" + charMatch + "*\\.js"
	matched, _ := regexp.MatchString("^\\/("+htmlMatch+"|"+cssMatch+"|"+
		scriptMatch+")$", req.AdminPath)

	if matched {
		// Serve static file
		_, filename, _, _ := runtime.Caller(1)
		actualPath := path.Join(path.Dir(filename), "../static"+req.AdminPath)
		http.ServeFile(req.Response, req.Request, actualPath)
		return true
	} else {
		return false
	}
}
