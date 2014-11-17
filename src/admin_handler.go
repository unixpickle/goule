package goule

import (
	"net/http"
	"strings"
)

func AdminHandler(res http.ResponseWriter, req *http.Request, path string) {
	if strings.HasPrefix(path, "/api/") {
		apiName := path[5 : len(path)]
		AdminAPICall(res, req, apiName)
		return
	}
	res.Header().Set("Content-Type", "text/plain")
	res.Write([]byte("Welcome to the admin site! Path:" + path))
}

func AdminAPICall(res http.ResponseWriter, req *http.Request, api string) {
	res.Header().Set("Content-Type", "application/json")
	res.Write([]byte("\"Hi there, this is json!\""))
}
