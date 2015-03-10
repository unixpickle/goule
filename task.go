package main

import (
	"os/exec"
	"time"
)

const (
	taskActionStart  = iota
	taskActionStop   = iota
	taskActionStatus = iota
)

const (
	TaskStatusStopped    = iota
	TaskStatusRunning    = iota
	TaskStatusRestarting = iota
)

// A Task runs an executable in the background. Tasks each have their own
// background loop. While a task's background loop is running, its fields
// should not be modified.
type Task struct {
	Args     []string
	Dir      string
	Env      map[string]string
	GID      int
	Interval int
	UID      int
	Relaunch bool
	SetGID   bool
	SetUID   bool
	
	actions chan<- taskAction
}

// NewTask creates an empty task. The task's background loop will not be running
// until StartLoop() is called.
func NewTask() *Task {
	return &Task{}
}

// Start begins executing a command for the task. If the task is executing, this
// has no effect.
func (t *Task) Start() {
	resp := make(chan interface{})
	t.actions <- taskAction{taskActionStart, resp}
	<-resp
}

// StartLoop starts the task's background Goroutine. You must call this before
// using the Start(), Stop(), and Status() methods.
func (t *Task) StartLoop() {
	if t.actions != nil {
		panic("task's loop is already running")
	}
	ch := make(chan taskAction)
	t.actions = ch
	go t.loop(ch)
}

// Status returns the task's current state. Possible values are
// TaskStatusStopped, TaskStatusRunning, and TaskStatusRestarting.
func (t *Task) Status() int {
	resp := make(chan interface{})
	t.actions <- taskAction{taskActionStatus, resp}
	return (<-resp).(int)
}

// Stop terminates the task's command. If the task is not executing, this has no
// effect. This blocks to wait for the task to stop executing.
func (t *Task) Stop() {
	resp := make(chan interface{})
	t.actions <- taskAction{taskActionStop, resp}
	<-resp
}

// StopLoop stops a task's background Goroutine. You must call this after you
// are done using a task.
//
// If the task is executing, this will terminate the process and block until it
// has stopped.
func (t *Task) StopLoop() {
	if t.actions == nil {
		panic("task's loop is not running")
	}
	t.Stop()
	close(t.actions)
	t.actions = nil
}

func (t *Task) cmd() *exec.Cmd {
	task := exec.Command(t.Args[0], t.Args[1:]...)
	for key, value := range t.Env {
		task.Env = append(task.Env, key+"="+value)
	}
	task.Dir = t.Dir

	// TODO: here, set UID and GID

	return task
}

func (t *Task) loop(actions <-chan taskAction) {
	for {
		if val, ok := <-actions; !ok {
			return
		} else if val.action == taskActionStatus {
			val.resp <- TaskStatusStopped
		} else if val.action == taskActionStart {
			close(val.resp)
			if t.Relaunch {
				t.runRestart(actions)
			} else {
				t.runOnce(actions)
			}
		} else {
			close(val.resp)
		}
	}
}

func (t *Task) runOnce(actions <-chan taskAction) {
	cmd := t.cmd()
	if err := cmd.Start(); err != nil {
		return
	}

	doneChan := make(chan struct{})
	go func() {
		cmd.Wait()
		close(doneChan)
	}()

	// Wait for commands or termination.
	for {
		select {
		case <-doneChan:
			return
		case val, ok := <-actions:
			if !ok || val.action == taskActionStop {
				cmd.Process.Kill()
				// Wait for the task to die before closing the response channel.
				<-doneChan
				if ok {
					close(val.resp)
				}
				return
			} else if val.action == taskActionStatus {
				val.resp <- TaskStatusRunning
			} else {
				close(val.resp)
			}
		}
	}
}

func (t *Task) runRestart(actions <-chan taskAction) {
	cmd := t.cmd()
	if err := cmd.Start(); err != nil {
		return
	}

	// Wait for termination in the background
	doneChan := make(chan struct{})
	go func() {
		cmd.Wait()
		doneChan <- struct{}{}
	}()

	// Wait for commands and restart the task if it stops.
	for {
		select {
		case <-doneChan:
			// Wait for the timeout and then start again.
			if !t.waitTimeout(actions) {
				return
			}
			// TODO: see if I need to re-create cmd every time.
			cmd = t.cmd()
			go func() {
				cmd.Run()
				doneChan <- struct{}{}
			}()
		case val, ok := <-actions:
			if !ok || val.action == taskActionStop {
				cmd.Process.Kill()
				// Wait for the task to die before closing the response channel.
				<-doneChan
				if ok {
					close(val.resp)
				}
				return
			} else if val.action == taskActionStatus {
				val.resp <- TaskStatusRunning
			} else {
				close(val.resp)
			}
		}
	}
}

func (t *Task) waitTimeout(actions <-chan taskAction) bool {
	for {
		select {
		case <-time.After(time.Second * time.Duration(t.Interval)):
			return true
		case val, ok := <-actions:
			if !ok || val.action == taskActionStop {
				if ok {
					close(val.resp)
				}
				return false
			} else if val.action == taskActionStatus {
				val.resp <- TaskStatusRestarting
			} else if val.action == taskActionStart {
				close(val.resp)
				return true
			}
		}
	}
}

type taskAction struct {
	action int
	resp   chan<- interface{}
}
