package config

import (
	"bytes"
	"encoding/json"
	"github.com/unixpickle/goule/pkg/server"
	"io/ioutil"
)

type Config struct {
	LoadedPath    string         `json:"-"`
	Services      []Service      `json:"services"`
	TLS           server.TLSInfo `json:"tls"`
	HTTPSettings  ServerSettings `json:"http"`
	HTTPSSettings ServerSettings `json:"https"`
	Admin         AdminSettings  `json:"admin"`
}

func NewConfig() *Config {
	return &Config{}
}

func (self *Config) Copy() Config {
	services := make([]Service, len(self.Services))
	for i := range self.Services {
		services[i] = self.Services[i].Copy()
	}
	return Config{self.LoadedPath, services, self.TLS.Copy(),
		self.HTTPSettings, self.HTTPSSettings, self.Admin.Copy()}
}

func (self *Config) Read(path string) error {
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

func (self *Config) Save() error {
	if data, err := json.Marshal(self); err == nil {
		var out bytes.Buffer
		json.Indent(&out, data, "", "  ")
		return ioutil.WriteFile(self.LoadedPath, out.Bytes(), 0700)
	} else {
		return err
	}
}
