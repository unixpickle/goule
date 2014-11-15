package goule

type SourceURL struct {
	Protocol     string `json:"protocol"`
	Hostname     string `json:"hostname"`
	Path         string `json:"path"`
}

type DestinationURL struct {
	Protocol     string `json:"protocol"`
	Hostname     string `json:"hostname"`
	Port         int    `json:"port"`
	Path         string `json:"path"`
}

type ForwardRule struct {
	From SourceURL      `json:"from"`
	To   DestinationURL `json:"to"`
}

type Executable struct {
	Arguments   []string          `json:"arguments"`
	Environment map[string]string `json:"environment"`
}

type Task struct {
	Name             string        `json:"name"`
	Dirname          string        `json:"dirname"`
	Autolaunch       bool          `json:"autolaunch"`
	Relaunch         bool          `json:"relaunch"`
	RelaunchInterval int           `json:"relaunch_interval"`
	SetGroupId       bool          `json:"set_group_id"`
	SetUserId        bool          `json:"set_user_id"`
	GroupId          int           `json:"group_id"`
	UserId           int           `json:"user_id"`
	LogStdout        bool          `json:"log_stdout"`
	LogStderr        bool          `json:"log_stderr"`
	ForwardRules     []ForwardRule `json:"forward_rules"`
	Executables      []Executable  `json:"executables"`
}

type Certificate struct {
	Hostname    string   `json:"hostname"`
	Certificate string   `json:"certificate"`
	Key         string   `json:"key"`
	Authorities []string `json:"authorities"`
}

type Configuration struct {
	Tasks        []Task        `json:"tasks"`
	Certificates []Certificate `json:"certificates"`
	ServeHTTP    bool          `json:"serve_http"`
	ServeHTTPS   bool          `json:"serve_https"`
	HTTPPort     int           `json:"http_port"`
	HTTPSPort    int           `json:"https_port"`
	AdminRules   []SourceURL   `json:"admin_rules"`
	AdminHash    string        `json:"admin_hash"`
}

func MakeConfiguration() *Configuration {
	return &Configuration{[]Task{}, []Certificate{}, true, true, 80, 443,
	    []SourceURL{},
		"5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"}
}
