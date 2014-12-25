package goule

import (
	"bytes"
	"encoding/json"
	"github.com/unixpickle/ezserver"
	"github.com/unixpickle/reverseproxy"
	"io/ioutil"
	pathlib "path"
	"runtime"
)

// Config stores the global configuration for a Goule instance
type Config struct {
	Path       string              `json:"-"`
	Services   map[string]Service  `json:"services"`
	Rules      []reverseproxy.Rule `json:"rules"`
	Admin      Admin               `json:"admin"`
	ServeHTTP  bool                `json:"serve_http"`
	HTTPPort   int                 `json:"http_port"`
	ServeHTTPS bool                `json:"serve_https"`
	HTTPSPort  int                 `json:"https_port"`
	Websockets bool                `json:"websockets"`
	TLS        ezserver.TLSConfig  `json:"tls"`
}

// NewConfig creates a new configuration with reasonable defaults.
func NewConfig() *Config {
	res := new(Config)
	res.Services = map[string]Service{}
	res.Rules = []reverseproxy.Rule{}
	res.Admin.Hash = Hash("password")
	res.Admin.Port = 8080
	res.Admin.Timeout = 60 * 60

	// Get the "assets" path.
	_, curPath, _, _ := runtime.Caller(0)
	res.Admin.Assets = pathlib.Join(pathlib.Dir(curPath), "assets")
	return res
}

// ReadConfig creates a new Config and reads a file into it.
func ReadConfig(path string) (*Config, error) {
	cfg := new(Config)
	if err := cfg.Read(path); err != nil {
		return nil, err
	}
	return cfg, nil
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
