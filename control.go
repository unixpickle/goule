package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/hoisie/mustache"
)

var Store = sessions.NewCookieStore(securecookie.GenerateRandomKey(16),
	securecookie.GenerateRandomKey(16))

// Control is an http.Handler which serves the web control panel.
type Control struct {
	Config *Config
	Server *Server
}

// ServeAddTask serves the add-task page.
func (c Control) ServeAddTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		serveTemplate(w, r, "add_task", map[string]interface{}{})
		return
	}
	taskJSON := r.PostFormValue("task")
	task := &Task{}
	if err := json.Unmarshal([]byte(taskJSON), task); err != nil {
		serveTemplate(w, r, "add_task", map[string]interface{}{"error": err.Error()})
		return
	}

	c.Config.Lock()
	c.Config.Tasks = append([]*Task{task}, c.Config.Tasks...)
	task.StartLoop()
	if task.AutoRun {
		task.Start()
	}
	c.Config.Save()
	c.Config.Unlock()

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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

// ServeChpass serves the change password POST target.
func (c Control) ServeChpass(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "You must POST to this API", http.StatusMethodNotAllowed)
		return
	}
	old := r.PostFormValue("old")
	newPass := r.PostFormValue("new")
	confirm := r.PostFormValue("confirm")
	c.Config.Lock()
	defer c.Config.Unlock()
	if HashPassword(old) != c.Config.AdminHash {
		http.Redirect(w, r, "/general?error=Password%20incorrect",
			http.StatusTemporaryRedirect)
		return
	}
	if newPass != confirm {
		http.Redirect(w, r, "/general?error=Passwords%20did%20not%20match",
			http.StatusTemporaryRedirect)
		return
	}
	c.Config.AdminHash = HashPassword(newPass)
	c.Config.Save()
	http.Redirect(w, r, "/general?success=Password%20changed",
		http.StatusTemporaryRedirect)
}

// ServeEditTask serves the task editor.
func (c Control) ServeEditTask(w http.ResponseWriter, r *http.Request) {
	index, err := strconv.Atoi(r.URL.Query().Get("index"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		c.Config.Lock()
		defer c.Config.Unlock()
		if index < 0 || index >= len(c.Config.Tasks) {
			http.Error(w, "Invalid task index", http.StatusBadRequest)
			return
		}

		oldStatus := c.Config.Tasks[index].Status()
		newTask := &Task{}
		if err := json.Unmarshal([]byte(r.PostFormValue("task")), newTask); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		c.Config.Tasks[index].StopLoop()
		c.Config.Tasks[index] = newTask
		c.Config.Save()
		newTask.StartLoop()
		if oldStatus != TaskStatusStopped {
			newTask.Start()
		}
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	c.Config.RLock()
	defer c.Config.RUnlock()
	if index < 0 || index >= len(c.Config.Tasks) {
		http.Error(w, "Invalid task index", http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(c.Config.Tasks[index])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	serveTemplate(w, r, "edit_task", map[string]interface{}{"taskData": string(data),
		"index": strconv.Itoa(index)})
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
		c.Config.Save()
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

	query := r.URL.Query()
	if errMsg := query.Get("error"); errMsg != "" {
		template["chpassError"] = errMsg
	} else if msg := query.Get("success"); msg != "" {
		template["chpassSuccess"] = msg
	}

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
		"/chpass": c.ServeChpass, "/": c.ServeRoot,
		"/setrules": c.ServeSetRules, "/add_task": c.ServeAddTask,
		"/start_task": c.ServeStartTask, "/stop_task": c.ServeStopTask,
		"/edit_task": c.ServeEditTask}
	handler, ok := pages[urlPath]
	if !ok {
		handler = http.NotFound
	}
	handler(w, r)
}

// ServeHTTPConfig provides a basic link-driven API for controlling the HTTP server.
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

// ServeHTTPSConfig provides a basic link-driven API for controlling the HTTPS server.
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
		c.Config.RLock()
		realHash := c.Config.AdminHash
		c.Config.RUnlock()
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
	c.Config.RLock()
	objects := make([]map[string]string, len(c.Config.Tasks))
	for i, task := range c.Config.Tasks {
		status := task.Status()
		statusStr := []string{"stopped", "running", "restarting"}[status]
		action := []string{"start", "stop", "stop"}[status]
		actionName := []string{"Start", "Stop", "Restarting"}[status]
		args := strings.Join(task.Args, " ")
		objects[i] = map[string]string{"action": action, "status": statusStr, "args": args,
			"actionName": actionName, "index": strconv.Itoa(i)}
	}
	template["tasks"] = objects
	c.Config.RUnlock()

	serveTemplate(w, r, "tasks", template)
}

// ServeRules serves requests for the rules page.
func (c Control) ServeRules(w http.ResponseWriter, r *http.Request) {
	template := map[string]interface{}{}

	// Encode the rules as JSON and put them in the template.
	c.Config.RLock()
	encoded, _ := json.Marshal(c.Config.Rules)
	c.Config.RUnlock()
	template["rules"] = string(encoded)

	serveTemplate(w, r, "rules", template)
}

// ServeSetRules serves requests for the page that sets the rules.
func (c Control) ServeSetRules(w http.ResponseWriter, r *http.Request) {
	// Get rules from the request.
	rulesData := r.URL.Query().Get("rules")
	var decoded map[string][]string
	if err := json.Unmarshal([]byte(rulesData), &decoded); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set rules in the configuration and server.
	c.Config.Lock()
	c.Config.Rules = decoded
	c.Server.Proxy.SetRuleTable(decoded)
	c.Config.Save()
	c.Config.Unlock()

	http.Redirect(w, r, "/rules", http.StatusTemporaryRedirect)
}

// ServeStartTask starts a task given its index.
func (c Control) ServeStartTask(w http.ResponseWriter, r *http.Request) {
	c.ServeTaskAction(w, r, true)
}

// ServeStopTask starts a task given its index.
func (c Control) ServeStopTask(w http.ResponseWriter, r *http.Request) {
	c.ServeTaskAction(w, r, false)
}

// ServeTLS serves requests for the TLS settings page.
func (c Control) ServeTLS(w http.ResponseWriter, r *http.Request) {
	template := map[string]interface{}{}
	// TODO: fill template
	serveTemplate(w, r, "tls", template)
}

// ServeTaskAction serves the start_task and stop_task pages.
func (c Control) ServeTaskAction(w http.ResponseWriter, r *http.Request, start bool) {
	index, err := strconv.Atoi(r.URL.Query().Get("index"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c.Config.Lock()
	defer c.Config.Unlock()
	if index < 0 || index >= len(c.Config.Tasks) {
		http.Error(w, "Invalid task index", http.StatusBadRequest)
		return
	}

	if start {
		c.Config.Tasks[index].Start()
	} else {
		c.Config.Tasks[index].Stop()
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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
