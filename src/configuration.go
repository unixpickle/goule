package goule

import (
	"encoding/json"
	"io/ioutil"
)

type LogSettings struct {
	Enabled bool `json:"enabled"`
	CapSize bool `json:"cap_size"`
	MaxSize int  `json:"max_size"`
}

type ExecutableInfo struct {
	Dirname          string            `json:"dirname"`
	LogId            string            `json:"log_id"`
	Stdout           LogSettings       `json:"stdout"`
	Stderr           LogSettings       `json:"stderr"`
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

type ServiceInfo struct {
	Name         string           `json:"name"`
	ForwardRules []ForwardRule    `json:"forward_rules"`
	Executables  []ExecutableInfo `json:"executables"`
}

type ServerSettings struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
}

type AdminSettings struct {
	Rules        []SourceURL `json:"rules"`
	PasswordHash string      `json:"password_hash"`
}

type Configuration struct {
	LoadedPath    string         `json:"-"`
	Services      []ServiceInfo  `json:"services"`
	TLS           TLSInfo        `json:"tls"`
	HTTPSettings  ServerSettings `json:"http"`
	HTTPSSettings ServerSettings `json:"https"`
	Admin         AdminSettings  `json:"admin"`
}

func NewConfiguration() *Configuration {
	return &Configuration{}
}

func (self *Configuration) Read(path string) error {
	configData, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(configData, self); err != nil {
		return err
	}
	self.LoadedPath = path
	return nil
}
