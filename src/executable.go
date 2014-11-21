package goule

import (
	"errors"
	"os/exec"
	"sync"
	"time"
)

const (
	EXECUTABLE_HALTED     = iota
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

func NewExecutable(info ExecutableInfo) *Executable {
	return &Executable{info, sync.Mutex{}, nil, EXECUTABLE_HALTED, nil}
}

func (self *Executable) Start() {
	self.globalLock.Lock()
	defer self.globalLock.Unlock()

	if self.bgLock != nil {
		if self.bgLock.Lock() {
			self.bgLock.SkipWait()
			self.bgLock.Unlock()
		} else {
			self.bgLock = nil
			self.startTask()
		}
	} else {
		self.startTask()
	}
}

func (self *Executable) Stop() {
	self.globalLock.Lock()
	defer self.globalLock.Unlock()

	if self.bgLock == nil {
		return
	}

	if self.bgLock.Lock() {
		self.bgLock.Stop()
		self.bgLock = nil
	} else {
		self.bgLock = nil
	}
}

func (self *Executable) GetInfo() ExecutableInfo {
	return self.info
}

func (self *Executable) GetStatus() ExecutableStatus {
	self.globalLock.Lock()
	defer self.globalLock.Unlock()
	if self.bgLock == nil {
		return EXECUTABLE_HALTED
	}
	if !self.bgLock.Lock() {
		self.bgLock = nil
		return EXECUTABLE_HALTED
	}
	defer self.bgLock.Unlock()
	return self.status
}

func (self *Executable) runTask() {
	// NOTE: callLock is still held when this is called.

	self.bgLock = NewStoppableLock()

	// TODO: Start the task here and run it
}

func (self *Executable) createCmd() error {
	task := exec.Command(self.info.Arguments[0],
		self.info.Arguments[1:len(self.info.Arguments)]...)
	for key, value := range self.info.Environment {
		task.Env = append(task.Env, key+"="+value)
	}

	// TODO: here, set UID and GID

	var err error
	// Attempt to pipe to the log files
	if task.Stdout, err = createLogStdout(self.info); err != nil {
		return err
	}
	if task.Stderr, err = createLogStderr(self.info); err != nil {
		return err
	}
	// Run the task
	if err = task.Start(); err != nil {
		return err
	}
	self.command = task
	return nil
}
