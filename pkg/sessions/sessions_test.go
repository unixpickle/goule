package sessions

import (
	"testing"
	"time"
)

func TestLoginLogout(t *testing.T) {
	manager := NewManager()
	key1 := manager.Login()
	key2 := manager.Login()
	if !manager.Validate(key1) {
		t.Error("New key is not valid.")
	}
	if !manager.Validate(key2) {
		t.Error("New key is not valid.")
	}
	manager.Logout(key1)
	if manager.Validate(key1) {
		t.Error("Logged-out key is still valid.")
	}
	if !manager.Validate(key2) {
		t.Error("New key is not valid after Logout of another key.")
	}
	manager.Logout(key2)
	if manager.Validate(key2) {
		t.Error("Logged-out key is still valid.")
	}
}

func TestTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Cannot test timeout in short mode.")
	}
	manager := NewManager()
	key := manager.Login()
	if !manager.Validate(key) {
		t.Error("New key is not valid.")
	}
	manager.SetTimeout(1)
	time.Sleep(time.Second * 2)
	if manager.Validate(key) {
		t.Error("Key did not timeout properly.")
	}
}
