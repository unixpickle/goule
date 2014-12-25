package goule

import (
	"encoding/json"
	"github.com/unixpickle/executor"
	"github.com/unixpickle/ezserver"
	"github.com/unixpickle/gohttputil"
	"github.com/unixpickle/reverseproxy"
	"net/http"
	"reflect"
)

type api struct {
	*Goule
	w http.ResponseWriter
	r *http.Request
}

// AddRuleAPI adds a new proxy rule.
func (a *api) AddRuleAPI(rule reverseproxy.Rule) {
	a.config.Rules = append(a.config.Rules, rule)
	a.config.Save()
}

// AddServiceAPI adds a new service and possibly starts it.
func (a *api) AddServiceAPI(name string, cfg Service) error {
	if _, ok := a.services[name]; ok {
		return ErrNameTaken
	}

	// Create the executor.Service and possibly start it
	excService := cfg.ToExecutorService()
	a.services[name] = excService
	if cfg.Autolaunch {
		excService.Start()
	}

	// Update configuration
	a.config.Services[name] = cfg
	a.config.Save()
	return nil
}

// AuthAPI returns whether the given password is correct.
func (a *api) AuthAPI(password string) bool {
	if !a.config.Admin.Try(password) {
		return false
	}
	// Create a new cookie and set it.
	id := a.sessions.login()
	cookie := &http.Cookie{Name: SessionIdCookie, Value: id}
	http.SetCookie(a.w, cookie)
	return true
}

// Call performs an API.
func (a *api) Call(name string, body []byte) (int, error) {
	// Find the method for the given API.
	method := reflect.ValueOf(a).MethodByName(name + "API")
	if !method.IsValid() {
		return http.StatusNotFound, ErrUnknownAPI
	}

	// Decode the array of JSON-encoded arguments.
	var rawArgs []string
	if err := json.Unmarshal(body, &rawArgs); err != nil {
		return http.StatusBadRequest, err
	}

	// Decode the exact arguments.
	args, err := decodeArgs(method, rawArgs)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Lock the mutex in the appropriate way
	if isWriteAPI(name) {
		a.mutex.Lock()
		defer a.mutex.Unlock()
	} else {
		a.mutex.RLock()
		defer a.mutex.RUnlock()
	}

	// Convert the return value to an array of serializable objects.
	res := method.Call(args)
	resList := make([]interface{}, len(res))
	for i, val := range res {
		rawValue := val.Interface()
		// Convert errors to strings
		if err, ok := rawValue.(error); ok {
			rawValue = err.Error()
		}
		resList[i] = rawValue
	}

	// Encode the result
	gohttputil.RespondJSON(a.w, http.StatusOK, resList)
	return 0, nil
}

// ConfigAPI returns the full server configuration.
func (a *api) ConfigAPI() *Config {
	return a.config
}

// DeauthAPI does nothing.
func (a *api) DeauthAPI() {
	// Invalidate the current session
	cookie, _ := a.r.Cookie(SessionIdCookie)
	a.sessions.logout(cookie.Value)

	// Delete the cookie on the client-side
	content := SessionIdCookie + "=deleted; " +
		"expires=Thu, 01 Jan 1970 00:00:00 GMT"
	a.w.Header()["Set-Cookie"] = []string{content}
}

// DeleteRuleAPI deletes a rule by value
func (a *api) DeleteRuleAPI(rule reverseproxy.Rule) error {
	for i, r := range a.config.Rules {
		if rulesEqual(r, rule) {
			// Remove the rule
			a.config.Rules = append(a.config.Rules[0:i],
				a.config.Rules[i+1:]...)
			a.config.Save()
			return nil
		}
	}
	return ErrRuleNotFound
}

// DeleteServiceAPI deletes a service by name.
func (a *api) DeleteServiceAPI(name string) error {
	service, ok := a.services[name]
	if !ok {
		return ErrServiceNotFound
	}
	service.Stop()
	delete(a.services, name)
	delete(a.config.Services, name)
	a.config.Save()
	return nil
}

// Handle handles the API call and writes a JSON response.
func (a *api) Handle() {
	// The path is "/api/APINAME"
	name := a.r.URL.Path[5:]

	// Make sure they are authorized to make this request.
	authed := a.w.Header().Get("Set-Cookie") != ""
	if !authed && name != "Auth" {
		gohttputil.RespondJSON(a.w, http.StatusForbidden, ErrPermissionsDenied)
		return
	}

	// Read the contents of the request
	contents, err := gohttputil.ReadRequest(a.r, 0x10000)
	if err != nil {
		gohttputil.RespondJSON(a.w, http.StatusBadRequest, err.Error())
		return
	}

	// Run the call
	if code, err := a.Call(name, contents); err != nil {
		gohttputil.RespondJSON(a.w, code, err.Error())
		return
	}
}

