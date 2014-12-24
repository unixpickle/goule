package goule

import (
	"github.com/unixpickle/executor"
	"sync"
	"time"
)

// Service is an interface for a service which can be started, stopped, etc.
type Service interface {
	Config() *ServiceConfig
	SkipWait() error
	Start() error
	Status() executor.Status
	Stop() error
}

func NewService(config *ServiceConfig) Service {
	if config.Relaunch {
		return &relaunchService{sync.Mutex{}, nil, config.Clone(),
			config.Clone()}
	} else {
		return &jobService{sync.Mutex{}, config.ToJob(), false, nil,
			config.Clone()}
	}
}

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

type jobService struct {
	mutex   sync.Mutex
	job     executor.Job
	running bool
	done    chan struct{}
	config  *ServiceConfig
}

func (j *jobService) Config() *ServiceConfig {
	return j.config
}

func (r *jobService) SkipWait() error {
	return executor.ErrNotWaiting
}

func (j *jobService) Start() error {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	if j.running {
		return executor.ErrAlreadyRunning
	}
	ch := make(chan struct{})
	go func() {
		runJob(j.job)
		j.mutex.Lock()
		j.running = false
		j.done = nil
		j.mutex.Unlock()
		close(ch)
	}()
	j.running = true
	j.done = ch
	return nil
}

func (j *jobService) Status() executor.Status {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	if j.running {
		return executor.STATUS_RUNNING
	} else {
		return executor.STATUS_STOPPED
	}
}

func (j *jobService) Stop() error {
	j.mutex.Lock()
	if !j.running {
		j.mutex.Unlock()
		return executor.ErrNotRunning
	}
	if err := j.job.Stop(); err != nil {
		j.mutex.Unlock()
		return err
	}
	ch := j.done
	j.mutex.Unlock()
	<-ch
	return nil
}

type relaunchService struct {
	mutex      sync.Mutex
	relauncher *executor.Relauncher
	config     *ServiceConfig
	configCopy *ServiceConfig
}

func (r *relaunchService) Config() *ServiceConfig {
	return r.configCopy
}

func (r *relaunchService) SkipWait() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.relauncher == nil {
		return executor.ErrNotRunning
	}
	return r.relauncher.SkipWait()
}

func (r *relaunchService) Start() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.relauncher != nil {
		return executor.ErrAlreadyRunning
	}
	dur := time.Duration(r.config.Interval * float64(time.Second))
	r.relauncher = executor.Relaunch(r.config.ToJob(), dur)
	return nil
}

func (r *relaunchService) Status() executor.Status {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.relauncher == nil {
		return executor.STATUS_STOPPED
	}
	return r.relauncher.Status()
}

func (r *relaunchService) Stop() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.relauncher == nil {
		return executor.ErrNotRunning
	}
	r.relauncher.Stop()
	r.relauncher = nil
	return nil
}

func runJob(j executor.Job) {
	if err := j.Start(); err == nil {
		j.Wait()
	}
}
