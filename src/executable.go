package goule

import (
	"os/exec"
	"sync"
)

type Executable struct {
	mutex   sync.RWMutex
	info    ExecutableInfo
	running *Cmd
}

func NewExecutable() *Executable {
	return new(Executable)
}

func (self *Executable) Start() error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if self.running != nil {
		return errors.New("Executable already running.")
	}
	return errors.New("NYI")
	// self.running = exec.Command(self.info.arguments[0])

	// TODO: add the rest of the arguments
	// TODO: set up stdout and stderr to get the output
	// TODO: set the user ID and group ID
	// TODO: set up the environment
	// TODO: run the process etc.
}

func (self *Executable) Stop() {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if self.running.Process != nil {
		self.running.Process.Kill()
	}
	self.running = nil
}

func (self *Executable) IsRunning() bool {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.running != nil
}
