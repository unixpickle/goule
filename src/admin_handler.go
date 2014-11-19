package goule

import (
	"net/http"
	pathlib "path"
	"regexp"
	"runtime"
	"strings"
)

func AdminHandler(res http.ResponseWriter, req *http.Request, path string) {
	if path == "/" || path == "" {
		AdminHandler(res, req, "/index.html")
		return
	}

	if strings.HasPrefix(path, "/api/") {
		apiName := path[5:len(path)]
		AdminAPICall(res, req, apiName)
		return
	}

	// Validate the path
	charMatch := "[a-zA-Z0-9\\-]"
	htmlMatch := charMatch + "*\\.html"
	cssMatch := "style\\/" + charMatch + "*\\.css"
	matched, _ := regexp.MatchString("^\\/("+htmlMatch+"|"+cssMatch+
		")$", path)

	if matched {
		// Serve static file
		_, filename, _, _ := runtime.Caller(1)
		actualPath := pathlib.Join(pathlib.Dir(filename), "../static"+path)
		http.ServeFile(res, req, actualPath)
	} else {
		res.Header().Set("Content-Type", "text/plain")
		res.Write([]byte("Invalid path: " + path))
	}
}

func AdminAPICall(res http.ResponseWriter, req *http.Request, api string) {
	res.Header().Set("Content-Type", "application/json")
	res.Write([]byte("\"Hi there, this is json!\""))
}
