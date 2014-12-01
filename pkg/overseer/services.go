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
	// I manually lock here for the following two reasons:
	// - GetSet() would save unnecessarily in the case of a failure.
	// - setService() does the opposite of what is needed.
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
	return self.setService(name, func(index int) {
		// Remove the service from the configuration
		last := len(self.configuration.Services) - 1
		self.configuration.Services[index] = self.configuration.Services[last]
		self.configuration.Services = self.configuration.Services[0:last]

		// Remove the executables associated with the service
		self.groups.Remove(name)
	})
}

// RenameService renames a service without stopping its execution.
// Returns false if and only if the named service does not exist.
// This is thread-safe.
func (self *Overseer) RenameService(oldName string, newName string) bool {
	return self.setService(oldName, func(index int) {
		self.configuration.Services[index].Name = newName
		self.groups.Rename(oldName, newName)
	})
}

// SetServiceRules sets the forward rules for a service.
// Returns false if and only if the named service does not exist.
// This is thread-safe.
func (self *Overseer) SetServiceRules(name string,
	rules []config.ForwardRule) bool {
	return self.setService(name, func(index int) {
		self.configuration.Services[index].ForwardRules =
			make([]config.ForwardRule, len(rules))
		copy(self.configuration.Services[index].ForwardRules, rules)
	})
}

// SetServiceExecutables sets the executables for a service.
// If the service had pre-existing executables, they will be stopped and
// replaced.
// The newly added executables will not automatically be executed.
// Returns false if and only if the named service does not exist.
// This is thread-safe.
func (self *Overseer) SetServiceExecutables(name string,
	execs []exec.Settings) bool {
	return self.setService(name, func(index int) {
		service := &self.configuration.Services[index]

		// Update the configuration.
		service.Executables = make([]exec.Settings, len(execs))
		for i := range execs {
			service.Executables[i] = execs[i].Copy()
		}

		// Update the executable group.
		self.groups.Remove(name)
		self.groups.Add(name, exec.NewGroup(execs))
	})
}

// GetServiceInfos returns ServiceInfo objects for every service.
// This is thread-safe.
func (self *Overseer) GetServiceInfos() []ServiceInfo {
	return self.Get(func() interface{} {
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
	}).([]ServiceInfo)
}

func (self *Overseer) indexOfService(name string) int {
	for i, x := range self.configuration.Services {
		if x.Name == name {
			return i
		}
	}
	return -1
}

func (self *Overseer) setService(name string, f func(int)) bool {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if idx := self.indexOfService(name); idx < 0 {
		return false
	} else {
		f(idx)
		self.configuration.Save()
		return true
	}
}
