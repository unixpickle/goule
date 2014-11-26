package exec

// LogSettings stores the configuration for the logging facility.
type LogSettings struct {
	Enabled bool `json:"enabled"`
	CapSize bool `json:"cap_size"`
	MaxSize int  `json:"max_size"`
}

// UserIdentity stores the UID and GID to use for an executable.
type UserIdentity struct {
	SetGroupId bool `json:"set_group_id"`
	SetUserId  bool `json:"set_user_id"`
	GroupId    int  `json:"group_id"`
	UserId     int  `json:"user_id"`
}

// Settings holds the configuration of an executable.
type Settings struct {
	Dirname          string            `json:"dirname"`
	LogId            string            `json:"log_id"`
	Stdout           LogSettings       `json:"stdout"`
	Stderr           LogSettings       `json:"stderr"`
	Identity         UserIdentity      `json:"identity"`
	Arguments        []string          `json:"arguments"`
	Environment      map[string]string `json:"environment"`
	Autolaunch       bool              `json:"autolaunch"`
	Relaunch         bool              `json:"relaunch"`
	RelaunchInterval int               `json:"relaunch_interval"`
}

func (self *Settings) Copy() Settings {
	res := *self
	res.Arguments = make([]string, len(self.Arguments))
	copy(res.Arguments, self.Arguments)
	res.Environment = make(map[string]string)
	for key, val := range self.Environment {
		res.Environment[key] = val
	}
	return res
}
