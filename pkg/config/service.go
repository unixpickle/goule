package config

import "github.com/unixpickle/goule/pkg/exec"

type Service struct {
	Name         string          `json:"name"`
	ForwardRules []ForwardRule   `json:"forward_rules"`
	Executables  []exec.Settings `json:"executables"`
}

func (self *Service) Copy() Service {
	rules := make([]ForwardRule, len(self.ForwardRules))
	copy(rules, self.ForwardRules)
	execs := make([]exec.Settings, len(self.Executables))
	for i := range self.Executables {
		execs[i] = self.Executables[i].Copy()
	}
	return Service{self.Name, rules, execs}
}
