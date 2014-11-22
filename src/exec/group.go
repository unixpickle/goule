package exec

import (
	"errors"
)

// GroupMap maps Groups to string identifiers.
type GroupMap map[string]Group

// NewGroupMap creates an empty GroupMap
func NewGroupMap() GroupMap {
	return GroupMap{}
}

// Remove deletes a Group from the map and stops it.
// This is not thread-safe.
func (self GroupMap) Remove(name string) {
	if group, ok := self[name]; ok {
		group.StopAll()
		delete(self, name)
	}
}

// Add adds a group to the list.
// Returns true if and only if the group was added.
// This is not thread-safe.
func (self GroupMap) Add(name string, group Group) bool {
	if _, ok := self[name]; ok {
		return false
	}
	self[name] = group
	return true
}

// StartAutolaunch calls StartAutolaunch on all the contained groups.
// This is not thread-safe.
func (self GroupMap) StartAutolaunch() {
	for _, group := range self {
		group.StartAutolaunch()
	}
}

// StopAll calls StopAll on all the contained groups.
// This is not thread-safe.
func (self GroupMap) StopAll() {
	for _, group := range self {
		group.StopAll()
	}
}

// Group is an ordered list of Exec pointers.
type Group []*Exec

// NewGroup creates a new Group from a list of executable settings.
func NewGroup(settings []Settings) Group {
	result := make(Group, len(settings))
	for i, setting := range settings {
		result[i] = NewExec(setting)
	}
	return result
}

// StopAll stops all the executables that belong to this group.
// This is not thread-safe.
func (self Group) StopAll() {
	for _, exc := range self {
		exc.Stop()
	}
}

// StartAll starts all the executables that belong to this group.
// This is not thread-safe.
func (self Group) StartAll() {
	for _, exc := range self {
		exc.Start()
	}
}

// StartAutolaunch starts all the executables that belong to this group which
// have the Autolaunch flag set.
// This is not thread-safe.
func (self Group) StartAutolaunch() {
	for _, exc := range self {
		if exc.GetSettings().Autolaunch {
			exc.Start()
		}
	}
}

// StartAt starts the executable at a given index.
// Returns an error if the index is out of bounds.
// This is not thread-safe.
func (self Group) StartAt(idx int) error {
	if idx < 0 || idx >= len(self) {
		return errors.New("Invalid executable index.")
	}
	self[idx].Start()
	return nil
}

// StopAt stops the executable at a given index.
// Returns an error if the index is out of bounds.
// This is not thread-safe.
func (self Group) StopAt(idx int) error {
	if idx < 0 || idx >= len(self) {
		return errors.New("Invalid executable index.")
	}
	self[idx].Stop()
	return nil
}

// GetInfos returns the info for each Exec in this group.
// This is not thread-safe.
func (self Group) GetInfos() []Info {
	result := make([]Info, len(self))
	for i, item := range self {
		result[i] = item.GetInfo()
	}
	return result
}
