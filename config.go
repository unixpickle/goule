package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"github.com/unixpickle/ezserver"
	"github.com/unixpickle/reverseproxy"
)

// Config encompasses the collective configuration of a Goule server.
type Config struct {
	sync.RWMutex
	HTTPPort   int
	HTTPSPort  int
	AdminHash  string
	Rules      reverseproxy.RuleTable
	StartHTTP  bool
	StartHTTPS bool
	Tasks      []*Task
	TLS        *ezserver.TLSConfig
	LastTaskID int64
	path       string
}

// LoadConfig reads a configuration from a JSON file and returns the result.
// The resulting Config will have zero or more Tasks.
// None of these tasks will have a running loop.
func LoadConfig(path string) (*Config, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig(path), nil
		}
		return nil, err
	}

	var res Config
	if err := json.Unmarshal(contents, &res); err != nil {
		return nil, err
	}
	res.path = path
	return &res, nil
}

// Save writes the configuration to its file.
// The Config should be locked (a read-only lock is sufficient).
func (c *Config) Save() error {
	encoded, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.path, encoded, os.FileMode(0600))
}

// defaultConfig creates the default configuration.
// The password for this configuration is "password".
func defaultConfig(path string) *Config {
	tls := ezserver.TLSConfig{
		map[string]ezserver.KeyCert{}, []string{}, ezserver.KeyCert{},
	}

	hash := HashPassword("password")

	return &Config{Rules: reverseproxy.RuleTable{}, Tasks: []*Task{},
		AdminHash: hash, TLS: &tls, path: path}
}
