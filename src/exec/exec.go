package exec

import (
	execlib "os/exec"
	"time"
)

const (
	HALTED     = iota
	STARTING   = iota
	RUNNING    = iota
	RESTARTING = iota
)

// Status indicates the state of an Exec.
// Valid values are: HALTED, STARTED, RUNNING, RESTARTING.
type Status int

// Info includes statistics about an Exec's lifetime, its current status,
// and its settings.
type Info struct {
	Status     Status    `json:"status"`
	LastTerm   time.Time `json:"last_term"`
	LastLaunch time.Time `json:"last_launch"`
	LastError  time.Time `json:"last_error"`
	Error      string    `json:"error"`
	Settings   Settings  `json:"settings"`
}

func (self *Info) Copy() Info {
	res := *self
	res.Settings = self.Settings.Copy()
	return res
}

// An Exec can be started, stopped, restarted, etc.
// It represents a program or the potential to run a program.
type Exec struct {
	bgLock  *Lock
	info    Info
	command *execlib.Cmd
}

// NewExec creates a new executable which is not running.
func NewExec(info *Settings) *Exec {
	result := new(Exec)
	result.info.Settings = info.Copy()
	return result
}

// Start starts the executable if it is not currently running or if it is in the
// process of restarting.
// This is not thread-safe.
func (self *Exec) Start() {
	if self.attemptLock() {
		self.bgLock.SkipWait()
		self.bgLock.Unlock()
	} else {
		self.bgLock = NewLock()
		self.info.Status = STARTING
		go self.backgroundThread(self.bgLock)
	}
}

// Stop stops the executable if it is currently running.
// This is not thread-safe.
func (self *Exec) Stop() {
	if !self.attemptLock() {
		return
	}

	self.info.LastTerm = time.Now()

	// Kill the task if it's running
	if self.command != nil && self.command.Process != nil {
		self.command.Process.Kill()
	}

	// Tell the background thread to stop.
	self.bgLock.Stop()
	self.bgLock = nil
}

// GetSettings returns the settings for the executable.
// This is thread-safe.
func (self *Exec) GetSettings() Settings {
	return self.info.Settings.Copy()
}

// GetInfo returns the info for the executable.
// This is not thread-safe.
func (self *Exec) GetInfo() Info {
	if !self.attemptLock() {
		info := self.info.Copy()
		info.Status = HALTED
		return info
	}
	defer self.bgLock.Unlock()
	return self.info.Copy()
}

func (self *Exec) attemptLock() bool {
	if self.bgLock == nil {
		return false
	}
	if !self.bgLock.Lock() {
		self.bgLock = nil
		return false
	} else {
		return true
	}
}

func (self *Exec) createCommand() (*execlib.Cmd, error) {
	task := execlib.Command(self.info.Settings.Arguments[0],
		self.info.Settings.Arguments[1:]...)
	for key, value := range self.info.Settings.Environment {
		task.Env = append(task.Env, key+"="+value)
	}

	// TODO: here, set UID and GID

	task.Dir = self.info.Settings.Dirname

	var err error
	// Attempt to pipe to the log files
	if task.Stdout, err = createLogStdout(self.info.Settings); err != nil {
		return nil, err
	}
	if task.Stderr, err = createLogStderr(self.info.Settings); err != nil {
		return nil, err
	}
	// Run the task
	if err = task.Start(); err != nil {
		return nil, err
	}
	return task, nil
}

func (self *Exec) backgroundThread(lock *Lock) {
	if !lock.Lock() {
		return
	}
	for {
		cmd, err := self.createCommand()
		if cmd != nil {
			self.info.Status = RUNNING
			self.info.LastLaunch = time.Now()
			self.command = cmd
			lock.Unlock()
			cmd.Wait()
			if !lock.Lock() {
				return
			}
			self.info.LastTerm = time.Now()
		} else {
			self.info.LastError = time.Now()
			self.info.Error = err.Error()
		}
		if !self.info.Settings.Relaunch {
			lock.Stop()
			return
		}
		self.info.Status = RESTARTING
		duration := time.Duration(self.info.Settings.RelaunchInterval)
		if !lock.Wait(time.Second * duration) {
			return
		}
	}
}
