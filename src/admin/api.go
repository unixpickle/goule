package admin

import (
	"../"
	"../httputil"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type apiFunc func(*Context, []byte) (interface{}, error)

// TryAPI runs an API if applicable and returns whether or not it performed an
// API call.
func TryAPI(ctx *Context) bool {
	// The API path must start with "/api/"
	if !strings.HasPrefix(ctx.Admin.Path, "/api/") {
		return false
	}

	// Get the API name from the URL path
	api := ctx.Admin.Path[5:]

	// Read the request body
	contents, err := httputil.ReadRequest(ctx.Request, 0x10000)
	if err != nil {
		httputil.RespondJSON(ctx.Response, http.StatusBadRequest, err.Error())
	} else {
		RunAPICall(ctx, api, contents)
	}
	return true
}

// RunAPICall runs an API call.
// Returns false if and only if an API error occurred.
// If the API returns a value which cannot be marshaled to JSON, RunAPICall
// returns true even though it responds with an error code.
func RunAPICall(ctx *Context, api string, contents []byte) bool {
	// Prevent unauthorized requests
	if !ctx.Authorized && api != "auth" {
		httputil.RespondJSON(ctx.Response, http.StatusUnauthorized,
			"Permissions denied.")
		return false
	}
	// Lookup the API and find the associated function
	handlers := map[string]apiFunc{"auth": AuthAPI,
		"services": ListServicesAPI, "change_password": ChangePasswordAPI,
		"set_http": SetHTTPAPI, "set_https": SetHTTPSAPI}
	handler, ok := handlers[api]
	if !ok {
		httputil.RespondJSON(ctx.Response, http.StatusNotFound, "No API: "+api)
		return false
	}
	// Run the API
	reply, err := handler(ctx, contents)
	if err != nil {
		httputil.RespondJSON(ctx.Response, http.StatusBadRequest, err.Error())
		return false
	} else {
		httputil.RespondJSON(ctx.Response, http.StatusOK, reply)
		return true
	}
}

// AuthAPI is the interface for the "auth" API call.
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

// ListServicesAPI is the interface for the "services" API call.
func ListServicesAPI(ctx *Context, body []byte) (interface{}, error) {
	return ctx.Overseer.GetServiceInfos(), nil
}

// ChangePasswordAPI is the interface for the "change_password" API call.
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

// SetHTTPAPI is the interface for the "set_http" API call.
func SetHTTPAPI(ctx *Context, body []byte) (interface{}, error) {
	var settings goule.ServerSettings
	if err := json.Unmarshal(body, &settings); err != nil {
		return nil, err
	}
	ctx.Overseer.SetHTTPSettings(settings)
	return true, nil
}

// SetHTTPSAPI is the interface for the "set_https" API call.
func SetHTTPSAPI(ctx *Context, body []byte) (interface{}, error) {
	var settings goule.ServerSettings
	if err := json.Unmarshal(body, &settings); err != nil {
		return nil, err
	}
	ctx.Overseer.SetHTTPSSettings(settings)
	return true, nil
}
