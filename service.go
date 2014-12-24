package goule

import (
	"github.com/unixpickle/executor"
	"time"
)

// Service is a wrapper around executor.Service which includes a ServiceConfig.
type Service struct {
	executor.Service
	config *ServiceConfig
}

// NewService creates a new Service with a given ServiceConfig.
// Even if Autolaunch is set, the returned Service will not be started.
func NewService(config *ServiceConfig) Service {
	job := config.ToCmd().ToJob()
	if config.Relaunch {
		// Convert the interval to a time.Duration
		dur := time.Duration(config.Interval * float64(time.Second))
		excService := executor.RelaunchService(job, dur)
		return Service{excService, config}
	} else {
		return Service{executor.JobService(job), config}
	}
}

// Config returns the service's configuration object.
// You should not modify this object, but if you do it will not affect the
// running service.
func (s Service) Config() *ServiceConfig {
	return s.config
}

// ServiceConfig stores configuration for a service.
type ServiceConfig struct {
	// Stdout stores the standard output configuration.
	Stdout Log `json:"stdout"`

	// Stderr stores the standard error configuration.
	Stderr Log `json:"stderr"`

	// Directory is the working directory for the command
	Directory string `json:"directory"`

	// SetUID specifies whether or not the UID field should be used.
	SetUID bool `json:"set_uid"`

	// UID is the UID to run the command under.
	UID int `json:"uid"`

	// SetGID specifies whether or not the GID field should be used.
	SetGID bool `json:"set_gid"`

	// GID is the GID to run the command under.
	GID int `json:"gid"`

	// Arguments is the command-line arguments for the command.
	Arguments []string `json:"arguments"`

	// Environment is a mapping of environment variables for the command.
	Environment map[string]string `json:"environment"`

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

// ToCmd creates an executor.Cmd from a ServiceConfig.
func (s *ServiceConfig) ToCmd() *executor.Cmd {
	return &executor.Cmd{&s.Stdout, &s.Stderr, s.Directory, s.SetUID, s.UID,
		s.SetGID, s.GID, s.Arguments, s.Environment}
}