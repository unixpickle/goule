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

type apiFunc func(*Context, []byte) (interface{}, error)

// TryAdminAPI checks if the passed context corresponds to an API call.
// Returns true if and only if the context was an API call.
// This will send a response and process the request synchronously.
func TryAdminAPI(ctx *Context) bool {
	// The API path must start with "/api/"
	if !strings.HasPrefix(ctx.Admin.Path, "/api/") {
		return false
	}

	// Get the API name from the URL path
	api := ctx.Admin.Path[5:]

	// Read the request body
	contents, err := readRequest(ctx.Request)
	if err != nil {
		respondJSON(ctx.Response, http.StatusBadRequest, err.Error())
		return true
	}

	RunAPICall(ctx, contents, api)
	return true
}

// RunAPICall runs an API call on a given context, given the request data.
// Returns false if any sort of API error occurred.
func RunAPICall(ctx *Context, contents []byte, api string) bool {
	// Prevent unauthorized requests
	if !ctx.Admin.Authorized && api != "auth" {
		respondJSON(ctx.Response, http.StatusUnauthorized,
			"Permissions denied.")
		return false
	}
	// Lookup the API and find the associated function
	handlers := map[string]apiFunc{"auth": AuthAPI,
		"services": ListServicesAPI, "change_password": ChangePasswordAPI,
		"set_http": SetHTTPAPI, "set_https": SetHTTPSAPI}
	handler, ok := handlers[api]
	if !ok {
		respondJSON(ctx.Response, http.StatusNotFound, "No such API: "+api)
		return false
	}
	// Run the API
	reply, err := handler(ctx, contents)
	if err != nil {
		respondJSON(ctx.Response, http.StatusBadRequest, err.Error())
		return false
	}
	// Send the APIs response
	respondJSON(ctx.Response, http.StatusOK, reply)
	return true
}

// The interface for the "auth" API call.
func AuthAPI(ctx *Context, body []byte) (interface{}, error) {
	var password string
	if err := json.Unmarshal(body, &password); err != nil {
		return nil, err
	}

	// Check the password
	hash := sha256.Sum256([]byte(password))
	hex := hex.EncodeToString(hash[:])
	adminHash := ctx.Overseer.GetConfiguration().Admin.PasswordHash
	if strings.ToLower(hex) != strings.ToLower(adminHash) {
		return nil, errors.New("The provided password was incorrect.")
	}

	// Create a new session
	sessionId := ctx.Overseer.GetSessions().Login()
	cookie := &http.Cookie{Name: SessionIdCookie, Value: sessionId,
		Path: ctx.Admin.Rule.Path, Domain: ctx.Admin.Rule.Hostname}
	http.SetCookie(ctx.Response, cookie)
	return true, nil
}

// The interface for the "services" API call.
func ListServicesAPI(ctx *Context, body []byte) (interface{}, error) {
	return ctx.Overseer.GetServiceDescriptions(), nil
}

// The interface for the "change_password" API call.
func ChangePasswordAPI(ctx *Context, body []byte) (interface{}, error) {
	var password string
	if err := json.Unmarshal(body, &password); err != nil {
		return nil, err
	}

	// Hash the password and save it
	hash := sha256.Sum256([]byte(password))
	hex := hex.EncodeToString(hash[:])
	ctx.Overseer.SetPasswordHash(strings.ToLower(hex))
	return true, nil
}

func SetHTTPAPI(ctx *Context, body []byte) (interface{}, error) {
	var settings ServerSettings
	if err := json.Unmarshal(body, &settings); err != nil {
		return nil, err
	}
	ctx.Overseer.SetHTTPSettings(settings)
	return true, nil
}

func SetHTTPSAPI(ctx *Context, body []byte) (interface{}, error) {
	var settings ServerSettings
	if err := json.Unmarshal(body, &settings); err != nil {
		return nil, err
	}
	ctx.Overseer.SetHTTPSSettings(settings)
	return true, nil
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
