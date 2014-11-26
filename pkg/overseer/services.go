package overseer

import (
	"github.com/unixpickle/goule/pkg/config"
	"github.com/unixpickle/goule/pkg/exec"
)

// ServiceInfo contains the info contained in Service as well as live info about
// each executable.
type ServiceInfo struct {
	Name         string               `json:"name"`
	ForwardRules []config.ForwardRule `json:"forward_rules"`
	Executables  []exec.Info          `json:"executables"`
}

// AddService adds a service.
// Returns false if and only if the new service's name conflicts with an
// existing service.
// This is thread-safe.
func (self *Overseer) AddService(service *config.Service) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	// Make sure the service does not exist.
	if self.indexOfService(service.Name) >= 0 {
		return false
	}

	// Add the service to the configuration.
	self.configuration.Services = append(self.configuration.Services,
		service.Copy())
	self.configuration.Save()

	// Run the new group.
	group := exec.NewGroup(service.Executables)
	self.groups.Add(service.Name, group)
	group.StartAutolaunch()

	return true
}

// RemoveService removes a service by name.
// Returns false if and only if the service could not be found.
// This is thread-safe.
func (self *Overseer) RemoveService(name string) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	// Find the service.
	index := self.indexOfService(name)
	if index < 0 {
		return false
	}

	// Remove the service from the configuration
	last := len(self.configuration.Services) - 1
	self.configuration.Services[index] = self.configuration.Services[last]
	self.configuration.Services = self.configuration.Services[0:last]
	self.configuration.Save()

	// Remove the executables associated with the service
	self.groups.Remove(name)
	return true
}

// RenameService renames a service without stopping its execution.
// Returns false if and only if the named service does not exist.
// This is thread-safe.
func (self *Overseer) RenameService(oldName string, newName string) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	index := self.indexOfService(oldName)
	if index < 0 {
		return false
	}

	// Update the configuration.
	self.configuration.Services[index].Name = newName
	self.configuration.Save()

	// Update the executable group.
	self.groups.Rename(oldName, newName)
	return true
}

// SetServiceRules sets the forward rules for a service.
// Returns false if and only if the named service does not exist.
// This is thread-safe.
func (self *Overseer) SetServiceRules(name string,
	rules []config.ForwardRule) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	index := self.indexOfService(name)
	if index < 0 {
		return false
	}

	self.configuration.Services[index].ForwardRules =
		make([]config.ForwardRule, len(rules))
	copy(self.configuration.Services[index].ForwardRules, rules)
	self.configuration.Save()
	return true
}

// SetServiceExecutables sets the executables for a service.
// If the service had pre-existing executables, they will be stopped and
// replaced.
// The newly added executables will not automatically be executed.
// Returns false if and only if the named service does not exist.
// This is thread-safe.
func (self *Overseer) SetServiceExecutables(name string,
	execs []exec.Settings) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	index := self.indexOfService(name)
	if index < 0 {
		return false
	}

	service := &self.configuration.Services[index]

	// Update the configuration.
	service.Executables = make([]exec.Settings, len(execs))
	for i := range execs {
		service.Executables[i] = execs[i].Copy()
	}
	self.configuration.Save()

	// Update the executable group.
	self.groups.Remove(name)
	self.groups.Add(name, exec.NewGroup(execs))
	return true
}

// GetServiceInfos returns ServiceInfo objects for every service.
// This is thread-safe.
func (self *Overseer) GetServiceInfos() []ServiceInfo {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	result := []ServiceInfo{}
	for _, info := range self.configuration.Services {
		if group, ok := self.groups[info.Name]; ok {
			rules := make([]config.ForwardRule, len(info.ForwardRules))
			copy(rules, info.ForwardRules)
			desc := ServiceInfo{info.Name, rules, group.GetInfos()}
			result = append(result, desc)
		}
	}
	return result
}

func (self *Overseer) indexOfService(name string) int {
	for i, x := range self.configuration.Services {
		if x.Name == name {
			return i
		}
	}
	return -1
}
