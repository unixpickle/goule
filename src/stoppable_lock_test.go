package goule

import (
	"testing"
	"time"
)

func TestStop(t *testing.T) {
	// Serial stops
	lock := NewStoppableLock()
	if !lock.Lock() {
		t.Fatal("Lock should have succeeded.")
	}
	lock.Stop()
	if lock.Lock() {
		t.Error("Lock should have failed.")
	}
	
	// Parallel stops
	lock = NewStoppableLock()
	channel := make(chan bool)
	fn := func() {
		if lock.Lock() {
			channel <- true
			lock.Stop()
		} else {
			channel <- false
		}
	}
	go fn()
	go fn()
	a := <-channel
	b := <-channel
	if !a || b {
		t.Error("Parallel Stop() failed.")
	}
}

func TestSkipWait(t *testing.T) {
	lock := NewStoppableLock()
	waitingChannel := make(chan struct{})
	channel := make(chan bool)
	
	go func() {
		lock.Lock()
		waitingChannel <- struct{}{}
		res := lock.Wait(time.Hour)
		if res {
			lock.Unlock()
		}
		channel <- res
	}()
	
	<-waitingChannel
	lock.Lock()
	lock.SkipWait()
	lock.Unlock()
	
	select {
	case val := <-channel:
		if !val {
			t.Fatal("Wait returned false after SkipWait")
		}
	case <-time.After(time.Second):
		t.Fatal("Wait timed out after SkipWait call.")
	}
}

func TestStopWait(t *testing.T) {
	lock := NewStoppableLock()
	waitingChannel := make(chan struct{})
	channel := make(chan bool)
	
	go func() {
		lock.Lock()
		waitingChannel <- struct{}{}
		res := lock.Wait(time.Hour)
		if res {
			lock.Unlock()
		}
		channel <- res
	}()
	
	<-waitingChannel
	lock.Lock()
	lock.Stop()
	
	select {
	case val := <-channel:
		if val {
			t.Fatal("Wait returned true after Stop")
		}
	case <-time.After(time.Second):
		t.Fatal("Wait timed out after SkipWait call.")
	}
}
