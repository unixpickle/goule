package goule

import (
	"errors"
	"sync"
)

type Service struct {
	mutex        sync.RWMutex
	name         string
	forwardRules []ForwardRule
	executables  []*Executable
}

func (self *Service) GetName() string {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	return self.name
}

func (self *Service) SetName(name string) {
	self.mutex.Lock()
	self.name = name
	self.mutex.Unlock()
}

func (self *Service) GetForwardRules() []ForwardRule {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	result := make([]ForwardRule, len(self.forwardRules))
	copy(result, self.forwardRules)
	return result
}

func (self *Service) SetForwardRules(rules []ForwardRule) {
	self.mutex.Lock()
	self.forwardRules = make([]ForwardRule, len(rules))
	copy(self.forwardRules, rules)
	self.mutex.Unlock()
}

func (self *Service) GetExecutables() ([]ExecutableInfo, []ExecutableStats) {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	list := make([]ExecutableInfo, len(self.executables))
	stats := make([]ExecutableStats, len(self.executables))
	for i, exc := range self.executables {
		list[i] = exc.GetInfo()
		stats[i] = exc.GetStats()
	}
	return list, stats
}

func (self *Service) UpdateExecutables(infos []ExecutableInfo) {
	self.mutex.Lock()
	self.stopAllInternal()
	self.executables = make([]*Executable, len(infos))
	for i, info := range infos {
		self.executables[i] = NewExecutable(info)
	}
	self.mutex.Unlock()
}

func (self *Service) StopAll() {
	self.mutex.Lock()
	self.stopAllInternal()
	self.mutex.Unlock()
}

func (self *Service) StartAll() {
	self.mutex.Lock()
	for _, exc := range self.executables {
		exc.Start()
	}
	self.mutex.Unlock()
}

func (self *Service) StartAutolaunch() {
	self.mutex.Lock()
	for _, exc := range self.executables {
		if exc.GetInfo().Autolaunch {
			exc.Start()
		}
	}
	self.mutex.Unlock()
}

func (self *Service) StartAt(idx int) error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if idx < 0 || idx >= len(self.executables) {
		return errors.New("Invalid executable index.")
	}
	self.executables[idx].Start()
	return nil
}

func (self *Service) StopAt(idx int) error {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if idx < 0 || idx >= len(self.executables) {
		return errors.New("Invalid executable index.")
	}
	self.executables[idx].Stop()
	return nil
}

func (self *Service) stopAllInternal() {
	for _, exc := range self.executables {
		exc.Stop()
	}
}
