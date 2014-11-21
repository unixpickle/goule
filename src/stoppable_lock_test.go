package goule

import "testing"

func TestStop(t *testing.T) {
	// Serial stops
	lock := NewStoppableLock()
	if !lock.Lock() {
		t.Fatal("Lock should have succeeded.")
	}
	lock.Stop()
	if lock.Lock() {
		t.Fatal("Lock should have failed.")
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