// ServicesAPI returns an array of serviceDesc objects.
func (a *api) ServicesAPI() map[string]serviceDesc {
	res := map[string]serviceDesc{}
	for name, es := range a.services {
		s := a.config.Services[name]
		history, status := es.HistoryStatus()
		res[name] = serviceDesc{s, float64(history.LastStart.Unix()),
			float64(history.LastStop.Unix()), float64(history.LastError.Unix()),
			history.Error.Error(), status}
	}
	return res
}

// SetAdminPortAPI updates the admin port.
func (a *api) SetAdminPortAPI(port int) error {
	a.admin.Stop()
	if err := a.admin.Start(port); err != nil {
		// Attempt to restart it on the old port.
		a.admin.Start(a.config.Admin.Port)
		return err
	}
	// Port change was successful; save configuration
	a.config.Admin.Port = port
	a.config.Save()
	return nil
}

// SetAssetsAPI sets the admin assets path.
func (a *api) SetAssetsAPI(path string) {
	a.config.Admin.Assets = path
	a.config.Save()
}

// SetHTTP sets the HTTP port and enables/disables it.
func (a *api) SetHTTP(enable bool, port int) error {
	return a.setServer(a.http, &a.config.ServeHTTP, &a.config.HTTPPort, enable,
		port)
}

// SetHTTPS sets the HTTPS port and enables/disables it.
func (a *api) SetHTTPS(enable bool, port int) error {
	return a.setServer(a.https, &a.config.ServeHTTPS, &a.config.HTTPSPort,
		enable, port)
}

// SetPasswordAPI sets the new administrative password.
func (a *api) SetPasswordAPI(password string) {
	a.config.Admin.Hash = Hash(password)
	a.config.Save()
}

// SetRuleAPI replaces an old rule with a new rule
func (a *api) SetRuleAPI(old, rule reverseproxy.Rule) {
	for i, r := range a.config.Rules {
		if rulesEqual(r, old) {
			a.config.Rules[i] = rule
			a.config.Save()
			return
		}
	}
}

// SetSessionTimeoutAPI sets the session timeout in seconds.
func (a *api) SetSessionTimeoutAPI(timeout int) {
	a.config.Admin.Timeout = timeout
	a.config.Save()
}

// SetTLSAPI sets the TLS configuration for HTTPS.
func (a *api) SetTLSAPI(tls ezserver.TLSConfig) {
	a.https.SetTLSConfig(&tls)
	a.config.TLS = tls
	a.config.Save()
}

// StartAPI starts a service by name
func (a *api) StartAPI(name string) error {
	service, ok := a.services[name]
	if !ok {
		return ErrServiceNotFound
	}
	return service.Start()
}

// StopAPI stops a service by name
func (a *api) StopAPI(name string) error {
	service, ok := a.services[name]
	if !ok {
		return ErrServiceNotFound
	}
	return service.Stop()
}

// UpdateServiceAPI updates a service by name.
func (a *api) UpdateServiceAPI(name string, service Service) error {
	oldServ, ok := a.services[name]
	if !ok {
		return ErrServiceNotFound
	}
	oldServ.Stop()
	a.services[name] = service.ToExecutorService()
	a.config.Services[name] = service
	a.config.Save()
	return nil
}

// setServer sets settings on a given server.
// This exists to prevent repetition for HTTP and HTTPS server settings.
func (a *api) setServer(s ezserver.Server, enable *bool, port *int,
	newEnable bool, newPort int) error {
	s.Stop()
	// If disabled, we're done.
	if !newEnable {
		*enable = false
		a.config.Save()
		return nil
	}
	// Attempt to start the server on the new port
	if err := s.Start(newPort); err != nil {
		if *enable {
			s.Start(*port)
		}
		return err
	}
	// Save the configuration
	*enable = true
	*port = newPort
	a.config.Save()
	return nil
}

type serviceDesc struct {
	Service   Service         `json:"service"`
	LastStart float64         `json:"last_start"`
	LastStop  float64         `json:"last_stop"`
	LastError float64         `json:"last_error"`
	Error     string          `json:"error"`
	Status    executor.Status `json:"status"`
}

func decodeArgs(method reflect.Value, raw []string) ([]reflect.Value, error) {
	// Make sure they passed the right number of arguments
	if method.Type().NumIn() != len(raw) {
		return nil, ErrArgumentCount
	}

	// Decode each argument separately.
	res := make([]reflect.Value, len(raw))
	for i, rawArg := range raw {
		inputType := method.Type().In(i)
		dec := reflect.New(inputType)
		if err := json.Unmarshal([]byte(rawArg), dec.Interface()); err != nil {
			return nil, err
		}
		res[i] = reflect.Indirect(dec)
	}

	return res, nil
}

func isWriteAPI(name string) bool {
	return name != "Config" && name != "Services"
}

func rulesEqual(r1 reverseproxy.Rule, r2 reverseproxy.Rule) bool {
	return reflect.DeepEqual(r1, r2)
}
