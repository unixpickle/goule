package handler

import (
	"encoding/json"
	"errors"
	"github.com/unixpickle/goule/pkg/config"
	"github.com/unixpickle/goule/pkg/exec"
	"github.com/unixpickle/goule/pkg/httputil"
	"github.com/unixpickle/goule/pkg/proxy"
	"github.com/unixpickle/goule/pkg/server"
	"net/http"
	"reflect"
	"strings"
)

type apiFunc func(*Context, []byte) (interface{}, error)

// TryAPI runs an API if applicable and returns whether or not it performed an
// API call.
func TryAPI(ctx *Context) bool {
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
// Returns false if and only if any sort of error occurred.
func RunAPICall(ctx *Context, api string, body []byte) bool {
	// Prevent unauthorized requests
	if !ctx.Authorized && api != "Auth" {
		httputil.RespondJSON(ctx.Response, http.StatusUnauthorized,
			"Permissions denied.")
		return false
	}
	// Find the method corresponding to the API call
	ctxVal := reflect.ValueOf(ctx)
	method := ctxVal.MethodByName(api + "API")
	if !method.IsValid() {
		httputil.RespondJSON(ctx.Response, http.StatusNotFound, "No API: "+api)
		return false
	}
	args := make([]reflect.Value, method.Type().NumIn())
	if method.Type().NumIn() == 1 {
		inputType := method.Type().In(0)
		dec := reflect.New(inputType)
		if err := json.Unmarshal(body, dec.Interface()); err != nil {
			httputil.RespondJSON(ctx.Response, http.StatusBadRequest,
				err.Error())
			return false
		}
		args[0] = reflect.Indirect(dec)
	}
	res := method.Call(args)
	if len(res) == 0 {
		// Empty return value => send an empty JSON object "{}"
		return httputil.RespondJSON(ctx.Response, http.StatusOK,
			map[string]string{})
	} else if len(res) == 1 {
		// Single return value could be an error.
		retVal := res[0].Interface()
		if err, ok := retVal.(error); ok {
			return httputil.RespondJSON(ctx.Response, http.StatusBadRequest,
				err)
		} else {
			return httputil.RespondJSON(ctx.Response, http.StatusOK, retVal)
		}
	} else {
		if !res[1].IsNil() {
			err := res[1].Interface().(error)
			return httputil.RespondJSON(ctx.Response, http.StatusBadRequest,
				err.Error())
		} else {
			return httputil.RespondJSON(ctx.Response, http.StatusOK,
				res[0].Interface())
		}
	}
}

// AuthAPI authenticates the user by checking a submitted password.
func (ctx *Context) AuthAPI(password string) error {
	// Check the password
	admin := ctx.Overseer.GetConfiguration().Admin
	if !admin.CheckPassword(password) {
		return errors.New("The provided password was incorrect.")
	}

	// Create a new session
	sessionId := ctx.Overseer.GetSessions().Login()
	cookie := &http.Cookie{Name: SessionIdCookie, Value: sessionId,
		Path: ctx.Rule.Path, Domain: ctx.Rule.Hostname}
	if cookie.Domain == "localhost" {
		cookie.Domain = ""
	}
	http.SetCookie(ctx.Response, cookie)
	return nil
}

// ListServicesAPI returns a list of service infos.
func (ctx *Context) ListServicesAPI() interface{} {
	return ctx.Overseer.GetServiceInfos()
}

// ChangePasswordAPI changes the admin password.
func (ctx *Context) ChangePasswordAPI(password string) {
	// Hash the password and save it
	ctx.Overseer.SetPasswordHash(config.HashPassword(password))
}

// SetHTTPAPI sets the HTTP server settings.
func (ctx *Context) SetHTTPAPI(settings config.ServerSettings) {
	ctx.Overseer.SetHTTPSettings(settings)
}

// SetHTTPSAPI is the interface for the "set_https" API call.
func (ctx *Context) SetHTTPSAPI(settings config.ServerSettings) {
	ctx.Overseer.SetHTTPSSettings(settings)
}

func (ctx *Context) SetTLSAPI(tls server.TLSInfo) {
	ctx.Overseer.SetTLS(tls)
}

func (ctx *Context) SetAdminRulesAPI(rules []config.SourceURL) {
	ctx.Overseer.SetAdminRules(rules)
}

func (ctx *Context) RenameServiceAPI(oldNew []string) error {
	if len(oldNew) != 2 {
		return errors.New("Expecting two array items.")
	}
	if ctx.Overseer.RenameService(oldNew[0], oldNew[1]) {
		return nil
	} else {
		return errors.New("Named service does not exist.")
	}
}

func (ctx *Context) SetServiceRulesAPI(info setRulesCall) error {
	if ctx.Overseer.SetServiceRules(info.name, info.rules) {
		return nil
	} else {
		return errors.New("Named service does not exist.")
	}
}

func (ctx *Context) SetServiceExecutablesAPI(info setExecutablesCall) error {
	if ctx.Overseer.SetServiceExecutables(info.name, info.execs) {
		return nil
	} else {
		return errors.New("Named service does not exist.")
	}
}

func (ctx *Context) GetConfigurationAPI() interface{} {
	return ctx.Overseer.GetConfiguration()
}

func (ctx *Context) SetAdminSessionTimeoutAPI(num int) {
	ctx.Overseer.SetSessionTimeout(num)
}

func (ctx *Context) AddServiceAPI(service config.Service) error {
	if ctx.Overseer.AddService(&service) {
		return nil
	} else {
		return errors.New("Named service already exists.")
	}
}

func (ctx *Context) SetProxyAPI(settings proxy.Settings) {
	ctx.Overseer.SetProxySettings(settings)
}

type setRulesCall struct {
	name  string               `json:"name"`
	rules []config.ForwardRule `json:"rules"`
}

type setExecutablesCall struct {
	name  string          `json:"name"`
	execs []exec.Settings `json:"settings"`
}
