package goule

import (
	"github.com/unixpickle/executor"
	"time"
)

// Service stores configuration for a service.
type Service struct {
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

	// Relaunch specifiecs whether the service should automatically be restarted
	// at the given interval.
	Relaunch bool `json:"relaunch"`

	// Autolaunch sets whether the service should be launched when Goule starts.
	Autolaunch bool `json:"autolaunch"`

	// Interval stores the relaunch interval in seconds.
	Interval float64 `json:"interval"`
}

func (s *Service) ToExecutorService() executor.Service {
	cmd := &executor.Cmd{&s.Stdout, &s.Stderr, s.Directory, s.SetUID, s.UID,
		s.SetGID, s.GID, s.Arguments, s.Environment}
	job := cmd.ToJob()
	if s.Relaunch {
		// Convert the interval to a time.Duration
		dur := time.Duration(s.Interval * float64(time.Second))
		return executor.RelaunchService(job, dur)
	} else {
		return executor.JobService(job)
	}
}
