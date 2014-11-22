package goule

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type apiFunc func(*RouteContext, []byte) (interface{}, error)

func RouteAdminAPI(ctx *RouteContext) bool {
	// The API path must start with "/api/"
	if !strings.HasPrefix(ctx.AdminPath, "/api/") {
		return false
	}

	// Get the API name from the URL path
	api := ctx.AdminPath[5:]

	// Read the request body
	contents, err := readRequest(ctx.Request)
	if err != nil {
		respondJSON(ctx.Response, http.StatusBadRequest, err.Error())
		return true
	}

	runAPI(ctx, contents, api)
	return true
}

func runAPI(ctx *RouteContext, contents []byte, api string) {
	// Prevent unauthorized requests
	if !ctx.Authorized && api != "auth" {
		respondJSON(ctx.Response, http.StatusUnauthorized,
			"Permissions denied.")
		return
	}
	// Lookup the API and find the associated function
	handlers := map[string]apiFunc{"auth": AuthAPI,
		"services": ListServicesAPI}
	handler, ok := handlers[api]
	if !ok {
		respondJSON(ctx.Response, http.StatusNotFound, "No such API: "+api)
		return
	}
	// Run the API
	reply, err := handler(ctx, contents)
	if err != nil {
		respondJSON(ctx.Response, http.StatusBadRequest, err.Error())
		return
	}
	// Send the APIs response
	respondJSON(ctx.Response, http.StatusOK, reply)
}

func AuthAPI(ctx *RouteContext, body []byte) (interface{}, error) {
	var password string
	if err := json.Unmarshal(body, &password); err != nil {
		return nil, err
	}

	// Check the password
	hash := sha256.Sum256([]byte(password))
	hex := hex.EncodeToString(hash[:])
	adminHash := ctx.Overseer.GetAdminSettings().PasswordHash
	if strings.ToLower(hex) != strings.ToLower(adminHash) {
		return nil, errors.New("The provided password was incorrect.")
	}

	// Create a new session
	sessionId := ctx.Overseer.GetSessions().Login()
	cookie := &http.Cookie{Name: SessionIdCookie, Value: sessionId,
		Path: ctx.AdminRule.Path}
	http.SetCookie(ctx.Response, cookie)
	return "Authentication successful.", nil
}

func ListServicesAPI(ctx *RouteContext, body []byte) (interface{}, error) {
	return ctx.Overseer.GetServices(), nil
}

func readRequest(ctx *http.Request) ([]byte, error) {
	// Cap the data limit
	response := []byte{}
	for {
		next := make([]byte, 4096)
		num, err := ctx.Body.Read(next)
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

func respondJSON(res http.ResponseWriter, code int, msg interface{}) {
	res.Header().Set("Content-Type", "application/json")
	if marshaled, err := json.Marshal(msg); err == nil {
		res.Header().Set("Content-Length", strconv.Itoa(len(marshaled)))
		res.WriteHeader(code)
		res.Write(marshaled)
	} else {
		data := []byte("Failed to encode object")
		res.Header().Set("Content-Length", strconv.Itoa(len(data)))
		// Preserve error code (if there is one)
		if code == 200 {
			res.WriteHeader(http.StatusInternalServerError)
		} else {
			res.WriteHeader(code)
		}
		res.Write(data)
	}

}
