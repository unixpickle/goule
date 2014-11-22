package goule

import (
	"errors"
)

// ExecutableGroups maps a service name to an ExecutableGroup.
type ExecutableGroups map[string]ExecutableGroup

// NewExecutableGroups creates an empty ExecutableGroups
func NewExecutableGroups(infos []ServiceInfo) ExecutableGroups {
	res := ExecutableGroups{}
	for _, info := range infos {
		res.Add(info)
	}
	return res
}

// Remove deletes an executable group from the list and stops it.
// This method is not thread-safe.
func (self ExecutableGroups) Remove(name string) {
	if group, ok := self[name]; ok {
		group.StopAll()
		delete(self, name)
	}
}

// Add adds an executable group to the list if it is not already present.
// Returns true if the group was added.
// This method is not thread-safe.
func (self ExecutableGroups) Add(info ServiceInfo) bool {
	if _, ok := self[info.Name]; ok {
		return false
	}
	group := make(ExecutableGroup, len(info.Executables))
	for i, info := range info.Executables {
		group[i] = NewExecutable(info)
	}
	self[info.Name] = group
	return true
}

// Autolaunch calls StartAutolaunch on all the receiver's groups.
// This method is not thread-safe.
func (self ExecutableGroups) Autolaunch() {
	for _, group := range self {
		group.StartAutolaunch()
	}
}

// StopAll calls StopAll on all the receiver's groups.
// This method is not thread-safe.
func (self ExecutableGroups) StopAll() {
	for _, group := range self {
		group.StopAll()
	}
}

// ExecutableGroup is an ordered list of executables
type ExecutableGroup []*Executable

// StopAll stops all the executables that belong to this group.
// This method is not thread-safe.
func (self ExecutableGroup) StopAll() {
	for _, exc := range self {
		exc.Stop()
	}
}

// StartAll starts all the executables that belong to this group.
// This method is not thread-safe.
func (self ExecutableGroup) StartAll() {
	for _, exc := range self {
		exc.Start()
	}
}

// StartAutolaunch starts all the executables that belong to this group which
// have the Autolaunch flag set.
// This method is not thread-safe.
func (self ExecutableGroup) StartAutolaunch() {
	for _, exc := range self {
		if exc.GetInfo().Autolaunch {
			exc.Start()
		}
	}
}

// StartAt starts the executable at a given index.
// Returns an error if the index is out of bounds.
// This method is not thread-safe.
func (self ExecutableGroup) StartAt(idx int) error {
	if idx < 0 || idx >= len(self) {
		return errors.New("Invalid executable index.")
	}
	self[idx].Start()
	return nil
}

// StopAt stops the executable at a given index.
// Returns an error if the index is out of bounds.
// This method is not thread-safe.
func (self ExecutableGroup) StopAt(idx int) error {
	if idx < 0 || idx >= len(self) {
		return errors.New("Invalid executable index.")
	}
	self[idx].Stop()
	return nil
}

// GetDescriptions returns the description for each element in this group.
func (self ExecutableGroup) GetDescriptions() []ExecutableDescription {
	result := make([]ExecutableDescription, len(self))
	for i, item := range self {
		result[i] = item.GetDescription()
	}
	return result
}
