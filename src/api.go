package goule

import (
	"./exec"
	"./httputil"
	"./server"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type apiFunc func(*AdminContext, []byte) (interface{}, error)

// TryAPI runs an API if applicable and returns whether or not it performed an
// API call.
func TryAPI(ctx *AdminContext) bool {
	// The API path must start with "/api/"
	if !strings.HasPrefix(ctx.Path, "/api/") {
		return false
	}

	// Get the API name from the URL path
	api := ctx.Path[5:]

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
func RunAPICall(ctx *AdminContext, api string, contents []byte) bool {
	// Prevent unauthorized requests
	if !ctx.Authorized && api != "auth" {
		httputil.RespondJSON(ctx.Response, http.StatusUnauthorized,
			"Permissions denied.")
		return false
	}
	// Lookup the API and find the associated function
	handlers := map[string]apiFunc{"auth": AuthAPI,
		"services": ListServicesAPI, "change_password": ChangePasswordAPI,
		"set_http": SetHTTPAPI, "set_https": SetHTTPSAPI,
		"set_admin_rules": SetAdminRulesAPI, "rename": RenameServiceAPI,
		"set_service_rules": SetServiceRulesAPI,
		"set_service_execs": SetServiceExecsAPI}
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
func AuthAPI(ctx *AdminContext, body []byte) (interface{}, error) {
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
		Path: ctx.Rule.Path, Domain: ctx.Rule.Hostname}
	http.SetCookie(ctx.Response, cookie)
	return true, nil
}

// ListServicesAPI is the interface for the "services" API call.
func ListServicesAPI(ctx *AdminContext, body []byte) (interface{}, error) {
	return ctx.Overseer.GetServiceInfos(), nil
}

// ChangePasswordAPI is the interface for the "change_password" API call.
func ChangePasswordAPI(ctx *AdminContext, body []byte) (interface{}, error) {
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
func SetHTTPAPI(ctx *AdminContext, body []byte) (interface{}, error) {
	var settings ServerSettings
	if err := json.Unmarshal(body, &settings); err != nil {
		return nil, err
	}
	ctx.Overseer.SetHTTPSettings(settings)
	return true, nil
}

// SetHTTPSAPI is the interface for the "set_https" API call.
func SetHTTPSAPI(ctx *AdminContext, body []byte) (interface{}, error) {
	var settings ServerSettings
	if err := json.Unmarshal(body, &settings); err != nil {
		return nil, err
	}
	ctx.Overseer.SetHTTPSSettings(settings)
	return true, nil
}

func SetTLSAPI(ctx *AdminContext, body []byte) (interface{}, error) {
	var tls server.TLSInfo
	if err := json.Unmarshal(body, &tls); err != nil {
		return nil, err
	}
	ctx.Overseer.SetTLS(tls)
	return true, nil
}

func SetAdminRulesAPI(ctx *AdminContext, body []byte) (interface{}, error) {
	var rules []SourceURL
	if err := json.Unmarshal(body, &rules); err != nil {
		return nil, err
	}
	ctx.Overseer.SetAdminRules(rules)
	return true, nil
}

func RenameServiceAPI(ctx *AdminContext, body []byte) (interface{}, error) {
	var oldNew []string
	if err := json.Unmarshal(body, &oldNew); err != nil {
		return nil, err
	}
	if len(oldNew) != 2 {
		return nil, errors.New("Expecting two array items.")
	}
	res := ctx.Overseer.RenameService(oldNew[0], oldNew[1])
	return res, nil
}

func SetServiceRulesAPI(ctx *AdminContext, body []byte) (interface{}, error) {
	var info setRulesCall
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	res := ctx.Overseer.SetServiceRules(info.name, info.rules)
	return res, nil
}

func SetServiceExecsAPI(ctx *AdminContext, body []byte) (interface{}, error) {
	var info setExecutablesCall
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	res := ctx.Overseer.SetServiceExecutables(info.name, info.execs)
	return res, nil
}

type setRulesCall struct {
	name  string        `json:"name"`
	rules []ForwardRule `json:"rules"`
}

type setExecutablesCall struct {
	name  string          `json:"name"`
	execs []exec.Settings `json:"settings"`
}
