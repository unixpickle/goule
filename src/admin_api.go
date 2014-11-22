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

type handlerFunc func(*RouteRequest, []byte) ([]byte, error)

func RouteAdminAPI(req *RouteRequest) bool {
	if !strings.HasPrefix(req.AdminPath, "/api/") {
		return false
	}
	api := req.AdminPath[5:]

	// Prevent unauthorized requests
	if !req.Authorized && api != "auth" {
		respondError(req.Response, http.StatusUnauthorized,
			"Permissions denied.")
		return true
	}

	// Get the handler to use
	handlers := map[string]handlerFunc{"auth": AuthAPI,
		"services": ListServicesAPI}
	handler, ok := handlers[api]
	if !ok {
		respondError(req.Response, http.StatusNotFound, "No such API: "+api)
		return true
	}

	// Read the request
	contents, err := readRequest(req.Request)
	if err != nil {
		respondError(req.Response, http.StatusBadRequest, err.Error())
		return true
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
	return true
}

func AuthAPI(req *RouteRequest, body []byte) ([]byte, error) {
	var password string
	if err := json.Unmarshal(body, &password); err != nil {
		return nil, err
	}

	// Check the password
	hash := sha256.Sum256([]byte(password))
	hex := hex.EncodeToString(hash[:])
	adminHash := req.Overseer.GetAdminSettings().PasswordHash
	if strings.ToLower(hex) != strings.ToLower(adminHash) {
		return nil, errors.New("The provided password was incorrect.")
	}

	// Create a new session
	sessionId := req.Overseer.GetSessions().Login()
	cookie := &http.Cookie{Name: SessionIdCookie, Value: sessionId,
		Path: req.AdminRule.Path}
	http.SetCookie(req.Response, cookie)
	return []byte("\"Authentication successful.\""), nil
}

func ListServicesAPI(req *RouteRequest, body []byte) ([]byte, error) {
	return json.Marshal(req.Overseer.GetServices())
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
