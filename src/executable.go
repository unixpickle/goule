package goule

import (
	"errors"
	"os/exec"
	"sync"
)

const (
	EXECUTABLE_HALTED     = iota
	EXECUTABLE_RUNNING    = iota
	EXECUTABLE_RESTARTING = iota
)

type ExecutableStatus int

type Executable struct {
	info      ExecutableInfo
	callLock  sync.Mutex
	stateLock sync.RWMutex
	status    ExecutableStatus
	trigger   chan chan struct{}
	command   *exec.Cmd
}

func NewExecutable(info ExecutableInfo) *Executable {
	return &Executable{info, sync.Mutex{}, sync.RWMutex{}, EXECUTABLE_HALTED,
		nil, nil}
}

func (self *Executable) Start() error {
	self.callLock.Lock()
	defer self.callLock.Unlock()

	// Perform different actions depending on the state.
	self.stateLock.Lock()
	switch self.status {
	case EXECUTABLE_RUNNING:
		// The thing is already running
		self.stateLock.Unlock()
		return errors.New("Already running.")
	case EXECUTABLE_RESTARTING:
		// Trigger a restart
		response := make(chan struct{}, 1)
		self.trigger <- response
		self.stateLock.Unlock()
		_ = <-response
		return nil
	case EXECUTABLE_HALTED:
		return self.runTask()
	default:
		self.stateLock.Unlock()
		return errors.New("Unknown state.")
	}
}

func (self *Executable) Stop() error {
	self.callLock.Lock()
	defer self.callLock.Unlock()

	// Perform different actions depending on the state.
	self.stateLock.Lock()
	switch self.status {
	case EXECUTABLE_RUNNING:
		// Terminate the executable and wait for the task to die.
		if self.command.Process != nil {
			self.command.Process.Kill()
		}
		response := make(chan struct{}, 1)
		self.trigger <- response
		self.stateLock.Unlock()
		_ = <-response
		return nil
	case EXECUTABLE_RESTARTING:
		// Cancel the restart process and return immediately.
		close(self.trigger)
		self.trigger = nil
		self.stateLock.Unlock()
		return nil
	case EXECUTABLE_HALTED:
		self.stateLock.Unlock()
		return errors.New("Executable is not running.")
	default:
		self.stateLock.Unlock()
		return errors.New("Unknown state.")
	}
}

func (self *Executable) GetInfo() ExecutableInfo {
	return self.info
}

func (self *Executable) GetStatus() ExecutableStatus {
	self.callLock.Lock()
	defer self.callLock.Unlock()
	self.stateLock.RLock()
	defer self.stateLock.RUnlock()
	return self.status
}

func (self *Executable) runTask() error {
	// TODO: this is where the magic happens
	self.stateLock.Unlock()
	return errors.New("runTask() - NYI")
}
