package main

import (
	"bytes"
	"os/exec"
	"sync"
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

const MaxBacklogSize = 1000

const (
	BacklogLineStdout = iota
	BacklogLineStderr = iota
	BacklogLineStatus = iota
)

// A BacklogLine represents a line of output from a task. A line can be a normal line of output, a
// status message, or an error from stdout.
type BacklogLine struct {
	// Type is either BacklogLineStdout, BacklogLineStatus, or BacklogLineStderr.
	Type int

	// Data is the actual message that was output by the task.
	Data string

	// Time is the UNIX timestamp in milliseconds when the message was logged.
	Time int64
}

// A Task runs an executable in the background. Tasks each have their own background loop.
// While a task's background loop is running, its fields should not be modified.
type Task struct {
	Args     []string
	AutoRun  bool
	Dir      string
	Env      map[string]string
	GID      int
	Interval int
	UID      int
	Relaunch bool
	SetGID   bool
	SetUID   bool

	backlogLock sync.RWMutex
	backlog     []BacklogLine

	actions chan<- taskAction
}

// NewTask creates an empty task. The task's background loop will not be running
// until StartLoop() is called.
func NewTask() *Task {
	return &Task{}
}

// Backlog returns a copy of the command's backlog.
func (t *Task) Backlog() []BacklogLine {
	t.backlogLock.RLock()
	defer t.backlogLock.RUnlock()
	backlog := make([]BacklogLine, len(t.backlog))
	for i, x := range t.backlog {
		backlog[i] = x
	}
	return backlog
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

func (t *Task) generateStreams(cmd *exec.Cmd, doneChan <-chan struct{}) {
	stdoutStream := make(chan string)
	stderrStream := make(chan string)
	stdout := &lineForwarder{sendTo: stdoutStream}
	stderr := &lineForwarder{sendTo: stderrStream}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	go func() {
	Loop:
		for {
			select {
			case line := <-stdoutStream:
				t.pushBacklog(BacklogLineStdout, line)
			case line := <-stderrStream:
				t.pushBacklog(BacklogLineStderr, line)
			case <-doneChan:
				break Loop
			}
		}
		stdout.FlushIfNotEmpty()
		stderr.FlushIfNotEmpty()
	MissedItemLoop:
		for {
			select {
			case line := <-stdoutStream:
				t.pushBacklog(BacklogLineStdout, line)
			case line := <-stderrStream:
				t.pushBacklog(BacklogLineStderr, line)
			default:
				break MissedItemLoop
			}
		}
	}()
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

func (t *Task) pushBacklog(typeNum int, data string) {
	line := BacklogLine{typeNum, data, time.Now().UnixNano() / 1000000}
	t.backlogLock.Lock()
	if len(t.backlog) < MaxBacklogSize {
		t.backlog = append(t.backlog, line)
	} else {
		for i := 1; i < len(t.backlog); i++ {
			t.backlog[i-1] = t.backlog[i]
		}
		t.backlog[MaxBacklogSize-1] = line
	}
	t.backlogLock.Unlock()
}

func (t *Task) runOnce(actions <-chan taskAction) {
	doneChan := make(chan struct{})
	cmd := t.cmd()
	t.generateStreams(cmd, doneChan)

	if err := cmd.Start(); err != nil {
		t.pushBacklog(BacklogLineStatus, "error starting: "+err.Error())
		return
	}

	t.pushBacklog(BacklogLineStatus, "started task")

	go func() {
		cmd.Wait()
		close(doneChan)
	}()

	for {
		select {
		case <-doneChan:
			t.pushBacklog(BacklogLineStatus, "task exited")
			return
		case val, ok := <-actions:
			if !ok || val.action == taskActionStop {
				t.pushBacklog(BacklogLineStatus, "task stopped")
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
	doneChan := make(chan struct{})
	cmd := t.cmd()
	t.generateStreams(cmd, doneChan)

	if err := cmd.Start(); err != nil {
		t.pushBacklog(BacklogLineStatus, "error starting: "+err.Error())
		return
	}

	t.pushBacklog(BacklogLineStatus, "started task")

	go func() {
		cmd.Wait()
		close(doneChan)
	}()

	for {
		select {
		case <-doneChan:
			t.pushBacklog(BacklogLineStatus, "task exited; waiting to restart")

			if !t.waitTimeout(actions) {
				return
			}

			t.pushBacklog(BacklogLineStatus, "restarting...")
			cmd = t.cmd()
			doneChan = make(chan struct{})
			t.generateStreams(cmd, doneChan)
			go func() {
				if err := cmd.Run(); err != nil {
					t.pushBacklog(BacklogLineStatus, "error restarting: "+err.Error())
				}
				close(doneChan)
			}()
		case val, ok := <-actions:
			if !ok || val.action == taskActionStop {
				t.pushBacklog(BacklogLineStatus, "task stopped")
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
	timeoutChannel := time.After(time.Second * time.Duration(t.Interval))
	for {
		select {
		case <-timeoutChannel:
			return true
		case val, ok := <-actions:
			if !ok || val.action == taskActionStop {
				t.pushBacklog(BacklogLineStatus, "stopped during relaunch")
				if ok {
					close(val.resp)
				}
				return false
			} else if val.action == taskActionStatus {
				val.resp <- TaskStatusRestarting
			} else if val.action == taskActionStart {
				t.pushBacklog(BacklogLineStatus, "relaunch wait bypassed")
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

// A lineForwarder is an io.Writer which buffers lines and sends them over a channel.
type lineForwarder struct {
	sendTo chan<- string
	buffer bytes.Buffer
}

func (l *lineForwarder) FlushIfNotEmpty() {
	if l.buffer.Len() > 0 {
		l.FlushLine()
	}
}

func (l *lineForwarder) FlushLine() {
	l.sendTo <- l.buffer.String()
	l.buffer.Reset()
}

func (l *lineForwarder) Write(p []byte) (n int, err error) {
	for _, ch := range p {
		if ch == '\n' {
			l.FlushLine()
		} else {
			l.buffer.WriteByte(ch)
		}
	}
	return len(p), nil
}
