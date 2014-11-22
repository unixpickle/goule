package goule

import (
	"net/http"
	"path"
	"regexp"
	"runtime"
	"strings"
)

func AdminHandler(req *RouteRequest) {
	if req.AdminPath == "/" || req.AdminPath == "" {
		req.AdminPath = "/index.html"
		AdminHandler(req)
		return
	}

	if strings.HasPrefix(req.AdminPath, "/api/") {
		apiName := req.AdminPath[5:]
		AdminAPICall(req, apiName)
		return
	}

	// Validate the path
	charMatch := "[a-zA-Z0-9\\-]"
	htmlMatch := charMatch + "*\\.html"
	cssMatch := "style\\/" + charMatch + "*\\.css"
	matched, _ := regexp.MatchString("^\\/("+htmlMatch+"|"+cssMatch+
		")$", req.AdminPath)

	if matched {
		// Serve static file
		_, filename, _, _ := runtime.Caller(1)
		actualPath := path.Join(path.Dir(filename), "../static"+req.AdminPath)
		http.ServeFile(req.Response, req.Request, actualPath)
	} else {
		// TODO: send a nice 404 page here
		req.Response.Header().Set("Content-Type", "text/plain")
		req.Response.Write([]byte("Invalid path: " + req.AdminPath))
	}
}

func AdminAPICall(req *RouteRequest, api string) {
	req.Response.Header().Set("Content-Type", "application/json")
	req.Response.Write([]byte("\"Hi there! This is json.\""))
}
