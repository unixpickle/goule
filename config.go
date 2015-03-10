package main

import (
	"encoding/json"
	"github.com/unixpickle/ezserver"
	"github.com/unixpickle/reverseproxy"
	"io/ioutil"
	"os"
	"sync"
)

// Config encompasses the collective configuration of a Goule server.
type Config struct {
	sync.RWMutex
	HTTPPort   int
	HTTPSPort  int
	AdminHash  string
	AdminPort  int
	Rules      reverseproxy.RuleTable
	StartHTTP  bool
	StartHTTPS bool
	Tasks      []*Task
	TLS        *ezserver.TLSConfig
}

// LoadConfig reads a configuration from a JSON file and returns the result.
// The resulting Config will have zero or more Tasks. None of these tasks will
// have a running loop.
func LoadConfig(path string) (*Config, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	var res Config
	if err := json.Unmarshal(contents, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Save writes the configuration to a JSON file. The Config should be
// locked (a read-only lock is sufficient).
func (c *Config) Save(path string) error {
	encoded, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, encoded, os.FileMode(0600))
}
