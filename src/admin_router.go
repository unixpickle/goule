package goule

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func AdminHandler(req *RouteRequest) {
	// Forward "/" to index page.
	if req.AdminPath == "/" || req.AdminPath == "" {
		req.AdminPath = "/index.html"
		AdminHandler(req)
		return
	}

	// Send the session cookie in the response.
	if req.Authorized {
		for _, cookie := range req.Request.Cookies() {
			http.SetCookie(req.Response, cookie)
		}
	}

	// Handle API calls
	if strings.HasPrefix(req.AdminPath, "/api/") {
		apiName := req.AdminPath[5:]
		RouteAPICall(req, apiName)
		return
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
	} else {
		// TODO: send a nice 404 page here
		req.Response.Header().Set("Content-Type", "text/plain")
		req.Response.WriteHeader(404)
		req.Response.Write([]byte("Invalid path: " + req.AdminPath))
	}
}

func RouteAPICall(req *RouteRequest, api string) {
	// Prevent unauthorized requests
	if !req.Authorized && api != "auth" {
		respondError(req.Response, http.StatusUnauthorized,
			"Permissions denied.")
		return
	}

	// Get the handler to use
	handlers := APIHandlers()
	handler, ok := handlers[api]
	if !ok {
		respondError(req.Response, http.StatusNotFound, "No such API: "+api)
		return
	}

	// Read the request
	contents, err := readRequest(req.Request)
	if err != nil {
		respondError(req.Response, http.StatusBadRequest, err.Error())
		return
	}

	// Run the API and return its response.
	responseData, err := handler(req, contents)
	if err != nil {
		respondError(req.Response, http.StatusBadRequest, err.Error())
	} else {
		req.Response.Header().Set("Content-Type", "application/json")
		req.Response.Header().Set("Content-Length",
			strconv.Itoa(len(responseData)))
		req.Response.Write(responseData)
	}
}

func readRequest(req *http.Request) ([]byte, error) {
	// Cap the data limit
	response := []byte{}
	for {
		next := make([]byte, 4096)
		num, err := req.Body.Read(next)
		response = append(response, next[0:num]...)
		if len(response) > 0x10000 {
			return nil, errors.New("Request exceeded 0x10000 bytes.")
		}
		if err != nil {
			break
		}
	}
	return response, nil
}

func respondError(res http.ResponseWriter, code int, msg string) {
	res.Header().Set("Content-Type", "application/json")
	data := []byte("\"Unable to encode error!\"")
	if marshaled, err := json.Marshal(msg); err == nil {
		data = marshaled
	}
	res.Header().Set("Content-Length", strconv.Itoa(len(data)))
	res.WriteHeader(code)
	res.Write(data)
}
