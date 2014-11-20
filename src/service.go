package goule

import "sync"

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

func (self *Service) GetExecutableInfos() ([]ExecutableInfo, []ExecutableStatus) {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	list := make([]ExecutableInfo, len(self.executables))
	stats := make([]ExecutableStatus, len(self.executables))
	for i, exc := range self.executables {
		list[i] = exc.GetInfo()
		stats[i] = exc.GetStatus()
	}
	return list, stats
}

func (self *Service) UpdateExecutableInfos(infos []ExecutableInfo) {
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

func (self *Service) stopAllInternal() {
	for _, exc := range self.executables {
		exc.Stop()
	}
}
