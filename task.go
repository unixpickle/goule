package goule

import (
	"os/exec"
	"time"
)

const (
	taskActionStart = iota
	taskActionStop = iota
	taskActionStatus = iota
)

const (
	TaskStatusStopped = iota
	TaskStatusRunning = iota
	TaskStatusRestarting = iota
)

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

// StartLoop starts the task's background Goroutine. You must call this before
// using the Task.
func (t *Task) StartLoop() {
	if t.actions != nil {
		panic("task's loop is already running")
	}
	ch := make(chan taskAction)
	t.actions = ch
	go t.loop(ch)
}

// StopLoop stops a task's background Goroutine. You must call this after you
// are done using a task.
func (t *Task) StopLoop() {
	if t.actions == nil {
		panic("task's loop is not running")
	}
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
				<-doneChan
				close(val.resp)
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
				close(val.resp)
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
				close(val.resp)
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

