package goule

import (
	"encoding/json"
	"errors"
	"github.com/unixpickle/gohttputil"
	"net/http"
	"reflect"
	"strings"
)

func (g *Goule) apiHandler(w http.ResponseWriter, r *http.Request) {
	// The path is "/api/APINAME"
	name := r.URL.Path[5:]
	
	// Make sure they are authorized to make this request.
	authed := w.Header().Get("Set-Cookie") != ""
	if !authed && name != "Auth" {
		gohttputil.RespondJSON(w, http.StatusForbidden, "Permissions denied.")
		return
	}

	// Read the contents of the request
	contents, err := gohttputil.ReadRequest(r, 0x10000)
	if err != nil {
		gohttputil.RespondJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Run the call
	ctx := &api{g, r, w}
	values, code, err := ctx.Do(name, contents)
	if err != nil {
		gohttputil.RespondJSON(w, code, err.Error())
		return
	}
	gohttputil.RespondJSON(w, http.StatusOK, values)
}

type api struct {
	*Goule
	r *http.Request
	w http.ResponseWriter
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

// Do performs an API.
func (a *api) Do(name string, body []byte) ([]interface{}, int, error) {
	// Find the method for the given API.
	method := reflect.ValueOf(a).MethodByName(name + "API")
	if !method.IsValid() {
		return nil, http.StatusNotFound, errors.New("Unknown API: " + name)
	}

	// Decode the array of JSON-encoded arguments.
	var rawArgs []string
	if err := json.Unmarshal(body, &rawArgs); err != nil {
		return nil, http.StatusBadRequest, err
	}

	// Decode the exact arguments.
	args, err := decodeArgs(method, rawArgs)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	// Run the call
	var res []reflect.Value
	if name == "Auth" || name == "Deauth" || strings.HasPrefix(name, "Set") {
		a.mutex.Lock()
		res = method.Call(args)
		a.mutex.Unlock()
	} else {
		a.mutex.RLock()
		res = method.Call(args)
		a.mutex.RUnlock()
	}

	// Convert the return value to an array of serializable objects.
	resList := make([]interface{}, len(res))
	for i, val := range res {
		rawValue := val.Interface()
		// Convert errors to strings
		if err, ok := rawValue.(error); ok {
			rawValue = err.Error()
		}
		resList[i] = rawValue
	}

	return resList, 0, nil
}

// SetPasswordAPI sets the new administrative password.
func (a *api) SetPasswordAPI(password string) {
	a.config.Admin.Hash = Hash(password)
	a.config.Save()
}

func decodeArgs(method reflect.Value, raw []string) ([]reflect.Value, error) {
	// Make sure they passed the right number of arguments
	if method.Type().NumIn() != len(raw) {
		return nil, errors.New("Invalid number of arguments.")
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
