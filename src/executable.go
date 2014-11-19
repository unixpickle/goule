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

	task := exec.Command(self.info.arguments...)
	for key, value := range self.info.environment {
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
