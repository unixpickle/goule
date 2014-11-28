package proxy

type Settings struct {
	Websockets  bool
	RewriteHost bool
}

func NewSettings() *Settings {
	return &Settings{true, true}
}