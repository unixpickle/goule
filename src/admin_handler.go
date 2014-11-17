package goule

import (
	"net/http"
	"runtime"
	"strings"
	pathlib "path"
)

func AdminHandler(res http.ResponseWriter, req *http.Request, path string) {
	if path == "/" || path == "" {
		AdminHandler(res, req, "/index.html")
		return
	}
	
	if strings.HasPrefix(path, "/api/") {
		apiName := path[5 : len(path)]
		AdminAPICall(res, req, apiName)
		return
	}
	
	// TODO: here, validate the path before serving the file; this is a security
	// concern.
	
	// Serve static file
	_, filename, _, _ := runtime.Caller(1)
	actualPath := pathlib.Join(pathlib.Dir(filename), "../static" + path)
	http.ServeFile(res, req, actualPath)
}

func AdminAPICall(res http.ResponseWriter, req *http.Request, api string) {
	res.Header().Set("Content-Type", "application/json")
	res.Write([]byte("\"Hi there, this is json!\""))
}
