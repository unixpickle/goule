package goule

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

type DestinationURL struct {
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
	Port     int    `json:"port"`
	Path     string `json:"path"`
}

type ForwardRule struct {
	From SourceURL      `json:"from"`
	To   DestinationURL `json:"to"`
}

type Executable struct {
	Dirname          string            `json:"dirname"`
	LogStdout        bool              `json:"log_stdout"`
	LogStderr        bool              `json:"log_stderr"`
	SetGroupId       bool              `json:"set_group_id"`
	SetUserId        bool              `json:"set_user_id"`
	GroupId          int               `json:"group_id"`
	UserId           int               `json:"user_id"`
	Arguments        []string          `json:"arguments"`
	Environment      map[string]string `json:"environment"`
	Autolaunch       bool              `json:"autolaunch"`
	Relaunch         bool              `json:"relaunch"`
	RelaunchInterval int               `json:"relaunch_interval"`
}

type Service struct {
	Name         string        `json:"name"`
	ForwardRules []ForwardRule `json:"forward_rules"`
	Executables  []Executable  `json:"executables"`
}

type Certificate struct {
	Hostname    string   `json:"hostname"`
	Certificate string   `json:"certificate"`
	Key         string   `json:"key"`
	Authorities []string `json:"authorities"`
}

type Configuration struct {
	lock         *sync.RWMutex
	ConfigPath   string        `json:"-"`
	Services     []Service     `json:"services"`
	Certificates []Certificate `json:"certificates"`
	ServeHTTP    bool          `json:"serve_http"`
	ServeHTTPS   bool          `json:"serve_https"`
	HTTPPort     int           `json:"http_port"`
	HTTPSPort    int           `json:"https_port"`
	AdminRules   []SourceURL   `json:"admin_rules"`
	AdminHash    string        `json:"admin_hash"`
}

func MakeConfiguration() *Configuration {
	// The default password, by the way, is "password".
	return &Configuration{&sync.RWMutex{}, "", []Service{}, []Certificate{},
		true, true, 80, 443, []SourceURL{SourceURL{"http", "localhost", ""}},
		"5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"}
}

func ReadConfiguration(path string) (*Configuration, error) {
	configData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := MakeConfiguration()
	if err := json.Unmarshal(configData, config); err != nil {
		return nil, err
	}
	config.ConfigPath = path
	return config, nil
}

func (self *Configuration) Lock() {
	self.lock.Lock()
}

func (self *Configuration) Unlock() {
	self.lock.Unlock()
}

func (self *Configuration) RLock() {
	self.lock.RLock()
}

func (self *Configuration) RUnlock() {
	self.lock.RUnlock()
}
