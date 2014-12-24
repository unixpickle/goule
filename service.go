package goule

import (
	"github.com/unixpickle/executor"
)

// ServiceConfig stores configuration for a service.
type ServiceConfig struct {
	*executor.Cmd

	// Identifier is a unique identifier for a service.
	Identifier string `json:"id"`

	// Relaunch specifiecs whether the service should automatically be restarted
	// at the given interval.
	Relaunch bool `json:"relaunch"`

	// Autolaunch sets whether the service should be launched when Goule starts.
	Autolaunch bool `json:"autolaunch"`

	// Interval stores the relaunch interval in seconds.
	Interval float64 `json:"interval"`
}

// Clone returns a deep copy of a ServiceConfig.
func (s *ServiceConfig) Clone() *ServiceConfig {
	cpy := *s
	res := &cpy
	res.Cmd = res.Cmd.Clone()
	return res
}
