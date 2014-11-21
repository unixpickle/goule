package goule

import (
	"os/exec"
	"sync"
	"time"
)

const (
	EXECUTABLE_HALTED     = iota
	EXECUTABLE_STARTING   = iota
	EXECUTABLE_RUNNING    = iota
	EXECUTABLE_RESTARTING = iota
)

type ExecutableStatus int

type Executable struct {
	info       ExecutableInfo
	globalLock sync.Mutex
	bgLock     *StoppableLock
	status     ExecutableStatus
	command    *exec.Cmd
}

// NewExecutable creates a new executable which is not running.
func NewExecutable(info ExecutableInfo) *Executable {
	return &Executable{info, sync.Mutex{}, nil, EXECUTABLE_HALTED, nil}
}

// Start starts the executable if it is not currently running or if it is in the
// process of restarting.
func (self *Executable) Start() {
	self.globalLock.Lock()
	defer self.globalLock.Unlock()

	if self.attemptLock() {
		self.bgLock.SkipWait()
		self.bgLock.Unlock()
	} else {
		self.bgLock = NewStoppableLock()
		self.status = EXECUTABLE_STARTING
		go self.backgroundThread(self.bgLock)
	}
}

// Stop stops the executable if it is currently running.
func (self *Executable) Stop() {
	self.globalLock.Lock()
	defer self.globalLock.Unlock()

	if !self.attemptLock() {
		return
	}

	// Kill the task if it's running
	if self.command != nil && self.command.Process != nil {
		self.command.Process.Kill()
	}

	// Tell the background thread to stop.
	self.bgLock.Stop()
	self.bgLock = nil
}

// GetInfo returns the current executable info in a thread-safe manner.
func (self *Executable) GetInfo() ExecutableInfo {
	return self.info
}

// GetStatus returns the current status of the executable in a thread-safe
// manner.
func (self *Executable) GetStatus() ExecutableStatus {
	self.globalLock.Lock()
	defer self.globalLock.Unlock()
	if !self.attemptLock() {
		return EXECUTABLE_HALTED
	}
	defer self.bgLock.Unlock()
	return self.status
}

// attemptLock attempts to lock bgLock.
// This method assumes that self is already globally locked.
func (self *Executable) attemptLock() bool {
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

func (self *Executable) createCommand() (*exec.Cmd, error) {
	task := exec.Command(self.info.Arguments[0],
		self.info.Arguments[1:len(self.info.Arguments)]...)
	for key, value := range self.info.Environment {
		task.Env = append(task.Env, key+"="+value)
	}

	// TODO: here, set UID and GID

	var err error
	// Attempt to pipe to the log files
	if task.Stdout, err = createLogStdout(self.info); err != nil {
		return nil, err
	}
	if task.Stderr, err = createLogStderr(self.info); err != nil {
		return nil, err
	}
	// Run the task
	if err = task.Start(); err != nil {
		return nil, err
	}
	return task, nil
}

func (self *Executable) backgroundThread(lock *StoppableLock) {
	if !lock.Lock() {
		return
	}
	for {
		cmd, _ := self.createCommand()
		if cmd != nil {
			self.status = EXECUTABLE_RUNNING
			self.command = cmd
			lock.Unlock()
			cmd.Wait()
			if !lock.Lock() {
				return
			}
		}
		if !self.info.Relaunch {
			lock.Stop()
			return
		}
		self.status = EXECUTABLE_RESTARTING
		if !lock.Wait(time.Second * time.Duration(self.info.RelaunchInterval)) {
			return
		}
	}
}
