package goule

import "net/http"

func AdminHandler(res http.ResponseWriter, req *http.Request, path string) {
	res.Header().Set("Content-Type", "text/plain")
	res.Write([]byte("Welcome to the admin site! Path:" + path))
}
