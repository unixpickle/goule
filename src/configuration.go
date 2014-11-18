package goule

import (
	"encoding/json"
	"io/ioutil"
	"sync"
)

type ExecutableInfo struct {
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

type ServiceInfo struct {
	Name         string           `json:"name"`
	ForwardRules []ForwardRule    `json:"forward_rules"`
	Executables  []ExecutableInfo `json:"executables"`
}

type CertInfo struct {
	CertPath       string   `json:"cert_path"`
	KeyPath        string   `json:"key_path"`
	AuthorityPaths []string `json:"authority_paths"`
}

type ServerSetting struct {
	Enabled bool
	Port    int
}

type Configuration struct {
	LoadedPath   string              `json:"-"`
	Services     []ServiceInfo       `json:"services"`
	Certs        map[string]CertInfo `json:"certs"`
	DefaultCert  CertInfo            `json:"default_cert"`
	HTTPSetting  ServerSetting       `json:"http"`
	HTTPSSetting ServerSetting       `json:"https"`
	AdminRules   []SourceURL         `json:"admin_rules"`
	AdminHash    string              `json:"admin_hash"`
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
