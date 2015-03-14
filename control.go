package main

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/hoisie/mustache"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"
)

var Store = sessions.NewCookieStore(securecookie.GenerateRandomKey(16),
	securecookie.GenerateRandomKey(16))

// Control is an http.Handler which serves the web control panel.
type Control struct {
	Config *Config
	Server *Server
}

// ServeAsset serves a static asset.
func (c Control) ServeAsset(w http.ResponseWriter, r *http.Request) {
	urlPath := path.Clean(r.URL.Path)
	if data, err := Asset(urlPath[1:]); err != nil {
		http.NotFound(w, r)
	} else {
		mimeType := mime.TypeByExtension(path.Ext(urlPath))
		if mimeType == "" {
			mimeType = "text/plain"
		}
		w.Header().Set("Content-Type", mimeType)
		w.Write(data)
	}
}

// ServeGeneral serves requests for the general settings page.
func (c Control) ServeGeneral(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Use posted form data to update configuration.
		httpPort := r.PostFormValue("http")
		httpsPort := r.PostFormValue("https")
		startHTTP := r.PostFormValue("starthttp")
		startHTTPS := r.PostFormValue("starthttps")
		c.Config.Lock()
		c.Config.HTTPPort, _ = strconv.Atoi(httpPort)
		c.Config.HTTPSPort, _ = strconv.Atoi(httpsPort)
		c.Config.StartHTTP = (startHTTP == "On")
		c.Config.StartHTTPS = (startHTTPS == "On")
		c.Config.Unlock()
	}
	
	template := map[string]interface{}{}
	
	// Put server settings in template.
	c.Config.RLock()
	template["http"] = c.Config.HTTPPort
	template["https"] = c.Config.HTTPSPort
	template["startHTTP"] = c.Config.StartHTTP
	template["startHTTPS"] = c.Config.StartHTTPS
	c.Config.RUnlock()
	
	template["httpRunning"], template["httpPort"] = c.Server.HTTP.Status()
	template["httpsRunning"], template["httpsPort"] = c.Server.HTTPS.Status()
	
	serveTemplate(w, r, "general", template)
}

// ServeHTTP serves the web control panel.
func (c Control) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := path.Clean(r.URL.Path)
	if urlPath == "/login" {
		c.ServeLogin(w, r)
		return
	} else if strings.HasPrefix(urlPath, "/assets/") {
		c.ServeAsset(w, r)
		return
	} else if !isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}
	
	// Page routing for authenticated clients.
	pages := map[string]func(http.ResponseWriter, *http.Request){
		"/general": c.ServeGeneral, "/rules": c.ServeRules, "/tls": c.ServeTLS,
		"/http": c.ServeHTTPConfig, "/https": c.ServeHTTPSConfig,
		"/": c.ServeRoot}
	handler, ok := pages[urlPath]
	if !ok {
		handler = http.NotFound
	}
	handler(w, r)
}

// ServeHTTPConfig provides a basic link-driven API for controlling the HTTP
// server.
func (c Control) ServeHTTPConfig(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	switch query.Get("action") {
	case "start":
		c.Config.RLock()
		port := c.Config.HTTPPort
		c.Config.RUnlock()
		c.Server.HTTP.Start(port)
	case "stop":
		c.Server.HTTP.Stop()
	default:
		http.Error(w, "Invalid action.", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/general", http.StatusTemporaryRedirect)
}

// ServeHTTPSConfig provides a basic link-driven API for controlling the HTTPS
// server.
func (c Control) ServeHTTPSConfig(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	switch query.Get("action") {
	case "start":
		c.Config.RLock()
		port := c.Config.HTTPSPort
		c.Config.RUnlock()
		c.Server.HTTPS.Start(port)
	case "stop":
		c.Server.HTTPS.Stop()
	default:
		http.Error(w, "Invalid action.", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/general", http.StatusTemporaryRedirect)
}

// ServeLogin serves the login page.
func (c Control) ServeLogin(w http.ResponseWriter, r *http.Request) {
	template := map[string]interface{}{"error": false}
	if r.Method == "POST" {
		// Get their submitted hash and the real hash.
		password := r.PostFormValue("password")
		hash := HashPassword(password)
		GlobalConfig.RLock()
		realHash := GlobalConfig.AdminHash
		GlobalConfig.RUnlock()
		// Check if they got the password correct.
		if hash == realHash {
			s, _ := Store.Get(r, "sessid")
			s.Values["authenticated"] = true
			s.Save(r, w)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		template["error"] = true
	}

	// Serve login page with no template.
	data, err := Asset("templates/login.mustache")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	content := mustache.Render(string(data), template)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// ServeRoot serves the homepage (task list).
func (c Control) ServeRoot(w http.ResponseWriter, r *http.Request) {
	template := map[string]interface{}{}
	GlobalConfig.RLock()
	objects := make([]map[string]string, len(GlobalConfig.Tasks))
	for i, task := range GlobalConfig.Tasks {
		status := task.Status()
		statusStr := []string{"stopped", "running", "restarting"}[status]
		action := []string{"Start", "Stop", "Restarting"}[status]
		args := strings.Join(task.Args, " ")
		objects[i] = map[string]string{"action": action, "status": statusStr,
			"args": args}
	}
	template["tasks"] = objects
	GlobalConfig.RUnlock()

	serveTemplate(w, r, "tasks", template)
}

// ServeRules serves requests for the rules page.
func (c Control) ServeRules(w http.ResponseWriter, r *http.Request) {
	template := map[string]interface{}{}
	// TODO: fill template
	serveTemplate(w, r, "rules", template)
}

// ServeTLS serves requests for the TLS settings page.
func (c Control) ServeTLS(w http.ResponseWriter, r *http.Request) {
	template := map[string]interface{}{}
	// TODO: fill template
	serveTemplate(w, r, "tls", template)
}

// HashPassword returns the SHA-256 hash of a string.
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return strings.ToLower(hex.EncodeToString(hash[:]))
}

// isAuthenticated returns whether or not a request was authenticated.
func isAuthenticated(r *http.Request) bool {
	s, _ := Store.Get(r, "sessid")
	val, ok := s.Values["authenticated"].(bool)
	return ok && val
}

// serveTemplate serves a mustache template asset.
func serveTemplate(w http.ResponseWriter, r *http.Request, name string,
	info interface{}) {
	data, err := Asset("templates/" + name + ".mustache")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	content := mustache.Render(string(data), info)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}
