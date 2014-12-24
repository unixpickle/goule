package goule

import (
	"bytes"
	"encoding/json"
	"github.com/unixpickle/reverseproxy"
	"io/ioutil"
	"time"
)

type Config struct {
	Path       string              `json:"-"`
	Services   []ServiceConfig     `json:"services"`
	Rules      []reverseproxy.Rule `json:"rules"`
	Auth       Auth                `json:"auth"`
	ServeHTTP  bool                `json:"serve_http"`
	HTTPPort   int                 `json:"http_port"`
	ServeHTTPS bool                `json:"serve_https"`
	HTTPSPort  int                 `json:"https_port"`
}

func NewConfig() *Config {
	res := new(Config)
	res.Services = []ServiceConfig{}
	res.Rules = []reverseproxy.Rule{}
	res.Auth.Hash = Hash("password")
	res.Auth.Port = 8080
	res.Auth.Timeout = time.Hour
	return res
}

func (c *Config) Read(path string) error {
	configData, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(configData, c); err != nil {
		return err
	}
	c.Path = path
	return nil
}

func (c *Config) Save() error {
	if data, err := json.Marshal(c); err == nil {
		var out bytes.Buffer
		json.Indent(&out, data, "", "  ")
		return ioutil.WriteFile(c.Path, out.Bytes(), 0600)
	} else {
		return err
	}
}
