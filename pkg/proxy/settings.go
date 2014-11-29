package proxy

type Settings struct {
	Websockets  bool `json:"websockets"`
	RewriteHost bool `json:"rewrite_host"`
}

func NewSettings() *Settings {
	return &Settings{true, true}
}