package goule

import (
	"encoding/json"
	"errors"
	"github.com/unixpickle/gohttputil"
	"net/http"
	"reflect"
	"strings"
)

// APIHandler handles an API HTTP request.
func (g *Goule) APIHandler(w http.ResponseWriter, r *http.Request) {
	// The path is "/api/APINAME"
	api := r.URL.Path[5:]

	// Make sure they are authorized to make this request.
	authed := w.Header().Get("Set-Cookie") != ""
	if !authed && api != "Auth" {
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
	values, code, err := g.APICall(api, contents)
	if err != nil {
		gohttputil.RespondJSON(w, code, err.Error())
		return
	}

	// The "Auth" call is special--it creates a new cookie.
	if api == "Auth" && values[0].(bool) {
		id := g.sessions.login()
		cookie := &http.Cookie{Name: SessionIdCookie, Value: id}
		http.SetCookie(w, cookie)
	}

	gohttputil.RespondJSON(w, http.StatusOK, values)
}

// APICall runs an API call on the Goule object.
func (g *Goule) APICall(name string, body []byte) ([]interface{}, int, error) {
	// Find the method for the given API.
	method := reflect.ValueOf(g).MethodByName(name + "API")
	if !method.IsValid() {
		return nil, http.StatusNotFound, errors.New("Unknown API: " + name)
	}

	// Decode the raw arguments.
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
	if strings.HasPrefix(name, "Set") || name == "Auth" {
		g.mutex.Lock()
		res = method.Call(args)
		g.mutex.Unlock()
	} else {
		g.mutex.RLock()
		res = method.Call(args)
		g.mutex.RUnlock()
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

// AuthAPI returns whether the given password is correct.
func (g *Goule) AuthAPI(password string) bool {
	return g.config.Admin.Try(password)
}

// SetPasswordAPI sets the new administrative password.
func (g *Goule) SetPasswordAPI(password string) {
	g.config.Admin.Hash = Hash(password)
	g.config.Save()
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
