package goule

import (
	"sync"
)

const (
	EXECUTABLE_HALTED = iota
	EXECUTABLE_RUNNING = iota
	EXECUTABLE_RESTARTING = iota
)

type ExecutableStatus int

type Executable struct {
	info      ExecutableInfo
	callLock  sync.Mutex
	
	stateLock sync.RWMutex
	status    ExecutableStatus
	restart   chan chan struct{}
}

func NewExecutable(info ExecutableInfo) *Executable {
	return &Executable{info, sync.Mutex{}, sync.RWMutex{}, EXECUTABLE_HALTED, nil}
}

func (self *Executable) Start() error {
	self.callLock.Lock()
	defer self.callLock.Unlock()
	
	self.stateLock.Lock()
	if self.status == EXECUTABLE_RUNNING {
		// The thing is already running
		self.stateLock.Unlock()
		return errors.New("Already running.")
	} else if self.status == EXECUTABLE_RESTARTING {
		// Trigger a restart
		response := make(chan struct{}, 1)
		self.restart <- response
		self.stateLock.Unlock()
		res := <- response
		return nil
	} else if self.status == EXECUTABLE_HALTED {
		// TODO: Launch the task here
		self.stateLock.Unlock()
	}
	self.stateLock.Unlock()
	return errors.New("Unknown state.")
}

func (self *Executable) Stop() error {
	// TODO: stop the executable if it is not already running
}

func (self *Executable) GetInfo() ExecutableInfo {
	// TODO: return the info contained in the executable
	return ExecutableInfo{}
}

func (self *Executable) GetStatus() ExecutableStatus {
	// TODO: return the current status of the executable
	return 0
}
