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
		if os.IsNotExist(err) {
			tls := ezserver.TLSConfig{
				map[string]ezserver.KeyCert{}, []string{}, ezserver.KeyCert{},
			}
			// Hash of "password"
			hash := "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62" +
				"a11ef721d1542d8"
			// Create the default configuration.
			return &Config{Rules: reverseproxy.RuleTable{}, Tasks: []*Task{},
				AdminHash: hash, TLS: &tls}, nil
		}
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
