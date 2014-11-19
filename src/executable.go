package goule

import (
	"errors"
	"os/exec"
	"sync"
)

type Executable struct {
	info    ExecutableInfo
	mutex   sync.RWMutex
	running *exec.Cmd
}

func NewExecutable(info ExecutableInfo) *Executable {
	return &Executable{info, sync.RWMutex{}, nil}
}

func (self *Executable) Start() error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if self.running != nil {
		return errors.New("Executable already running.")
	}

	task := exec.Command(self.info.Arguments[0],
		self.info.Arguments[1:len(self.info.Arguments)]...)
	for key, value := range self.info.Environment {
		task.Env = append(task.Env, key+"="+value)
	}

	var err error
	// Attempt to pipe to the log files
	if task.Stdout, err = createLogStdout(self); err != nil {
		return err
	}
	if task.Stderr, err = createLogStderr(self); err != nil {
		return err
	}
	// Run the task
	if err = task.Start(); err != nil {
		return err
	}
	self.running = task

	// In the background, wait for the task to stop
	go func() {
		task.Wait()
		self.mutex.Lock()
		if self.running == task {
			self.running = nil
		}
		self.mutex.Unlock()
	}()

	return nil
}

func (self *Executable) Stop() {
	self.mutex.Lock()
	if self.running.Process != nil {
		self.running.Process.Kill()
	}
	self.running = nil
	self.mutex.Unlock()
}

func (self *Executable) IsRunning() bool {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.running != nil
}

func (self *Executable) GetInfo() ExecutableInfo {
	return self.info
}
