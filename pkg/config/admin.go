package config

type AdminSettings struct {
	Rules          []SourceURL `json:"rules"`
	PasswordHash   string      `json:"password_hash"`
	SessionTimeout int         `json:"session_timeout"`
}

func (self *AdminSettings) Copy() AdminSettings {
	rules := make([]SourceURL, len(self.Rules))
	copy(rules, self.Rules)
	return AdminSettings{rules, self.PasswordHash, self.SessionTimeout}
}
