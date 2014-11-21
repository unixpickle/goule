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

type controlStruct struct {
	stop     bool
	response chan struct{}
}

type Executable struct {
	info      ExecutableInfo
	callLock  sync.Mutex
	stateLock sync.RWMutex
	status    ExecutableStatus
	control   chan controlStruct
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
		self.control <- controlStruct{false, response}
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
		self.control <- controlStruct{true, response}
		self.stateLock.Unlock()
		_ = <-response
		return nil
	case EXECUTABLE_RESTARTING:
		// Cancel the restart process and return once the thread notices.
		response := make(chan struct{}, 1)
		self.control <- controlStruct{true, response}
		self.stateLock.Unlock()
		_ = <-response
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
	// NOTE: since callLock is still held, no other call will be able to interfere
	// with this executable.
	self.stateLock.Unlock()

	// Start the new task on a background thread; wait until it launches.
	channel := make(chan error)
	go func() {
		if err := self.createCmd(); err != nil {
			channel <- err
			close(channel)
			return
		} else {
			self.control = make(chan controlStruct, 1)
			self.status = EXECUTABLE_RUNNING
			close(channel)
			self.backgroundLoop()
		}
	}()
	return <-channel
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

func (self *Executable) backgroundLoop() {
	for {
		self.command.Wait()
		for {
			relaunch := self.awaitRelaunch()
			if !relaunch {
				return
			}
			// After we lock the state, we need to check if we've gotten a stop
			// message.
			self.stateLock.Lock()
			if self.gotStopMessage() {
				return
			}
			// If we successfully start the command, we can set our status to
			// RUNNING finally.
			if err := self.createCmd(); err == nil {
				self.status = EXECUTABLE_RUNNING
				break
			}
		}
	}
}

func (self *Executable) awaitRelaunch() bool {
	self.stateLock.Lock()
	self.command = nil

	// We select{} before checking self.info.Relaunch so we get the opportunity to
	// respond over the control channel.

	if self.gotStopMessage() {
		return false
	}

	if !self.info.Relaunch {
		// Relaunching is not enabled anyway
		self.halt()
		return false
	}

	// Begin the relaunch process, and accept stop/start messages along the way.
	self.status = EXECUTABLE_RESTARTING
	self.stateLock.Unlock()
	duration := time.Duration(self.info.RelaunchInterval)
	select {
	case ctrl := <-self.control:
		if ctrl.stop {
			self.stateLock.Lock()
			self.halt()
			close(ctrl.response)
			return false
		} else {
			close(ctrl.response)
		}
	case <-time.After(time.Second * duration):
	}

	return true
}

func (self *Executable) halt() {
	self.status = EXECUTABLE_HALTED
	close(self.control)
	self.control = nil
	self.stateLock.Unlock()
}

func (self *Executable) gotStopMessage() bool {
	select {
	case ctrl := <-self.control:
		if ctrl.stop {
			self.halt()
			close(ctrl.response)
			return true
		} else {
			close(ctrl.response)
		}
	default:
	}
	return false
}
