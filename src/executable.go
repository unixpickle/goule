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
	self.running = exec.Command(self.info.arguments...)
	for key, value := range self.info.environment {
		self.running.Env = append(self.running.Env, key+"="+value)
	}
	// TODO: set up stdout and stderr to get the output
	// TODO: set the user ID and group ID
	if err := self.running.Start(); err != nil {
		self.running = nil
		return err
	}
	return nil
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
